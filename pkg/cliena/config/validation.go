package config

import (
	"fmt"
	"strings"
)

// ValidationError is one validation finding (error or warning).
type ValidationError struct {
	Field   string
	Key     string
	Message string
}

// ValidationResult collects validation errors and warnings.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationError
}

// HasErrors reports whether any errors were recorded.
func (vr *ValidationResult) HasErrors() bool { return len(vr.Errors) > 0 }

// HasWarnings reports whether any warnings were recorded.
func (vr *ValidationResult) HasWarnings() bool { return len(vr.Warnings) > 0 }

func (vr *ValidationResult) addError(field, key, msg string) {
	vr.Errors = append(vr.Errors, ValidationError{Field: field, Key: key, Message: msg})
}

func (vr *ValidationResult) addWarning(field, key, msg string) {
	vr.Warnings = append(vr.Warnings, ValidationError{Field: field, Key: key, Message: msg})
}

// namedSection pairs a config section with its TOML field name for reporting.
type namedSection struct {
	name  string
	binds map[string][]string
}

// validatedSections returns the sections that get per-key validation and the
// macOS alt+ advisory. Prefix/terminal/navigation sections are excluded to
// mirror the original surface.
func validatedSections(cfg *UserConfig) []namedSection {
	kb := &cfg.Keybindings
	return []namedSection{
		{"window_management", kb.WindowManagement},
		{"workspaces", kb.Workspaces},
		{"layout", kb.Layout},
		{"mode_control", kb.ModeControl},
		{"system", kb.System},
		{"prefix_mode", kb.PrefixMode},
		{"window_prefix", kb.WindowPrefix},
		{"minimize_prefix", kb.MinimizePrefix},
		{"workspace_prefix", kb.WorkspacePrefix},
	}
}

// ValidateConfig checks keybindings for malformed keys, empty actions, in-context
// conflicts, missing essentials, and (on macOS) alt+ usage.
func ValidateConfig(cfg *UserConfig) *ValidationResult {
	result := &ValidationResult{}
	normalizer := NewKeyNormalizer()

	sections := validatedSections(cfg)

	// per-key validation; empty action lists are a soft warning
	for _, s := range sections {
		for action, keys := range s.binds {
			if len(keys) == 0 {
				result.addWarning(s.name, action, fmt.Sprintf("Action '%s' has no keybindings", action))
				continue
			}
			for _, key := range keys {
				if ok, reason := normalizer.ValidateKey(key); !ok {
					result.addError(s.name, key, reason)
				}
			}
		}
	}

	if cfg.Keybindings.LeaderKey != "" {
		if ok, reason := normalizer.ValidateKey(cfg.Keybindings.LeaderKey); !ok {
			result.addError("keybindings", "leader_key", reason)
		}
	}

	for key, actions := range findConflicts(cfg, normalizer) {
		result.addWarning("keybindings", key,
			fmt.Sprintf("Key '%s' is bound to multiple actions: %s", key, strings.Join(actions, ", ")))
	}

	// essentials make cliena usable; warn (not error) if any is unbound
	essentials := []struct{ action, section string }{
		{"new_window", "window_management"},
		{"close_window", "window_management"},
		{"enter_terminal_mode", "mode_control"},
		{"enter_window_mode", "mode_control"},
		{"quit", "mode_control"},
	}
	for _, e := range essentials {
		if !hasKeybinding(cfg, e.section, e.action) {
			result.addWarning(e.section, e.action,
				fmt.Sprintf("Essential action '%s' has no keybinding - cliena may be difficult to use", e.action))
		}
	}

	if normalizer.IsMacOS() {
		for _, s := range sections {
			for action, keys := range s.binds {
				for _, key := range keys {
					if strings.HasPrefix(strings.ToLower(strings.TrimSpace(key)), "alt+") {
						result.addWarning(s.name, key, fmt.Sprintf(
							"Action '%s': On macOS, consider using 'opt+' instead of 'alt+' for consistency with your keyboard (⌥ Option key)", action))
					}
				}
			}
		}
	}

	return result
}

