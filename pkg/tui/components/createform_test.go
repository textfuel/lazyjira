package components

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func makeTestFields() []CreateFormField {
	return []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText, Required: true, DisplayValue: testSummaryText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: testFieldID, Name: testFieldName, Type: CFFieldSingleSelect, AllowedValues: []ModalItem{
			{ID: testItem1ID, Label: testItem1Label},
			{ID: testItem2ID, Label: testItem2Label},
		}},
	}
}

func TestCreateForm_NewCreateForm(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	testkit.AssertEqual(t, "initially invisible", form.IsVisible(), false)
}

func TestCreateForm_ShowForm_MakesVisible(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	testkit.AssertEqual(t, "visible after ShowForm", form.IsVisible(), true)
}

func TestCreateForm_PauseAndResume(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Pause()
	testkit.AssertEqual(t, "paused", form.paused, true)
	form.Resume()
	testkit.AssertEqual(t, "resumed", form.paused, false)
}

func TestCreateForm_FocusedPanel(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	testkit.AssertEqual(t, "default panel is summary", form.FocusedPanel(), CreatePanelSummary)
}

func TestCreateForm_Hide(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Hide()
	testkit.AssertEqual(t, "hidden after Hide", form.IsVisible(), false)
}

func TestCreateForm_SetFieldValue(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.SetFieldValue(2, "high", testItem1Label)
	field := form.FieldAt(2)
	if field == nil {
		t.Fatal("expected field at index 2")
	}
	testkit.AssertEqual(t, "field value set", field.DisplayValue, testItem1Label)
}

func TestCreateForm_IndexGuards(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		run  func(form *CreateForm)
	}{
		{
			name: "FieldAt negative index returns nil",
			run: func(form *CreateForm) {
				testkit.AssertEqual(t, "nil for negative index", form.FieldAt(-1), (*CreateFormField)(nil))
			},
		},
		{
			name: "FieldAt large index returns nil",
			run: func(form *CreateForm) {
				testkit.AssertEqual(t, "nil for large index", form.FieldAt(100), (*CreateFormField)(nil))
			},
		},
		{
			name: "SetFieldValue out-of-range leaves existing fields unchanged",
			run: func(form *CreateForm) {
				before := form.FieldAt(2).DisplayValue
				form.SetFieldValue(99, "x", "x")
				testkit.AssertEqual(t, "field at index 2 unchanged", form.FieldAt(2).DisplayValue, before)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			form := NewCreateForm(nil)
			form.SetSize(120, 40)
			form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
			tc.run(&form)
		})
	}
}

func TestCreateForm_SetError(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.SetError("something failed")
	testkit.AssertEqual(t, "error set", form.errorMsg, "something failed")
	testkit.AssertEqual(t, "loading cleared", form.loading, false)
}

func TestCreateForm_SetLoading(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.SetLoading(true)
	testkit.AssertEqual(t, "loading set", form.loading, true)
	testkit.AssertEqual(t, "visible when loading", form.IsVisible(), true)
}

func TestCreateForm_SetSize(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(100, 30)
	testkit.AssertEqual(t, "width", form.width, 100)
	testkit.AssertEqual(t, "height", form.height, 30)
}

func TestCreateForm_IsFiltering_FilterQuery_FilterBarView(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	testkit.AssertEqual(t, "not filtering initially", form.IsFiltering(), false)
	testkit.AssertEqual(t, "empty filter query", form.FilterQuery(), "")
	form.filtering = true
	testkit.AssertEqual(t, "filtering", form.IsFiltering(), true)
	barView := form.FilterBarView()
	if !strings.Contains(stripANSI(barView), "/") {
		t.Errorf("expected '/' in filter bar, got %q", stripANSI(barView))
	}
}

func TestCreateForm_SetDescRenderer(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	called := false
	form.SetDescRenderer(func(text string, width int) []string {
		called = true
		return []string{text}
	})
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText, DisplayValue: "some desc"},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Render("", 120, 40)
	testkit.AssertEqual(t, "desc renderer called during render", called, true)
}

func TestCreateForm_InterceptTabCyclesPanels(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	testkit.AssertEqual(t, "starts on summary", form.FocusedPanel(), CreatePanelSummary)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "tab moves to description", form.FocusedPanel(), CreatePanelDescription)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "tab moves to fields", form.FocusedPanel(), CreatePanelFields)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "tab wraps to summary", form.FocusedPanel(), CreatePanelSummary)
}

