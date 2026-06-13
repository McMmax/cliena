// Package config holds tunable constants, runtime appearance flags, user
// settings loading, and the keybinding registry that drives the TUI.
package config

import (
	"time"

	"charm.land/lipgloss/v2"
	"github.com/lrstanley/go-nf/glyphs/fa"
	"github.com/lrstanley/go-nf/glyphs/ple"
)

// Window sizing defaults, in terminal cells.
const (
	// DefaultWindowWidth is the spawn width for a new floating window.
	DefaultWindowWidth = 20
	// DefaultWindowHeight is the spawn height for a new floating window.
	DefaultWindowHeight = 5
	// MinWindowWidth is the floor a window can be resized to.
	MinWindowWidth = 10
	// MinWindowHeight is the floor a window can be resized to.
	MinWindowHeight = 3
)

// Animation timing for minimize, restore, snap, and notification effects.
const (
	// DefaultAnimationDuration covers minimize/restore transitions.
	DefaultAnimationDuration = 300 * time.Millisecond
	// FastAnimationDuration covers snapping and window swaps.
	FastAnimationDuration = 200 * time.Millisecond
	// NotificationFadeOutDuration is how long a notification takes to fade.
	NotificationFadeOutDuration = 500 * time.Millisecond
	// NotificationDuration is how long a notification stays fully visible.
	NotificationDuration = 1500 * time.Millisecond
)

// Timeouts and polling intervals across the runtime.
const (
	// PrefixCommandTimeout aborts prefix mode if no follow-up key arrives.
	PrefixCommandTimeout = 2 * time.Second
	// CPUUpdateInterval gates how often the CPU graph resamples.
	CPUUpdateInterval = 500 * time.Millisecond
	// ProcessWaitDelay is the grace pause while reaping a child process.
	ProcessWaitDelay = 50 * time.Millisecond
	// WhichKeyDelay defers the which-key overlay so fast chords don't flash it.
	WhichKeyDelay = 500 * time.Millisecond
	// ProcessShutdownTimeout bounds a graceful child shutdown before SIGKILL.
	ProcessShutdownTimeout = 500 * time.Millisecond
)

// Render rates. mutable — appearance.max_fps can raise NormalFPS at load.
var (
	// NormalFPS is the steady-state render rate. capped 10..120 via max_fps.
	NormalFPS = 60
	// InteractionFPS throttles renders during drag/resize for input responsiveness.
	InteractionFPS = 30
	// IdleFPS is the near-zero-cost rate once output stalls.
	IdleFPS = 10
	// IdleThresholdFrames is how many idle NormalFPS frames trigger the drop to IdleFPS (~500ms).
	IdleThresholdFrames = 30
	// BackgroundWindowUpdateCycle skips this many cycles between background window repaints.
	BackgroundWindowUpdateCycle = 3
)

// UI layout dimensions, in cells.
const (
	// DockHeight is the dock strip height.
	DockHeight = 2
	// StatusBarLeftWidth is the left status segment width.
	StatusBarLeftWidth = 30
	// LogViewerWidth is the log overlay width.
	LogViewerWidth = 80
	// CPUGraphWidth is the CPU graph width including its label.
	CPUGraphWidth = 19
	// CPUGraphBars is how many bars the CPU graph draws.
	CPUGraphBars = 10
	// CPUGraphScale maps a 0..100 percent onto 8 block glyphs (100/8).
	CPUGraphScale = 12.5
	// LeftInfoWidth is the dock left info width.
	LeftInfoWidth = 30
	// RightInfoWidth is the dock right info width.
	RightInfoWidth = 20
	// DockItemWidth is a single dock item's base width.
	DockItemWidth = 6
	// MaxNotificationWidth caps notification width.
	MaxNotificationWidth = 60
	// MinNotificationWidth floors notification width.
	MinNotificationWidth = 20
	// NotificationMargin is the gap from the screen edge.
	NotificationMargin = 8
	// NotificationSpacing is the vertical gap between stacked notifications.
	NotificationSpacing = 4
	// MaxVisibleNotifications caps simultaneously shown notifications.
	MaxVisibleNotifications = 3
	// AnimationMargin pads the cull box for animating windows.
	AnimationMargin = 20
	// VisibilityMargin pads the cull box for static windows.
	VisibilityMargin = 5
	// MaxNameLengthDock truncates window names shown in the dock.
	MaxNameLengthDock = 12
	// MinimizedDockWidth is the minimized-window thumbnail width.
	MinimizedDockWidth = 5
	// MinimizedDockHeight is the minimized-window thumbnail height.
	MinimizedDockHeight = 3
)

