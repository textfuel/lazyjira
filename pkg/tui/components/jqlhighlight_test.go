package components

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/tui/theme"
)

func segmentTexts(segments []StyledSegment) []string {
	texts := make([]string, len(segments))
	for i, segment := range segments {
		texts[i] = segment.Text
	}
	return texts
}

func TestHighlightJQL_Segments(t *testing.T) {
	t.Parallel()

	segments := HighlightJQL([]rune(`project = "Foo Bar"`))

	want := []string{"project", " ", "=", " ", `"Foo Bar"`}
	got := segmentTexts(segments)
	if len(got) != len(want) {
		t.Fatalf("segments = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("segment %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func assertForeground(t *testing.T, label string, style lipgloss.Style, want lipgloss.TerminalColor) {
	t.Helper()
	if got := style.GetForeground(); got != want {
		t.Errorf("%s foreground = %v, want %v", label, got, want)
	}
}

func TestClassifyJQLWord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		word     string
		fullText string
		pos      int
		want     lipgloss.TerminalColor
	}{
		{"known field", "project", "project = x", 0, theme.ColorBlue},
		{"customfield prefix", "customfield_10001", "customfield_10001 = x", 0, theme.ColorBlue},
		{"keyword", "and", "a and b", 2, theme.ColorMagenta},
		{"word operator", "in", "status in (x)", 7, theme.ColorGreen},
		{"unknown word before operator is a field", "myfield", "myfield = x", 0, theme.ColorBlue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			style := classifyJQLWord(tt.word, tt.fullText, tt.pos)
			assertForeground(t, tt.name, style, tt.want)
		})
	}
}
