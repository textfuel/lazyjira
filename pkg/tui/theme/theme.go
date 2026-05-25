package theme

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Semantic accent slots.
	Accent   lipgloss.Style // active borders/tabs/markers, key labels
	Muted    lipgloss.Style // separators, placeholders, inactive tab text
	IssueKey lipgloss.Style // default colour for an issue key

	// SelectedForeground, when non-empty, is applied to selected rows
	// rendered with plain (uncolored) cells.
	SelectedForeground lipgloss.Color

	// Per-value style maps. Lookup helpers fall back to TypeFallback or
	// the relevant style slot when a name is missing.
	TypeColors     map[string]lipgloss.Style
	TypeFallback   lipgloss.Style
	PriorityColors map[string]lipgloss.Style
	StatusColors   map[string]lipgloss.Style // keyed on status category key

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

func fg(c lipgloss.Color) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(c)
}

// buildTheme constructs a Theme from a color palette and author palette.
// Used by builtin themes; YAML-loaded themes populate the struct directly.
func buildTheme(p ColorPalette, authors []lipgloss.Color) *Theme {
	return &Theme{
		Title:          lipgloss.NewStyle().Bold(true).Foreground(p.Green),
		Subtitle:       fg(p.Gray),
		HintBar:        fg(p.Gray),
		SelectedItem:   lipgloss.NewStyle().Bold(true).Background(p.Highlight),
		NormalItem:     lipgloss.NewStyle(),
		ActiveBorder:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Green),
		InactiveBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.None),
		ErrorText:      lipgloss.NewStyle().Foreground(p.Red).Bold(true),
		SuccessText:    fg(p.Green),
		WarningText:    fg(p.Yellow),
		KeyStyle:       fg(p.Green),
		ValueStyle:     lipgloss.NewStyle(),

		Accent:   fg(p.Green),
		Muted:    fg(p.Gray),
		IssueKey: fg(p.Cyan),

		TypeColors: map[string]lipgloss.Style{
			"Bug":         fg(p.Red),
			"Story":       fg(p.Green),
			"Epic":        fg(p.Magenta),
			"Task":        fg(p.Blue),
			"Sub-task":    fg(p.Gray),
			"Improvement": fg(p.Cyan),
			"New Feature": fg(p.Orange),
		},
		TypeFallback: fg(p.Gray),

		PriorityColors: map[string]lipgloss.Style{
			"Highest": fg(p.Red),
			"High":    fg(p.Orange),
			"Medium":  fg(p.Yellow),
			"Low":     fg(p.Green),
			"Lowest":  fg(p.Gray),
		},

		StatusColors: map[string]lipgloss.Style{
			"done":          fg(p.Green),
			"indeterminate": fg(p.Yellow),
			"new":           fg(p.Blue),
		},

		Colors:        p,
		AuthorPalette: authors,
	}
}

// TypeStyle returns the style for an issue type name, falling back to
// TypeFallback when the name is not enumerated.
func (t *Theme) TypeStyle(name string) lipgloss.Style {
	if name != "" {
		if s, ok := t.TypeColors[name]; ok {
			return s
		}
	}
	return t.TypeFallback
}

// PriorityStyle returns the style for a priority name. Matching is
// case-insensitive against the configured keys; falls back to the Muted
// style when no entry is present.
func (t *Theme) PriorityStyle(name string) lipgloss.Style {
	if name == "" {
		return t.Muted
	}
	if s, ok := t.PriorityColors[name]; ok {
		return s
	}
	low := strings.ToLower(name)
	for k, v := range t.PriorityColors {
		if strings.ToLower(k) == low {
			return v
		}
	}
	// Legacy aliases used by existing call sites.
	switch low {
	case "critical", "blocker":
		if s, ok := t.PriorityColors["Highest"]; ok {
			return s
		}
	}
	return t.Muted
}

// StatusStyle returns the style for a Jira status category key
// ("done", "indeterminate", "new"). Falls back to Muted.
func (t *Theme) StatusStyle(categoryKey string) lipgloss.Style {
	if s, ok := t.StatusColors[categoryKey]; ok {
		return s
	}
	return t.Muted
}

// ActiveBorderColor returns the foreground color of the active border.
func (t *Theme) ActiveBorderColor() lipgloss.Color {
	if c, ok := t.ActiveBorder.GetBorderTopForeground().(lipgloss.Color); ok {
		return c
	}
	return t.Colors.Green
}

// InactiveBorderColor returns the foreground color of the inactive border.
func (t *Theme) InactiveBorderColor() lipgloss.Color {
	if c, ok := t.InactiveBorder.GetBorderTopForeground().(lipgloss.Color); ok {
		return c
	}
	return t.Colors.None
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

	authorPalette = Default.AuthorPalette
	authorCache = make(map[string]lipgloss.Style)
}

// themesDir returns the directory where YAML themes are stored.
// It is overridable by tests via the LAZYJIRA_THEMES_DIR env var.
func themesDir() string {
	if dir := os.Getenv("LAZYJIRA_THEMES_DIR"); dir != "" {
		return dir
	}
	if dir := os.Getenv("CONFIG_DIR"); dir != "" {
		return filepath.Join(dir, "themes")
	}
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "lazyjira", "themes")
	}
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "lazyjira", "themes")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "lazyjira", "themes")
}

// SetTheme selects a theme by name and updates the global Default instance
// along with all package-level color variables. Must be called before the
// TUI starts.
//
// Builtins: "default", "catppuccin-latte", "catppuccin-frappe",
// "catppuccin-macchiato", "catppuccin-mocha". Any other name is looked up
// as <themes-dir>/<name>.yml.
func SetTheme(name string) error {
	switch name {
	case "", "default":
		Default = defaultTheme()
	case "catppuccin-latte":
		Default = catppuccinLatte()
	case "catppuccin-frappe":
		Default = catppuccinFrappe()
	case "catppuccin-macchiato":
		Default = catppuccinMacchiato()
	case "catppuccin-mocha":
		Default = catppuccinMocha()
	default:
		path := filepath.Join(themesDir(), name+".yml")
		th, err := loadYAMLTheme(path)
		if err != nil {
			return fmt.Errorf("theme %q: %w", name, err)
		}
		Default = th
	}
	syncColors()
	return nil
}

// PriorityStyled applies priority color based on name
func PriorityStyled(name string) string {
	return Default.PriorityStyle(name).Render(name)
}

// StatusColor returns a style for a Jira status category key. Retained for
// call sites that have not yet moved to Theme.StatusStyle.
func StatusColor(categoryKey string) lipgloss.Style {
	return Default.StatusStyle(categoryKey)
}
