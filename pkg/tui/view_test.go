package tui

import (
	"regexp"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func plain(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func appForView(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	app.updateFocusState()
	return app
}

func TestView_LoadingBeforeWidthSet(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})

	output := app.View()

	if output != "Loading..." {
		t.Errorf("expected Loading..., got %q", output)
	}
}

func TestView_HorizontalLayout(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: testSummary}})

	output := app.View()

	if output == "" {
		t.Fatal("View() returned empty string")
	}
	stripped := plain(output)
	if stripped == "" {
		t.Error("stripped output should not be empty")
	}
}

func TestView_VerticalLayout(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 60
	app.height = 40
	app.layoutPanels()
	app.updateFocusState()
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: testSummary}})

	output := app.View()

	if output == "" {
		t.Fatal("View() returned empty string")
	}
	if !app.isVerticalLayout() {
		t.Error("expected vertical layout for 60-col terminal")
	}
}

func TestView_ShowHelpOverlay(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.showHelp = true

	output := app.View()

	stripped := plain(output)
	if stripped == "" {
		t.Error("help overlay output should not be empty")
	}
}

func TestView_ShowHelpOverlay_WithFilter(t *testing.T) {
	t.Parallel()
	app := appForView(t)
	app.showHelp = true
	app.helpFilter = string(ActQuit)

	output := app.View()

	stripped := plain(output)
	if stripped == "" {
		t.Error("filtered help overlay output should not be empty")
	}
}

func TestView_NoPanic_AllFocusStates(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		setup func(*App)
	}{
		{
			name:  "left issues",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusIssues },
		},
		{
			name:  "left info",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusInfo },
		},
		{
			name:  "left projects",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusProjects },
		},
		{
			name:  "left status",
			setup: func(app *App) { app.side = sideLeft; app.leftFocus = focusStatus },
		},
		{
			name:  "right side",
			setup: func(app *App) { app.side = sideRight },
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := appForView(t)
			testCase.setup(app)
			app.updateFocusState()

			output := app.View()
			if output == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}
