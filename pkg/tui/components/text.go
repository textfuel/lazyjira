package components

import "github.com/charmbracelet/lipgloss"

// TruncateEnd truncates s to fit within maxWidth display columns,
// appending "…" if needed. Respects multi-byte UTF-8 characters.
func TruncateEnd(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i])
		if lipgloss.Width(candidate)+1 <= maxWidth { // +1 for "…"
			return candidate + "…"
		}
	}
	return "…"
}

// TruncateMiddle truncates keeping start and end visible: "abcdef...xyz"
func TruncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 5 {
		return s[:maxLen]
	}
	side := (maxLen - 3) / 2
	return s[:side+1] + "..." + s[len(s)-side:]
}

// Truncate truncates s to max bytes, appending "…" if needed.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}

// PanelDimensions computes usable content width and inner height from total panel dimensions.
func PanelDimensions(width, height int) (contentWidth, innerHeight int) {
	return max(width-2, 10), max(height-2, 1)
}
