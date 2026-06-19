package theme

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorPalette holds the semantic color values for a theme.
// The default theme uses ANSI 16 codes; Catppuccin themes use hex values.
type ColorPalette struct {
	Green     lipgloss.Color
	Blue      lipgloss.Color
	Red       lipgloss.Color
	Yellow    lipgloss.Color
	Cyan      lipgloss.Color
	Magenta   lipgloss.Color
	White     lipgloss.Color
	Gray      lipgloss.Color
	Orange    lipgloss.Color
	None      lipgloss.Color
	Highlight lipgloss.Color // selection/cursor background
}

// Package-level color variables. These are kept in sync with Default.Colors
// by SetTheme so that existing call sites (theme.ColorBlue, etc.) continue
// to work without changes.
var (
	ColorGreen     = lipgloss.Color("2")   // ANSI green — active borders, accents
	ColorBlue      = lipgloss.Color("4")   // ANSI blue — help bar, selected bg
	ColorRed       = lipgloss.Color("1")   // ANSI red — errors, unstaged
	ColorYellow    = lipgloss.Color("3")   // ANSI yellow — warnings, in-progress
	ColorCyan      = lipgloss.Color("6")   // ANSI cyan — search mode
	ColorMagenta   = lipgloss.Color("5")   // ANSI magenta — JQL keywords
	ColorWhite     = lipgloss.Color("7")   // ANSI white (light gray)
	ColorGray      = lipgloss.Color("8")   // ANSI bright black (dark gray)
	ColorOrange    = lipgloss.Color("208") // ANSI 256 orange — secondary accent (names, metadata)
	ColorNone      = lipgloss.Color("-1")  // default terminal color
	ColorHighlight = lipgloss.Color("4")   // selection/cursor background (same as blue by default)
)

type Theme struct {
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	HintBar        lipgloss.Style
	SelectedItem   lipgloss.Style
	NormalItem     lipgloss.Style
	ActiveBorder   lipgloss.Style
	InactiveBorder lipgloss.Style
	ErrorText      lipgloss.Style
	SuccessText    lipgloss.Style
	WarningText    lipgloss.Style
	KeyStyle       lipgloss.Style
	ValueStyle     lipgloss.Style
	PriorityHigh   lipgloss.Style
	PriorityMedium lipgloss.Style
	PriorityLow    lipgloss.Style

	Colors        ColorPalette
	AuthorPalette []lipgloss.Color
}

// Default is the singleton theme instance
var Default = defaultTheme()

// DefaultTheme returns the singleton theme. Kept for compatibility
func DefaultTheme() *Theme { return Default }

// defaultPalette returns the ANSI 16 color palette used by the default theme.
func defaultPalette() ColorPalette {
	return ColorPalette{
		Green:     lipgloss.Color("2"),
		Blue:      lipgloss.Color("4"),
		Red:       lipgloss.Color("1"),
		Yellow:    lipgloss.Color("3"),
		Cyan:      lipgloss.Color("6"),
		Magenta:   lipgloss.Color("5"),
		White:     lipgloss.Color("7"),
		Gray:      lipgloss.Color("8"),
		Orange:    lipgloss.Color("208"),
		None:      lipgloss.Color("-1"),
		Highlight: lipgloss.Color("4"), // same as blue for default theme
	}
}

// defaultAuthorPalette returns the ANSI 256 author colors for the default theme.
func defaultAuthorPalette() []lipgloss.Color {
	return []lipgloss.Color{
		lipgloss.Color("208"), // orange
		lipgloss.Color("176"), // pink/magenta
		lipgloss.Color("114"), // light green
		lipgloss.Color("216"), // salmon
		lipgloss.Color("81"),  // sky blue
		lipgloss.Color("222"), // gold
		lipgloss.Color("183"), // lavender
		lipgloss.Color("150"), // sage
		lipgloss.Color("209"), // coral
		lipgloss.Color("117"), // light cyan
		lipgloss.Color("180"), // tan
		lipgloss.Color("147"), // periwinkle
	}
}

func defaultTheme() *Theme {
	return buildTheme(defaultPalette(), defaultAuthorPalette())
}

