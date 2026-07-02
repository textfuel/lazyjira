package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func TestHandleSearchChanged_FiltersFocusedPanel(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: "alpha"}, {Key: "PLAT-2", Summary: "beta"}})
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, _ = app.handleSearchChanged(components.SearchChangedMsg{Query: "alpha"})

	if sel := app.issuesList.SelectedIssue(); sel == nil || sel.Key != testKey {
		t.Errorf("filtered selection = %v, want %s", sel, testKey)
	}
}

func TestHandleSearchConfirmed_IssuesSelectsAndFetches(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, cmd := app.handleSearchConfirmed()

	if cmd == nil {
		t.Error("expected a fetch command for the selected issue")
	}
}

func TestHandleSearchConfirmed_UpdatesPreviewToLandedIssue(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: "alpha"}, {Key: "PLAT-2", Summary: "beta"}})
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.previewKey = testKey

	app.issuesList.SetFilter("beta")
	_, _ = app.handleSearchConfirmed()

	if selected := app.issuesList.SelectedIssue(); selected == nil || selected.Key != "PLAT-2" {
		t.Fatalf("cursor landed on %v, want PLAT-2", selected)
	}
	if current := app.currentIssue(); current == nil || current.Key != "PLAT-2" {
		t.Errorf("currentIssue() = %v, want PLAT-2 so actions target the searched ticket", current)
	}
}

func TestHandleSearchConfirmed_ProjectsSwitches(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "1"}})
	app.side = sideLeft
	app.leftFocus = focusProjects

	_, _ = app.handleSearchConfirmed()

	if app.projectKey != testProject {
		t.Errorf("projectKey = %q, want %s after confirming project search", app.projectKey, testProject)
	}
}

func TestHandleSearchCancelled_ClearsFilters(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: "alpha"}, {Key: "PLAT-2", Summary: "beta"}})
	app.issuesList.SetFilter("alpha")

	_, cmd := app.handleSearchCancelled()

	if cmd != nil {
		t.Error("expected nil cmd")
	}
}

func TestHandleAutoFetch_SchedulesNextTick(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})

	_, cmd := app.handleAutoFetch()

	if cmd == nil {
		t.Error("auto fetch should always schedule the next tick")
	}
}

func TestRouteToPanel_ForwardsToFocusedPanel(t *testing.T) {
	t.Parallel()

	t.Run("left issues panel receives input", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.keymap = DefaultKeymap()
		app.issuesList.ResolveNav = app.keymap.MatchNav
		app.issuesList.SetFocused(true)
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}, {Key: "PLAT-2"}})
		app.side = sideLeft
		app.leftFocus = focusIssues

		_ = app.routeToPanel(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

		if sel := app.issuesList.SelectedIssue(); sel == nil || sel.Key != "PLAT-2" {
			t.Errorf("expected cursor to move down to PLAT-2, got %v", sel)
		}
	})

	t.Run("right detail panel receives input", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.side = sideRight

		app.routeToPanel(tea.KeyMsg{Type: tea.KeyDown})

		if app.side != sideRight {
			t.Errorf("side = %v, want right after routing to detail panel", app.side)
		}
	})
}
