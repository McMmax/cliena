// Command clienactl is a minimal, non-TUI control client for the cliena daemon.
//
// It exists for two reasons:
//
//  1. As a connection smoke-test: prove that an external process can reach the
//     daemon, complete the JSON handshake, and read live session/window state —
//     locally over the unix socket, or remotely over an SSH-forwarded TCP port.
//  2. As the executable spec for the JetBrains extension. The extension is just
//     another control client speaking this same JSON protocol; whatever this
//     program does over the wire, the Kotlin side mirrors.
//
// Unlike session.Client (which dials only the local unix socket, speaks gob, and
// is built to attach and stream a full TUI), clienactl dials either transport,
// speaks JSON, and only issues control queries — it never attaches as a TUI.
//
// Usage:
//
//	clienactl                      # local unix socket (default), list sessions + windows
//	clienactl -addr localhost:7070 # remote, via `ssh -L 7070:<remote socket> user@host`
//	clienactl -socket /path/to.sock
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gaddox/cliena/pkg/cliena/session"
)

func main() {
	var (
		addr    = flag.String("addr", "", "TCP address of the daemon (e.g. localhost:7070, for an SSH-forwarded socket). Overrides -socket.")
		socket  = flag.String("socket", "", "Unix socket path of the daemon (defaults to the local cliena socket).")
		timeout = flag.Duration("timeout", 5*time.Second, "Dial/response timeout.")
	)
	flag.Parse()

	if err := run(*addr, *socket, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, "clienactl:", err)
		os.Exit(1)
	}
}

func run(addr, socket string, timeout time.Duration) error {
	conn, where, err := dial(addr, socket, timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Printf("connected to %s\n", where)

	// The control client speaks JSON. The daemon now bootstraps its per-connection
	// codec from the Hello's wire codec byte, so this handshake works even though
	// the daemon's internal default is gob.
	codec := session.GetCodec(session.CodecJSON)

	welcome, err := handshake(conn, codec, timeout)
	if err != nil {
		return fmt.Errorf("handshake: %w", err)
	}
	fmt.Printf("daemon version %q, codec %q, %d session(s) advertised\n",
		welcome.Version, welcome.Codec, len(welcome.SessionNames))

	// List sessions — answered by the daemon directly, no attach required.
	sessions, err := listSessions(conn, codec, timeout)
	if err != nil {
		return fmt.Errorf("list sessions: %w", err)
	}
	if len(sessions) == 0 {
		fmt.Println("no sessions yet — start `devspace` to create one, then re-run")
		return nil
	}
	fmt.Printf("\nsessions (%d):\n", len(sessions))
	for _, s := range sessions {
		fmt.Printf("  - %-20s windows=%d attached=%v %dx%d\n",
			s.Name, s.WindowCount, s.Attached, s.Width, s.Height)
	}

	// Query windows of the most-recently-active session. handleQueryWindows reads
	// from the daemon's stored state, so it works with or without a TUI attached.
	windows, err := queryWindows(conn, codec, timeout)
	if err != nil {
		return fmt.Errorf("query windows: %w", err)
	}
	fmt.Printf("\nwindows in most-recent session (%d):\n", len(windows))
	for _, w := range windows {
		name, _ := w["display_name"].(string)
		ws, _ := w["workspace"].(float64) // JSON numbers decode as float64
		focused, _ := w["focused"].(bool)
		fmt.Printf("  - %-24s workspace=%v focused=%v\n", name, int(ws), focused)
	}
	return nil
}

// dial connects over TCP (if addr is set) or a unix socket, returning the conn
// and a human-readable description of the endpoint.
func dial(addr, socket string, timeout time.Duration) (net.Conn, string, error) {
	if addr != "" {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		return conn, "tcp " + addr, err
	}
	if socket == "" {
		var err error
		socket, err = session.GetSocketPath()
		if err != nil {
			return nil, "", fmt.Errorf("resolve default socket: %w", err)
		}
	}
	conn, err := net.DialTimeout("unix", socket, timeout)
	return conn, "unix " + socket, err
}

func handshake(conn net.Conn, codec session.Codec, timeout time.Duration) (*session.WelcomePayload, error) {
	hello := &session.HelloPayload{
		Version:        "clienactl",
		Term:           "xterm-256color",
		Width:          80,
		Height:         24,
		PreferredCodec: "json",
	}
	if err := writeMsg(conn, session.MsgHello, hello, codec, timeout); err != nil {
		return nil, err
	}
	resp, err := readMsg(conn, timeout)
	if err != nil {
		return nil, err
	}
	if resp.Type != session.MsgWelcome {
		return nil, fmt.Errorf("expected Welcome, got message type %d", resp.Type)
	}
	var welcome session.WelcomePayload
	if err := resp.ParsePayloadWithCodec(&welcome, codec); err != nil {
		return nil, err
	}
	return &welcome, nil
}

func listSessions(conn net.Conn, codec session.Codec, timeout time.Duration) ([]session.SessionInfo, error) {
	if err := writeMsg(conn, session.MsgList, nil, codec, timeout); err != nil {
		return nil, err
	}
	resp, err := readMsg(conn, timeout)
	if err != nil {
		return nil, err
	}
	if resp.Type != session.MsgSessionList {
		return nil, fmt.Errorf("expected SessionList, got message type %d", resp.Type)
	}
	var payload session.SessionListPayload
	if err := resp.ParsePayloadWithCodec(&payload, codec); err != nil {
		return nil, err
	}
	return payload.Sessions, nil
}

func queryWindows(conn net.Conn, codec session.Codec, timeout time.Duration) ([]map[string]any, error) {
	req := &session.QueryWindowsPayload{RequestID: "clienactl-1"}
	if err := writeMsg(conn, session.MsgQueryWindows, req, codec, timeout); err != nil {
		return nil, err
	}
	resp, err := readMsg(conn, timeout)
	if err != nil {
		return nil, err
	}
	if resp.Type != session.MsgCommandResult {
		return nil, fmt.Errorf("expected CommandResult, got message type %d", resp.Type)
	}
	var result session.CommandResultPayload
	if err := resp.ParsePayloadWithCodec(&result, codec); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, fmt.Errorf("daemon: %s", result.Message)
	}
	raw, _ := result.Data["windows"].([]any)
	windows := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			windows = append(windows, m)
		}
	}
	return windows, nil
}

func writeMsg(conn net.Conn, t session.MessageType, payload any, codec session.Codec, timeout time.Duration) error {
	msg, err := session.NewMessageWithCodec(t, payload, codec)
	if err != nil {
		return err
	}
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))
	return session.WriteMessageWithCodec(conn, msg, codec)
}

func readMsg(conn net.Conn, timeout time.Duration) (*session.Message, error) {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	msg, _, err := session.ReadMessageWithCodec(conn)
	return msg, err
}