// conflictContexts groups actions that are mutually exclusive at runtime. Two
// actions only conflict on a shared key if they live in the same group, or if
// both are context-free (no group).
var conflictContexts = [][]string{
	{ // tiling mode
		"select_window_1", "select_window_2", "select_window_3", "select_window_4",
		"select_window_5", "select_window_6", "select_window_7", "select_window_8", "select_window_9",
		"swap_left", "swap_right", "swap_up", "swap_down",
	},
	{ // floating (non-tiling) mode
		"snap_corner_1", "snap_corner_2", "snap_corner_3", "snap_corner_4",
		"snap_left", "snap_right", "snap_fullscreen", "unsnap",
	},
	{ // selection mode
		"toggle_selection", "toggle_selection_term", "copy_selection", "paste_clipboard", "clear_selection",
	},
}

// contextOf returns the index of the conflict group containing action, or -1
// for a context-free action.
func contextOf(action string) int {
	for i, group := range conflictContexts {
		for _, a := range group {
			if a == action {
				return i
			}
		}
	}
	return -1
}

// actionsConflict reports whether two actions sharing a key really clash: same
// context group, or both context-free.
func actionsConflict(a, b string) bool {
	ca, cb := contextOf(a), contextOf(b)
	if ca == -1 && cb == -1 {
		return true // both always-active
	}
	return ca == cb && ca != -1
}

// findConflicts returns keys bound to multiple actions that conflict in context.
func findConflicts(cfg *UserConfig, normalizer *KeyNormalizer) map[string][]string {
	kb := &cfg.Keybindings
	scanned := []map[string][]string{
		kb.WindowManagement,
		kb.Workspaces,
		kb.Layout,
		kb.ModeControl,
		kb.System,
	}

	// key → set of distinct actions bound to it
	keyActions := make(map[string]map[string]struct{})
	for _, section := range scanned {
		for action, keys := range section {
			for _, key := range normalizer.ExpandKeys(keys) {
				nk := conflictKey(key)
				if keyActions[nk] == nil {
					keyActions[nk] = make(map[string]struct{})
				}
				keyActions[nk][action] = struct{}{}
			}
		}
	}

	conflicts := make(map[string][]string)
	for key, set := range keyActions {
		if len(set) < 2 {
			continue
		}
		actions := make([]string, 0, len(set))
		for a := range set {
			actions = append(actions, a)
		}
		if anyPairConflicts(actions) {
			conflicts[key] = actions
		}
	}
	return conflicts
}

// anyPairConflicts reports whether any two actions in the slice truly conflict.
func anyPairConflicts(actions []string) bool {
	for i := range actions {
		for j := i + 1; j < len(actions); j++ {
			if actionsConflict(actions[i], actions[j]) {
				return true
			}
		}
	}
	return false
}

// conflictKey normalizes a key for conflict comparison: single letters keep
// their case (M ≠ m), everything else lowercases.
func conflictKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if singleLetter(trimmed) {
		return trimmed
	}
	return strings.ToLower(trimmed)
}

// hasKeybinding reports whether the named section binds action to ≥1 key.
func hasKeybinding(cfg *UserConfig, sectionName, action string) bool {
	kb := &cfg.Keybindings
	var section map[string][]string
	switch sectionName {
	case "window_management":
		section = kb.WindowManagement
	case "workspaces":
		section = kb.Workspaces
	case "layout":
		section = kb.Layout
	case "mode_control":
		section = kb.ModeControl
	case "system":
		section = kb.System
	case "prefix_mode":
		section = kb.PrefixMode
	case "window_prefix":
		section = kb.WindowPrefix
	case "minimize_prefix":
		section = kb.MinimizePrefix
	case "workspace_prefix":
		section = kb.WorkspacePrefix
	default:
		return false
	}
	return len(section[action]) > 0
}
