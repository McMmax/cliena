package config

import (
	"log"

	"github.com/gaddox/cliena/pkg/cliena/theme"
)

// Overrides carries CLI-flag values layered over user config. A zero value
// means the flag was not passed, so the user config default should win.
type Overrides struct {
	ASCIIOnly           bool   // use ASCII glyphs instead of Nerd Font
	BorderStyle         string // window border style
	DockbarPosition     string // dock position
	HideWindowButtons   bool   // hide window control buttons
	HideScrollbar       bool   // hide the scrollbar thumb
	WindowTitlePosition string // window title position
	HideClock           bool   // hide the clock (deprecated, use ShowClock)
	ShowClock           bool   // show the clock overlay
	ShowCPU             bool   // show the dock CPU graph
	ShowRAM             bool   // show dock RAM usage
	SharedBorders       bool   // share borders between tiled windows
	ScrollbackLines     int    // scrollback depth (0 = use default)
	NoAnimations        bool   // disable UI animations
	ConfirmQuit         bool   // always show the quit dialog
	ThemeName           string // theme to load
	ZoomMaxWidth        int    // zoom width cap (0 = fullscreen)
}

// scrollbackMin and scrollbackMax bound a valid scrollback depth.
const (
	scrollbackMin = 100
	scrollbackMax = 1000000
)

// ApplyOverrides folds CLI flags and user config into the runtime globals.
// CLI flags take precedence; user config fills the rest. A nil userConfig
// applies CLI flags only. Resolution rules differ per field and are noted inline.
func ApplyOverrides(overrides Overrides, userConfig *UserConfig) {
	if overrides.ASCIIOnly {
		UseASCIIOnly = true
	}

	// string fields: CLI flag wins, else user config when non-empty
	if overrides.BorderStyle != "" {
		BorderStyle = overrides.BorderStyle
	} else if ap := appearance(userConfig); ap != nil && ap.BorderStyle != "" {
		BorderStyle = ap.BorderStyle
	}
	if overrides.DockbarPosition != "" {
		DockbarPosition = overrides.DockbarPosition
	} else if ap := appearance(userConfig); ap != nil && ap.DockbarPosition != "" {
		DockbarPosition = ap.DockbarPosition
	}
	if overrides.WindowTitlePosition != "" {
		WindowTitlePosition = overrides.WindowTitlePosition
	} else if ap := appearance(userConfig); ap != nil && ap.WindowTitlePosition != "" {
		WindowTitlePosition = ap.WindowTitlePosition
	}

	// boolean toggles: OR of CLI flag and user config
	if ap := appearance(userConfig); ap != nil {
		HideWindowButtons = overrides.HideWindowButtons || ap.HideWindowButtons
		HideScrollbar = overrides.HideScrollbar || ap.HideScrollbar
		HideClock = overrides.HideClock || ap.HideClock
		ShowClock = overrides.ShowClock || ap.ShowClock
		ShowCPU = overrides.ShowCPU || ap.ShowCPU
		ShowRAM = overrides.ShowRAM || ap.ShowRAM
	} else {
		HideWindowButtons = overrides.HideWindowButtons
		HideScrollbar = overrides.HideScrollbar
		HideClock = overrides.HideClock
		ShowClock = overrides.ShowClock
		ShowCPU = overrides.ShowCPU
		ShowRAM = overrides.ShowRAM
	}

	// scrollback: CLI flag (clamped) wins, else positive user config value
	if overrides.ScrollbackLines > 0 {
		ScrollbackLines = clamp(overrides.ScrollbackLines, scrollbackMin, scrollbackMax)
	} else if ap := appearance(userConfig); ap != nil && ap.ScrollbackLines > 0 {
		ScrollbackLines = ap.ScrollbackLines
	}

	if userConfig != nil && userConfig.Keybindings.LeaderKey != "" {
		LeaderKey = userConfig.Keybindings.LeaderKey
	}

	// one-way flags: only flip when the flag is set, never reset
	if overrides.NoAnimations {
		AnimationsEnabled = false
	}
	if overrides.ConfirmQuit {
		AlwaysConfirmQuit = true
	}

	// shared borders: CLI flag wins, else explicit user config pointer
	if overrides.SharedBorders {
		SharedBorders = true
	} else if ap := appearance(userConfig); ap != nil && ap.SharedBorders != nil {
		SharedBorders = *ap.SharedBorders
	}

	applyTheme(overrides, userConfig)

	if overrides.ZoomMaxWidth > 0 {
		ZoomMaxWidth = overrides.ZoomMaxWidth
	} else if ap := appearance(userConfig); ap != nil && ap.ZoomMaxWidth > 0 {
		ZoomMaxWidth = ap.ZoomMaxWidth
	}

	if ap := appearance(userConfig); ap != nil {
		if ap.NiriReverseScroll {
			NiriReverseScroll = true
		}
		if ap.MaxFPS > 0 {
			NormalFPS = clamp(ap.MaxFPS, 10, 120)
		}
	}
}

// applyTheme resolves the theme name (CLI flag over user config) and loads it.
func applyTheme(overrides Overrides, userConfig *UserConfig) {
	name := overrides.ThemeName
	if name == "" {
		if ap := appearance(userConfig); ap != nil {
			name = ap.Theme
		}
	}
	if name == "" {
		return
	}
	if err := theme.Initialize(name); err != nil {
		log.Printf("Warning: Failed to load theme '%s': %v", name, err)
	}
}

// appearance returns the user config's appearance block, or nil when unset.
func appearance(cfg *UserConfig) *AppearanceConfig {
	if cfg == nil {
		return nil
	}
	return &cfg.Appearance
}

// clamp bounds v to [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
