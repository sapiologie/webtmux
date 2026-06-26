// Package voice turns a spoken instruction into a single validated terminal
// action. It transcribes audio with Mistral Voxtral, routes the transcript to
// an action with a Mistral chat model, and validates that action against a
// fixed allowlist before it is handed back to the frontend for dispatch.
package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	transcribeURL = "https://api.mistral.ai/v1/audio/transcriptions"
	chatURL       = "https://api.mistral.ai/v1/chat/completions"
)

// Action is a single validated terminal action produced from a voice command.
// It is the contract shared with the frontend dispatcher (dispatchVoiceAction).
type Action struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`
	Target string `json:"target,omitempty"`
	Name   string `json:"name,omitempty"`
	Index  int    `json:"index"`
	Dir    string `json:"dir,omitempty"`
	Amount int    `json:"amount,omitempty"`
	Scope  string `json:"scope,omitempty"`
}

// Client talks to the Mistral API for transcription and action routing.
type Client struct {
	apiKey          string
	transcribeModel string
	routerModel     string
	http            *http.Client
}

// NewClient builds a voice client. transcribeModel defaults to voxtral-mini-latest
// and routerModel to mistral-small-latest when empty.
func NewClient(apiKey, transcribeModel, routerModel string) *Client {
	if transcribeModel == "" {
		transcribeModel = "voxtral-mini-latest"
	}
	if routerModel == "" {
		routerModel = "mistral-small-latest"
	}
	return &Client{
		apiKey:          apiKey,
		transcribeModel: transcribeModel,
		routerModel:     routerModel,
		http:            &http.Client{Timeout: 60 * time.Second},
	}
}

// Transcribe sends audio to Mistral Voxtral and returns the transcript text.
func (c *Client) Transcribe(ctx context.Context, audio []byte, filename, contentType string) (string, error) {
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)

	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(audio); err != nil {
		return "", err
	}
	if err := mw.WriteField("model", c.transcribeModel); err != nil {
		return "", err
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, transcribeURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("transcription API status %d: %s", resp.StatusCode, string(data))
	}

	var parsed struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("failed to parse transcription response: %w", err)
	}
	return strings.TrimSpace(parsed.Text), nil
}

// Route asks the router model to map a transcript to a single action. The result
// is unvalidated: callers must pass it through Validate before use.
func (c *Client) Route(ctx context.Context, transcript string, sessions []string) (Action, error) {
	reqBody := map[string]interface{}{
		"model": c.routerModel,
		"messages": []map[string]string{
			{"role": "system", "content": buildRouterPrompt(sessions)},
			{"role": "user", "content": transcript},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return Action{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatURL, bytes.NewReader(buf))
	if err != nil {
		return Action{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return Action{}, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return Action{}, fmt.Errorf("router API status %d: %s", resp.StatusCode, string(data))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Action{}, fmt.Errorf("failed to parse router response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return Action{}, fmt.Errorf("router returned no choices")
	}

	var action Action
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &action); err != nil {
		return Action{}, fmt.Errorf("failed to parse action JSON %q: %w", parsed.Choices[0].Message.Content, err)
	}
	return action, nil
}

func buildRouterPrompt(sessions []string) string {
	sessionList := "none"
	if len(sessions) > 0 {
		sessionList = strings.Join(sessions, ", ")
	}
	return `You convert a single spoken instruction (already transcribed) into exactly one terminal action, returned as a JSON object and nothing else.

