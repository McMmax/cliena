package config

import "strings"

// KeybindRegistry resolves keys to action names. Normal-mode sections are
// flattened into a reverse key→action map at build time; prefix-mode sections
// are looked up on demand so their keys never clash with normal-mode binds.
type KeybindRegistry struct {
	keyToAction map[string]string
	config      *UserConfig
	normalizer  *KeyNormalizer
}

// NewKeybindRegistry builds a registry from cfg with its reverse map populated.
func NewKeybindRegistry(cfg *UserConfig) *KeybindRegistry {
	r := &KeybindRegistry{
		config:     cfg,
		normalizer: NewKeyNormalizer(),
	}
	r.buildMappings()
	return r
}

// normalModeSections lists the sections active without a prefix. Prefix
// sections (PrefixMode, *Prefix) are intentionally excluded — they resolve via
// their dedicated Get*PrefixAction methods only.
func (r *KeybindRegistry) normalModeSections() []map[string][]string {
	kb := &r.config.Keybindings
	return []map[string][]string{
		kb.WindowManagement,
		kb.Workspaces,
		kb.Layout,
		kb.ModeControl,
		kb.System,
		kb.Navigation,
		kb.RestoreMinimized,
	}
}

// buildMappings (re)builds the reverse key→action map from normal-mode sections.
func (r *KeybindRegistry) buildMappings() {
	r.keyToAction = make(map[string]string)
	for _, section := range r.normalModeSections() {
		r.indexSection(r.keyToAction, section)
	}
}

// indexSection writes every expanded key of a section into dst as key→action.
// Keys keep their normalized case so M (shift+m) and m stay distinct.
func (r *KeybindRegistry) indexSection(dst map[string]string, section map[string][]string) {
	for action, keys := range section {
		for _, key := range r.normalizer.ExpandKeys(keys) {
			dst[key] = action
		}
	}
}

// GetAction returns the normal-mode action bound to key, or "".
func (r *KeybindRegistry) GetAction(key string) string {
	return r.lookupKey(key, r.keyToAction)
}

// GetPrefixAction returns the action for key in main prefix mode (Ctrl+B).
func (r *KeybindRegistry) GetPrefixAction(key string) string {
	return r.lookupInSection(key, r.config.Keybindings.PrefixMode)
}

// GetWindowPrefixAction returns the action for key in window prefix mode (Ctrl+B, t).
func (r *KeybindRegistry) GetWindowPrefixAction(key string) string {
	return r.lookupInSection(key, r.config.Keybindings.WindowPrefix)
}

// GetMinimizePrefixAction returns the action for key in minimize prefix mode (Ctrl+B, m).
func (r *KeybindRegistry) GetMinimizePrefixAction(key string) string {
	return r.lookupInSection(key, r.config.Keybindings.MinimizePrefix)
}

// GetWorkspacePrefixAction returns the action for key in workspace prefix mode (Ctrl+B, w).
func (r *KeybindRegistry) GetWorkspacePrefixAction(key string) string {
	return r.lookupInSection(key, r.config.Keybindings.WorkspacePrefix)
}

// GetDebugPrefixAction returns the action for key in debug prefix mode (Ctrl+B, D).
func (r *KeybindRegistry) GetDebugPrefixAction(key string) string {
	return r.lookupInSection(key, r.config.Keybindings.DebugPrefix)
}

// GetTapePrefixAction returns the action for key in tape prefix mode (Ctrl+B, T).
func (r *KeybindRegistry) GetTapePrefixAction(key string) string {
	return r.lookupInSection(key, r.config.Keybindings.TapePrefix)
}

// lookupInSection expands a single section on demand and resolves key against it.
func (r *KeybindRegistry) lookupInSection(key string, section map[string][]string) string {
	tmp := make(map[string]string)
	r.indexSection(tmp, section)
	return r.lookupKey(key, tmp)
}

// lookupKey resolves key in keyMap. Single letters are matched case-sensitively
// first (so M ≠ m), then fall back to lowercase; chords match lowercased.
func (r *KeybindRegistry) lookupKey(key string, keyMap map[string]string) string {
	key = strings.TrimSpace(key)
	if singleLetter(key) {
		if action, ok := keyMap[key]; ok {
			return action
		}
		return keyMap[strings.ToLower(key)]
	}
	return keyMap[strings.ToLower(key)]
}

