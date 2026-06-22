// Command clienad runs the cliena daemon headless — no TUI, no terminal required.
//
// On a CRUISER server this is what hosts sessions so external control clients
// (clienactl, the JetBrains extension over an SSH-forwarded socket) have a daemon
// to talk to without a developer keeping `devspace` open in a terminal. The
// daemon owns PTYs and session state; clients attach over the socket.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gaddox/cliena/pkg/cliena/session"
)

func main() {
	socket := flag.String("socket", "", "Unix socket path to listen on (defaults to the standard cliena socket).")
	flag.Parse()

	cfg := &session.DaemonConfig{
		Version:    "clienad",
		SocketPath: *socket, // empty = daemon's default path
		Foreground: true,
	}

	d := session.NewDaemon(cfg)
	fmt.Fprintln(os.Stderr, "clienad: starting daemon...")
	// Run blocks until the process is signalled, then shuts down cleanly.
	if err := d.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "clienad:", err)
		os.Exit(1)
	}
}
