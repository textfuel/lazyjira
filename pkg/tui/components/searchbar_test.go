package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestSearchBar_InactiveViewIsEmpty(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	testkit.AssertEqual(t, "inactive view", bar.View(), "")
}

func TestSearchBar_ActivateShowsView(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.Activate()
	testkit.AssertEqual(t, "active after activate", bar.IsActive(), true)
	out := bar.View()
	if !strings.Contains(stripANSI(out), "/") {
		t.Errorf("expected '/' prefix in active view, got %q", stripANSI(out))
	}
}

func TestSearchBar_DeactivateClearsQuery(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.Activate()
	bar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	bar.Deactivate()
	testkit.AssertEqual(t, "inactive after deactivate", bar.IsActive(), false)
	testkit.AssertEqual(t, "query cleared", bar.Query(), "")
}

func TestSearchBar_EnterConfirmsSearch(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.Activate()
	bar, _ = bar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	bar, _ = bar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("i")})
	bar, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyEnter})
	testkit.AssertEqual(t, "inactive after enter", bar.IsActive(), false)
	if cmd == nil {
		t.Fatal("expected command from Enter")
	}
	msg := cmd()
	confirmed, ok := msg.(SearchConfirmedMsg)
	if !ok {
		t.Fatalf("expected SearchConfirmedMsg, got %T", msg)
	}
	testkit.AssertEqual(t, "confirmed query", confirmed.Query, "hi")
}

func TestSearchBar_EscCancelsSearch(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.Activate()
	bar, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "inactive after esc", bar.IsActive(), false)
	if cmd == nil {
		t.Fatal("expected command from Esc")
	}
	msg := cmd()
	if _, ok := msg.(SearchCancelledMsg); !ok {
		t.Fatalf("expected SearchCancelledMsg, got %T", msg)
	}
}

func TestSearchBar_TypeChangesQuery(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.Activate()
	bar, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd == nil {
		t.Fatal("expected SearchChangedMsg command after typing")
	}
	msg := cmd()
	changed, ok := msg.(SearchChangedMsg)
	if !ok {
		t.Fatalf("expected SearchChangedMsg, got %T", msg)
	}
	testkit.AssertEqual(t, "changed query", changed.Query, "x")
	testkit.AssertEqual(t, "bar query", bar.Query(), "x")
}

func TestSearchBar_UpdateInactiveIsNoop(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar2, cmd := bar.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	testkit.AssertEqual(t, "still inactive", bar2.IsActive(), false)
	testkit.AssertEqual(t, "no cmd", cmd == nil, true)
}

func TestSearchBar_SetWidth(t *testing.T) {
	t.Parallel()
	bar := NewSearchBar()
	bar.SetWidth(80)
	bar.Activate()
	out := bar.View()
	if out == "" {
		t.Error("expected non-empty view after SetWidth and Activate")
	}
}

func TestRenderFilterBarInput_ContainsSlash(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetValue("test")
	out := RenderFilterBarInput(&input)
	if !strings.Contains(stripANSI(out), "/") {
		t.Errorf("expected '/' in filter bar, got %q", stripANSI(out))
	}
}
