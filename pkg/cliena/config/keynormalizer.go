package config

import (
	"os"
	"runtime"
	"strings"
)

// KeyNormalizer canonicalizes user key strings and expands platform variants.
// On macOS the Option key emits unicode glyphs (e.g. opt+1 → ¡) and users tend
// to write "opt+" where the runtime reports "alt+", so a single binding must
// resolve to several accepted forms.
type KeyNormalizer struct {
	macOS bool
}

// NewKeyNormalizer builds a normalizer with platform detection baked in.
func NewKeyNormalizer() *KeyNormalizer {
	return &KeyNormalizer{macOS: detectMacOS()}
}

// IsMacOS reports whether key expansion targets macOS conventions.
func (kn *KeyNormalizer) IsMacOS() bool { return kn.macOS }

// detectMacOS trusts GOOS, then falls back to env hints for cross-compiled cases.
func detectMacOS() bool {
	if runtime.GOOS == "darwin" {
		return true
	}
	env := strings.ToLower(os.Getenv("GOOS") + " " + os.Getenv("OSTYPE"))
	return strings.Contains(env, "darwin")
}

// optAltGlyph holds the macOS expansion targets for one Option chord: the
// literal glyph the terminal emits plus the alt+ form the runtime may report.
type optAltGlyph struct {
	glyph string
	alt   string
}

// macOptionExpansions maps each lowercased opt/option chord to its glyph + alt
// equivalents. both "opt+" and "option+" spellings point at the same entry.
var macOptionExpansions = buildMacOptionExpansions()

func buildMacOptionExpansions() map[string]optAltGlyph {
	// number row: opt+N glyphs and opt+shift+N glyphs, index 0 unused.
	optNums := []string{"", "¡", "™", "£", "¢", "∞", "§", "¶", "•", "ª"}
	optShiftNums := []string{"", "⁄", "€", "‹", "›", "ﬁ", "ﬂ", "‡", "°", "·"}

	m := make(map[string]optAltGlyph, 40)
	put := func(chord, glyph, alt string) {
		m[chord] = optAltGlyph{glyph: glyph, alt: alt}
		// accept the long "option+" spelling too
		m[strings.Replace(chord, "opt+", "option+", 1)] = optAltGlyph{glyph: glyph, alt: alt}
	}

	for n := 1; n <= 9; n++ {
		d := string(rune('0' + n))
		put("opt+"+d, optNums[n], "alt+"+d)
		put("opt+shift+"+d, optShiftNums[n], "alt+shift+"+d)
	}
	put("opt+tab", "⇥", "alt+tab")
	put("opt+shift+tab", "⇤", "alt+shift+tab")
	return m
}

// singleLetter reports whether s is one ASCII letter, where case is significant
// (M and m are distinct keys to the input layer).
func singleLetter(s string) bool {
	if len(s) != 1 {
		return false
	}
	c := s[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// canonical returns the case-correct base form: single letters keep their case,
// everything else lowercases so chords compare uniformly.
func canonical(key string) string {
	if singleLetter(key) {
		return key
	}
	return strings.ToLower(key)
}

// NormalizeKey expands one key string into every accepted runtime form.
// The base canonical form always leads; macOS Option chords add their glyph and
// alt+ variants, and any alt+ chord gains opt+/option+ aliases.
func (kn *KeyNormalizer) NormalizeKey(key string) []string {
	key = strings.TrimSpace(key)
	lower := strings.ToLower(key)

	out := make([]string, 0, 4)
	add := uniqueAppender(&out)
	add(canonical(key))

	if kn.macOS {
		if exp, ok := macOptionExpansions[lower]; ok {
			add(strings.ToLower(exp.glyph))
			add(exp.alt)
		}
		// an alt+ binding should also answer to opt+/option+ on macOS
		if strings.HasPrefix(lower, "alt+") {
			add("opt+" + strings.TrimPrefix(lower, "alt+"))
			add("option+" + strings.TrimPrefix(lower, "alt+"))
		}
	}

	return out
}

// ExpandKeys normalizes a list of keys and flattens the results, de-duped.
func (kn *KeyNormalizer) ExpandKeys(keys []string) []string {
	out := make([]string, 0, len(keys)*2)
	add := uniqueAppender(&out)
	for _, key := range keys {
		for _, v := range kn.NormalizeKey(key) {
			add(v)
		}
	}
	return out
}

// uniqueAppender returns an append helper that drops values already in *dst.
func uniqueAppender(dst *[]string) func(string) {
	seen := make(map[string]struct{}, cap(*dst))
	for _, v := range *dst {
		seen[v] = struct{}{}
	}
	return func(v string) {
		if _, dup := seen[v]; dup {
			return
		}
		seen[v] = struct{}{}
		*dst = append(*dst, v)
	}
}

// validSpecialKeys is the set of recognized multi-char key names.
var validSpecialKeys = map[string]struct{}{
	"enter": {}, "return": {}, "esc": {}, "escape": {},
	"tab": {}, "space": {}, "backspace": {}, "delete": {},
	"up": {}, "down": {}, "left": {}, "right": {},
	"home": {}, "end": {}, "pgup": {}, "pageup": {},
	"pgdown": {}, "pagedown": {},
	"f1": {}, "f2": {}, "f3": {}, "f4": {}, "f5": {}, "f6": {},
	"f7": {}, "f8": {}, "f9": {}, "f10": {}, "f11": {}, "f12": {},
}

// ValidateKey checks a key string for the current platform, returning false and
// a reason when it is malformed or uses modifiers the terminal can't deliver.
func (kn *KeyNormalizer) ValidateKey(key string) (bool, string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return false, "key cannot be empty"
	}
	lower := strings.ToLower(key)

	// opt/option only mean anything on macOS
	if !kn.macOS && (strings.Contains(lower, "opt+") || strings.Contains(lower, "option+")) {
		return false, "opt/option keys are only valid on macOS, use alt+ instead"
	}

	parts := strings.Split(lower, "+")
	base := parts[len(parts)-1]
	if base == "" {
		return false, "key combination incomplete (ends with +)"
	}

	if len(parts) > 1 {
		if reason, ok := kn.checkModifiers(parts[:len(parts)-1]); !ok {
			return false, reason
		}
	}

	// single-char base keys (letters, digits, symbols) are always fine
	if len(base) == 1 {
		return true, ""
	}
	if _, ok := validSpecialKeys[base]; !ok {
		return false, "unknown special key: " + base
	}
	return true, ""
}

// checkModifiers validates the modifier segment of a chord: known names only,
// no duplicates, opt/option gated to macOS.
func (kn *KeyNormalizer) checkModifiers(mods []string) (string, bool) {
	valid := map[string]bool{
		"ctrl": true, "alt": true, "shift": true,
		"opt": kn.macOS, "option": kn.macOS,
	}
	seen := make(map[string]struct{}, len(mods))
	for _, mod := range mods {
		if !valid[mod] {
			if mod == "opt" || mod == "option" {
				return "opt/option modifiers are only valid on macOS", false
			}
			return "invalid modifier: " + mod, false
		}
		if _, dup := seen[mod]; dup {
			return "duplicate modifier: " + mod, false
		}
		seen[mod] = struct{}{}
	}
	return "", true
}
