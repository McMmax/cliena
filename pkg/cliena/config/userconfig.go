package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

// configRelPath is the config file location relative to the XDG config root.
const configRelPath = "cliena/config.toml"

// UserConfig is the top-level user-editable configuration tree.
type UserConfig struct {
	Appearance  AppearanceConfig  `toml:"appearance"`
	Keybindings KeybindingsConfig `toml:"keybindings"`
	Daemon      DaemonConfig      `toml:"daemon"`
	Hooks       HooksConfig       `toml:"hooks"`
}

// DaemonConfig configures the background session daemon.
type DaemonConfig struct {
	LogLevel     string `toml:"log_level"`     // off, errors, basic, messages, verbose, trace (default off)
	DefaultCodec string `toml:"default_codec"` // gob or json (default gob)
	SocketPath   string `toml:"socket_path"`   // empty → $XDG_RUNTIME_DIR/cliena/daemon.sock
}

// AppearanceConfig configures window chrome, dock, and render behavior.
// Pointer bools distinguish "unset" (nil → use default) from an explicit false.
type AppearanceConfig struct {
	BorderStyle          string `toml:"border_style"`           // rounded, normal, thick, double, hidden, block, ascii, outer-half-block, inner-half-block
	HideWindowButtons    bool   `toml:"hide_window_buttons"`    // hide minimize/maximize/close chrome
	HideScrollbar        bool   `toml:"hide_scrollbar"`         // hide the scrollbar thumb on the border
	ScrollbackLines      int    `toml:"scrollback_lines"`       // scrollback depth, clamped 100..1000000 (default 10000)
	DockbarPosition      string `toml:"dockbar_position"`       // bottom, top, hidden
	PreferredShell       string `toml:"preferred_shell"`        // empty → auto-detect per platform
	AnimationsEnabled    *bool  `toml:"animations_enabled"`     // default true; false for instant transitions
	ConfirmQuit          *bool  `toml:"confirm_quit"`           // default false; force quit dialog when true
	WhichKeyEnabled      *bool  `toml:"whichkey_enabled"`       // default true; which-key popup after leader
	WhichKeyPosition     string `toml:"whichkey_position"`      // bottom-right (default), bottom-left, top-right, top-left, center
	WindowTitlePosition  string `toml:"window_title_position"`  // bottom (default), top, hidden
	HideClock            bool   `toml:"hide_clock"`             // deprecated; prefer show_clock
	ShowClock            bool   `toml:"show_clock"`             // show clock overlay (default false)
	ShowCPU              bool   `toml:"show_cpu"`               // show dock CPU graph (default false)
	ShowRAM              bool   `toml:"show_ram"`               // show dock RAM usage (default false)
	Theme                string `toml:"theme"`                  // color theme name; empty → terminal default
	SharedBorders        *bool  `toml:"shared_borders"`         // share borders between adjacent tiled windows (default false)
	Gap                  int    `toml:"gap"`                    // cells of gap between tiled panes (default 0)
	BorderFocusedColor   string `toml:"border_focused_color"`   // hex color for the focused pane border
	BorderUnfocusedColor string `toml:"border_unfocused_color"` // hex color for unfocused pane borders
	WindowTitleFormat    string `toml:"window_title_format"`    // title template: {title}, {index}, {cwd}
	SeparatorStyle       string `toml:"separator_style"`        // rounded (default), powerline, flat, none
	ZoomMaxWidth         int    `toml:"zoom_max_width"`         // zoom width cap in cells (0 = fullscreen)
	NiriReverseScroll    bool   `toml:"niri_reverse_scroll"`    // flip wheel direction in niri scrolling (default false)
	MaxFPS               int    `toml:"max_fps"`                // render FPS cap (default 60, max 120)
}

// HooksConfig maps event names to shell command hooks.
type HooksConfig map[string]any

