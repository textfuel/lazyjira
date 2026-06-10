package components

import (
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestTextInput_SetWidthAndView(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetWidth(40)
	input.SetValue("hello")
	out := input.View()
	if out == "" {
		t.Error("expected non-empty view")
	}
	plain := stripANSI(out)
	if !strings.Contains(plain, "hello") {
		t.Errorf("expected 'hello' in view, got %q", plain)
	}
}

func TestTextInput_ViewWithCursorInMiddle(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetWidth(20)
	input.SetValue("hello")
	input.setCursor(2)
	out := input.View()
	testkit.AssertEqual(t, "view non-empty", out != "", true)
}

func TestTextInput_ViewWithHighlighter(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetWidth(40)
	input.Highlighter = func(text []rune) []StyledSegment {
		return []StyledSegment{{Text: string(text)}}
	}
	input.SetValue("query = PROJ")
	out := input.View()
	if !strings.Contains(stripANSI(out), "query") {
		t.Errorf("expected 'query' in highlighted view, got %q", stripANSI(out))
	}
}

func TestTextInput_ViewHighlightedWithCursorInSegment(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetWidth(40)
	input.Highlighter = func(text []rune) []StyledSegment {
		return []StyledSegment{{Text: string(text)}}
	}
	input.SetValue("project")
	input.setCursor(3)
	out := input.View()
	testkit.AssertEqual(t, "non-empty highlighted view", out != "", true)
}

func TestTextInput_ViewCursorAtEnd(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetWidth(40)
	input.SetValue("ab")
	out := input.View()
	testkit.AssertEqual(t, "non-empty view cursor at end", out != "", true)
}

func TestTextInput_ViewLongTextScrolls(t *testing.T) {
	t.Parallel()
	var input TextInput
	input.SetWidth(10)
	input.SetValue("this is a very long string that exceeds width")
	out := input.View()
	testkit.AssertEqual(t, "view non-empty for long text", out != "", true)
}
