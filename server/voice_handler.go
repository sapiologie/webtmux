package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"webtmux/pkg/voice"
)

// maxVoiceUpload caps the size of an uploaded audio clip (push-to-talk clips are
// small; this is a safety bound, not an expected size).
const maxVoiceUpload = 10 << 20 // 10 MB

type voiceResponse struct {
	Transcript string       `json:"transcript"`
	Action     voice.Action `json:"action"`
}

// handleVoice transcribes an uploaded audio clip, routes it to a single
// validated terminal action, and returns {transcript, action}. It never touches
// the PTY: the frontend dispatches the action through the existing controls.
// Only read-only tmux (capture-pane for copy) happens here.
func (server *Server) handleVoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if server.options.MistralAPIKey == "" {
		http.Error(w, "Voice control is not configured (missing Mistral API key)", http.StatusServiceUnavailable)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxVoiceUpload)
	if err := r.ParseMultipartForm(maxVoiceUpload); err != nil {
		http.Error(w, "Invalid upload", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Missing audio field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	audio, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read audio", http.StatusBadRequest)
		return
	}

	filename := header.Filename
	if filename == "" {
		filename = "audio.webm"
	}

	client := voice.NewClient(server.options.MistralAPIKey, server.options.MistralModel, server.options.MistralRouterModel)

	transcript, err := client.Transcribe(r.Context(), audio, filename, header.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("voice: transcription failed: %v", err)
		http.Error(w, "Transcription failed", http.StatusBadGateway)
		return
	}

	sessions := server.sessionNames()

	action, err := client.Route(r.Context(), transcript, sessions)
	if err != nil {
		log.Printf("voice: routing failed, falling back to dictate: %v", err)
		action = voice.Action{Type: "dictate", Text: transcript}
	}
	action = voice.Validate(action, sessions, transcript)

	// copy is the one action that needs server-side tmux: read the pane and hand
	// the text back so the browser can put it on the device clipboard.
	if action.Type == "copy" && server.tmuxCtrl != nil {
		text, err := server.tmuxCtrl.CapturePane(action.Scope == "all")
		if err != nil {
			log.Printf("voice: capture-pane failed: %v", err)
		} else {
			action.Text = text
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(voiceResponse{Transcript: transcript, Action: action}); err != nil {
		log.Printf("voice: failed to encode response: %v", err)
	}
}

// sessionNames returns the live tmux session names from the controller, used as
// context for routing and as the allowlist for switch_session.
func (server *Server) sessionNames() []string {
	if server.tmuxCtrl == nil {
		return nil
	}
	layout := server.tmuxCtrl.GetLayout()
	if layout == nil {
		return nil
	}
	names := make([]string, 0, len(layout.Sessions))
	for _, s := range layout.Sessions {
		names = append(names, s.Name)
	}
	return names
}