// Nerd Font dock glyphs. populated in init() since the glyph values are runtime
// strings from go-nf, not compile-time constants.
var (
	// DockPillLeftChar is the left pill cap.
	DockPillLeftChar string
	// DockPillRightChar is the right pill cap.
	DockPillRightChar string
	// DockModeIconWindow marks window mode.
	DockModeIconWindow string
	// DockModeIconTerminal marks terminal mode.
	DockModeIconTerminal string
	// DockModeIconTiling marks tiling mode.
	DockModeIconTiling string
	// DockIconTerminalCount precedes the terminal count.
	DockIconTerminalCount string
	// DockIconWorkspaceCount precedes the workspace count.
	DockIconWorkspaceCount string
	// DockSeparator divides dock sections.
	DockSeparator = "  " // two spaces, breathing room between segments
	// WindowPillLeft is the left pill cap on window chrome.
	WindowPillLeft string
	// WindowPillRight is the right pill cap on window chrome.
	WindowPillRight string
)

func init() {
	DockPillLeftChar = ple.LeftHalfCircleThick.String()
	DockPillRightChar = ple.RightHalfCircleThick.String()
	DockModeIconWindow = " " + fa.WindowRestore.String() + " "
	DockModeIconTerminal = " " + fa.Terminal.String() + " "
	DockModeIconTiling = " " + fa.Th.String() + " "
	DockIconTerminalCount = fa.Terminal.String()
	DockIconWorkspaceCount = fa.ThLarge.String()
	WindowPillLeft = ple.LeftHalfCircleThick.String()
	WindowPillRight = ple.RightHalfCircleThick.String()
}

// ASCII fallbacks for the dock glyphs, used under --ascii-only.
const (
	// DockPillLeftCharASCII is the ASCII left pill cap.
	DockPillLeftCharASCII = "["
	// DockPillRightCharASCII is the ASCII right pill cap.
	DockPillRightCharASCII = "]"
	// DockModeIconWindowASCII marks window mode in ASCII.
	DockModeIconWindowASCII = " W "
	// DockModeIconTerminalASCII marks terminal mode in ASCII.
	DockModeIconTerminalASCII = " T "
	// DockModeIconTilingASCII marks tiling mode in ASCII.
	DockModeIconTilingASCII = " # "
	// DockIconTerminalCountASCII precedes the terminal count in ASCII.
	DockIconTerminalCountASCII = "win"
	// DockIconWorkspaceCountASCII precedes the workspace count in ASCII.
	DockIconWorkspaceCountASCII = "ws"
	// DockSeparatorASCII divides dock sections in ASCII.
	DockSeparatorASCII = " | "
)

// Tape manager labels.
const (
	// TapeManagerTitle titles the tape manager overlay.
	TapeManagerTitle = "Tape Manager"
	// TapeRecordingIndicator flags an active recording.
	TapeRecordingIndicator = "[REC]"
	// TapeSuccessIcon flags a completed action.
	TapeSuccessIcon = "[OK]"
	// TapeSelectedIcon marks the selected tape row.
	TapeSelectedIcon = ">"
)

// Notification glyphs, ASCII-safe so they render everywhere.
const (
	// NotificationIconError prefixes error notifications.
	NotificationIconError = "[X]"
	// NotificationIconWarning prefixes warning notifications.
	NotificationIconWarning = "[!]"
	// NotificationIconSuccess prefixes success notifications.
	NotificationIconSuccess = "[OK]"
	// NotificationIconInfo prefixes info notifications.
	NotificationIconInfo = "[i]"
)

// Dock mode indicator colors, hex.
const (
	// DockColorWindow tints the window-mode indicator.
	DockColorWindow = "#4865f2"
	// DockColorTerminal tints the terminal-mode indicator.
	DockColorTerminal = "#4ade80"
	// DockColorCopy tints the copy-mode indicator.
	DockColorCopy = "#fb923c"
)

// Runtime appearance flags. mutable — set by CLI flags then user config at startup.

// UseASCIIOnly swaps Nerd Font glyphs for ASCII. set via --ascii-only.
var UseASCIIOnly = false

// AnimationsEnabled toggles UI animations. set via --no-animations or config.
var AnimationsEnabled = true