// KeybindingsConfig holds every keybinding section. Each section maps an action
// name to the keys that trigger it. Pointer-free maps keep TOML round-tripping simple.
type KeybindingsConfig struct {
	LeaderKey        string              `toml:"leader_key"` // prefix key for command mode (default ctrl+b)
	WindowManagement map[string][]string `toml:"window_management"`
	Workspaces       map[string][]string `toml:"workspaces"`
	Layout           map[string][]string `toml:"layout"`
	ModeControl      map[string][]string `toml:"mode_control"`
	System           map[string][]string `toml:"system"`
	Navigation       map[string][]string `toml:"navigation"`
	RestoreMinimized map[string][]string `toml:"restore_minimized"`
	PrefixMode       map[string][]string `toml:"prefix_mode"`
	WindowPrefix     map[string][]string `toml:"window_prefix"`
	MinimizePrefix   map[string][]string `toml:"minimize_prefix"`
	WorkspacePrefix  map[string][]string `toml:"workspace_prefix"`
	DebugPrefix      map[string][]string `toml:"debug_prefix"`
	TapePrefix       map[string][]string `toml:"tape_prefix"`
	TerminalMode     map[string][]string `toml:"terminal_mode"` // direct binds in terminal mode, no prefix
}

// numberedBinds builds a "<prefix>_N" → {keys...} table for N in 1..9, taking
// the per-index key list from the provided generator. cuts repetitive literals.
func numberedBinds(actionPrefix string, keys func(n int) []string) map[string][]string {
	m := make(map[string][]string, 9)
	for n := 1; n <= 9; n++ {
		m[fmt.Sprintf("%s_%d", actionPrefix, n)] = keys(n)
	}
	return m
}

// digit returns the decimal digit n as a one-key slice.
func digit(n int) []string { return []string{fmt.Sprintf("%d", n)} }

// shiftedDigits are the shifted number-row symbols, indexed 1..9.
var shiftedDigits = [10]string{"", "!", "@", "#", "$", "%", "^", "&", "*", "("}

// DefaultConfig returns the built-in configuration: every section populated with
// the default keys the TUI ships with.
func DefaultConfig() *UserConfig {
	windowMgmt := map[string][]string{
		"new_window":      {"n"},
		"close_window":    {"w", "x"},
		"rename_window":   {"r"},
		"minimize_window": {"m"},
		"restore_all":     {"M"},
		"toggle_zoom":     {"z"},
		"next_window":     {"tab"},
		"prev_window":     {"shift+tab"},
	}
	for n := 1; n <= 9; n++ {
		windowMgmt[fmt.Sprintf("select_window_%d", n)] = digit(n)
	}

	prefixMode := map[string][]string{
		"prefix_new_window":       {"c"},
		"prefix_close_window":     {"x"},
		"prefix_rename_window":    {",", "r"},
		"prefix_next_window":      {"n", "tab"},
		"prefix_prev_window":      {"p", "shift+tab"},
		"prefix_toggle_tiling":    {"space"},
		"prefix_workspace":        {"w"},
		"prefix_minimize":         {"m"},
		"prefix_window":           {"t"},
		"prefix_detach":           {"d", "esc"},
		"prefix_selection":        {"["},
		"prefix_help":             {"?"},
		"prefix_debug":            {"D"},
		"prefix_tape":             {"T"},
		"prefix_quit":             {"q"},
		"prefix_fullscreen":       {"z"},
		"prefix_split_horizontal": {"-"},
		"prefix_split_vertical":   {"|", "\\"},
		"prefix_rotate_split":     {"R"},
		"prefix_equalize_splits":  {"="},
		"prefix_scrollback":       {"s"},
	}
	// prefix window jumps 0..9
	for n := 0; n <= 9; n++ {
		prefixMode[fmt.Sprintf("prefix_select_%d", n)] = digit(n)
	}

	minimizePrefix := map[string][]string{
		"minimize_prefix_focused":     {"m"},
		"minimize_prefix_restore_all": {"M"},
		"minimize_prefix_cancel":      {"esc"},
	}
	for n := 1; n <= 9; n++ {
		minimizePrefix[fmt.Sprintf("minimize_prefix_restore_%d", n)] = digit(n)
	}

	workspacePrefix := map[string][]string{"workspace_prefix_cancel": {"esc"}}
	for n := 1; n <= 9; n++ {
		workspacePrefix[fmt.Sprintf("workspace_prefix_switch_%d", n)] = digit(n)
		workspacePrefix[fmt.Sprintf("workspace_prefix_move_%d", n)] = []string{shiftedDigits[n]}
	}

	cfg := &UserConfig{
		Appearance: AppearanceConfig{
			BorderStyle:       "rounded",
			HideWindowButtons: false,
			ScrollbackLines:   10000,
			DockbarPosition:   "bottom",
			PreferredShell:    "",
		},
		Daemon: DaemonConfig{
			LogLevel:     "off",
			DefaultCodec: "gob",
			SocketPath:   "", // empty → default XDG path
		},
		Keybindings: KeybindingsConfig{
			LeaderKey:        "ctrl+b",
			WindowManagement: windowMgmt,
			Workspaces:       defaultWorkspaceBinds(),
			Layout:           defaultLayoutBinds(),
			ModeControl: map[string][]string{
				"enter_terminal_mode": {"i", "enter"},
				"enter_window_mode":   {"esc"},
				"toggle_help":         {"?"},
				"quit":                {"q"},
			},
			// debug/log toggles live behind the Ctrl+B D submenu, not bindable here
			System: map[string][]string{},
			Navigation: map[string][]string{
				"nav_up":       {"up"},
				"nav_down":     {"down"},
				"nav_left":     {"left"},
				"nav_right":    {"right"},
				"extend_up":    {"shift+up"},
				"extend_down":  {"shift+down"},
				"extend_left":  {"shift+left"},
				"extend_right": {"shift+right"},
			},
			RestoreMinimized: numberedBinds("restore_minimized", func(n int) []string {
				return []string{fmt.Sprintf("shift+%d", n), shiftedDigits[n]}
			}),
			PrefixMode: prefixMode,
			WindowPrefix: map[string][]string{
				"window_prefix_new":    {"n"},
				"window_prefix_close":  {"x"},
				"window_prefix_rename": {"r"},
				"window_prefix_next":   {"tab"},
				"window_prefix_prev":   {"shift+tab"},
				"window_prefix_tiling": {"t"},
				"window_prefix_cancel": {"esc"},
			},
			MinimizePrefix:  minimizePrefix,
			WorkspacePrefix: workspacePrefix,
			DebugPrefix: map[string][]string{
				"debug_prefix_logs":       {"l"},
				"debug_prefix_cache":      {"c"},
				"debug_prefix_animations": {"a"},
				"debug_prefix_cancel":     {"esc"},
			},
			TapePrefix: map[string][]string{
				"tape_prefix_manager": {"m"},
				"tape_prefix_record":  {"r"},
				"tape_prefix_stop":    {"s"},
				"tape_prefix_cancel":  {"esc"},
			},
			TerminalMode: defaultTerminalModeBinds(),
		},
	}
	return cfg
}

