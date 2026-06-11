package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func jqlApp(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	return app
}

func TestHandleJQLSubmit_SetsLoadingAndReturnsCmd(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, _ string, _, _ int) (*jira.SearchResult, error) {
		return &jira.SearchResult{}, nil
	}
	app := jqlApp(t)
	app.client = fake

	_, cmd := app.handleJQLSubmit(components.JQLSubmitMsg{Query: "project = X"})

	if cmd == nil {
		t.Fatal("expected a fetch cmd")
	}
}

func TestHandleJQLSearchResult_AddsTabAndFocusesIssues(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	app := jqlApp(t)
	app.projectKey = testProject

	_, _ = app.handleJQLSearchResult(jqlSearchResultMsg{
		issues: []jira.Issue{{Key: testKey, Summary: testSummary}},
		jql:    "project = " + testProject,
	})

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusIssues)
	if !app.issuesList.IsJQLTab() {
		t.Error("JQL tab should be added after search result")
	}
}

func TestHandleJQLSearchError_ShowsErrorInModal(t *testing.T) {
	t.Parallel()
	app := jqlApp(t)
	app.jqlModal.Show("", nil)

	_, _ = app.handleJQLSearchError(jqlSearchErrorMsg{err: "bad jql"})

	if !app.jqlModal.IsVisible() {
		t.Error("modal should remain visible after error")
	}
	app.jqlModal.SetSize(80, 24)
	if view := app.jqlModal.View(); !strings.Contains(view, "bad jql") {
		t.Errorf("modal view should contain error text, got: %q", view)
	}
}

func TestHandleJQLFieldsLoaded_CachesFields(t *testing.T) {
	t.Parallel()
	app := jqlApp(t)
	fields := []jira.AutocompleteField{{Value: "summary"}, {Value: "assignee"}}

	_, _ = app.handleJQLFieldsLoaded(jqlFieldsLoadedMsg{fields: fields})

	if len(app.jqlFields) != 2 {
		t.Errorf("jqlFields len = %d, want 2", len(app.jqlFields))
	}
}

func TestHandleJQLSuggestions_UpdatesWhenVisible(t *testing.T) {
	t.Parallel()
	app := jqlApp(t)
	app.jqlModal.Show("", nil)
	app.jqlModal.SetSize(80, 24)

	suggestions := []jira.AutocompleteSuggestion{{Value: "open"}, {Value: "done"}}
	_, _ = app.handleJQLSuggestions(jqlSuggestionsMsg{suggestions: suggestions})

	if view := app.jqlModal.View(); !strings.Contains(view, "open") {
		t.Errorf("modal view should contain suggestion text after update, got: %q", view)
	}
}

func TestHandleJQLSuggestions_NoopWhenHidden(t *testing.T) {
	t.Parallel()
	app := jqlApp(t)

	_, cmd := app.handleJQLSuggestions(jqlSuggestionsMsg{
		suggestions: []jira.AutocompleteSuggestion{{Value: "open"}},
	})
	if cmd != nil {
		t.Error("expected nil cmd when modal is not visible")
	}
}

func TestHandleJQLInputChanged_FieldMode(t *testing.T) {
	t.Parallel()
	app := jqlApp(t)
	app.jqlModal.Show("", nil)
	app.jqlModal.SetSize(80, 24)
	app.jqlFields = []jira.AutocompleteField{{Value: "summary"}}

	_, _ = app.handleJQLInputChanged(components.JQLInputChangedMsg{Text: "sum", CursorPos: 3})

	if view := app.jqlModal.View(); !strings.Contains(view, "summary") {
		t.Errorf("modal view should contain suggestion after field-mode input change, got: %q", view)
	}
}

func TestHandleJQLInputChanged_ValueMode(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetJQLAutocompleteSuggestionsFunc = func(_ context.Context, _, _ string) ([]jira.AutocompleteSuggestion, error) {
		return nil, errors.New("network error")
	}
	app := jqlApp(t)
	app.client = fake

	_, cmd := app.handleJQLInputChanged(components.JQLInputChangedMsg{Text: "status = ", CursorPos: 9})

	if cmd == nil {
		t.Error("expected a suggestions fetch cmd in value mode")
	}
}

func TestHandleJQLInputChanged_DefaultMode(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())
	app := jqlApp(t)

	_, cmd := app.handleJQLInputChanged(components.JQLInputChangedMsg{Text: "", CursorPos: 0})

	if cmd != nil {
		t.Error("expected nil cmd in default mode")
	}
}
