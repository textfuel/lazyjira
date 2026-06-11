package views

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/tui/theme"
)

func TestNoneStyle_GrayForeground(t *testing.T) {
	t.Parallel()
	var wantGray lipgloss.TerminalColor = theme.ColorGray
	testkit.AssertEqual(t, "foreground", noneStyle().GetForeground(), wantGray)
}
