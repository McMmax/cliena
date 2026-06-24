// Command devspace is cliena's first consumer: a developer-workspace shell. It
// wires the cliena window-manager framework to terminal panes and a non-PTY log
// pane — proving the content-agnostic Pane seam end to end — and is the
// reference for adding git/build/IDE panes later.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/gaddox/cliena/cmd/devspace/panes"
	"github.com/gaddox/cliena/pkg/cliena"
	"github.com/gaddox/cliena/pkg/cliena/terminal"
)

func main() {
	daemonMode := flag.Bool("daemon", false, "Attach to the cliena session daemon and run the full window manager — the TUI the JetBrains extension drives.")
	sessionName := flag.String("session", "main", "Daemon session name to attach to or create (used with -daemon).")
	flag.Parse()

	// Daemon mode runs the full app-package window manager attached to the
	// daemon (see daemon.go). Default mode runs the lightweight framework demo.
	if *daemonMode {
		if err := runDaemon(*sessionName); err != nil {
			fmt.Fprintln(os.Stderr, "devspace -daemon:", err)
			os.Exit(1)
		}
		return
	}

	prog := cliena.NewProgram()

	// NewPane is the framework's default content factory. Here every new window
	// ('n') is a shell terminal; the stateSync channel doubles as the PTY's
	// render-signal channel so output drives the event-driven repaint loop.
	prog.OS.NewPane = func(stateSync chan struct{}, exitChan chan string, w, h int) (cliena.Pane, string) {
		id := uuid.NewString()
		pane := terminal.NewTerminalPane(id, "", w, h, exitChan, stateSync)
		if pane == nil {
			return nil, ""
		}
		return pane, id
	}

	// Openers register extra pane types on their own keys. 'o' opens a LogPane:
	// a non-PTY Pane that renders and updates through the same WM loop as a
	// terminal — the proof that cliena windows are content-agnostic.
	prog.OS.Openers = map[string]cliena.Opener{
		"o": {Label: "log", New: func(stateSync chan struct{}, _ chan string, _, _ int) (cliena.Pane, string) {
			id := uuid.NewString()
			lp := panes.NewLogPane("log", stateSync)
			lp.Append("devspace log pane — a non-PTY cliena.Pane")
			lp.Append("enter Content mode (i), then: j/k scroll, PgUp/PgDn, g/G ends")
			startHeartbeat(lp) // demo producer; real panes tail files/commands/events
			return lp, id
		}},
	}

	if err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "devspace:", err)
		os.Exit(1)
	}
}

// startHeartbeat feeds a LogPane a timestamped line every two seconds until the
// pane closes, so the live, event-driven update path is visible. A production
// consumer would instead point lp.Tail at a command's stdout or a tailed file.
func startHeartbeat(lp *panes.LogPane) {
	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		n := 0
		for {
			select {
			case <-lp.Closed():
				return
			case now := <-t.C:
				n++
				lp.Append(fmt.Sprintf("[%s] heartbeat #%d", now.Format("15:04:05"), n))
			}
		}
	}()
}