// modKey returns "opt+suffix" on macOS, "alt+suffix" elsewhere. macOS users
// think in the Option key; normalization later expands opt+ into its variants.
func modKey(suffix string) string {
	if isMacOS() {
		return "opt+" + suffix
	}
	return "alt+" + suffix
}

// defaultTerminalModeBinds returns the no-prefix binds active inside a terminal.
func defaultTerminalModeBinds() map[string][]string {
	return map[string][]string{
		"terminal_next_window": {modKey("n")},
		"terminal_prev_window": {modKey("p")},
		"terminal_exit_mode":   {modKey("esc")},
	}
}

// defaultWorkspaceBinds returns workspace switch/move binds, keyed off the
// platform modifier (opt on macOS, alt elsewhere).
func defaultWorkspaceBinds() map[string][]string {
	m := make(map[string][]string, 18)
	for n := 1; n <= 9; n++ {
		d := fmt.Sprintf("%d", n)
		m[fmt.Sprintf("switch_workspace_%d", n)] = []string{modKey(d)}
		m[fmt.Sprintf("move_and_follow_%d", n)] = []string{modKey("shift+" + d)}
	}
	return m
}

// defaultLayoutBinds returns snap/swap/resize/BSP binds, including
// platform-specific preselect chords.
func defaultLayoutBinds() map[string][]string {
	layout := map[string][]string{
		"snap_left":                 {"h"},
		"snap_right":                {"l"},
		"snap_fullscreen":           {"f"},
		"unsnap":                    {"u"},
		"snap_corner_1":             {"1"},
		"snap_corner_2":             {"2"},
		"snap_corner_3":             {"3"},
		"snap_corner_4":             {"4"},
		"toggle_tiling":             {"t"},
		"swap_left":                 {"H", "ctrl+left"},
		"swap_right":                {"L", "ctrl+right"},
		"swap_up":                   {"K", "ctrl+up"},
		"swap_down":                 {"J", "ctrl+down"},
		"resize_master_shrink":      {"<", "shift+,"},
		"resize_master_grow":        {">", "shift+."},
		"resize_height_shrink":      {"{", "shift+["},
		"resize_height_grow":        {"}", "shift+]"},
		"resize_master_shrink_left": {","},
		"resize_master_grow_left":   {"."},
		"resize_height_shrink_top":  {"["},
		"resize_height_grow_top":    {"]"},
		"split_horizontal":          {"-"},
		"split_vertical":            {"|", "\\"},
		"rotate_split":              {"R"},
		"equalize_splits":           {"="},
	}
	layout["preselect_left"] = []string{modKey("h")}
	layout["preselect_right"] = []string{modKey("l")}
	layout["preselect_up"] = []string{modKey("k")}
	layout["preselect_down"] = []string{modKey("j")}
	return layout
}

