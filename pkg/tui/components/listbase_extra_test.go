package components

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestListBase_SetSizeAndSetFocused(t *testing.T) {
	t.Parallel()
	var lb ListBase
	lb.SetSize(80, 20)
	testkit.AssertEqual(t, "width", lb.Width, 80)
	testkit.AssertEqual(t, "height", lb.Height, 20)
	lb.SetFocused(true)
	testkit.AssertEqual(t, "focused", lb.Focused, true)
}

func TestListBase_ItemCount(t *testing.T) {
	t.Parallel()
	var lb ListBase
	lb.ResolveNav = func(string) NavAction { return NavNone }
	lb.SetSize(80, 10)
	lb.SetItemCount(5)
	testkit.AssertEqual(t, "item count", lb.ItemCount(), 5)
}

func TestListBase_ItemCountClampsCursor(t *testing.T) {
	t.Parallel()
	var lb ListBase
	lb.SetSize(80, 10)
	lb.SetItemCount(10)
	lb.Cursor = 8
	lb.SetItemCount(3)
	testkit.AssertEqual(t, "cursor clamped", lb.Cursor, 2)
}
