package webtty

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"webtmux/pkg/tmux"
)

// TmuxController interface for tmux operations
type TmuxController interface {
	GetLayout() *tmux.Layout
	RefreshLayout() error
	SelectPane(paneID string) error
	SelectWindow(windowID string) error
	SwitchSession(sessionName string) error
	SplitPane(horizontal bool) error
	ClosePane(paneID string) error
	EnterCopyMode() error
	ExitCopyMode() error
	ScrollUp(lines int) error
	ScrollDown(lines int) error
	NewWindow() error
	NewSession(name string) error
	RenameSession(target, newName string) error
	Events() <-chan tmux.Event
}

// SetTmuxController sets the tmux controller for the WebTTY instance
func (wt *WebTTY) SetTmuxController(tc TmuxController) {
	wt.tmuxCtrl = tc
}

// SendTmuxLayout sends the current tmux layout to the client
func (wt *WebTTY) SendTmuxLayout() error {
	if wt.tmuxCtrl == nil {
		return nil
	}

	layout := wt.tmuxCtrl.GetLayout()
	if layout == nil {
		return nil
	}

	data, err := json.Marshal(layout)
	if err != nil {
		return errors.Wrap(err, "failed to marshal tmux layout")
	}

	return wt.masterWrite(append([]byte{TmuxLayoutUpdate}, data...))
}

// SendTmuxLayoutData sends an already-marshaled layout to the client, avoiding a
// second json.Marshal on the 500ms poll hot path.
func (wt *WebTTY) SendTmuxLayoutData(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	return wt.masterWrite(append([]byte{TmuxLayoutUpdate}, data...))
}

// SendTmuxError reports a failed tmux command to the client so it surfaces to
// the user without tearing down the connection.
func (wt *WebTTY) SendTmuxError(msg string) error {
	return wt.masterWrite(append([]byte{TmuxError}, []byte(msg)...))
}

// SendTmuxModeUpdate sends the copy mode state to the client
func (wt *WebTTY) SendTmuxModeUpdate(inCopyMode bool) error {
	state := tmux.ModeState{
		InCopyMode: inCopyMode,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "failed to marshal tmux mode state")
	}

	return wt.masterWrite(append([]byte{TmuxModeUpdate}, data...))
}

// handleTmuxMessage handles tmux-specific messages from the client. A failed
// tmux command is reported to the client but does NOT return an error, so a
// user mistake (e.g. a duplicate session name) cannot tear down the websocket.
func (wt *WebTTY) handleTmuxMessage(msgType byte, payload []byte) error {
	if wt.tmuxCtrl == nil {
		return nil // Silently ignore if no tmux controller
	}

	// run reports a failed operation to the client and keeps the connection
	// alive; on success it pushes the updated layout. The only error it returns
	// is a master-write failure (a genuinely dead connection).
	run := func(err error) error {
		if err != nil {
			log.Printf("tmux operation failed: %v", err)
			return wt.SendTmuxError(err.Error())
		}
		return wt.SendTmuxLayout()
	}

	switch msgType {
	case TmuxSelectPane:
		return run(wt.tmuxCtrl.SelectPane(string(payload)))

	case TmuxSelectWindow:
		return run(wt.tmuxCtrl.SelectWindow(string(payload)))

	case TmuxSplitPane:
		return run(wt.tmuxCtrl.SplitPane(string(payload) == "h"))

	case TmuxClosePane:
		return run(wt.tmuxCtrl.ClosePane(string(payload)))

	case TmuxNewWindow:
		return run(wt.tmuxCtrl.NewWindow())

	case TmuxSwitchSession:
		return run(wt.tmuxCtrl.SwitchSession(string(payload)))

	case TmuxNewSession:
		return run(wt.tmuxCtrl.NewSession(string(payload)))

	case TmuxRenameSession:
		parts := strings.SplitN(string(payload), "\n", 2)
		if len(parts) != 2 {
			return wt.SendTmuxError("rename session requires a target and a new name")
		}
		return run(wt.tmuxCtrl.RenameSession(parts[0], parts[1]))

	case TmuxCopyMode:
		enter := string(payload) == "1"
		var err error
		if enter {
			err = wt.tmuxCtrl.EnterCopyMode()
		} else {
			err = wt.tmuxCtrl.ExitCopyMode()
		}
		if err != nil {
			log.Printf("tmux copy-mode failed: %v", err)
			return nil
		}
		return wt.SendTmuxModeUpdate(enter)

	case TmuxScrollUp:
		lines, _ := strconv.Atoi(string(payload))
		if err := wt.tmuxCtrl.ScrollUp(lines); err != nil {
			log.Printf("tmux scroll-up failed: %v", err)
		}
		return nil

	case TmuxScrollDown:
		lines, _ := strconv.Atoi(string(payload))
		if err := wt.tmuxCtrl.ScrollDown(lines); err != nil {
			log.Printf("tmux scroll-down failed: %v", err)
		}
		return nil

	default:
		return errors.Errorf("unknown tmux message type: %c", msgType)
	}
}

// isTmuxMessage returns true if the message type is a tmux-specific message
func isTmuxMessage(msgType byte) bool {
	switch msgType {
	case TmuxSelectPane, TmuxSelectWindow, TmuxSplitPane, TmuxClosePane,
		TmuxCopyMode, TmuxSendCommand, TmuxScrollUp, TmuxScrollDown, TmuxNewWindow,
		TmuxSwitchSession, TmuxNewSession, TmuxRenameSession:
		return true
	default:
		return false
	}
}
