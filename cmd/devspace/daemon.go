package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"golang.org/x/term"

	"github.com/gaddox/cliena/pkg/cliena"
	"github.com/gaddox/cliena/pkg/cliena/session"
	"github.com/gaddox/cliena/pkg/cliena/terminal"
)

// runDaemon runs the cliena framework as the session daemon's TUI client, so the
// JetBrains extension can observe and control it.
//
// Design: control-relay, not PTY-relocation. The framework keeps its panes LOCAL
// (real shells in this process, exactly as plain devspace does). The daemon is
// used only as a control channel: we publish the window list to it (so the IDE's
// queryWindows reflects what's on screen) and accept the IDE's remote commands,
// injecting them into the MVU loop. The terminal content stays where the TUI
// renders — correct, since the IDE monitors and commands but does not render.
func runDaemon(sessionName string) error {
	width, height := daemonTerminalSize()

	// Ensure a daemon is up (in-process if none), then attach as a TUI client so
	// the daemon will route the IDE's remote commands to us.
	if err := session.EnsureDaemonRunning(); err != nil {
		return fmt.Errorf("start daemon: %w", err)
	}
	client := session.NewTUIClient()
	if err := client.Connect("cliena", width, height); err != nil {
		return fmt.Errorf("connect to daemon: %w", err)
	}
	defer client.Close()
	if _, err := client.AttachSession(sessionName, true, width, height); err != nil {
		return fmt.Errorf("attach session %q: %w", sessionName, err)
	}

	// Build the framework with local terminal panes (same factory as plain mode).
	prog := cliena.NewProgram()
	prog.OS.NewPane = func(stateSync chan struct{}, exitChan chan string, w, h int) (cliena.Pane, string) {
		id := uuid.NewString()
		pane := terminal.NewTerminalPane(id, "", w, h, exitChan, stateSync)
		if pane == nil {
			return nil, ""
		}
		return pane, id
	}

	// Publish the window list whenever it changes. OnWindowsChanged fires on the
	// model goroutine, so reading OS state here is race-free.
	publish := func() {
		_ = client.UpdateState(buildSessionState(sessionName, prog.OS, width, height))
	}
	prog.OS.OnWindowsChanged = publish

	// Route the IDE's remote commands (relayed by the daemon as RemoteCommand
	// payloads) into the loop, and report the result back so the IDE gets "OK".
	client.OnRemoteCommand(func(p *session.RemoteCommandPayload) error {
		switch {
		case p.CommandType == "tape_command" && p.TapeCommand == "NewWindow":
			prog.Send(cliena.RemoteCommandMsg{Kind: "new_window"})
		case p.CommandType == "tape_command" && p.TapeCommand == "CloseWindow":
			prog.Send(cliena.RemoteCommandMsg{Kind: "close_window"})
		case p.CommandType == "send_keys":
			prog.Send(cliena.RemoteCommandMsg{Kind: "send_keys", Keys: p.Keys})
		default:
			return client.SendCommandResult(p.RequestID, false,
				fmt.Sprintf("unsupported command: %s/%s", p.CommandType, p.TapeCommand))
		}
		return client.SendCommandResult(p.RequestID, true, "ok")
	})

	client.StartReadLoop()
	publish() // initial state so the session is visible before any change

	return prog.Run()
}

// buildSessionState snapshots the framework's window list into the daemon's wire
// state, which the daemon stores and serves to the IDE's queryWindows.
func buildSessionState(name string, o *cliena.OS, w, h int) *session.SessionState {
	snap := o.WindowSnapshot()
	windows := make([]session.WindowState, 0, len(snap))
	for _, wi := range snap {
		windows = append(windows, session.WindowState{
			ID:        wi.ID,
			Title:     wi.Title,
			X:         wi.X,
			Y:         wi.Y,
			Width:     wi.W,
			Height:    wi.H,
			Workspace: wi.Workspace,
		})
	}
	return &session.SessionState{
		Name:             name,
		Windows:          windows,
		FocusedWindowID:  o.FocusedWindowID(),
		CurrentWorkspace: o.Active + 1,
		Width:            w,
		Height:           h,
	}
}

func daemonTerminalSize() (int, int) {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		return 80, 24
	}
	return w, h
}