// isMacOS detects macOS, trusting GOOS then env hints for cross-compilation.
func isMacOS() bool {
	if runtime.GOOS == "darwin" {
		return true
	}
	env := strings.ToLower(os.Getenv("GOOS") + " " + os.Getenv("OSTYPE"))
	return strings.Contains(env, "darwin")
}

// LoadUserConfig reads the user config from XDG, creating defaults if absent.
// Missing sections are backfilled, the result validated, and fatal validation
// errors are returned; warnings are surfaced but non-fatal.
func LoadUserConfig() (*UserConfig, error) {
	configPath, err := xdg.SearchConfigFile(configRelPath)
	if err != nil {
		return createDefaultConfig() // no file yet → write and return defaults
	}

	// #nosec G304 - configPath comes from XDG search; reading user config is intended
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg UserConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	applyDefaults(&cfg)

	validation := ValidateConfig(&cfg)
	if validation.HasErrors() {
		for _, e := range validation.Errors {
			fmt.Fprintf(os.Stderr, "Config error in [%s]: %s - %s\n", e.Field, e.Key, e.Message)
		}
		return nil, fmt.Errorf("configuration has %d error(s), please fix and restart", len(validation.Errors))
	}
	for _, w := range validation.Warnings {
		tea.Println(fmt.Sprintf("Config warning in [%s]: %s - %s", w.Field, w.Key, w.Message))
	}

	return &cfg, nil
}

// applyDefaults backfills every section of cfg from DefaultConfig in place.
func applyDefaults(cfg *UserConfig) {
	def := DefaultConfig()
	fillMissingAppearance(cfg, def)
	fillMissingDaemon(cfg, def)
	fillMissingKeybinds(cfg, def)
}

