package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestTextInput_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initial     string
		cursor      int
		key         tea.KeyType
		runes       string
		wantValue   string
		wantCursor  int
		wantChanged bool
	}{
		{name: "left moves cursor", initial: "hello", cursor: 3, key: tea.KeyLeft, wantValue: "hello", wantCursor: 2},
		{name: "right moves cursor", initial: "hello", cursor: 3, key: tea.KeyRight, wantValue: "hello", wantCursor: 4},
		{name: "right stops at end", initial: "hi", cursor: 2, key: tea.KeyRight, wantValue: "hi", wantCursor: 2},
		{name: "home jumps to start", initial: "hello", cursor: 3, key: tea.KeyHome, wantValue: "hello", wantCursor: 0},
		{name: "end jumps to end", initial: "hello", cursor: 1, key: tea.KeyEnd, wantValue: "hello", wantCursor: 5},
		{name: "backspace deletes before cursor", initial: "hello", cursor: 3, key: tea.KeyBackspace, wantValue: "helo", wantCursor: 2, wantChanged: true},
		{name: "backspace at start is noop", initial: "hello", cursor: 0, key: tea.KeyBackspace, wantValue: "hello", wantCursor: 0},
		{name: "delete removes under cursor", initial: "hello", cursor: 1, key: tea.KeyDelete, wantValue: "hllo", wantCursor: 1, wantChanged: true},
		{name: "delete at end is noop", initial: "hi", cursor: 2, key: tea.KeyDelete, wantValue: "hi", wantCursor: 2},
		{name: "ctrl+a jumps to start", initial: "hello", cursor: 3, key: tea.KeyCtrlA, wantValue: "hello", wantCursor: 0},
		{name: "ctrl+e jumps to end", initial: "hello", cursor: 1, key: tea.KeyCtrlE, wantValue: "hello", wantCursor: 5},
		{name: "ctrl+w deletes word", initial: "foo bar", cursor: 7, key: tea.KeyCtrlW, wantValue: "foo ", wantCursor: 4, wantChanged: true},
		{name: "ctrl+k kills to end", initial: "hello", cursor: 2, key: tea.KeyCtrlK, wantValue: "he", wantCursor: 2, wantChanged: true},
		{name: "ctrl+u kills to start", initial: "hello", cursor: 2, key: tea.KeyCtrlU, wantValue: "llo", wantCursor: 0, wantChanged: true},
		{name: "space inserts", initial: "ab", cursor: 1, key: tea.KeySpace, wantValue: "a b", wantCursor: 2, wantChanged: true},
		{name: "rune inserts at cursor", initial: "ab", cursor: 1, key: tea.KeyRunes, runes: "X", wantValue: "aXb", wantCursor: 2, wantChanged: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			input := NewTextInput()
			input.SetValue(tt.initial)
			input.setCursor(tt.cursor)

			msg := tea.KeyMsg{Type: tt.key}
			if tt.runes != "" {
				msg.Runes = []rune(tt.runes)
			}
			got, changed := input.Update(msg)

			testkit.AssertEqual(t, "value", got.Value(), tt.wantValue)
			testkit.AssertEqual(t, "cursor", got.CursorPos(), tt.wantCursor)
			testkit.AssertEqual(t, "changed", changed, tt.wantChanged)
		})
	}
}

func TestTextInput_InsertAtCursorAndSetValue(t *testing.T) {
	t.Parallel()

	input := NewTextInput()
	input.SetValue("ac")
	testkit.AssertEqual(t, "cursor at end after SetValue", input.CursorPos(), 2)

	input.setCursor(1)
	input.InsertAtCursor("b")
	testkit.AssertEqual(t, "value after insert", input.Value(), "abc")
	testkit.AssertEqual(t, "cursor after insert", input.CursorPos(), 2)
}
