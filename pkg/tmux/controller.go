package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// Controller manages tmux interactions for a session
type Controller struct {
	sessionName string
	sessionMu   sync.RWMutex

	layoutCache *Layout
	layoutMu    sync.RWMutex

	eventChan chan Event
	closeChan chan struct{}
}

// NewController creates a new tmux controller for the given session
func NewController(sessionName string) (*Controller, error) {
	c := &Controller{
		sessionName: sessionName,
		eventChan:   make(chan Event, 100),
		closeChan:   make(chan struct{}),
	}

	return c, nil
}

// session returns the current session name under a read lock. sessionName is
// read by the 500ms poller and the voice HTTP handler while being written by
// the websocket goroutine (Switch/New/RenameSession), so all access is guarded.
func (c *Controller) session() string {
	c.sessionMu.RLock()
	defer c.sessionMu.RUnlock()
	return c.sessionName
}

func (c *Controller) setSession(name string) {
	c.sessionMu.Lock()
	c.sessionName = name
	c.sessionMu.Unlock()
}

// SanitizeSessionName turns a user-supplied (typed or spoken) name into a
// tmux-safe session name. tmux uses '.' and ':' in target syntax, leading '-'
// is parsed as a flag, and control characters (incl. the newline used as the
// rename wire delimiter) must never reach a command, so all are collapsed to
// '-' and stripped from the ends.
func SanitizeSessionName(name string) string {
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == '.' || r == ':' || r == ' ' || r == '\t' {
			return '-'
		}
		return r
	}, name)
	return strings.Trim(name, "-")
}

// Start initializes the controller and gets initial layout
func (c *Controller) Start() error {
	session := c.session()
	// Check if tmux session exists, create if not
	cmd := exec.Command("tmux", "has-session", "-t", session)
	if err := cmd.Run(); err != nil {
		// Session doesn't exist, create it
		createCmd := exec.Command("tmux", "new-session", "-d", "-s", session)
		if createErr := createCmd.Run(); createErr != nil {
			return fmt.Errorf("failed to create tmux session %s: %w", session, createErr)
		}
	}

	// Get initial layout
	if err := c.RefreshLayout(); err != nil {
		return fmt.Errorf("failed to get initial layout: %w", err)
	}

	return nil
}

// Stop closes the controller
func (c *Controller) Stop() error {
	close(c.closeChan)
	return nil
}

// Events returns the channel for tmux events
func (c *Controller) Events() <-chan Event {
	return c.eventChan
}

// GetLayout returns the cached layout
func (c *Controller) GetLayout() *Layout {
	c.layoutMu.RLock()
	defer c.layoutMu.RUnlock()
	return c.layoutCache
}

// RefreshLayout fetches the current tmux layout
func (c *Controller) RefreshLayout() error {
	current := c.session()

	// Get session info
	sessionOut, err := c.runTmux("display-message", "-t", current, "-p", "#{session_id},#{session_name}")
	if err != nil {
		return err
	}
	sessionParts := strings.Split(strings.TrimSpace(sessionOut), ",")
	if len(sessionParts) < 2 {
		return fmt.Errorf("invalid session output: %s", sessionOut)
	}

	layout := &Layout{
		SessionID:   sessionParts[0],
		SessionName: sessionParts[1],
	}

	// Get all sessions
	sessionsOut, err := c.runTmux("list-sessions", "-F", "#{session_id},#{session_name},#{session_windows},#{session_attached}")
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(sessionsOut), "\n") {
			if line == "" {
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) < 4 {
				continue
			}
			winCount, _ := strconv.Atoi(parts[2])
			attached := parts[3] == "1"
			sess := Session{
				ID:       parts[0],
				Name:     parts[1],
				Windows:  winCount,
				Attached: attached,
				Active:   parts[1] == current,
			}
			layout.Sessions = append(layout.Sessions, sess)
		}
	}

	// Get windows
	windowsOut, err := c.runTmux("list-windows", "-t", current, "-F", "#{window_id},#{window_name},#{window_index},#{window_active}")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(strings.TrimSpace(windowsOut), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 4 {
			continue
		}

		idx, _ := strconv.Atoi(parts[2])
		active := parts[3] == "1"

		win := Window{
			ID:     parts[0],
			Name:   parts[1],
			Index:  idx,
			Active: active,
		}

		if active {
			layout.ActiveWinID = win.ID
		}

		// Get panes for this window
		panesOut, err := c.runTmux("list-panes", "-t", win.ID, "-F",
			"#{pane_id},#{pane_index},#{pane_active},#{pane_width},#{pane_height},#{pane_top},#{pane_left},#{pane_current_command},#{mouse_any_flag},#{pane_title}")
		if err != nil {
			continue
		}

		for _, paneLine := range strings.Split(strings.TrimSpace(panesOut), "\n") {
			if paneLine == "" {
				continue
			}
			paneParts := strings.Split(paneLine, ",")
			if len(paneParts) < 10 {
				continue
			}

			paneIdx, _ := strconv.Atoi(paneParts[1])
			paneActive := paneParts[2] == "1"
			width, _ := strconv.Atoi(paneParts[3])
			height, _ := strconv.Atoi(paneParts[4])
			top, _ := strconv.Atoi(paneParts[5])
			left, _ := strconv.Atoi(paneParts[6])

			pane := Pane{
				ID:      paneParts[0],
				Index:   paneIdx,
				Active:  paneActive,
				Width:   width,
				Height:  height,
				Top:     top,
				Left:    left,
				Command: paneParts[7],
				MouseOn: paneParts[8] == "1",
				Title:   paneParts[9],
			}

			if paneActive && active {
				layout.ActivePaneID = pane.ID
			}

			win.Panes = append(win.Panes, pane)
		}

		layout.Windows = append(layout.Windows, win)
	}

	c.layoutMu.Lock()
	c.layoutCache = layout
	c.layoutMu.Unlock()

	return nil
}