func TestCreateForm_InterceptShiftTabCyclesBackward(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyShiftTab})
	testkit.AssertEqual(t, "shift-tab wraps to fields", form.FocusedPanel(), CreatePanelFields)
}

func TestCreateForm_InterceptEscOnSummaryHides(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	cmd, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "consumed", consumed, true)
	testkit.AssertEqual(t, "hidden after esc on summary", form.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel cmd")
	}
	if _, ok := cmd().(CreateFormCancelMsg); !ok {
		t.Error("expected CreateFormCancelMsg")
	}
}

func TestCreateForm_InterceptSummaryBackspaceDeleteLeft(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	initialLen := len(form.summaryText)
	form.Intercept(tea.KeyMsg{Type: tea.KeyBackspace})
	testkit.AssertEqual(t, "backspace removes char", len(form.summaryText), initialLen-1)
}

func TestCreateForm_InterceptSummaryDeleteForward(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.summaryCursor = 0
	initialLen := len(form.summaryText)
	form.Intercept(tea.KeyMsg{Type: tea.KeyDelete})
	testkit.AssertEqual(t, "delete removes char", len(form.summaryText), initialLen-1)
}

func TestCreateForm_InterceptSummaryLeftRight(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	endPos := form.summaryCursor
	form.Intercept(tea.KeyMsg{Type: tea.KeyLeft})
	testkit.AssertEqual(t, "left moves cursor back", form.summaryCursor, endPos-1)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRight})
	testkit.AssertEqual(t, "right moves cursor forward", form.summaryCursor, endPos)
}

func TestCreateForm_InterceptSummaryHomeEnd(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyHome})
	testkit.AssertEqual(t, "cursor at start", form.summaryCursor, 0)
	form.Intercept(tea.KeyMsg{Type: tea.KeyEnd})
	testkit.AssertEqual(t, "cursor at end", form.summaryCursor, len(form.summaryText))
}

func TestCreateForm_InterceptSummaryCtrlAE(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyCtrlA})
	testkit.AssertEqual(t, "ctrl+a moves to start", form.summaryCursor, 0)
	form.Intercept(tea.KeyMsg{Type: tea.KeyCtrlE})
	testkit.AssertEqual(t, "ctrl+e moves to end", form.summaryCursor, len(form.summaryText))
}

func TestCreateForm_InterceptSummaryCtrlUK(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyCtrlU})
	testkit.AssertEqual(t, "ctrl+u clears to start", len(form.summaryText), 0)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.summaryCursor = 0
	form.Intercept(tea.KeyMsg{Type: tea.KeyCtrlK})
	testkit.AssertEqual(t, "ctrl+k clears to end when cursor at start", len(form.summaryText), 0)
}

func TestCreateForm_InterceptSummarySpace(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	initialLen := len(form.summaryText)
	form.Intercept(tea.KeyMsg{Type: tea.KeySpace})
	testkit.AssertEqual(t, "space appended", len(form.summaryText), initialLen+1)
}

func TestCreateForm_InterceptSummaryRunes(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	initialLen := len(form.summaryText)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("AB")})
	testkit.AssertEqual(t, "runes appended", len(form.summaryText), initialLen+2)
}

func TestCreateForm_InterceptDescriptionScrollJK(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 15)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText, DisplayValue: strings.Repeat("long line\n", 30)},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "desc offset increased", form.descOffset > 0, true)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	testkit.AssertEqual(t, "desc offset decreased", form.descOffset, 0)
}

func TestCreateForm_InterceptDescriptionGG(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 15)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText, DisplayValue: strings.Repeat("line\n", 30)},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	testkit.AssertEqual(t, "G scrolls to bottom", form.descOffset > 0, true)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	testkit.AssertEqual(t, "g scrolls to top", form.descOffset, 0)
}

func TestCreateForm_InterceptDescriptionCtrlDU(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 15)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText, DisplayValue: strings.Repeat("x\n", 50)},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyCtrlD)})
	testkit.AssertEqual(t, "ctrl+d scrolls down", form.descOffset > 0, true)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyCtrlU)})
	testkit.AssertEqual(t, "ctrl+u scrolls up", form.descOffset, 0)
}

func TestCreateForm_InterceptDescriptionEscHides(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "hidden after desc esc", form.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel cmd from desc esc")
	}
	if _, ok := cmd().(CreateFormCancelMsg); !ok {
		t.Error("expected CreateFormCancelMsg from desc esc")
	}
}

