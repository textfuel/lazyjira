package theme

import "github.com/charmbracelet/lipgloss"

// ResolveColor turns a YAML-or-config color string into a lipgloss.Color.
// If value matches a palette key it returns that color; otherwise it is
// passed straight through to lipgloss (which accepts hex, ANSI numeric, and
// named colors). An empty string returns the zero-value Color, which callers
// should treat as "unset".
func ResolveColor(value string, palette map[string]lipgloss.Color) lipgloss.Color {
	if value == "" {
		return lipgloss.Color("")
	}
	if palette != nil {
		if c, ok := palette[value]; ok {
			return c
		}
	}
	return lipgloss.Color(value)
}