// keysLookupSections lists the sections GetKeys scans for an action's raw keys.
// Note: includes the prefix sections and TerminalMode, but not Navigation,
// DebugPrefix, or TapePrefix — matching how callers query bindings for display.
func (r *KeybindRegistry) keysLookupSections() []map[string][]string {
	kb := &r.config.Keybindings
	return []map[string][]string{
		kb.WindowManagement,
		kb.Workspaces,
		kb.Layout,
		kb.ModeControl,
		kb.System,
		kb.Navigation,
		kb.RestoreMinimized,
		kb.PrefixMode,
		kb.WindowPrefix,
		kb.MinimizePrefix,
		kb.WorkspacePrefix,
		kb.TerminalMode,
	}
}

// GetKeys returns the configured keys for an action (first matching section wins).
func (r *KeybindRegistry) GetKeys(action string) []string {
	for _, section := range r.keysLookupSections() {
		if keys, ok := section[action]; ok {
			return keys
		}
	}
	return nil
}

// GetKeysForDisplay formats an action's keys for help, collapsing 3+ to "a, b, ...".
func (r *KeybindRegistry) GetKeysForDisplay(action string) string {
	keys := r.GetKeys(action)
	switch {
	case len(keys) == 0:
		return ""
	case len(keys) > 2:
		return strings.Join(keys[:2], ", ") + ", ..."
	default:
		return strings.Join(keys, ", ")
	}
}

// HasAction reports whether the action has any configured keys.
func (r *KeybindRegistry) HasAction(action string) bool {
	return len(r.GetKeys(action)) > 0
}

// Reload swaps in a new config and rebuilds the reverse map.
func (r *KeybindRegistry) Reload(cfg *UserConfig) {
	r.config = cfg
	r.buildMappings()
}

// GetConfig returns the registry's backing config.
func (r *KeybindRegistry) GetConfig() *UserConfig { return r.config }

// explicitActionSections maps actions whose presence in the user's file should
// disable hard-coded fallbacks, to the config section they live in.
var explicitActionSections = map[string]string{
	"new_window":      "window_management",
	"close_window":    "window_management",
	"rename_window":   "window_management",
	"minimize_window": "window_management",
	"restore_all":     "window_management",
	"next_window":     "window_management",
	"prev_window":     "window_management",
}

// IsActionExplicitlyConfigured reports whether the user's config defines the
// action (even with empty keys), signalling that defaults should not apply.
func (r *KeybindRegistry) IsActionExplicitlyConfigured(action string) bool {
	section := explicitActionSections[action]
	if section == "" {
		return false
	}

	var m map[string][]string
	switch section {
	case "window_management":
		m = r.config.Keybindings.WindowManagement
	case "workspaces":
		m = r.config.Keybindings.Workspaces
	case "layout":
		m = r.config.Keybindings.Layout
	case "mode_control":
		m = r.config.Keybindings.ModeControl
	case "system":
		m = r.config.Keybindings.System
	default:
		return false
	}

	_, ok := m[action]
	return ok
}

// ActionDescriptions maps action names to human-readable help text.
var ActionDescriptions = buildActionDescriptions()

