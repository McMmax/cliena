# cliena

A content-agnostic TUI **window-manager framework** in Go. cliena positions,
tiles, focuses, and animates windows; each window owns a **Pane** that fills its
rectangle with whatever it likes — a terminal, a log tailer, a build view, an
IDE bridge. The framework never knows what's inside a pane, which is the whole
point: new content types plug in without touching the core.

It also speaks a tmux-style **daemon protocol**, so a running cliena can be
observed and driven from outside — including from a JetBrains IDE via the
companion **cliena-idea** plugin.

> Status: early. The framework core, terminal panes, tiling/animation, and
> daemon-mode IDE control work end to end. APIs may still move.

## Layout

```
cmd/devspace/         reference app ("devspace") that consumes the framework
pkg/cliena/           framework core (OS, windows, panes, update/render, tiling)
pkg/cliena/terminal/  TerminalPane — the first Pane impl (PTY + VT emulator)
pkg/cliena/session/   the daemon/client protocol (tmux-style, gob + json codecs)
cmd/clienad/          headless daemon entrypoint (for servers)
cmd/clienactl/        JSON control probe / reference client
```

## Build & run

```bash
go build ./...

# local window manager (the reference consumer):
go run ./cmd/devspace      # n new window, x close, 1-9 workspaces,
                           # tab/h/l focus, t float, i/enter content mode, ctrl+b back
```

## Daemon mode (IDE control)

`devspace -daemon` runs the framework attached to the session daemon so an
external client can monitor and drive it. The daemon is a **control relay** —
panes stay local to the TUI; the client sends commands and reads the window list.

```bash
# on the host (a real terminal — it's a TUI):
go run ./cmd/devspace -daemon          # ensures a daemon, attaches as client

# from another machine, reach the daemon over SSH (the socket is local-only):
ssh -L 7070:/run/user/$(id -u)/cliena/cliena.sock  user@host
#   (or /tmp/cliena-$(id -u)/cliena.sock when XDG_RUNTIME_DIR is unset)
```

Then the JetBrains plugin (or `clienactl -addr localhost:7070`) can connect, list
windows, and issue `NewWindow` / `CloseWindow` / `SendKeys`. SendKeys accepts
literal text or a key name (`enter`, `tab`, `esc`, `ctrl+c`, arrows…).

`cmd/clienad` is a headless daemon if you'd rather host one without a TUI; note
that control commands require an attached TUI client (`devspace -daemon`) to run.

## Built with

cliena stands on the Charm terminal stack and a few focused libraries:

- [Bubble Tea](https://charm.land) — the Model-View-Update runtime
- [Lip Gloss](https://charm.land) — styling and layout
- [ultraviolet](https://github.com/charmbracelet/ultraviolet) — terminal cell primitives
- [x/vt](https://github.com/charmbracelet/x) — the VT terminal emulator
- [x/ansi](https://github.com/charmbracelet/x) — ANSI sequence parsing
- [x/xpty](https://github.com/charmbracelet/x) — cross-platform PTYs
- [colorprofile](https://github.com/charmbracelet/colorprofile) — terminal color detection
- [bubbletint](https://github.com/lrstanley/bubbletint) — color themes
- [go-toml](https://github.com/pelletier/go-toml) — config parsing
- [xdg](https://github.com/adrg/xdg) — config/runtime paths
- [fsnotify](https://github.com/fsnotify/fsnotify) — live config reload
- [uuid](https://github.com/google/uuid) — window/session ids

## License

MIT — see [LICENSE](LICENSE).
