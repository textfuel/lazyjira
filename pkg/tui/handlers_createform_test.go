package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func TestHandleCreateFormEditExternal_NilFieldIsNoop(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, cmd := app.handleCreateFormEditExternal(components.CreateFormEditExternalMsg{FieldIndex: 99})

	if cmd != nil {
		t.Error("expected nil cmd with nil field")
	}
}

func TestHandleCreateFormEditExternal_LaunchesEditorForDescription(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.projectKey = testProject
	app.createCtx = createCtx{
		intent:     true,
		projectKey: testProject,
	}
	form := components.NewCreateForm(nil)
	form.ShowForm([]components.CreateFormField{
		{
			FieldID:      fldDescription,
			Name:         "Description",
			Type:         components.CFFieldMultiText,
			DisplayValue: "initial description",
		},
	}, "Story", testProject)
	app.createForm = form

	_, cmd := app.handleCreateFormEditExternal(components.CreateFormEditExternalMsg{FieldIndex: 0})

	if cmd == nil {
		t.Error("expected editor launch cmd for description field")
	}
}
