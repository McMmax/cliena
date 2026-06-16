// Package anim provides the cliena animation primitives: framerate-fixed
// ease-out interpolation for window transitions (move, resize, snap, open).
//
// The scheduler property that matters most is negative: it only advances while
// something is in flight. The OS ticks at Interval solely when a window's
// rendered bounds differ from its target, and stops ticking the moment they all
// converge — so an idle workspace emits no frames and burns ~0 CPU, exactly
// like the PTY-driven render path.
package anim

import "math"

// Interval is the animation frame period (~60fps). The OS schedules a tick at
// this cadence only while at least one window is mid-transition.
const IntervalMS = 16

// factor is the per-frame fraction of the remaining distance to close. ~0.35
// yields a snappy ease-out over roughly 8-10 frames (~150ms).
const factor = 0.35

// snap is the distance (cells) within which a value jumps to its target,
// ending the tween cleanly instead of asymptotically creeping.
const snap = 1

// StepInt advances cur one frame toward target with ease-out smoothing. It
// returns the next value and whether the tween has reached the target. A
// minimum one-cell step guarantees progress so rounding can never stall it.
func StepInt(cur, target int) (next int, done bool) {
	d := target - cur
	if d >= -snap && d <= snap {
		return target, true
	}
	step := int(math.Round(float64(d) * factor))
	if step == 0 {
		if d > 0 {
			step = 1
		} else {
			step = -1
		}
	}
	return cur + step, false
}

// EaseOutCubic is the canonical ease-out curve, exposed for callers that prefer
// explicit time-parameterized tweening over the per-frame StepInt smoothing.
func EaseOutCubic(t float64) float64 {
	t -= 1
	return t*t*t + 1
}
