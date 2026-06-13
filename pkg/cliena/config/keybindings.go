package config

import "fmt"

// Keybinding is one key→description row for the help and prefix overlays.
type Keybinding struct {
	Key         string
	Description string
}

// KeybindingSection groups related Keybindings under a title. Condition gates
// visibility: "" always, "tiling" only in tiling mode, "!tiling" only outside it.
type KeybindingSection struct {
	Title     string
	Condition string
	Bindings  []Keybinding
}

// GetPrefixKeybindings returns the rows shown in a prefix-mode overlay.
// isDaemonSession (optional, first value used) tweaks the detach/quit wording.
func GetPrefixKeybindings(prefixType string, isDaemonSession ...bool) []Keybinding {
	daemon := len(isDaemonSession) > 0 && isDaemonSession[0]

	switch prefixType {
	case "workspace":
		return []Keybinding{
			{"1-9", "Switch to workspace"},
			{"Shift+1-9", "Move window to workspace"},
			{"Esc", "Cancel"},
		}
	case "minimize":
		return []Keybinding{
			{"m", "Minimize focused window"},
			{"1-9", "Restore window"},
			{"Shift+M", "Restore all"},
			{"Esc", "Cancel"},
		}
	case "window":
		return []Keybinding{
			{"n", "New window"},
			{"x", "Close window"},
			{"r", "Rename window"},
			{"Tab", "Next window"},
			{"Shift+Tab", "Previous window"},
			{"t", "Toggle tiling mode"},
			{"Esc", "Cancel"},
		}
	case "debug":
		return []Keybinding{
			{"l", "Toggle log viewer"},
			{"c", "Toggle cache statistics"},
			{"k", "Toggle showkeys overlay"},
			{"a", "Toggle animations"},
			{"Esc", "Cancel"},
		}
	case "tape":
		return []Keybinding{
			{"m", "Open tape manager"},
			{"r", "Start recording"},
			{"s", "Stop recording"},
			{"Esc", "Cancel"},
		}
	case "layout":
		return []Keybinding{
			{"l", "Load layout"},
			{"s", "Save layout"},
			{"Esc", "Cancel"},
		}
	default:
		return generalPrefixBindings(daemon)
	}
}

// generalPrefixBindings builds the main Ctrl+B overlay. The detach/quit rows
// read differently in daemon vs local sessions.
func generalPrefixBindings(daemon bool) []Keybinding {
	bindings := []Keybinding{
		{"c", "Create window"},
		{"x", "Close window"},
		{",", "Rename window"},
		{"n", "Next window"},
		{"p", "Previous window"},
		{"0-9", "Jump to window"},
		{"z", "Toggle zoom"},
		{"space", "Toggle tiling"},
		{"-", "Split horizontal (top/bottom)"},
		{"|/\\", "Split vertical (left/right)"},
		{"R", "Rotate split direction"},
		{"=", "Equalize splits"},
		{"w", "Workspace commands..."},
		{"m", "Minimize commands..."},
		{"t", "Window commands..."},
		{"D", "Debug commands..."},
		{"T", "Tape manager..."},
		{"P", "Command palette"},
		{"S", "Session switcher"},
		{"L", "Layout commands..."},
	}

	if daemon {
		// in daemon mode d detaches and Esc only switches mode
		bindings = append(bindings,
			Keybinding{"d", "Detach (session keeps running)"},
			Keybinding{"Esc", "Window management mode"},
		)
	} else {
		bindings = append(bindings, Keybinding{"d/Esc", "Window management mode"})
	}

	bindings = append(bindings,
		Keybinding{"[", "Scrollback mode"},
		Keybinding{"s", "Scrollback browser"},
		Keybinding{"?", "Toggle help"},
	)

	if daemon {
		bindings = append(bindings, Keybinding{"q", "Quit and kill session"})
	} else {
		bindings = append(bindings, Keybinding{"q", "Quit application"})
	}
	return bindings
}

// GetKeybindings builds the help-menu sections. With a registry the dynamic
// sections reflect the user's actual binds; a nil registry yields static defaults.
func GetKeybindings(registry *KeybindRegistry) []KeybindingSection {
	if registry == nil {
		return defaultHelpSections()
	}

	sections := make([]KeybindingSection, 0, 4)

	windowMgmt := KeybindingSection{Title: "WINDOW MANAGEMENT"}
	for _, b := range []struct{ action, desc string }{
		{"new_window", "New window"},
		{"close_window", "Close window"},
		{"rename_window", "Rename window"},
		{"minimize_window", "Minimize window"},
		{"restore_all", "Restore all"},
		{"next_window", "Next window"},
		{"prev_window", "Previous window"},
	} {
		appendBinding(&windowMgmt, registry, b.action, b.desc)
	}
	sections = appendNonEmpty(sections, windowMgmt)

	workspaces := KeybindingSection{Title: "WORKSPACES"}
	for i := 1; i <= 9; i++ {
		appendBinding(&workspaces, registry,
			fmt.Sprintf("switch_workspace_%d", i), fmt.Sprintf("Switch to workspace %d", i))
	}
	for i := 1; i <= 9; i++ {
		appendBinding(&workspaces, registry,
			fmt.Sprintf("move_and_follow_%d", i), fmt.Sprintf("Move to workspace %d and follow", i))
	}
	sections = appendNonEmpty(sections, workspaces)

	modes := KeybindingSection{Title: "MODES"}
	for _, b := range []struct{ action, desc string }{
		{"enter_terminal_mode", "Insert mode"},
		{"toggle_tiling", "Toggle tiling"},
		{"toggle_help", "Toggle help"},
	} {
		appendBinding(&modes, registry, b.action, b.desc)
	}
	sections = appendNonEmpty(sections, modes)

	return append(sections, staticHelpSections()...)
}

