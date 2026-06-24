// Package panes holds devspace's app-level cliena.Pane implementations. They
// are the proof that cliena windows are content-agnostic: a LogPane renders and
// updates through the exact same WM loop as a terminal, with no PTY in sight.
// It is also the template for the GitPane / BuildPane / JetBrainsPane to come.
package panes

import (
	"bufio"
	"io"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"

	"github.com/gaddox/cliena/pkg/cliena"
)

// logMaxLines bounds the in-memory scrollback of a LogPane.
const logMaxLines = 5000

// LogPane is a scrolling, append-only text view fed by a live stream. It is a
// non-PTY cliena.Pane: instead of a terminal emulator it keeps a ring of lines,
// and instead of a PTY reader it pokes the shared StateSync channel whenever a
// line arrives — so it drives the same event-driven repaint as terminals and
// adds zero idle cost when nothing is being written.
type LogPane struct {
	title string

	mu     sync.Mutex
	lines  []string
	follow bool // stick to the bottom as new lines arrive
	offset int  // lines scrolled up from the bottom (when not following)
	w, h   int
	dirty  bool

	sync      chan struct{}
	closeOnce sync.Once
	closed    chan struct{}
}

// compile-time check that LogPane is a Pane and an io.Writer.
var (
	_ cliena.Pane = (*LogPane)(nil)
	_ io.Writer   = (*LogPane)(nil)
)

// NewLogPane creates an empty LogPane that signals stateSync on each new line.
func NewLogPane(title string, stateSync chan struct{}) *LogPane {
	return &LogPane{
		title:  title,
		follow: true,
		dirty:  true,
		sync:   stateSync,
		closed: make(chan struct{}),
	}
}

// Closed reports when the pane has been closed, so producers can stop.
func (p *LogPane) Closed() <-chan struct{} { return p.closed }

// Append adds s (split on newlines) to the log, trims to the cap, and signals a
// repaint. Safe to call from any goroutine.
func (p *LogPane) Append(s string) {
	p.mu.Lock()
	for _, ln := range strings.Split(s, "\n") {
		p.lines = append(p.lines, sanitize(ln))
	}
	if len(p.lines) > logMaxLines {
		p.lines = p.lines[len(p.lines)-logMaxLines:]
	}
	p.dirty = true
	p.mu.Unlock()
	p.signal()
}

// Write implements io.Writer so a LogPane can be the sink of any stream
// (exec.Cmd.Stdout, an io.Pipe, a log.Logger, ...).
func (p *LogPane) Write(b []byte) (int, error) {
	p.Append(strings.TrimRight(string(b), "\n"))
	return len(b), nil
}

// Tail reads r line by line into the log until EOF or Close, in a goroutine.
func (p *LogPane) Tail(r io.Reader) {
	go func() {
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
		for sc.Scan() {
			select {
			case <-p.closed:
				return
			default:
			}
			p.Append(sc.Text())
		}
	}()
}

// Draw renders the visible slice of lines into a content-sized buffer.
func (p *LogPane) Draw(area cliena.Rect) *uv.Buffer {
	buf := uv.NewBuffer(area.W, area.H)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dirty = false

	n := len(p.lines)
	start := n - area.H - p.offset
	if p.follow {
		start = n - area.H
	}
	if start < 0 {
		start = 0
	}
	row := 0
	for i := start; i < n && row < area.H; i++ {
		drawText(buf, p.lines[i], row, area.W)
		row++
	}
	return buf
}

// HandleInput provides scrollback navigation while the pane is focused in
// Content mode: j/k or arrows by line, PgUp/PgDn by page, g/G to ends.
func (p *LogPane) HandleInput(msg tea.Msg) tea.Cmd {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	page := p.h - 1
	if page < 1 {
		page = 1
	}
	switch km.String() {
	case "up", "k":
		p.follow = false
		p.offset++
	case "down", "j":
		p.offset--
	case "pgup":
		p.follow = false
		p.offset += page
	case "pgdown":
		p.offset -= page
	case "g", "home":
		p.follow = false
		p.offset = len(p.lines)
	case "G", "end":
		p.follow = true
		p.offset = 0
	default:
		return nil
	}

	// clamp and re-engage follow at the bottom
	maxOff := len(p.lines) - p.h
	if maxOff < 0 {
		maxOff = 0
	}
	if p.offset >= maxOff {
		p.offset = maxOff
	}
	if p.offset <= 0 {
		p.offset = 0
		p.follow = true
	}
	p.dirty = true
	return nil
}

// Resize records the content size.
func (p *LogPane) Resize(w, h int) {
	p.mu.Lock()
	p.w, p.h = w, h
	p.dirty = true
	p.mu.Unlock()
}

// Dirty reports whether the view changed since the last Draw.
func (p *LogPane) Dirty() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.dirty
}

// Title returns the pane title.
func (p *LogPane) Title() string { return p.title }

// Close stops producers waiting on Closed.
func (p *LogPane) Close() error {
	p.closeOnce.Do(func() { close(p.closed) })
	return nil
}

// signal pokes the render channel without blocking (coalescing like the PTY path).
func (p *LogPane) signal() {
	if p.sync == nil {
		return
	}
	select {
	case p.sync <- struct{}{}:
	default:
	}
}

// drawText writes s into buf at row y, truncated to maxW cells.
func drawText(buf *uv.Buffer, s string, y, maxW int) {
	col := 0
	for _, r := range s {
		if col >= maxW {
			break
		}
		c := uv.NewCell(ansi.WcWidth, string(r))
		adv := c.Width
		if adv < 1 {
			adv = 1
		}
		buf.SetCell(col, y, c)
		col += adv
	}
}

// sanitize strips carriage returns and expands tabs so control bytes never
// corrupt the cell grid.
func sanitize(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "    ")
	return strings.Map(func(r rune) rune {
		if r < 0x20 {
			return -1
		}
		return r
	}, s)
}