// AnimationsSuppressed force-disables animations for the duration of a batch
// operation (e.g. replaying a remote command); wins over AnimationsEnabled.
var AnimationsSuppressed = false

// AlwaysConfirmQuit forces the quit dialog even with no live processes. set via confirm_quit.
var AlwaysConfirmQuit = false

// WhichKeyEnabled shows the which-key popup after the leader key.
var WhichKeyEnabled = true

// WhichKeyPosition anchors the which-key popup: bottom-right, bottom-left,
// top-right, top-left, or center.
var WhichKeyPosition = "bottom-right"

// GetAnimationDuration returns the standard animation length, or 0 (instant)
// when animations are off or suppressed.
func GetAnimationDuration() time.Duration {
	if !AnimationsEnabled || AnimationsSuppressed {
		return 0
	}
	return DefaultAnimationDuration
}

// GetFastAnimationDuration returns the fast animation length, or 0 (instant)
// when animations are off or suppressed.
func GetFastAnimationDuration() time.Duration {
	if !AnimationsEnabled || AnimationsSuppressed {
		return 0
	}
	return FastAnimationDuration
}

// SharedBorders collapses adjacent tiled borders into one shared line.
var SharedBorders = false

// BorderStyle names the active window border style.
var BorderStyle = "rounded"

// DockbarPosition places the dock: bottom, top, or hidden.
var DockbarPosition = "bottom"

// HideWindowButtons hides the minimize/maximize/close chrome.
var HideWindowButtons = false

// HideScrollbar hides the scrollbar thumb. effectively forced true when
// BorderStyle is "hidden" — no border exists to host the thumb.
var HideScrollbar = false

// WindowTitlePosition places the window title: bottom, top, or hidden.
var WindowTitlePosition = "bottom"

// HideClock hides the clock overlay.
// Deprecated: prefer ShowClock; HideClock wins when true.
var HideClock = false

// ShowClock shows the clock overlay (off by default).
var ShowClock = false

// ShowCPU shows the dock CPU graph (off by default).
var ShowCPU = false

// ShowRAM shows dock RAM usage (off by default).
var ShowRAM = false

// NeedsDockTick reports whether any live dock widget needs periodic refresh.
func NeedsDockTick() bool {
	return ShowClock || ShowCPU || ShowRAM
}

// ScrollbackLines is the per-window scrollback buffer depth.
var ScrollbackLines = 10000

// NiriReverseScroll flips wheel direction in niri scrolling mode.
var NiriReverseScroll = false

// LeaderKey is the prefix key that opens command mode.
var LeaderKey = "ctrl+b"

// ZoomMaxWidth caps zoom/zen width in cells; 0 means fullscreen.
var ZoomMaxWidth = 0

// GetDockPillLeftChar returns the left pill cap for the active glyph set.
func GetDockPillLeftChar() string {
	if UseASCIIOnly {
		return DockPillLeftCharASCII
	}
	return DockPillLeftChar
}

// GetDockPillRightChar returns the right pill cap for the active glyph set.
func GetDockPillRightChar() string {
	if UseASCIIOnly {
		return DockPillRightCharASCII
	}
	return DockPillRightChar
}

// GetDockModeIconWindow returns the window-mode icon for the active glyph set.
func GetDockModeIconWindow() string {
	if UseASCIIOnly {
		return DockModeIconWindowASCII
	}
	return DockModeIconWindow
}

// GetDockModeIconTerminal returns the terminal-mode icon for the active glyph set.
func GetDockModeIconTerminal() string {
	if UseASCIIOnly {
		return DockModeIconTerminalASCII
	}
	return DockModeIconTerminal
}

// GetDockModeIconTiling returns the tiling-mode icon for the active glyph set.
func GetDockModeIconTiling() string {
	if UseASCIIOnly {
		return DockModeIconTilingASCII
	}
	return DockModeIconTiling
}

// GetDockIconTerminalCount returns the terminal-count icon for the active glyph set.
func GetDockIconTerminalCount() string {
	if UseASCIIOnly {
		return DockIconTerminalCountASCII
	}
	return DockIconTerminalCount
}

// GetDockIconWorkspaceCount returns the workspace-count icon for the active glyph set.
func GetDockIconWorkspaceCount() string {
	if UseASCIIOnly {
		return DockIconWorkspaceCountASCII
	}
	return DockIconWorkspaceCount
}