// appendBinding adds an action's display keys to section if any are configured.
func appendBinding(section *KeybindingSection, registry *KeybindRegistry, action, desc string) {
	if keys := registry.GetKeysForDisplay(action); keys != "" {
		section.Bindings = append(section.Bindings, Keybinding{Key: keys, Description: desc})
	}
}

// appendNonEmpty appends section only if it has at least one binding.
func appendNonEmpty(sections []KeybindingSection, section KeybindingSection) []KeybindingSection {
	if len(section.Bindings) > 0 {
		return append(sections, section)
	}
	return sections
}

// defaultHelpSections returns the hard-coded help used when no registry exists.
// %s placeholders are filled with the platform modifier key by the renderer.
func defaultHelpSections() []KeybindingSection {
	sections := []KeybindingSection{
		{
			Title: "WINDOW MANAGEMENT",
			Bindings: []Keybinding{
				{"n", "New window"},
				{"x", "Close window"},
				{"r", "Rename window"},
				{"m", "Minimize window"},
				{"Shift+M", "Restore all"},
				{"Tab", "Next window"},
				{"Shift+Tab", "Previous window"},
				{"1-9", "Select window"},
			},
		},
		{
			Title: "WORKSPACES",
			Bindings: []Keybinding{
				{"%s+1-9", "Switch workspace"},
				{"%s+Shift+1-9", "Move window and follow"},
				{"Ctrl+B, w, 1-9", "Switch workspace (prefix)"},
				{"Ctrl+B, w, Shift+1-9", "Move window (prefix)"},
			},
		},
		{
			Title: "MODES",
			Bindings: []Keybinding{
				{"i, Enter", "Insert mode"},
				{"t", "Toggle tiling"},
				{"?", "Toggle help"},
			},
		},
	}
	return append(sections, staticHelpSections()...)
}

// staticHelpSections returns help rows that never depend on user binds
// (mouse, special modes, the full prefix reference).
func staticHelpSections() []KeybindingSection {
	return []KeybindingSection{
		{
			Title:     "TILING:",
			Condition: "tiling",
			Bindings: []Keybinding{
				{"Shift+H/L, Ctrl+←/→", "Swap left/right"},
				{"Shift+K/J, Ctrl+↑/↓", "Swap up/down"},
				{"< / >", "Resize master width"},
				{"{ / }", "Resize focused window height"},
				{"Ctrl+B, -", "Split horizontal"},
				{"Ctrl+B, |/\\", "Split vertical"},
				{"Ctrl+B, R", "Rotate split"},
			},
		},
		{
			Title:     "WINDOW SNAPPING:",
			Condition: "!tiling",
			Bindings: []Keybinding{
				{"h, l", "Snap left/right"},
				{"1-4", "Snap to corners"},
				{"f", "Fullscreen"},
				{"u", "Unsnap"},
			},
		},
		{
			Title: "SCROLLBACK:",
			Bindings: []Keybinding{
				{"Ctrl+B, [", "Enter scrollback mode"},
				{"Mouse wheel ↑", "Enter scrollback mode"},
				{"↑/↓, j/k", "Scroll up/down one line"},
				{"PgUp/PgDn", "Scroll half screen"},
				{"Ctrl+U/D", "Scroll half screen"},
				{"g, Home", "Go to oldest line"},
				{"G, End", "Go to newest line (exit)"},
				{"q, Esc", "Exit scrollback mode"},
			},
		},
		{
			Title: "WINDOW NAVIGATION:",
			Bindings: []Keybinding{
				{"Ctrl+↑/↓", "Swap/maximize windows"},
			},
		},
		{
			Title: "SYSTEM:",
			Bindings: []Keybinding{
				{"Ctrl+L", "Toggle log viewer"},
			},
		},
		{
			Title: "PREFIX (Ctrl+B) - Works in all modes:",
			Bindings: []Keybinding{
				{"c", "Create window"},
				{"x", "Close window"},
				{",/r", "Rename window"},
				{"n/Tab", "Next window"},
				{"p/Shift+Tab", "Previous window"},
				{"0-9", "Jump to window"},
				{"z", "Toggle zoom"},
				{"space", "Toggle tiling"},
				{"-", "Split horizontal"},
				{"|/\\", "Split vertical"},
				{"R", "Rotate split"},
				{"=", "Equalize splits"},
				{"w", "Workspace commands"},
				{"m", "Minimize commands"},
				{"t", "Window commands"},
				{"P", "Command palette"},
				{"S", "Session switcher"},
				{"L", "Load layout"},
				{"D", "Debug commands"},
				{"T", "Tape manager"},
				{"d", "Detach (daemon) / Window mode (local)"},
				{"Esc", "Window management mode"},
				{"[", "Enter scrollback mode"},
				{"s", "Scrollback browser"},
				{"?", "Toggle help"},
				{"q", "Quit"},
				{"Ctrl+B", "Send literal Ctrl+B"},
			},
		},
		{
			Title: "WINDOW PREFIX (Ctrl+B, t):",
			Bindings: []Keybinding{
				{"n", "New window"},
				{"x", "Close window"},
				{"r", "Rename window"},
				{"Tab/Shift+Tab", "Next/Previous window"},
				{"t", "Toggle tiling mode"},
			},
		},
		{
			Title: "TAPE PREFIX (Ctrl+B, T):",
			Bindings: []Keybinding{
				{"m", "Open tape manager"},
				{"r", "Start recording"},
				{"s", "Stop recording"},
			},
		},
		{
			Title: "",
			Bindings: []Keybinding{
				{"q, Ctrl+C", "Quit"},
			},
		},
	}
}