func buildActionDescriptions() map[string]string {
	m := map[string]string{
		// window management
		"new_window":      "New window",
		"close_window":    "Close window",
		"rename_window":   "Rename window",
		"minimize_window": "Minimize window",
		"restore_all":     "Restore all minimized",
		"toggle_zoom":     "Toggle zoom (fullscreen)",
		"next_window":     "Next window",
		"prev_window":     "Previous window",

		// layout
		"snap_left":                 "Snap left",
		"snap_right":                "Snap right",
		"snap_fullscreen":           "Fullscreen",
		"unsnap":                    "Unsnap",
		"snap_corner_1":             "Snap to top-left",
		"snap_corner_2":             "Snap to top-right",
		"snap_corner_3":             "Snap to bottom-left",
		"snap_corner_4":             "Snap to bottom-right",
		"toggle_tiling":             "Toggle tiling mode",
		"swap_left":                 "Swap left",
		"swap_right":                "Swap right",
		"swap_up":                   "Swap up",
		"swap_down":                 "Swap down",
		"resize_master_shrink":      "Decrease master width",
		"resize_master_grow":        "Increase master width",
		"resize_height_shrink":      "Decrease focused window height",
		"resize_height_grow":        "Increase focused window height",
		"resize_master_shrink_left": "Decrease master width from left edge",
		"resize_master_grow_left":   "Increase master width from left edge",
		"resize_height_shrink_top":  "Decrease focused window height from top edge",
		"resize_height_grow_top":    "Increase focused window height from top edge",

		// BSP tiling
		"split_horizontal": "Split window horizontally (top/bottom)",
		"split_vertical":   "Split window vertically (left/right)",
		"rotate_split":     "Rotate split direction",
		"equalize_splits":  "Equalize all split ratios",
		"preselect_left":   "Preselect left for next window",
		"preselect_right":  "Preselect right for next window",
		"preselect_up":     "Preselect up for next window",
		"preselect_down":   "Preselect down for next window",

		// mode control
		"enter_terminal_mode": "Enter terminal mode",
		"enter_window_mode":   "Enter window management mode",
		"toggle_help":         "Toggle help",
		"quit":                "Quit",

		// clipboard
		"paste_clipboard": "Paste from clipboard",

		// system
		"toggle_logs":        "Toggle log viewer",
		"toggle_cache_stats": "Toggle cache statistics",
		"toggle_showkeys":    "Toggle showkeys overlay",

		// prefix mode (non-numeric)
		"prefix_new_window":       "Create new window",
		"prefix_close_window":     "Close current window",
		"prefix_rename_window":    "Rename window",
		"prefix_next_window":      "Next window",
		"prefix_prev_window":      "Previous window",
		"prefix_toggle_tiling":    "Toggle tiling mode",
		"prefix_workspace":        "Enter workspace prefix",
		"prefix_minimize":         "Enter minimize prefix",
		"prefix_window":           "Enter window prefix",
		"prefix_detach":           "Detach (daemon: session keeps running)",
		"prefix_selection":        "Enter copy/scrollback mode",
		"prefix_help":             "Toggle help",
		"prefix_logs":             "Toggle log viewer",
		"prefix_debug":            "Enter debug prefix",
		"prefix_tape":             "Enter tape manager prefix",
		"prefix_quit":             "Quit (daemon: kills session)",
		"prefix_fullscreen":       "Fullscreen current window",
		"prefix_split_horizontal": "Split window horizontally",
		"prefix_split_vertical":   "Split window vertically",
		"prefix_rotate_split":     "Rotate split direction",
		"prefix_equalize_splits":  "Equalize all splits",

		// tape prefix + actions
		"tape_prefix_manager": "Open tape manager",
		"tape_prefix_record":  "Start recording",
		"tape_prefix_stop":    "Stop recording",
		"tape_prefix_cancel":  "Cancel tape prefix",
		"toggle_tape_manager": "Toggle tape manager",
		"stop_recording":      "Stop tape recording",

		// debug prefix
		"debug_prefix_logs":       "Toggle log viewer",
		"debug_prefix_cache":      "Toggle cache statistics",
		"debug_prefix_animations": "Toggle animations",
		"debug_prefix_cancel":     "Cancel debug prefix",

		// terminal mode
		"terminal_next_window": "Next window (terminal mode)",
		"terminal_prev_window": "Previous window (terminal mode)",
		"terminal_exit_mode":   "Exit terminal mode (to window mode)",
	}

	// numeric families: select/switch/move/jump for 1..9 (and prefix jump 0).
	ordinals := []string{"", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for n := 1; n <= 9; n++ {
		o := ordinals[n]
		m["select_window_"+o] = "Select window " + o
		m["switch_workspace_"+o] = "Switch to workspace " + o
		m["move_and_follow_"+o] = "Move to workspace " + o + " and follow"
	}
	for n := 0; n <= 9; n++ {
		d := string(rune('0' + n))
		m["prefix_select_"+d] = "Jump to window " + d
	}
	return m
}
