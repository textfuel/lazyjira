package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestResolveColorPaletteHit(t *testing.T) {
	pal := map[string]lipgloss.Color{
		"green": lipgloss.Color("#a6e3a1"),
	}
	got := ResolveColor("green", pal)
	if got != lipgloss.Color("#a6e3a1") {
		t.Errorf("ResolveColor(green) = %q, want #a6e3a1", got)
	}
}

func TestResolveColorHexPassthrough(t *testing.T) {
	got := ResolveColor("#123456", nil)
	if got != lipgloss.Color("#123456") {
		t.Errorf("ResolveColor(#123456) = %q, want #123456", got)
	}
}

func TestResolveColorANSIPassthrough(t *testing.T) {
	got := ResolveColor("208", nil)
	if got != lipgloss.Color("208") {
		t.Errorf("ResolveColor(208) = %q, want 208", got)
	}
}

func TestResolveColorEmpty(t *testing.T) {
	got := ResolveColor("", map[string]lipgloss.Color{"a": lipgloss.Color("#fff")})
	if got != lipgloss.Color("") {
		t.Errorf("ResolveColor(empty) = %q, want empty", got)
	}
}

func TestResolveColorUnknownNameFallsThrough(t *testing.T) {
	pal := map[string]lipgloss.Color{"green": lipgloss.Color("#0f0")}
	got := ResolveColor("magenta", pal)
	// Not in palette — passed to lipgloss as a named color literal.
	if got != lipgloss.Color("magenta") {
		t.Errorf("ResolveColor(magenta) = %q, want magenta", got)
	}
}