func TestCreateForm_InterceptDescriptionEEditExternal(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	testkit.AssertEqual(t, "consumed", consumed, true)
	if cmd == nil {
		t.Fatal("expected external edit cmd")
	}
	if _, ok := cmd().(CreateFormEditExternalMsg); !ok {
		t.Errorf("expected CreateFormEditExternalMsg, got %T", cmd())
	}
}

func TestCreateForm_InterceptFieldsJK(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "f1", Name: "Field1", Type: CFFieldSingleSelect, AllowedValues: []ModalItem{{ID: "a", Label: "A"}}},
		{FieldID: "f2", Name: "Field2", Type: CFFieldSingleSelect, AllowedValues: []ModalItem{{ID: "b", Label: "B"}}},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "on fields panel", form.FocusedPanel(), CreatePanelFields)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "field cursor moved down", form.fieldCursor, 1)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	testkit.AssertEqual(t, "field cursor moved up", form.fieldCursor, 0)
}

func TestCreateForm_InterceptFieldsGG(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "f1", Name: "F1", Type: CFFieldSingleSelect},
		{FieldID: "f2", Name: "F2", Type: CFFieldSingleSelect},
		{FieldID: "f3", Name: "F3", Type: CFFieldSingleSelect},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	testkit.AssertEqual(t, "G moves to last field", form.fieldCursor, 2)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	testkit.AssertEqual(t, "g moves to first field", form.fieldCursor, 0)
}

func TestCreateForm_InterceptFieldsCtrlDU(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := make([]CreateFormField, 0, 22)
	fields = append(fields, CreateFormField{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText})
	fields = append(fields, CreateFormField{FieldID: "description", Name: "Description", Type: CFFieldMultiText})
	for range 20 {
		fields = append(fields, CreateFormField{FieldID: "f", Name: "F", Type: CFFieldSingleSelect})
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyCtrlD)})
	testkit.AssertEqual(t, "ctrl+d moves cursor down", form.fieldCursor > 0, true)
	before := form.fieldCursor
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyCtrlU)})
	testkit.AssertEqual(t, "ctrl+u moves cursor up", form.fieldCursor < before, true)
}

func TestCreateForm_InterceptFieldsSlashEnablesFilter(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	testkit.AssertEqual(t, "filtering enabled", form.IsFiltering(), true)
}

func TestCreateForm_InterceptFieldsEscHides(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "hidden after fields esc", form.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel cmd from fields esc")
	}
}

func TestCreateForm_InterceptFieldsEditSingleSelect(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatal("expected command from e on single select")
	}
	if _, ok := cmd().(CreateFormPickerMsg); !ok {
		t.Errorf("expected CreateFormPickerMsg, got %T", cmd())
	}
}

func TestCreateForm_InterceptFieldsEditSingleText(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "f1", Name: "TextField", Type: CFFieldSingleText},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatal("expected command from e on text field")
	}
	if _, ok := cmd().(CreateFormEditTextMsg); !ok {
		t.Errorf("expected CreateFormEditTextMsg, got %T", cmd())
	}
}

func TestCreateForm_InterceptFieldsEditMultiText(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "f1", Name: "MultiTextField", Type: CFFieldMultiText},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatal("expected command from e on multi-text field")
	}
	if _, ok := cmd().(CreateFormEditExternalMsg); !ok {
		t.Errorf("expected CreateFormEditExternalMsg, got %T", cmd())
	}
}

func TestCreateForm_InterceptFieldsEditMultiSelect(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "f1", Name: "MultiSelect", Type: CFFieldMultiSelect, AllowedValues: []ModalItem{{ID: "a", Label: "A"}}},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	cmd, _ := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if cmd == nil {
		t.Fatal("expected command from e on multi-select field")
	}
	if _, ok := cmd().(CreateFormChecklistMsg); !ok {
		t.Errorf("expected CreateFormChecklistMsg, got %T", cmd())
	}
}

func TestCreateForm_InterceptFieldsEnterSubmitsMissingRequired(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: testFieldID, Name: testFieldName, Type: CFFieldSingleSelect, Required: true, DisplayValue: "", Value: nil},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	_, _ = form.Intercept(tea.KeyMsg{Type: tea.KeyEnter})
	testkit.AssertEqual(t, "error set for missing required", form.errorMsg != "", true)
}

func TestCreateForm_SubmitFormSucceeds(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText, DisplayValue: testSummaryText, Value: testSummaryText},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	cmd, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyEnter})
	testkit.AssertEqual(t, "consumed", consumed, true)
	if cmd == nil {
		t.Fatal("expected submit cmd")
	}
	msg, ok := cmd().(CreateFormSubmitMsg)
	if !ok {
		t.Fatalf("expected CreateFormSubmitMsg, got %T", cmd())
	}
	if _, hasSummary := msg.Fields["summary"]; !hasSummary {
		t.Error("expected summary field in submit fields")
	}
}

