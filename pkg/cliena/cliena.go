package cliena

import (
	tea "charm.land/bubbletea/v2"
)

// Program wraps a configured OS together with its Bubble Tea program so a
// consumer app can construct the model, tweak it (set NewPane, seed windows),
// and run it with a single call.
type Program struct {
	OS  *OS
	tea *tea.Program
}

// NewProgram builds an OS and a Bubble Tea program around it. Configure the
// returned Program.OS (e.g. set NewPane, add windows) before calling Run.
func NewProgram(opts ...tea.ProgramOption) *Program {
	o := New()
	// alt-screen, mouse, and focus reporting are driven from View() via the
	// tea.View fields in bubbletea v2 — no program options needed for them.
	return &Program{OS: o, tea: tea.NewProgram(o, opts...)}
}

// Run starts the event loop and blocks until the program exits.
func (p *Program) Run() error {
	_, err := p.tea.Run()
	return err
}

// Send injects a message into the running program's event loop from outside the
// MVU cycle — e.g. a daemon callback delivering a RemoteCommandMsg. Safe for
// concurrent use; the message is processed on the model goroutine.
func (p *Program) Send(msg tea.Msg) { p.tea.Send(msg) }
