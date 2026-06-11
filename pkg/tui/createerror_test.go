package tui

import (
	"errors"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestFormatCreateError_SubtaskWithAPIMessages(t *testing.T) {
	t.Parallel()

	err := &jira.APIError{
		Method:   "GET",
		Path:     "/issue/createmeta/DSOTEST/issuetypes/10054",
		Status:   404,
		Body:     `{"errorMessages":["You cannot create issues in this project."]}`,
		Messages: []string{"You cannot create issues in this project."},
	}

	got := formatCreateError(err, "DSOTEST", true)
	want := "Cannot create subtask in DSOTEST: you cannot create issues in this project"
	if got != want {
		t.Errorf("formatCreateError() = %q, want %q", got, want)
	}
}

func TestFormatCreateError_IssueNoun(t *testing.T) {
	t.Parallel()

	err := &jira.APIError{Status: 404, Messages: []string{"Boom happened."}}

	got := formatCreateError(err, "PLAT", false)
	want := "Cannot create issue in PLAT: boom happened"
	if got != want {
		t.Errorf("formatCreateError() = %q, want %q", got, want)
	}
}

func TestFormatCreateError_NonAPIErrorFallsBackToRaw(t *testing.T) {
	t.Parallel()

	err := errors.New("connection refused")

	got := formatCreateError(err, "PLAT", true)
	if got != "connection refused" {
		t.Errorf("formatCreateError() = %q, want raw error", got)
	}
}

func TestHandleCreatePreFormError_AbortsWithoutEmptyForm(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.createCtx = createCtx{projectKey: "DSOTEST", parentKey: testKey}
	app.createForm.SetLoading(true) // form is visible-loading before createmeta resolves
	if !app.createForm.IsVisible() {
		t.Fatal("precondition: loading form should be visible")
	}

	apiErr := &jira.APIError{
		Status:   404,
		Messages: []string{"You cannot create issues in this project."},
	}
	_, _ = app.handleCreatePreFormError(createPreFormErrorMsg{err: apiErr})

	if app.createForm.IsVisible() {
		t.Error("pre-form error must hide the form, not resume an empty one")
	}
	if app.createCtx.parentKey != "" {
		t.Errorf("createCtx should be cleared, got %+v", app.createCtx)
	}
	want := "Cannot create subtask in DSOTEST: you cannot create issues in this project"
	if got := app.statusPanel.ErrorMessage(); got != want {
		t.Errorf("status error = %q, want %q", got, want)
	}
}