func TestCreateForm_InterceptFilterEscRestores(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	testkit.AssertEqual(t, "filtering on", form.IsFiltering(), true)
	form.Intercept(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "filtering off", form.IsFiltering(), false)
}

func TestCreateForm_InterceptFilterEnterConfirms(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "prio", Name: testFieldName, Type: CFFieldSingleSelect},
		{FieldID: "assignee", Name: "Assignee", Type: CFFieldPerson},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("P")})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyEnter)})
	testkit.AssertEqual(t, "filtering off after enter", form.IsFiltering(), false)
}

func TestCreateForm_InterceptFilterJKNav(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText},
		{FieldID: "f1", Name: "FieldAlpha", Type: CFFieldSingleSelect},
		{FieldID: "f2", Name: "FieldBeta", Type: CFFieldSingleSelect},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyDown)})
	testkit.AssertEqual(t, "filter nav down", form.fieldCursor, 1)
	form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("up")})
	testkit.AssertEqual(t, "filter nav up", form.fieldCursor, 0)
}

func TestCreateForm_InterceptPausedIgnoresKeys(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.Pause()
	_, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyEnter})
	testkit.AssertEqual(t, "paused does not consume", consumed, false)
}

func TestCreateForm_InterceptLoadingEscCancels(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.loading = true
	cmd, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyEsc})
	testkit.AssertEqual(t, "consumed during loading", consumed, true)
	testkit.AssertEqual(t, "hidden after loading esc", form.IsVisible(), false)
	if cmd == nil {
		t.Fatal("expected cancel cmd during loading esc")
	}
	if _, ok := cmd().(CreateFormCancelMsg); !ok {
		t.Error("expected CreateFormCancelMsg")
	}
}

func TestCreateForm_InterceptLoadingOtherKeyIgnored(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.loading = true
	_, consumed := form.Intercept(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "consumed but no action", consumed, true)
	testkit.AssertEqual(t, "still visible", form.IsVisible(), true)
}

func TestCreateForm_RenderInvisibleReturnsBackground(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	bg := "background content"
	out := form.Render(bg, 80, 24)
	testkit.AssertEqual(t, "bg passthrough", out, bg)
}

func TestCreateForm_RenderLoadingNoFieldsShowsSpinner(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 24)
	form.SetLoading(true)
	bg := testkit.BlankCanvas(80, 24)
	out := form.Render(bg, 80, 24)
	plain := stripANSI(out)
	if !strings.Contains(plain, "Loading") {
		t.Errorf("expected 'Loading' in render output, got %q", plain)
	}
}

func TestCreateForm_RenderFormDraws(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	bg := testkit.BlankCanvas(120, 40)
	out := form.Render(bg, 120, 40)
	plain := stripANSI(out)
	if !strings.Contains(plain, "Summary") {
		t.Errorf("expected 'Summary' in rendered form, got %q", plain)
	}
}

func TestCreateForm_RenderWithError(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm(makeTestFields(), testIssueType, testProjectKey)
	form.SetError("1 required field(s) empty")
	bg := testkit.BlankCanvas(120, 40)
	out := form.Render(bg, 120, 40)
	plain := stripANSI(out)
	if !strings.Contains(plain, "required") {
		t.Errorf("expected error text in render, got %q", plain)
	}
}

func TestCreateForm_SubmitNamesEmptyRequiredFields(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(120, 40)
	form.ShowForm([]CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText, Required: true, DisplayValue: "filled"},
		{FieldID: "priority", Name: "Priority", Type: CFFieldSingleSelect, Required: true},
		{FieldID: "components", Name: "Components", Type: CFFieldMultiSelect, Required: true},
	}, testIssueType, testProjectKey)

	cmd, _ := form.submitForm()

	if cmd != nil {
		t.Fatal("expected no submit cmd while required fields are empty")
	}
	if !strings.Contains(form.errorMsg, "Priority") || !strings.Contains(form.errorMsg, "Components") {
		t.Errorf("error should name the empty required fields, got %q", form.errorMsg)
	}
}