// buildTheme constructs a Theme from a color palette and author palette.
func buildTheme(p ColorPalette, authors []lipgloss.Color) *Theme {
	return &Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Green),

		Subtitle: lipgloss.NewStyle().
			Foreground(p.Gray),

		HintBar: lipgloss.NewStyle().
			Foreground(p.Gray),

		SelectedItem: lipgloss.NewStyle().
			Bold(true).
			Background(p.Highlight),

		NormalItem: lipgloss.NewStyle(),

		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Green),

		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.None),

		ErrorText: lipgloss.NewStyle().
			Foreground(p.Red).
			Bold(true),

		SuccessText: lipgloss.NewStyle().
			Foreground(p.Green),

		WarningText: lipgloss.NewStyle().
			Foreground(p.Yellow),

		KeyStyle: lipgloss.NewStyle().
			Foreground(p.Green),

		ValueStyle: lipgloss.NewStyle(),

		PriorityHigh: lipgloss.NewStyle().
			Foreground(p.Red),

		PriorityMedium: lipgloss.NewStyle().
			Foreground(p.Yellow),

		PriorityLow: lipgloss.NewStyle().
			Foreground(p.Green),

		Colors:        p,
		AuthorPalette: authors,
	}
}

// syncColors updates the package-level color variables and the author palette
// to match the current Default theme.
func syncColors() {
	ColorGreen = Default.Colors.Green
	ColorBlue = Default.Colors.Blue
	ColorRed = Default.Colors.Red
	ColorYellow = Default.Colors.Yellow
	ColorCyan = Default.Colors.Cyan
	ColorMagenta = Default.Colors.Magenta
	ColorWhite = Default.Colors.White
	ColorGray = Default.Colors.Gray
	ColorOrange = Default.Colors.Orange
	ColorNone = Default.Colors.None
	ColorHighlight = Default.Colors.Highlight

	authorMutex.Lock()
	authorPalette = Default.AuthorPalette
	authorCache = make(map[string]lipgloss.Style)
	authorMutex.Unlock()
}

// Options configures the theme system. Values map 1:1 to the GUIConfig
// fields in pkg/config; the indirection keeps the theme package free of
// any config-package dependency.
//
// Precedence (low to high):
//  1. The selected preset's palette.
//  2. Colors (shared overrides applied to every preset).
//  3. ColorsDark or ColorsLight, whichever matches the preset's IsLight flag.
//
// Override map keys are lowercase palette field names: "green", "blue",
// "red", "yellow", "cyan", "magenta", "white", "gray", "orange", "highlight".
// Unknown keys and empty values are silently ignored so configs stay
// forward-compatible.
type Options struct {
	Preset      string
	Colors      map[string]string
	ColorsDark  map[string]string
	ColorsLight map[string]string
}

// Init selects a preset, applies any user overrides, and refreshes the
// global Default theme plus the package-level color variables. Must be
// called before the TUI starts.
//
// Preset resolution:
//   - "" (unset): the bundled "default" ANSI 16 preset, matching the
//     pre-themeing behavior.
//   - "auto": chosen at runtime from the terminal background
//     (DefaultDarkPresetName or DefaultLightPresetName).
//   - any other value: looked up case-insensitively; unknown names
//     return an error.
func Init(opts Options) error {
	var preset *Preset
	switch strings.ToLower(strings.TrimSpace(opts.Preset)) {
	case "":
		preset = FindPreset("default")
	case "auto":
		preset = autoDetectPreset()
	default:
		preset = FindPreset(opts.Preset)
		if preset == nil {
			return fmt.Errorf("unknown theme: %q", opts.Preset)
		}
	}
	if preset == nil {
		// "default" preset was stripped from presetList; fall back to
		// whatever ships first so the binary keeps rendering.
		preset = &presetList[0]
	}

	built := preset.Build()
	palette := applyOverrides(built.Colors, opts.Colors, "themeColors")
	if preset.IsLight {
		palette = applyOverrides(palette, opts.ColorsLight, "themeLight")
	} else {
		palette = applyOverrides(palette, opts.ColorsDark, "themeDark")
	}

	Default = buildTheme(palette, built.AuthorPalette)
	syncColors()
	return nil
}

