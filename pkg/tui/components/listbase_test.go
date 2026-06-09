package components

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func navResolver(key string) NavAction {
	switch key {
	case "j":
		return NavDown
	case "k":
		return NavUp
	case "g":
		return NavTop
	case "G":
		return NavBottom
	case "d":
		return NavHalfDown
	case "u":
		return NavHalfUp
	}
	return NavNone
}

func newListBase(itemCount, height, cursor int) *ListBase {
	list := &ListBase{ResolveNav: navResolver, Height: height}
	list.SetItemCount(itemCount)
	list.Cursor = cursor
	list.AdjustOffset()
	return list
}

func TestListBase_KeyNav(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cursor     int
		key        string
		wantCursor int
		wantMoved  bool
	}{
		{"down moves one", 0, "j", 1, true},
		{"down wraps at bottom", 9, "j", 0, true},
		{"up moves one", 5, "k", 4, true},
		{"up wraps at top", 0, "k", 9, true},
		{"top jumps to first", 5, "g", 0, true},
		{"bottom jumps to last", 0, "G", 9, true},
		{"half page down", 0, "d", 5, true},
		{"half page up", 9, "u", 4, true},
		{"unknown key does nothing", 3, "z", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			list := newListBase(10, 12, tt.cursor)

			moved := list.KeyNav(tt.key)

			testkit.AssertEqual(t, "moved", moved, tt.wantMoved)
			testkit.AssertEqual(t, "cursor", list.Cursor, tt.wantCursor)
		})
	}
}

func TestListBase_ScrollByClampsToBounds(t *testing.T) {
	t.Parallel()
	list := newListBase(10, 12, 5)

	list.ScrollBy(100)
	testkit.AssertEqual(t, "cursor clamped to last", list.Cursor, 9)

	list.ScrollBy(-100)
	testkit.AssertEqual(t, "cursor clamped to first", list.Cursor, 0)
}

func TestListBase_SetItemCountClampsCursor(t *testing.T) {
	t.Parallel()
	list := newListBase(10, 12, 8)

	list.SetItemCount(5)
	testkit.AssertEqual(t, "cursor clamped into range", list.Cursor, 4)
}

func TestListBase_ClickAt(t *testing.T) {
	t.Parallel()
	list := newListBase(10, 12, 0)

	testkit.AssertEqual(t, "first click is not double", list.ClickAt(3), false)
	testkit.AssertEqual(t, "cursor follows click", list.Cursor, 2)
	testkit.AssertEqual(t, "second click is double", list.ClickAt(3), true)
}

func TestListBase_ClickAtOutOfRangeIgnored(t *testing.T) {
	t.Parallel()
	list := newListBase(10, 12, 4)

	testkit.AssertEqual(t, "out of range click", list.ClickAt(99), false)
	testkit.AssertEqual(t, "cursor unchanged", list.Cursor, 4)
}

func TestListBase_VisibleRowsAndContentHeight(t *testing.T) {
	t.Parallel()

	list := newListBase(3, 12, 0)
	testkit.AssertEqual(t, "VisibleRows", list.VisibleRows(), 10)
	testkit.AssertEqual(t, "ContentHeight grows to items", list.ContentHeight(2), 5)
	testkit.AssertEqual(t, "ContentHeight honors minimum", list.ContentHeight(20), 20)

	tiny := newListBase(3, 1, 0)
	testkit.AssertEqual(t, "VisibleRows floors at one", tiny.VisibleRows(), 1)
}
