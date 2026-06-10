package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestHelpBar_ViewRendersItems(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar([]HelpItem{
		{Key: "q", Description: "quit"},
		{Key: "?", Description: "help"},
	})
	bar.SetWidth(120)
	out := bar.View()
	plain := stripANSI(out)
	if !strings.Contains(plain, "quit: q") {
		t.Errorf("expected 'quit: q' in output, got %q", plain)
	}
	if !strings.Contains(plain, "help: ?") {
		t.Errorf("expected 'help: ?' in output, got %q", plain)
	}
}

func TestHelpBar_SetItemsReplacesItems(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar([]HelpItem{{Key: "q", Description: "quit"}})
	bar.SetWidth(120)
	bar.SetItems([]HelpItem{{Key: "r", Description: "refresh"}})
	out := stripANSI(bar.View())
	if strings.Contains(out, "quit") {
		t.Error("old item should be replaced")
	}
	if !strings.Contains(out, "refresh: r") {
		t.Errorf("expected new item in output, got %q", out)
	}
}

func TestHelpBar_SetStatusMsgAppearsInView(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	bar.SetWidth(120)
	bar.SetStatusMsg("saved!")
	out := stripANSI(bar.View())
	if !strings.Contains(out, "saved!") {
		t.Errorf("expected status msg in output, got %q", out)
	}
}

func TestHelpBar_SetWidthUpdatesWidth(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	bar.SetWidth(80)
	testkit.AssertEqual(t, "width", bar.width, 80)
}

func TestHelpBar_UpdateHandlesWindowSize(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	updated, _ := bar.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	testkit.AssertEqual(t, "width after window size", updated.width, 100)
}

func TestHelpBar_TruncatesWhenTooNarrow(t *testing.T) {
	t.Parallel()
	items := make([]HelpItem, 10)
	for i := range items {
		items[i] = HelpItem{Key: "x", Description: "longdescription"}
	}
	bar := NewHelpBar(items)
	bar.SetWidth(20)
	out := bar.View()
	if !strings.Contains(stripANSI(out), "...") {
		t.Errorf("expected truncation marker '...' in narrow bar, got %q", stripANSI(out))
	}
}

func TestHelpBar_InitReturnsNil(t *testing.T) {
	t.Parallel()
	bar := NewHelpBar(nil)
	cmd := bar.Init()
	testkit.AssertEqual(t, "Init returns nil", cmd == nil, true)
}