// GetDockSeparator returns the dock section separator for the active glyph set.
func GetDockSeparator() string {
	if UseASCIIOnly {
		return DockSeparatorASCII
	}
	return DockSeparator
}

// Unicode window border glyphs (rounded set).
const (
	// WindowBorderTopLeft is the rounded top-left corner.
	WindowBorderTopLeft = "╭"
	// WindowBorderTopRight is the rounded top-right corner.
	WindowBorderTopRight = "╮"
	// WindowBorderBottomLeft is the rounded bottom-left corner.
	WindowBorderBottomLeft = "╰"
	// WindowBorderBottomRight is the rounded bottom-right corner.
	WindowBorderBottomRight = "╯"
	// WindowBorderHorizontal is the horizontal edge.
	WindowBorderHorizontal = "─"
	// WindowBorderVertical is the vertical edge.
	WindowBorderVertical = "│"
	// WindowButtonClose is the close button glyph.
	WindowButtonClose = " ⤫ "
	// WindowSeparatorChar divides window chrome elements.
	WindowSeparatorChar = "─"
)

// ASCII window border glyphs.
const (
	// WindowBorderTopLeftASCII is the ASCII top-left corner.
	WindowBorderTopLeftASCII = "+"
	// WindowBorderTopRightASCII is the ASCII top-right corner.
	WindowBorderTopRightASCII = "+"
	// WindowBorderBottomLeftASCII is the ASCII bottom-left corner.
	WindowBorderBottomLeftASCII = "+"
	// WindowBorderBottomRightASCII is the ASCII bottom-right corner.
	WindowBorderBottomRightASCII = "+"
	// WindowBorderHorizontalASCII is the ASCII horizontal edge.
	WindowBorderHorizontalASCII = "-"
	// WindowBorderVerticalASCII is the ASCII vertical edge.
	WindowBorderVerticalASCII = "|"
	// WindowButtonCloseASCII is the ASCII close button.
	WindowButtonCloseASCII = " X "
	// WindowPillLeftASCII is the ASCII left pill cap.
	WindowPillLeftASCII = "["
	// WindowPillRightASCII is the ASCII right pill cap.
	WindowPillRightASCII = "]"
	// WindowSeparatorCharASCII divides window chrome elements in ASCII.
	WindowSeparatorCharASCII = "-"
)

// GetBorderForStyle maps the active BorderStyle name to a lipgloss border.
// Falls back to rounded for any unknown name; ASCII mode short-circuits.
func GetBorderForStyle() lipgloss.Border {
	if UseASCIIOnly || BorderStyle == "ascii" {
		return lipgloss.ASCIIBorder()
	}
	switch BorderStyle {
	case "normal":
		return lipgloss.NormalBorder()
	case "thick":
		return lipgloss.ThickBorder()
	case "double":
		return lipgloss.DoubleBorder()
	case "hidden":
		return lipgloss.HiddenBorder()
	case "block":
		return lipgloss.BlockBorder()
	case "outer-half-block":
		return lipgloss.OuterHalfBlockBorder()
	case "inner-half-block":
		return lipgloss.InnerHalfBlockBorder()
	default: // rounded and anything unrecognized
		return lipgloss.RoundedBorder()
	}
}

// GetScrollbarThumbChar returns the scrollbar thumb glyph: a full block for
// visibility, or '#' under ASCII mode.
func GetScrollbarThumbChar() string {
	if UseASCIIOnly || BorderStyle == "ascii" {
		return "#"
	}
	return "█"
}

// GetWindowBorderTopLeft returns the active border's top-left corner.
func GetWindowBorderTopLeft() string { return GetBorderForStyle().TopLeft }

// GetWindowBorderTopRight returns the active border's top-right corner.
func GetWindowBorderTopRight() string { return GetBorderForStyle().TopRight }

// GetWindowBorderBottomLeft returns the active border's bottom-left corner.
func GetWindowBorderBottomLeft() string { return GetBorderForStyle().BottomLeft }

// GetWindowBorderBottomRight returns the active border's bottom-right corner.
func GetWindowBorderBottomRight() string { return GetBorderForStyle().BottomRight }

// GetWindowBorderTop returns the active border's top edge.
func GetWindowBorderTop() string { return GetBorderForStyle().Top }

// GetWindowBorderBottom returns the active border's bottom edge.
func GetWindowBorderBottom() string { return GetBorderForStyle().Bottom }

