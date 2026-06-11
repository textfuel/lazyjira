package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestModal_ShowReadOnlyRendersContent(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.ShowReadOnly(testTitle, items)
	testkit.AssertEqual(t, "visible", m.IsVisible(), true)
	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, testItem1Label) {
		t.Errorf("expected %q in read-only view, got %q", testItem1Label, plain)
	}
}

func TestModal_ShowErrorRendersWithErrorState(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{{ID: testItem1ID, Label: testItem1Label}}
	m.ShowError("Error occurred", items)
	testkit.AssertEqual(t, "visible", m.IsVisible(), true)
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view from ShowError")
	}
}

func TestModal_HideHidesModal(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})
	m.Hide()
	testkit.AssertEqual(t, "hidden", m.IsVisible(), false)
}

func TestModal_SearchViewRendersFilterBar(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})
	out := m.SearchView(80)
	plain := stripANSI(out)
	if !strings.Contains(plain, "/") {
		t.Errorf("expected '/' in search view, got %q", plain)
	}
}

func TestModal_SelectionContentW_DerivedFromItems(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	w := m.selectionContentW()
	if w <= 0 {
		t.Errorf("expected positive selection content width, got %d", w)
	}
}

func TestModal_HandleMouseWheelDown_MovesSelection(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
		{ID: testItem3ID, Label: testItem3Label},
	}
	m.Show(testTitle, items)
	initialCursor := m.cursor
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "cursor moved down", m.cursor > initialCursor, true)
}

func TestModal_HandleMouseWheelUp_MovesSelectionBack(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	before := m.cursor
	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "cursor moved up", m.cursor < before, true)
}

func TestModal_HandleMouseLeftClick_SelectsItem(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	mainBoxH := min(len(items)+4, 22) + 2
	topOffset := (24 - mainBoxH) / 2
	clickY := topOffset + 3
	m, cmd := m.Update(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      clickY,
	})
	if cmd == nil {
		t.Fatal("expected command from left click on item")
	}
	msg := cmd()
	sel, ok := msg.(ModalSelectedMsg)
	if !ok {
		t.Fatalf("expected ModalSelectedMsg, got %T", msg)
	}
	testkit.AssertEqual(t, "selected item ID", sel.Item.ID, testItem1ID)
}

func TestModal_ViewReadOnly_ScrollsWithJK(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := make([]ModalItem, 20)
	for i := range items {
		items[i] = ModalItem{ID: testItem1ID, Label: testItem1Label}
	}
	m.ShowReadOnly(testTitle, items)
	initialOffset := m.offset
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "offset increased", m.offset > initialOffset, true)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	testkit.AssertEqual(t, "offset decreased back", m.offset, initialOffset)
}

func TestModal_ViewSelectable_RendersItems(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, testItem1Label) {
		t.Errorf("expected %q in selectable view, got %q", testItem1Label, plain)
	}
}

func TestModal_ViewSelectable_FooterShowsCount(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "of 2") {
		t.Errorf("expected 'of 2' footer in view, got %q", plain)
	}
}

func TestModal_RenderDrawsOnBackground(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})
	bg := testkit.BlankCanvas(80, 24)
	out := m.Render(bg, 80, 24)
	plain := stripANSI(out)
	if !strings.Contains(plain, testItem1Label) {
		t.Errorf("expected item label in rendered output, got %q", plain)
	}
}

func TestModal_RenderInvisibleReturnsBackground(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	bg := "background content"
	out := m.Render(bg, 80, 24)
	testkit.AssertEqual(t, "bg passthrough", out, bg)
}

func TestModal_HintView_EmptyWhenNoHint(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})
	hint := m.HintView()
	testkit.AssertEqual(t, "no hint when item has no hint", hint, "")
}

func TestModal_HintView_ShowsHintWhenPresent(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label, Hint: "useful hint text"}})
	hint := m.HintView()
	plain := stripANSI(hint)
	if !strings.Contains(plain, "useful hint") {
		t.Errorf("expected hint text in HintView output, got %q", plain)
	}
}

func TestModal_ChecklistContentW_DerivedFromItems(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.ShowChecklist(testTitle, items, nil)
	w := m.checklistContentW()
	if w <= 0 {
		t.Errorf("expected positive checklist content width, got %d", w)
	}
}

func TestModal_InterceptConsumesKeyWhenVisible(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.Show(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})
	_, consumed := m.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "key consumed", consumed, true)
}

func TestModal_InterceptPassesThroughWhenHidden(t *testing.T) {
	t.Parallel()
	m := NewModal()
	_, consumed := m.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "not consumed when hidden", consumed, false)
}

func TestModal_SearchFilterAndConfirm(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	testkit.AssertEqual(t, "searching activated", m.IsSearching(), true)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	testkit.AssertEqual(t, "searching stopped after enter", m.IsSearching(), false)
}

func TestModal_SearchFilterEscRestores(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "searching stopped after esc", m.IsSearching(), false)
}

func TestModal_ChecklistToggleWithSpace(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.ShowChecklist(testTitle, items, nil)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	testkit.AssertEqual(t, "item toggled", m.selected[testItem1ID], true)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	testkit.AssertEqual(t, "item untoggled", m.selected[testItem1ID], false)
}

func TestModal_ChecklistConfirmSendsSelectedItems(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.ShowChecklist(testTitle, items, map[string]bool{testItem1ID: true})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from checklist confirm")
	}
	msg := cmd()
	confirmed, ok := msg.(ChecklistConfirmedMsg)
	if !ok {
		t.Fatalf("expected ChecklistConfirmedMsg, got %T", msg)
	}
	if len(confirmed.Selected) != 1 || confirmed.Selected[0].ID != testItem1ID {
		t.Errorf("expected selected item %q, got %v", testItem1ID, confirmed.Selected)
	}
}

func TestModal_ReadOnlyEnterHidesAndCancels(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	m.ShowReadOnly(testTitle, []ModalItem{{ID: testItem1ID, Label: testItem1Label}})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	testkit.AssertEqual(t, "hidden", m.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel command")
	}
	if _, ok := cmd().(ModalCancelledMsg); !ok {
		t.Error("expected ModalCancelledMsg")
	}
}

func TestModal_RenderItems_WithSeparatorAndActive(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label, Active: true},
		{Separator: true, Label: "---sep---"},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	view := m.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, testItem1Label) {
		t.Errorf("expected active item in view, got %q", plain)
	}
}

func TestModal_RenderItems_InternalItemRendersWithColor(t *testing.T) {
	t.Parallel()
	m := NewModal()
	m.SetSize(80, 24)
	items := []ModalItem{
		{ID: testItem1ID, Label: testItem1Label, Internal: true},
		{ID: testItem2ID, Label: testItem2Label},
	}
	m.Show(testTitle, items)
	view := m.View()
	if !strings.Contains(stripANSI(view), testItem1Label) {
		t.Errorf("expected internal item label %q in view, got %q", testItem1Label, stripANSI(view))
	}
	testkit.AssertEqual(t, "internal item rendered with color escape codes", view != stripANSI(view), true)
}
