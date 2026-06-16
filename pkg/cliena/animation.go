package cliena

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/gaddox/cliena/pkg/cliena/anim"
)

// AnimTickMsg is the animation frame pulse. It exists only while a transition
// is in flight; the loop re-arms it each frame until every window converges,
// then stops — so the ticker never runs at idle.
type AnimTickMsg struct{}

// animTick schedules the next animation frame.
func animTick() tea.Cmd {
	return tea.Tick(anim.IntervalMS*time.Millisecond, func(time.Time) tea.Msg {
		return AnimTickMsg{}
	})
}

// needsAnim reports whether any window on the active workspace has not yet
// reached its target bounds.
func (o *OS) needsAnim() bool {
	for _, w := range o.active().Windows {
		if w.Animating() {
			return true
		}
	}
	return false
}

// maybeAnim starts the animation tick loop if a transition is pending and one
// is not already running, batching the tick alongside cmd.
func (o *OS) maybeAnim(cmd tea.Cmd) tea.Cmd {
	if o.animating || !o.needsAnim() {
		return cmd
	}
	o.animating = true
	if cmd == nil {
		return animTick()
	}
	return tea.Batch(cmd, animTick())
}

// animStep advances every animating window one frame toward its target. It
// returns true while at least one window is still in motion.
func (o *OS) animStep() bool {
	any := false
	for _, w := range o.active().Windows {
		if !w.Animating() {
			continue
		}
		nb, done := stepRect(w.Bounds, w.Target)
		w.Bounds = nb
		if !done {
			any = true
		}
	}
	return any
}

// stepRect advances a rectangle one frame toward target, easing each edge
// independently. done is true only when all four components have arrived.
func stepRect(cur, target Rect) (Rect, bool) {
	x, dx := anim.StepInt(cur.X, target.X)
	y, dy := anim.StepInt(cur.Y, target.Y)
	w, dw := anim.StepInt(cur.W, target.W)
	h, dh := anim.StepInt(cur.H, target.H)
	return Rect{X: x, Y: y, W: w, H: h}, dx && dy && dw && dh
}