// GetWindowBorderLeft returns the active border's left edge.
func GetWindowBorderLeft() string { return GetBorderForStyle().Left }

// GetWindowBorderRight returns the active border's right edge.
func GetWindowBorderRight() string { return GetBorderForStyle().Right }

// GetWindowBorderHorizontal returns the top edge.
// Deprecated: use GetWindowBorderTop or GetWindowBorderBottom for half-block borders.
func GetWindowBorderHorizontal() string { return GetWindowBorderTop() }

// GetWindowBorderVertical returns the left edge.
// Deprecated: use GetWindowBorderLeft or GetWindowBorderRight for half-block borders.
func GetWindowBorderVertical() string { return GetWindowBorderLeft() }

// GetWindowButtonClose returns the close button glyph for the active glyph set.
func GetWindowButtonClose() string {
	if UseASCIIOnly {
		return WindowButtonCloseASCII
	}
	return WindowButtonClose
}

// GetWindowPillLeft returns the left pill cap for the active glyph set.
func GetWindowPillLeft() string {
	if UseASCIIOnly {
		return WindowPillLeftASCII
	}
	return WindowPillLeft
}

// GetWindowPillRight returns the right pill cap for the active glyph set.
func GetWindowPillRight() string {
	if UseASCIIOnly {
		return WindowPillRightASCII
	}
	return WindowPillRight
}

// GetWindowSeparatorChar returns the chrome separator for the active glyph set.
func GetWindowSeparatorChar() string {
	if UseASCIIOnly {
		return WindowSeparatorCharASCII
	}
	return WindowSeparatorChar
}

// Title-bar button column offsets, measured from the window's right edge.
const (
	// MinimizeButtonLeftNonTiling is the minimize button's left offset in float mode.
	MinimizeButtonLeftNonTiling = -11
	// MinimizeButtonRightNonTiling is the minimize button's right offset in float mode.
	MinimizeButtonRightNonTiling = -9
	// MaximizeButtonLeft is the maximize button's left offset.
	MaximizeButtonLeft = -8
	// MaximizeButtonRight is the maximize button's right offset.
	MaximizeButtonRight = -6
	// MinimizeButtonLeftTiling is the minimize button's left offset in tiling mode.
	MinimizeButtonLeftTiling = -8
	// MinimizeButtonRightTiling is the minimize button's right offset in tiling mode.
	MinimizeButtonRightTiling = -6
	// CloseButtonLeft is the close button's left offset (mode-independent).
	CloseButtonLeft = -5
	// CloseButtonRight is the close button's right offset (mode-independent).
	CloseButtonRight = -3
)

// Buffer and pool sizing.
const (
	// ByteSliceBufferSize is the pooled byte-slice size (32KB).
	ByteSliceBufferSize = 32 * 1024
	// WindowExitChannelBuffer is the depth of the window exit channel.
	WindowExitChannelBuffer = 10
	// LayerPoolInitialCapacity is the starting cap for pooled layer slices.
	LayerPoolInitialCapacity = 16
	// StringBuilderInitialCapacity is the starting cap for terminal content builders.
	StringBuilderInitialCapacity = 1000
)

// Runtime limits.
const (
	// MaxLogMessages caps in-memory log retention.
	MaxLogMessages = 100
	// MaxWorkspaces is the number of supported workspaces.
	MaxWorkspaces = 9
	// CPUHistorySize is how many CPU samples are retained.
	CPUHistorySize = 10
	// MaxDockItems caps minimized windows shown in the dock.
	MaxDockItems = 9
	// MaxGridColumns caps grid layout columns.
	MaxGridColumns = 3
	// MaxTwoColumnGridWindows is the window count that triggers a 2-column grid.
	MaxTwoColumnGridWindows = 6
	// MaxHelpLines estimates the help overlay's line budget.
	MaxHelpLines = 50
	// MaxSwapDistance bounds directional window swap targeting.
	MaxSwapDistance = 5
)