Available actions:
- {"type":"dictate"} - THE DEFAULT. Use for anything that is content to be typed into the terminal or into Claude Code rather than an explicit control command. Do NOT include or rewrite the text: the exact transcript is typed verbatim.
- {"type":"submit"} - press Enter to submit the current input.
- {"type":"key","name":"<enter|escape|ctrl_c|tab|backspace|up|down|left|right>"} - press one special key.
- {"type":"switch_session","target":"<session name>"} - switch to another tmux session.
- {"type":"new_session","target":"<new session name>"} - create a new tmux session with the given name and switch to it.
- {"type":"rename_session","target":"<new name>"} - rename the current session to the given name.
- {"type":"select_window","index":<n>} - switch to window number n.
- {"type":"scroll","dir":"<up|down>","amount":<lines>} - scroll the terminal history.
- {"type":"copy","scope":"<screen|all>"} - copy the visible screen or full scrollback to the clipboard.
- {"type":"paste"} - paste the clipboard into the terminal.

Available tmux sessions: ` + sessionList + `.

Rules:
- If the instruction is clearly a navigation, control, copy or paste command, emit that action.
- Otherwise treat the instruction as dictation by returning {"type":"dictate"} (the transcript is typed exactly as spoken, so never reword, summarize or correct it).
- When in doubt, prefer dictate.
- For switch_session, map spoken references like "instance two", "session two" or "number two" to the closest matching session name from the list above.
- For new_session and rename_session, derive a short, filesystem-friendly name from what the user said (e.g. "create a session for the build" -> "build").
- Respond with ONLY the JSON object.`
}

var allowedKeys = map[string]bool{
	"enter":     true,
	"escape":    true,
	"ctrl_c":    true,
	"tab":       true,
	"backspace": true,
	"up":        true,
	"down":      true,
	"left":      true,
	"right":     true,
}

// Validate enforces the action allowlist server-side. The model's output is
// untrusted: anything unrecognized or malformed falls back to dictating the
// verbatim transcript, and session targets and key names are re-checked here.
func Validate(a Action, sessions []string, transcript string) Action {
	dictate := Action{Type: "dictate", Text: transcript}

	switch a.Type {
	case "dictate":
		// Always dictate the verbatim transcript, never the router's text. The
		// router only decides whether something is dictation; it must not
		// rephrase the content (that caused garbled/rewritten typing).
		return Action{Type: "dictate", Text: transcript}

	case "submit":
		return Action{Type: "submit"}

	case "paste":
		return Action{Type: "paste"}

	case "key":
		if !allowedKeys[a.Name] {
			return dictate
		}
		return Action{Type: "key", Name: a.Name}

	case "switch_session":
		target := matchSession(a.Target, sessions)
		if target == "" {
			return dictate
		}
		return Action{Type: "switch_session", Target: target}

	case "new_session", "rename_session":
		// The controller sanitizes the name at execution time; here we only
		// confirm there is something to name, else fall back to dictation.
		name := strings.TrimSpace(a.Target)
		if name == "" {
			return dictate
		}
		return Action{Type: a.Type, Target: name}

	case "select_window":
		if a.Index < 0 {
			return dictate
		}
		return Action{Type: "select_window", Index: a.Index}

	case "scroll":
		dir := a.Dir
		if dir != "up" && dir != "down" {
			dir = "up"
		}
		amount := a.Amount
		if amount <= 0 {
			amount = 3
		}
		if amount > 200 {
			amount = 200
		}
		return Action{Type: "scroll", Dir: dir, Amount: amount}

	case "copy":
		scope := a.Scope
		if scope != "all" {
			scope = "screen"
		}
		return Action{Type: "copy", Scope: scope}

	default:
		return dictate
	}
}

// matchSession resolves a session reference returned by the router to a real
// session name. The router does the semantic mapping; this just confirms the
// target actually exists, tolerating case and surrounding whitespace.
func matchSession(target string, sessions []string) string {
	target = strings.TrimSpace(strings.ToLower(target))
	if target == "" {
		return ""
	}
	// Exact (case-insensitive) match only. The router is given the live session
	// list and returns a real name; a loose substring match could resolve to the
	// wrong session when one name is a substring of another (app2 vs app2-207).
	for _, s := range sessions {
		if strings.ToLower(s) == target {
			return s
		}
	}
	return ""
}
