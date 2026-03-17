package theme

import "github.com/charmbracelet/lipgloss"

// Lazygit-style colors: standard ANSI 16 palette.
const (
	ColorGreen  = lipgloss.Color("2")  // ANSI green — active borders, accents
	ColorBlue   = lipgloss.Color("4")  // ANSI blue — help bar, selected bg
	ColorRed    = lipgloss.Color("1")  // ANSI red — errors, unstaged
	ColorYellow = lipgloss.Color("3")  // ANSI yellow — warnings, in-progress
	ColorCyan   = lipgloss.Color("6")  // ANSI cyan — search mode
	ColorWhite  = lipgloss.Color("7")  // ANSI white (light gray)
	ColorGray   = lipgloss.Color("8")  // ANSI bright black (dark gray)
	ColorNone   = lipgloss.Color("-1") // default terminal color
)

type Theme struct {
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	StatusBar      lipgloss.Style
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
}

// DefaultTheme returns a lazygit-matching theme using ANSI 16 colors.
func DefaultTheme() *Theme {
	return &Theme{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGreen),

		Subtitle: lipgloss.NewStyle().
			Foreground(ColorGray),

		StatusBar: lipgloss.NewStyle().
			Foreground(ColorWhite),

		HintBar: lipgloss.NewStyle().
			Foreground(ColorGray),

		SelectedItem: lipgloss.NewStyle().
			Bold(true).
			Background(ColorBlue),

		NormalItem: lipgloss.NewStyle(),

		ActiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorGreen),

		InactiveBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorNone),

		ErrorText: lipgloss.NewStyle().
			Foreground(ColorRed).
			Bold(true),

		SuccessText: lipgloss.NewStyle().
			Foreground(ColorGreen),

		WarningText: lipgloss.NewStyle().
			Foreground(ColorYellow),

		KeyStyle: lipgloss.NewStyle().
			Foreground(ColorGreen),

		ValueStyle: lipgloss.NewStyle(),

		PriorityHigh: lipgloss.NewStyle().
			Foreground(ColorRed),

		PriorityMedium: lipgloss.NewStyle().
			Foreground(ColorYellow),

		PriorityLow: lipgloss.NewStyle().
			Foreground(ColorGreen),
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