// createDefaultConfig writes a commented default config to the XDG path and
// returns the in-memory defaults.
func createDefaultConfig() (*UserConfig, error) {
	cfg := DefaultConfig()

	configPath, err := xdg.ConfigFile(configRelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	contents := defaultConfigHeader(configPath) + string(data)
	if err := os.WriteFile(configPath, []byte(contents), 0600); err != nil {
		return nil, fmt.Errorf("failed to write config file: %w", err)
	}

	return cfg, nil
}

// defaultConfigHeader builds the comment banner prepended to a fresh config.
func defaultConfigHeader(configPath string) string {
	var b strings.Builder
	b.WriteString("# cliena Configuration File\n")
	b.WriteString("# This file allows you to customize appearance and keybindings\n")
	b.WriteString("#\n")
	b.WriteString("# Configuration location: " + configPath + "\n")
	b.WriteString("# For keybindings documentation, run: cliena keybinds list\n\n")
	b.WriteString("# ============================================================================\n")
	b.WriteString("# APPEARANCE SETTINGS\n")
	b.WriteString("# ============================================================================\n")
	b.WriteString("# border_style: rounded, normal, thick, double, hidden, block, ascii,\n")
	b.WriteString("#               outer-half-block, inner-half-block (default: rounded)\n")
	b.WriteString("# dockbar_position: bottom, top, hidden (default: bottom)\n")
	b.WriteString("# hide_window_buttons: hide minimize/maximize/close (default: false)\n")
	b.WriteString("# scrollback_lines: 100 to 1000000 (default: 10000)\n")
	b.WriteString("# shared_borders: share borders between tiled windows (default: false)\n")
	b.WriteString("# theme: color theme name, empty for terminal default. CLI --theme wins.\n")
	b.WriteString("#        custom themes live in ~/.config/cliena/themes/*.json\n")
	b.WriteString("# ============================================================================\n\n")
	return b.String()
}

// fillMissingAppearance backfills appearance fields and pushes explicitly-set
// values into the matching runtime globals.
func fillMissingAppearance(cfg, def *UserConfig) {
	ap := &cfg.Appearance

	if ap.BorderStyle == "" {
		ap.BorderStyle = def.Appearance.BorderStyle
	}
	if ap.DockbarPosition == "" {
		ap.DockbarPosition = def.Appearance.DockbarPosition
	}

	// scrollback: 0/negative → default, otherwise clamp to [100, 1000000]
	switch {
	case ap.ScrollbackLines <= 0:
		ap.ScrollbackLines = def.Appearance.ScrollbackLines
	case ap.ScrollbackLines < 100:
		ap.ScrollbackLines = 100
	case ap.ScrollbackLines > 1000000:
		ap.ScrollbackLines = 1000000
	}

	// pointer bools: only override the global when the user set them explicitly
	if ap.AnimationsEnabled != nil {
		AnimationsEnabled = *ap.AnimationsEnabled
	}
	if ap.ConfirmQuit != nil {
		AlwaysConfirmQuit = *ap.ConfirmQuit
	}
	if ap.SharedBorders != nil {
		SharedBorders = *ap.SharedBorders
	}
	if ap.WhichKeyEnabled != nil {
		WhichKeyEnabled = *ap.WhichKeyEnabled
	}

	if ap.WhichKeyPosition != "" {
		WhichKeyPosition = ap.WhichKeyPosition
	}
	// title position: respect a flag already set before load (still at default "bottom")
	if ap.WindowTitlePosition != "" && WindowTitlePosition == "bottom" {
		WindowTitlePosition = ap.WindowTitlePosition
	}
	// clock: a flag may have set it true already; don't clear that
	if !HideClock {
		HideClock = ap.HideClock
	}
	if ap.ZoomMaxWidth > 0 {
		ZoomMaxWidth = ap.ZoomMaxWidth
	}
}

// fillMissingDaemon backfills daemon fields. SocketPath stays empty to mean
// "use the XDG default", so it is never overridden here.
func fillMissingDaemon(cfg, def *UserConfig) {
	if cfg.Daemon.LogLevel == "" {
		cfg.Daemon.LogLevel = def.Daemon.LogLevel
	}
	if cfg.Daemon.DefaultCodec == "" {
		cfg.Daemon.DefaultCodec = def.Daemon.DefaultCodec
	}
}

// fillMissingKeybinds ensures every keybinding map is non-nil and backfills any
// action absent from the user's file with its default keys.
func fillMissingKeybinds(cfg, def *UserConfig) {
	kb := &cfg.Keybindings
	dkb := &def.Keybindings

	if kb.LeaderKey == "" {
		kb.LeaderKey = dkb.LeaderKey
	}

	// pair each user section with its default source, then merge
	pairs := []struct{ target, src *map[string][]string }{
		{&kb.WindowManagement, &dkb.WindowManagement},
		{&kb.Workspaces, &dkb.Workspaces},
		{&kb.Layout, &dkb.Layout},
		{&kb.ModeControl, &dkb.ModeControl},
		{&kb.System, &dkb.System},
		{&kb.Navigation, &dkb.Navigation},
		{&kb.RestoreMinimized, &dkb.RestoreMinimized},
		{&kb.PrefixMode, &dkb.PrefixMode},
		{&kb.WindowPrefix, &dkb.WindowPrefix},
		{&kb.MinimizePrefix, &dkb.MinimizePrefix},
		{&kb.WorkspacePrefix, &dkb.WorkspacePrefix},
		{&kb.DebugPrefix, &dkb.DebugPrefix},
		{&kb.TapePrefix, &dkb.TapePrefix},
		{&kb.TerminalMode, &dkb.TerminalMode},
	}
	for _, p := range pairs {
		if *p.target == nil {
			*p.target = make(map[string][]string)
		}
		for action, keys := range *p.src {
			if _, set := (*p.target)[action]; !set {
				(*p.target)[action] = keys
			}
		}
	}
}

// GetConfigPath returns the existing config path, or where it would be created.
func GetConfigPath() (string, error) {
	if path, err := xdg.SearchConfigFile(configRelPath); err == nil {
		return path, nil
	}
	return xdg.ConfigFile(configRelPath)
}
