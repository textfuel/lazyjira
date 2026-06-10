package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func appForKeybindings(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	return app
}

func TestContextBindings_ContainsQuit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "issues focus",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusIssues },
		},
		{
			name:  "info focus",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusInfo },
		},
		{
			name:  "projects focus",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusProjects },
		},
		{
			name:  "status focus",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusStatus },
		},
		{
			name:  "detail side",
			setup: func(a *App) { a.side = sideRight },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := appForKeybindings(t)
			tc.setup(app)
			bindings := app.ContextBindings()
			if len(bindings) == 0 {
				t.Fatal("expected non-empty bindings")
			}
			found := false
			for _, b := range bindings {
				if b.Description == string(ActQuit) {
					found = true
					break
				}
			}
			if !found {
				t.Error("quit binding missing from context bindings")
			}
		})
	}
}

func TestContextBindings_DetailCommentsIncludesEdit(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	app.side = sideRight
	app.detailView.SetIssue(&jira.Issue{Key: testKey})
	app.detailView.SetActiveTab(views.TabComments)

	bindings := app.ContextBindings()

	found := false
	for _, b := range bindings {
		if b.Description == "edit comment" {
			found = true
			break
		}
	}
	if !found {
		t.Error("edit comment binding missing when on comments tab")
	}
}

func TestHelpBarItems_NotEmpty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "issues panel",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusIssues },
		},
		{
			name:  "info panel",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusInfo },
		},
		{
			name:  "projects panel",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusProjects },
		},
		{
			name:  "status panel",
			setup: func(a *App) { a.side = sideLeft; a.leftFocus = focusStatus },
		},
		{
			name:  "detail right panel",
			setup: func(a *App) { a.side = sideRight },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := appForKeybindings(t)
			tc.setup(app)
			items := app.helpBarItems()
			if len(items) == 0 {
				t.Error("expected non-empty help bar items")
			}
		})
	}
}

func TestNavBindings_HasSixEntries(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	bindings := app.navBindings()
	if len(bindings) != 6 {
		t.Errorf("navBindings len = %d, want 6", len(bindings))
	}
}

func TestDetailScrollBindings_HasFourEntries(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	bindings := app.detailScrollBindings()
	if len(bindings) != 4 {
		t.Errorf("detailScrollBindings len = %d, want 4", len(bindings))
	}
}

func TestBind_ReturnsBindingWithDescription(t *testing.T) {
	t.Parallel()
	app := appForKeybindings(t)
	b := app.bind(ActQuit, "quit the app")
	if b.Description != "quit the app" {
		t.Errorf("description = %q, want %q", b.Description, "quit the app")
	}
	if b.Key == "" {
		t.Error("key should not be empty for ActQuit")
	}
}
