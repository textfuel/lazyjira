package tui

import (
	"context"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestIsCustomField(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fieldID string
		want    bool
	}{
		{"customfield_10001", true},
		{"customfield_99999", true},
		{fldPriority, false},
		{fldAssignee, false},
		{"status", false},
		{"summary", false},
	}

	for _, tc := range cases {
		t.Run(tc.fieldID, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "isCustomField", isCustomField(tc.fieldID), tc.want)
		})
	}
}

func TestFieldMultilineEnabled(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fields  []config.FieldConfig
		fieldID string
		want    bool
	}{
		{
			name:    "false by default",
			fields:  []config.FieldConfig{{ID: "customfield_10001", Name: "Story Points", Multiline: false}},
			fieldID: "customfield_10001",
			want:    false,
		},
		{
			name:    "true when set",
			fields:  []config.FieldConfig{{ID: "customfield_10001", Name: "Notes", Multiline: true}},
			fieldID: "customfield_10001",
			want:    true,
		},
		{
			name:    "false for unknown field",
			fields:  nil,
			fieldID: "customfield_99999",
			want:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp()
			app.cfg.Fields = tc.fields
			testkit.AssertEqual(t, "multiline", app.fieldMultilineEnabled(tc.fieldID), tc.want)
		})
	}
}

func TestConfiguredFieldType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		fields  []config.FieldConfig
		fieldID string
		want    string
	}{
		{
			name:    "returns configured type",
			fields:  []config.FieldConfig{{ID: "customfield_10001", Name: "Status", Type: "select"}},
			fieldID: "customfield_10001",
			want:    "select",
		},
		{
			name:    "empty for unknown",
			fields:  nil,
			fieldID: "customfield_99999",
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp()
			app.cfg.Fields = tc.fields
			testkit.AssertEqual(t, "field type", app.configuredFieldType(tc.fieldID), tc.want)
		})
	}
}

func TestIsPersonSchema(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		schemaType  string
		schemaItems string
		want        bool
	}{
		{
			name:        "user type",
			schemaType:  schemaUser,
			schemaItems: "",
			want:        true,
		},
		{
			name:        "array of users",
			schemaType:  schemaArray,
			schemaItems: schemaUser,
			want:        true,
		},
		{
			name:        "non-user type",
			schemaType:  "string",
			schemaItems: "",
			want:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := newTestApp()
			testkit.AssertEqual(t, "isPersonSchema", app.isPersonSchema(tc.schemaType, tc.schemaItems), tc.want)
		})
	}
}

func TestEditInfoField_NoSelectionIsNoop(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	issue := &jira.Issue{Key: testKey, Summary: testSummary}

	_, cmd := app.editInfoField(issue)

	if cmd != nil {
		t.Error("expected nil cmd with no field selection")
	}
}