func TestCreateForm_InterceptMouseWheelScrolls(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(nil)
	form.SetSize(80, 15)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText, DisplayValue: testSummaryText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText, DisplayValue: strings.Repeat("line\n", 30)},
		{FieldID: testFieldID, Name: testFieldName, Type: CFFieldSingleSelect},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	testkit.AssertEqual(t, "on description panel", form.FocusedPanel(), CreatePanelDescription)
	_, consumed := form.Intercept(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "wheel down consumed", consumed, true)
	testkit.AssertEqual(t, "desc offset increased after wheel down", form.descOffset > 0, true)
	offsetAfterDown := form.descOffset
	_, consumed = form.Intercept(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	testkit.AssertEqual(t, "wheel up consumed", consumed, true)
	testkit.AssertEqual(t, "desc offset decreased after wheel up", form.descOffset < offsetAfterDown, true)
}

func TestCreateForm_WrapTextLines(t *testing.T) {
	t.Parallel()
	lines := wrapTextLines("hello world", 5)
	if len(lines) == 0 {
		t.Error("expected non-empty lines")
	}
	for _, line := range lines {
		if lipgloss.Width(line) > 5 {
			t.Errorf("line %q exceeds maxWidth 5 display columns", line)
		}
	}
}

func TestCreateForm_WrapTextLinesEmpty(t *testing.T) {
	t.Parallel()
	lines := wrapTextLines("", 10)
	testkit.AssertEqual(t, "empty string returns one line", len(lines), 1)
}

func TestCreateForm_WrapTextLinesZeroWidth(t *testing.T) {
	t.Parallel()
	lines := wrapTextLines("hello", 0)
	testkit.AssertEqual(t, "zero width returns original", lines[0], "hello")
}

func TestCreateForm_NoneStyle(t *testing.T) {
	t.Parallel()
	style := noneStyle()
	out := style.Render("None")
	testkit.AssertEqual(t, "noneStyle applies color", out != stripANSI(out), true)
}

func TestCreateForm_StyleFieldValue_None(t *testing.T) {
	t.Parallel()
	field := CreateFormField{FieldID: "priority", Type: CFFieldSingleSelect}
	out := styleFieldValue(field, "None")
	plain := stripANSI(out)
	testkit.AssertEqual(t, "None rendered", plain, "None")
}

func TestCreateForm_StyleFieldValue_Priority(t *testing.T) {
	t.Parallel()
	field := CreateFormField{FieldID: "priority", Type: CFFieldSingleSelect}
	out := styleFieldValue(field, "High")
	testkit.AssertEqual(t, "priority value preserved", stripANSI(out), "High")
	testkit.AssertEqual(t, "priority color applied", out != stripANSI(out), true)
}

func TestCreateForm_StyleFieldValue_Person(t *testing.T) {
	t.Parallel()
	field := CreateFormField{FieldID: "assignee", Type: CFFieldPerson}
	out := styleFieldValue(field, "John Doe")
	testkit.AssertEqual(t, "person value preserved", stripANSI(out), "John Doe")
	testkit.AssertEqual(t, "person color applied", out != stripANSI(out), true)
}

func TestCreateForm_StyleFieldValue_SchemaUser(t *testing.T) {
	t.Parallel()
	field := CreateFormField{FieldID: "reporter", Type: CFFieldSingleSelect, SchemaItems: "user"}
	out := styleFieldValue(field, "Alice")
	testkit.AssertEqual(t, "schema user value preserved", stripANSI(out), "Alice")
	testkit.AssertEqual(t, "schema user color applied", out != stripANSI(out), true)
}

func TestCreateForm_StyleFieldValue_DefaultField(t *testing.T) {
	t.Parallel()
	field := CreateFormField{FieldID: "labels", Type: CFFieldMultiSelect}
	out := styleFieldValue(field, "backend")
	testkit.AssertEqual(t, "default field returns value", out, "backend")
}

func TestCreateForm_DescRenderer_ADF(t *testing.T) {
	t.Parallel()
	form := NewCreateForm(func(adf any, width int) []string {
		return []string{"adf rendered"}
	})
	form.SetSize(120, 40)
	fields := []CreateFormField{
		{FieldID: "summary", Name: "Summary", Type: CFFieldSingleText},
		{FieldID: "description", Name: "Description", Type: CFFieldMultiText, DisplayValue: "plain", Value: map[string]any{"type": "doc"}},
	}
	form.ShowForm(fields, testIssueType, testProjectKey)
	form.Intercept(tea.KeyMsg{Type: tea.KeyTab})
	bg := testkit.BlankCanvas(120, 40)
	out := form.Render(bg, 120, 40)
	plain := stripANSI(out)
	if !strings.Contains(plain, "adf rendered") {
		t.Errorf("expected ADF rendered content in output, got %q", plain)
	}
}