// Compositor z-index layers, low to high.
const (
	// ZIndexBase is the layer for ordinary windows.
	ZIndexBase = 0
	// ZIndexSeparators draws shared border lines above windows, below overlays.
	ZIndexSeparators = 998
	// ZIndexAnimating lifts a window while it animates.
	ZIndexAnimating = 999
	// ZIndexHelp is the help overlay layer.
	ZIndexHelp = 1000
	// ZIndexDock is the dock layer.
	ZIndexDock = 1000
	// ZIndexTime is the clock layer.
	ZIndexTime = 1001
	// ZIndexLogs is the log overlay layer.
	ZIndexLogs = 1001
	// ZIndexWhichKey is the which-key overlay layer.
	ZIndexWhichKey = 1002
	// ZIndexScrollbackBrowser is the scrollback browser layer.
	ZIndexScrollbackBrowser = 1003
	// ZIndexCommandPalette is the command palette layer.
	ZIndexCommandPalette = 1004
	// ZIndexSessionSwitcher is the session switcher layer.
	ZIndexSessionSwitcher = 1005
	// ZIndexLayoutPicker is the layout picker layer.
	ZIndexLayoutPicker = 1006
	// ZIndexNotifications is the top-most notification layer.
	ZIndexNotifications = 2000
)

// Assorted defaults.
const (
	// DefaultSSHPort is the SSH server's default port.
	DefaultSSHPort = "2222"
	// DefaultSSHHost is the SSH server's default host.
	DefaultSSHHost = "localhost"
	// DefaultTerminalWidth is the fallback width when the real size is unknown.
	DefaultTerminalWidth = 80
	// DefaultTerminalHeight is the fallback height when the real size is unknown.
	DefaultTerminalHeight = 24
	// MinTerminalWidth is the minimum usable terminal width.
	MinTerminalWidth = 1
	// MinTerminalHeight is the minimum usable terminal height.
	MinTerminalHeight = 1
)

// Fractional divisors used in layout math.
const (
	// HalfDivisor halves a dimension.
	HalfDivisor = 2
	// QuarterDivisor quarters a dimension.
	QuarterDivisor = 4
)

// Control characters and printable ASCII bounds.
const (
	// CtrlB is the Ctrl+B control code.
	CtrlB = 0x02
	// DEL is the delete code.
	DEL = 0x7f
	// ESC is the escape code.
	ESC = 0x1b
	// NUL is the null code.
	NUL = 0x00
	// Tab is the tab code.
	Tab = 0x09
	// CarriageReturn is the CR rune.
	CarriageReturn = '\r'
	// LineFeed is the LF rune.
	LineFeed = '\n'
	// Space is the space rune.
	Space = ' '
	// PrintableCharMin is the lowest printable ASCII code.
	PrintableCharMin = 32
	// PrintableCharMax is the highest printable ASCII code.
	PrintableCharMax = 126
	// ASCIICharMax is the highest single-byte ASCII code.
	ASCIICharMax = 127
)

// Border thickness and line caps used when sizing terminals.
const (
	// BorderWidth is the combined left+right border width.
	BorderWidth = 2
	// BorderHeight is the combined top+bottom border height.
	BorderHeight = 2
	// MaxLineLength caps a single display line's length.
	MaxLineLength = 2000
)

// ANSI modifier parameter values for CSI-encoded key modifiers.
const (
	// ModParamBase is the modifier parameter base.
	ModParamBase = 1
	// ModParamShift is the shift modifier parameter.
	ModParamShift = 2
	// ModParamAlt is the alt modifier parameter.
	ModParamAlt = 2
	// ModParamCtrl is the ctrl modifier parameter.
	ModParamCtrl = 4
)

// VT cell attribute bit flags.
const (
	// VTAttrBold marks bold text.
	VTAttrBold = 1
	// VTAttrFaint marks faint text.
	VTAttrFaint = 2
	// VTAttrItalic marks italic text.
	VTAttrItalic = 4
	// VTAttrReverse marks reverse video.
	VTAttrReverse = 32
	// VTAttrStrikethrough marks struck-through text.
	VTAttrStrikethrough = 128
)

// Tiling layout thresholds.
const (
	// TilingModeEnabledWorkspaces is how many workspaces allow tiling.
	TilingModeEnabledWorkspaces = MaxWorkspaces
	// GridLayoutThreshold is the window count that switches to grid layout.
	GridLayoutThreshold = 4
)

// Name formatting offsets.
const (
	// IDPrefixLength is how many UUID chars are shown as an ID prefix.
	IDPrefixLength = 8
	// MaxNameTruncateLength is the length past which a name is truncated.
	MaxNameTruncateLength = 12
	// EllipsisLength is the length of the truncation ellipsis.
	EllipsisLength = 3
	// MaxNameLengthBeforeEllipsis is the kept-name length before the ellipsis.
	MaxNameLengthBeforeEllipsis = MaxNameTruncateLength - EllipsisLength
)