// SelectPane switches to the specified pane
func (c *Controller) SelectPane(paneID string) error {
	_, err := c.runTmux("select-pane", "-t", paneID)
	if err != nil {
		return err
	}
	c.RefreshLayout()
	return nil
}

// SelectWindow switches to the specified window
func (c *Controller) SelectWindow(windowID string) error {
	_, err := c.runTmux("select-window", "-t", windowID)
	if err != nil {
		return err
	}
	c.RefreshLayout()
	return nil
}

// SwitchSession switches to the specified session
func (c *Controller) SwitchSession(sessionName string) error {
	_, err := c.runTmux("switch-client", "-t", sessionName)
	if err != nil {
		return err
	}
	c.setSession(sessionName)
	c.RefreshLayout()
	return nil
}

// SplitPane splits the current pane
func (c *Controller) SplitPane(horizontal bool) error {
	flag := "-v"
	if horizontal {
		flag = "-h"
	}
	_, err := c.runTmux("split-window", "-t", c.session(), flag)
	if err != nil {
		return err
	}
	c.RefreshLayout()
	return nil
}

// ClosePane closes the specified pane
func (c *Controller) ClosePane(paneID string) error {
	_, err := c.runTmux("kill-pane", "-t", paneID)
	if err != nil {
		return err
	}
	c.RefreshLayout()
	return nil
}

// EnterCopyMode enters copy mode on the active pane
func (c *Controller) EnterCopyMode() error {
	_, err := c.runTmux("copy-mode", "-t", c.session())
	return err
}

// ExitCopyMode exits copy mode
func (c *Controller) ExitCopyMode() error {
	_, err := c.runTmux("send-keys", "-t", c.session(), "-X", "cancel")
	return err
}

// scroll moves the copy-mode viewport by lines using a single command. A
// per-line loop spawned one tmux process and one redraw per line, which flooded
// the websocket on a fast swipe and dropped the connection.
func (c *Controller) scroll(command string, lines int) error {
	if lines <= 0 {
		lines = 1
	}
	_, err := c.runTmux("send-keys", "-t", c.session(), "-X", "-N", strconv.Itoa(lines), command)
	return err
}

// ScrollUp scrolls up by lines in copy mode.
func (c *Controller) ScrollUp(lines int) error {
	return c.scroll("scroll-up", lines)
}

// ScrollDown scrolls down by lines in copy mode.
func (c *Controller) ScrollDown(lines int) error {
	return c.scroll("scroll-down", lines)
}

// NewWindow creates a new window
func (c *Controller) NewWindow() error {
	_, err := c.runTmux("new-window", "-t", c.session())
	if err != nil {
		return err
	}
	c.RefreshLayout()
	return nil
}

// NewSession creates a new detached session with a sanitized name and switches
// the current client to it.
func (c *Controller) NewSession(name string) error {
	name = SanitizeSessionName(name)
	if name == "" {
		return fmt.Errorf("invalid session name")
	}
	if _, err := c.runTmux("new-session", "-d", "-s", name); err != nil {
		return err
	}
	if _, err := c.runTmux("switch-client", "-t", name); err != nil {
		return err
	}
	c.setSession(name)
	c.RefreshLayout()
	return nil
}

// RenameSession renames the target session to a sanitized newName. If the
// target is the current session, the tracked name is updated too.
func (c *Controller) RenameSession(target, newName string) error {
	newName = SanitizeSessionName(newName)
	if newName == "" {
		return fmt.Errorf("invalid session name")
	}
	if _, err := c.runTmux("rename-session", "-t", target, newName); err != nil {
		return err
	}
	if target == c.session() {
		c.setSession(newName)
	}
	c.RefreshLayout()
	return nil
}

// CapturePaneOf returns the text content of the active pane of the given
// session. If all is true the entire scrollback history is captured, otherwise
// only the visible screen.
func (c *Controller) CapturePaneOf(session string, all bool) (string, error) {
	args := []string{"capture-pane", "-p", "-t", session}
	if all {
		args = append(args, "-S", "-")
	}
	return c.runTmux(args...)
}

// CapturePane captures the controller's current session.
func (c *Controller) CapturePane(all bool) (string, error) {
	return c.CapturePaneOf(c.session(), all)
}

// runTmux executes a tmux command with the given arguments
func (c *Controller) runTmux(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("tmux command failed: %w", err)
	}
	return string(output), nil
}