func TestEditInfoField_StatusFieldFetchesTransitions(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetTransitionsFunc = func(_ context.Context, _ string) ([]jira.Transition, error) {
		return []jira.Transition{{ID: "1", Name: "Done"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	issue := &jira.Issue{
		Key:     testKey,
		Summary: testSummary,
		Status:  &jira.Status{Name: "Open"},
	}
	app.infoPanel.SetIssue(issue)

	_, cmd := app.editInfoField(issue)

	if cmd == nil {
		t.Error("expected transitions fetch cmd for status field")
	}
}

func TestEditInfoField_PriorityFieldFetchesPriorities(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetPrioritiesFunc = func(_ context.Context) ([]jira.Priority, error) {
		return []jira.Priority{{ID: "1", Name: "High"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	issue := &jira.Issue{
		Key:      testKey,
		Summary:  testSummary,
		Priority: &jira.Priority{Name: "Medium"},
	}
	app.infoPanel.SetIssue(issue)

	_, cmd := app.editInfoField(issue)

	if cmd == nil {
		t.Error("expected priorities fetch cmd for priority field")
	}
}

func TestEditInfoField_TextFieldShowsInputModal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	app.cfg.Fields = []config.FieldConfig{
		{ID: "customfield_10001", Name: "Story Points", Type: "text"},
	}
	issue := &jira.Issue{
		Key:       testKey,
		Summary:   testSummary,
		IssueType: &jira.IssueType{ID: "10000", Name: "Story"},
		CustomFields: map[string]any{
			"customfield_10001": "5",
		},
	}
	app.infoPanel.SetFields(app.cfg.Fields)
	app.infoPanel.SetIssue(issue)

	_, _ = app.editInfoField(issue)

	if !app.inputModal.IsVisible() {
		t.Error("expected input modal to be visible for text custom field")
	}
}

func TestHandleCustomFieldOptions_EmptyOptionsShowsInputModal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40

	msg := customFieldOptionsMsg{
		issueKey:  testKey,
		fieldID:   "customfield_10001",
		fieldName: "Sprint",
		fieldType: views.FieldSingleSelect,
		options:   nil,
	}

	_, _ = app.handleCustomFieldOptions(msg)

	if !app.inputModal.IsVisible() {
		t.Error("input modal should show when no options available")
	}
}

func TestHandleCustomFieldOptions_WithOptionsCachesAndShowsModal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40

	msg := customFieldOptionsMsg{
		issueKey:    testKey,
		fieldID:     "customfield_10001",
		fieldName:   "Type",
		fieldType:   views.FieldSingleSelect,
		options:     []jira.CreateMetaValue{{ID: "1", Name: "Bug"}, {ID: "2", Name: "Story"}},
		allFields:   []jira.CreateMetaField{{FieldID: "customfield_10001"}},
		issueTypeID: "10000",
		projectKey:  testProject,
	}

	_, _ = app.handleCustomFieldOptions(msg)

	if _, ok := app.createMetaCache[testProject+":10000"]; !ok {
		t.Error("createMetaCache should be populated when allFields provided")
	}
}

func TestHandleCustomFieldOptions_PersonSchemaFetchesUsers(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetUsersFunc = func(_ context.Context, _ string) ([]jira.User, error) {
		return []jira.User{{AccountID: "1", DisplayName: "Alice"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.projectKey = testProject

	msg := customFieldOptionsMsg{
		issueKey:    testKey,
		fieldID:     "customfield_10001",
		fieldName:   "Approver",
		fieldType:   views.FieldPerson,
		schemaType:  schemaUser,
		schemaItems: "",
		options:     []jira.CreateMetaValue{{ID: "1", Name: "Alice"}},
	}

	_, cmd := app.handleCustomFieldOptions(msg)

	if cmd == nil {
		t.Error("expected users fetch cmd for person schema field")
	}
}

func TestMakeFieldSelectCallback_BuildsUpdateCmd(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.UpdateIssueFunc = func(_ context.Context, _ string, _ map[string]any) error {
		return nil
	}
	app := newAppWithFake(t, fake)
	app.issueCache[testKey] = &jira.Issue{Key: testKey}

	cb := app.makeFieldSelectCallback(testKey, fldPriority)
	cmd := cb(components.ModalItem{ID: "1", Label: "High"})

	if cmd == nil {
		t.Error("expected update cmd from field select callback")
	}
}

func TestHandleGitBranchSwitch_UpdatesBranch(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleGitBranchSwitch("feature/new-branch")

	testkit.AssertEqual(t, "gitBranch", app.gitBranch, "feature/new-branch")
}

func TestHandleGitBranchSwitch_QuitsWhenConfigured(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.cfg.Git.CloseOnCheckout = true

	_, cmd := app.handleGitBranchSwitch("feature/new-branch")

	if cmd == nil {
		t.Error("expected tea.Quit cmd when CloseOnCheckout is true")
	}
}

func TestFetchCustomFieldOptionsForEdit_NoIssueTypeErrors(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40

	issue := &jira.Issue{Key: testKey, Summary: testSummary}
	field := &views.InfoField{
		FieldID: "customfield_10001",
		Name:    "Sprint",
		Type:    views.FieldSingleSelect,
	}

	_, _ = app.fetchCustomFieldOptionsForEdit(issue, field)

	if app.statusPanel.ErrorMessage() == "" {
		t.Error("expected status panel error when issue has no issue type")
	}
}

func TestFetchCustomFieldOptionsForEdit_TextTypeShowsModal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.cfg.Fields = []config.FieldConfig{
		{ID: "customfield_10001", Name: "Notes", Type: "text"},
	}

	issue := &jira.Issue{
		Key:       testKey,
		Summary:   testSummary,
		IssueType: &jira.IssueType{ID: "10000", Name: "Story"},
	}
	field := &views.InfoField{
		FieldID: "customfield_10001",
		Name:    "Notes",
		Type:    views.FieldSingleText,
	}

	_, _ = app.fetchCustomFieldOptionsForEdit(issue, field)

	if !app.inputModal.IsVisible() {
		t.Error("expected input modal to be visible for text type custom field")
	}
}

func TestFetchCustomFieldOptionsForEdit_CachedMetaUsed(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.projectKey = testProject

	issue := &jira.Issue{
		Key:       testKey,
		Summary:   testSummary,
		IssueType: &jira.IssueType{ID: "10000", Name: "Story"},
	}
	app.createMetaCache[testProject+":10000"] = []jira.CreateMetaField{
		{
			FieldID:       "customfield_10001",
			AllowedValues: []jira.CreateMetaValue{{ID: "1", Name: "Option A"}},
		},
	}
	field := &views.InfoField{
		FieldID: "customfield_10001",
		Name:    "Category",
		Type:    views.FieldSingleSelect,
	}

	_, _ = app.fetchCustomFieldOptionsForEdit(issue, field)

	if len(fake.GetCreateMetaCalls) != 0 {
		t.Errorf("GetCreateMeta called %d times, want 0 (cache hit)", len(fake.GetCreateMetaCalls))
	}
}
