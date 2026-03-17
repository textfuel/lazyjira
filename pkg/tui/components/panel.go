package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/pkg/tui/theme"
)

// RenderPanel draws a bordered panel with a title embedded in the top border.
// Like lazygit: ─[2] Issues────────────────
func RenderPanel(title, content string, width, innerHeight int, focused bool) string {
	th := theme.DefaultTheme()

	borderColor := theme.ColorNone
	if focused {
		borderColor = theme.ColorGreen
	}

	// Title styling.
	var styledTitle string
	if focused {
		styledTitle = th.Title.Render(title)
	} else {
		styledTitle = lipgloss.NewStyle().Foreground(borderColor).Render(title)
	}

	contentWidth := width - 2 // left + right border chars
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Build top border: ╭Title──────────╮
	// Total width = contentWidth + 2 (╭ and ╮).
	titleLen := lipgloss.Width(styledTitle) // rendered width for padding calc
	topPadding := contentWidth - titleLen
	if topPadding < 0 {
		topPadding = 0
	}
	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	topLine := borderStyle.Render("╭") +
		styledTitle +
		borderStyle.Render(strings.Repeat("─", topPadding) + "╮")

	// Content lines — pad/truncate to fit.
	lines := strings.Split(content, "\n")
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	borderVert := lipgloss.NewStyle().Foreground(borderColor).Render("│")
	var body strings.Builder
	for _, line := range lines {
		// Pad line to contentWidth using spaces.
		rendered := line
		lineWidth := lipgloss.Width(rendered)
		if lineWidth < contentWidth {
			rendered += strings.Repeat(" ", contentWidth-lineWidth)
		}
		body.WriteString(borderVert + rendered + borderVert + "\n")
	}

	// Bottom border: ╰──────────────╯
	bottomLine := lipgloss.NewStyle().Foreground(borderColor).Render(
		"╰" + strings.Repeat("─", contentWidth) + "╯")

	return topLine + "\n" + body.String() + bottomLine
}