// SetTheme is a thin wrapper around Init kept for callers (and tests) that
// only care about preset selection without overrides.
func SetTheme(name string) error {
	return Init(Options{Preset: name})
}

// autoDetectPreset returns the preset chosen when GUI.Theme is set to
// "auto". Falls back to the dark default if neither preset is registered,
// which should be impossible but keeps the binary working if someone
// strips presets.go.
func autoDetectPreset() *Preset {
	if lipgloss.HasDarkBackground() {
		if p := FindPreset(DefaultDarkPresetName); p != nil {
			return p
		}
	} else {
		if p := FindPreset(DefaultLightPresetName); p != nil {
			return p
		}
	}
	if p := FindPreset(DefaultDarkPresetName); p != nil {
		return p
	}
	return &presetList[0]
}

// ValidColor reports whether val is a color string lipgloss/termenv will
// render correctly. Accepted forms:
//   - hex: "#rgb", "#rrggbb", "#rrggbbaa" (case-insensitive)
//   - ANSI decimal: "0".."255"
//   - terminal default sentinel: "-1"
//
// Empty strings are rejected here; callers (e.g. applyOverrides) skip
// empty values before calling ValidColor so users can use "" to mean
// "leave the preset alone".
//
// Exported so other packages (e.g. pkg/tui/views/adf.go) can guard
// dynamic color strings from untrusted sources (Jira ADF marks, etc.)
// against malformed values.
func ValidColor(val string) bool {
	if val == "" {
		return false
	}
	if strings.HasPrefix(val, "#") {
		hex := val[1:]
		switch len(hex) {
		case 3, 6, 8:
		default:
			return false
		}
		for i := range len(hex) {
			c := hex[i]
			switch {
			case c >= '0' && c <= '9':
			case c >= 'a' && c <= 'f':
			case c >= 'A' && c <= 'F':
			default:
				return false
			}
		}
		return true
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return false
	}
	return n == -1 || (n >= 0 && n <= 255)
}

// applyOverrides patches a ColorPalette with any non-empty values from the
// provided map. Unknown keys are silently ignored for forward-compat. Values
// that are not a recognized color format fall back to the terminal default
// color and emit a slog.Warn.
//
// scope identifies the source map ("themeColors", "themeDark",
// "themeLight") so log lines can pinpoint the bad entry.
func applyOverrides(p ColorPalette, m map[string]string, scope string) ColorPalette {
	for key, val := range m {
		if val == "" {
			continue
		}
		c := lipgloss.Color(val)
		if !ValidColor(val) {
			slog.Warn("ignoring invalid theme color; falling back to terminal default",
				"scope", scope, "key", key, "value", val)
			c = lipgloss.Color("-1")
		}
		switch strings.ToLower(key) {
		case "green":
			p.Green = c
		case "blue":
			p.Blue = c
		case "red":
			p.Red = c
		case "yellow":
			p.Yellow = c
		case "cyan":
			p.Cyan = c
		case "magenta":
			p.Magenta = c
		case "white":
			p.White = c
		case "gray":
			p.Gray = c
		case "orange":
			p.Orange = c
		case "highlight":
			p.Highlight = c
		}
	}
	return p
}

// PriorityStyled applies priority color based on name
func PriorityStyled(name string) string {
	switch strings.ToLower(name) {
	case "highest", "high", "critical", "blocker":
		return Default.PriorityHigh.Render(name)
	case "medium":
		return Default.PriorityMedium.Render(name)
	default:
		return Default.PriorityLow.Render(name)
	}
}

func StatusColor(categoryKey string) lipgloss.Style {
	switch categoryKey {
	case "done":
		return lipgloss.NewStyle().Foreground(ColorGreen)
	case "indeterminate":
		return lipgloss.NewStyle().Foreground(ColorYellow)
	case "new":
		return lipgloss.NewStyle().Foreground(ColorBlue)
	default:
		return lipgloss.NewStyle().Foreground(ColorGray)
	}
}
