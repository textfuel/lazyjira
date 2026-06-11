package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestCreateForm_ScrollFocusedSummary(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 15)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.summaryCursor = 0
	form.Intercept(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "scroll down on summary moved cursor", form.summaryCursor > 0, true)
	form.summaryCursor = 5
	form.Intercept(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "scroll up on summary clamps to 0", form.summaryCursor, 0)
}

func TestCreateForm_ScrollFocusedFields(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 15)
	fields := make([]CreateFormField, 0, 17)
	fields = append(fields, CreateFormField{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText})
	fields = append(fields, CreateFormField{FieldID: "description", Name: "Description", Type: CFFieldMultiText})
	for range 15 {
		fields = append(fields, CreateFormField{FieldID: "f", Name: "F", Type: CFFieldSingleSelect})
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "on fields panel", form.FocusedPanel(), CreatePanelFields)
	form.Intercept(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "scroll down on fields moved cursor", form.fieldCursor > 0, true)
	form.Intercept(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "scroll up on fields decreased cursor", form.fieldCursor, 0)
}

func TestCreateForm_RenderSummaryPlainNotFocused(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	bg := testkit.BlankCanvas(120, 40)
	out := form.Render(bg, 120, 40)
	plain := stripANSI(out)
	if !strings.Contains(plain, testSummaryText) {
		t.Errorf("expected summary text in unfocused render, got %q", plain)
	}
}

func TestCreateForm_SetFieldValue_SummaryField(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.SetFieldValue(0, "new summary", "new summary")
	testkit.AssertEqual(t, "summary text updated", string(form.summaryText), "new summary")
}

func TestCreateForm_InterceptMouseClickOnSummaryPanel(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "focus is on description before click", form.FocusedPanel(), CreatePanelDescription)
	formW := min(max(120*6/10, 40), 120-2)
	formX := (120 - formW) / 2
	form.Intercept(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      formX + 5,
		Y:      16,
	})
	testkit.AssertEqual(t, "click on summary row switches focus to summary", form.FocusedPanel(), CreatePanelSummary)
}

func TestCreateForm_EditCurrentField_EmptyFiltered(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	testkit.AssertEqual(t, "consumed even with empty fields", consumed, true)
	testkit.AssertEqual(t, "nil cmd when no fields", cmd == nil, true)
}

func TestCreateForm_InterceptDescriptionQHides(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	testkit.AssertEqual(t, "consumed", consumed, true)
	testkit.AssertEqual(t, "hidden after q", form.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel cmd from q")
	}
	if _, ok := cmd().(CreateFormCancelMsg); !ok {
		t.Error("expected CreateFormCancelMsg from q")
	}
}

func TestCreateForm_InterceptFieldsQHides(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	testkit.AssertEqual(t, "hidden after q on fields", form.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel cmd from fields q")
	}
}

func TestCreateForm_FilteredFields_WithFilter(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "prio", Name: testFieldName, Type: CFFieldSingleSelect},
		{FieldID: "labels", Name: "Labels", Type: CFFieldMultiSelect},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	form.filterInput.SetValue("pri")
	filtered := form.filteredFields()
	testkit.AssertEqual(t, "filter returns matching fields", len(filtered), 1)
}

func TestCreateForm_EnsureFieldVisible_ScrollsOffset(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 12)
	fields := make([]CreateFormField, 0, 22)
	fields = append(fields, CreateFormField{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText})
	fields = append(fields, CreateFormField{FieldID: "description", Name: "Description", Type: CFFieldMultiText})
	for range 15 {
		fields = append(fields, CreateFormField{FieldID: "f", Name: "F", Type: CFFieldSingleSelect})
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.fieldCursor = 10
	form.ensureFieldVisible()
	testkit.AssertEqual(t, "offset adjusted to show cursor", form.fieldOffset > 0, true)
}
