package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func appWithPanelDims(t *testing.T, width int) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.width = width
	app.height = 40
	app.panelSideW = 30
	app.panelStatusH = 2
	app.panelIssuesH = 5
	app.panelInfoH = 4
	app.panelProjectsH = 6
	app.panelDetailH = 20
	app.panelLogH = 5
	return app
}

func TestHitTest_Horizontal(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 120)

	tests := []struct {
		name     string
		x, y     int
		want     panelID
		wantRelY int
	}{
		{"status", 5, 0, panelStatus, 0},
		{"issues", 5, 3, panelIssues, 1},
		{"info", 5, 8, panelInfo, 1},
		{"projects", 5, 15, panelProjects, 4},
		{"detail", 40, 5, panelDetail, 5},
		{"log", 40, 25, panelLog, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel, relY := app.hitTest(tt.x, tt.y)
			testkit.AssertEqual(t, "panel", panel, tt.want)
			testkit.AssertEqual(t, "relY", relY, tt.wantRelY)
		})
	}
}

func TestHitTest_Vertical(t *testing.T) {
	t.Parallel()
	app := appWithPanelDims(t, 60)

	tests := []struct {
		name     string
		y        int
		want     panelID
		wantRelY int
	}{
		{"status", 0, panelStatus, 0},
		{"issues", 3, panelIssues, 1},
		{"info", 8, panelInfo, 1},
		{"projects", 13, panelProjects, 2},
		{"detail", 20, panelDetail, 3},
		{"log", 38, panelLog, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel, relY := app.hitTest(2, tt.y)
			testkit.AssertEqual(t, "panel", panel, tt.want)
			testkit.AssertEqual(t, "relY", relY, tt.wantRelY)
		})
	}
}

func TestMouseScroll_FocusesPanel(t *testing.T) {
	t.Parallel()

	t.Run("issues panel scroll takes focus", func(t *testing.T) {
		t.Parallel()
		app := appWithPanelDims(t, 120)
		app.keymap = DefaultKeymap()
		app.side = sideRight
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}, {Key: "PLAT-2"}})

		_, _ = app.mouseScroll(panelIssues, 3)

		if app.side != sideLeft || app.leftFocus != focusIssues {
			t.Errorf("focus = (%v,%v), want left/issues", app.side, app.leftFocus)
		}
	})

	t.Run("detail panel scroll switches to right side", func(t *testing.T) {
		t.Parallel()
		app := appWithPanelDims(t, 120)
		app.keymap = DefaultKeymap()
		app.side = sideLeft

		_, _ = app.mouseScroll(panelDetail, 3)

		if app.side != sideRight {
			t.Errorf("side = %v, want right", app.side)
		}
	})
}
