package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRenderCollapsedBar_FocusedContainsTitle(t *testing.T) {
	t.Parallel()
	out := RenderCollapsedBar(testTitle, "", 40, true)
	if !strings.Contains(stripANSI(out), testTitle) {
		t.Errorf("expected title %q in output, got %q", testTitle, stripANSI(out))
	}
}

func TestRenderCollapsedBar_UnfocusedContainsTitle(t *testing.T) {
	t.Parallel()
	out := RenderCollapsedBar(testTitle, "", 40, false)
	if !strings.Contains(stripANSI(out), testTitle) {
		t.Errorf("expected title %q in output, got %q", testTitle, stripANSI(out))
	}
}

func TestRenderCollapsedBar_WithFooter(t *testing.T) {
	t.Parallel()
	out := RenderCollapsedBar(testTitle, "1 of 5", 40, false)
	plain := stripANSI(out)
	if !strings.Contains(plain, testTitle) {
		t.Errorf("expected title in output, got %q", plain)
	}
	if !strings.Contains(plain, "1 of 5") {
		t.Errorf("expected footer in output, got %q", plain)
	}
}

func TestRenderPanel_ContainsTitleAndContent(t *testing.T) {
	t.Parallel()
	out := RenderPanel(testTitle, "hello world", 40, 3, true)
	plain := stripANSI(out)
	if !strings.Contains(plain, testTitle) {
		t.Errorf("expected title %q in panel, got %q", testTitle, plain)
	}
	if !strings.Contains(plain, "hello world") {
		t.Errorf("expected content in panel, got %q", plain)
	}
}

func TestRenderPanelFull_FooterVisible(t *testing.T) {
	t.Parallel()
	out := RenderPanelFull(testTitle, "2 of 10", "line1\nline2", 40, 2, false, nil)
	plain := stripANSI(out)
	if !strings.Contains(plain, "2 of 10") {
		t.Errorf("expected footer in panel, got %q", plain)
	}
}

func TestRenderPanelWithColor_Renders(t *testing.T) {
	t.Parallel()
	out := RenderPanelWithColor(testTitle, "1 of 3", "content", 40, 3, nil, lipgloss.Color("9"))
	plain := stripANSI(out)
	if !strings.Contains(plain, testTitle) {
		t.Errorf("expected title in panel, got %q", plain)
	}
}

func TestRenderPanel_WithScrollInfo(t *testing.T) {
	t.Parallel()
	scroll := &ScrollInfo{Total: 20, Visible: 5, Offset: 3}
	out := RenderPanelFull(testTitle, "", "content", 40, 5, true, scroll)
	if len(out) == 0 {
		t.Error("expected non-empty panel with scrollbar")
	}
}

func TestRenderPanel_TruncatesExtraContentLines(t *testing.T) {
	t.Parallel()
	content := strings.Repeat("line\n", 20)
	out := RenderPanel(testTitle, content, 40, 5, false)
	lineCount := strings.Count(out, "\n")
	if lineCount > 10 {
		t.Errorf("expected at most ~8 lines for innerHeight=5, got %d", lineCount)
	}
}
