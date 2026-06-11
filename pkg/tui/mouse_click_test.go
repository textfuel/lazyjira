package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func mouseApp(t *testing.T) *App {
	t.Helper()
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	return app
}

func TestHandleMouse_WheelUpScrollsUp(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}, {Key: mainKey}, {Key: subKey1}})
	app.issuesList.Cursor = 1

	_, _ = app.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      3,
	})

	if app.issuesList.Cursor != 0 {
		t.Errorf("cursor = %d, want 0 after wheel up", app.issuesList.Cursor)
	}
}

func TestHandleMouse_WheelDownScrollsDown(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}, {Key: mainKey}, {Key: subKey1}})

	_, _ = app.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      3,
	})

	if app.issuesList.Cursor != 1 {
		t.Errorf("cursor = %d, want 1 after wheel down", app.issuesList.Cursor)
	}
}

func TestHandleMouse_LeftClickFocusesPanel(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideRight

	_, _ = app.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      0,
	})

	testkit.AssertEqual(t, "side after click on status", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus after status click", app.leftFocus, focusStatus)
}

func TestHandleMouse_MotionIsNoop(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, cmd := app.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonNone,
		Action: tea.MouseActionMotion,
		X:      5,
		Y:      3,
	})

	if cmd != nil {
		t.Error("mouse motion should produce no cmd")
	}
}

func TestMouseClick_StatusFocusesStatus(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideRight

	_, _ = app.mouseClick(panelStatus, 0, 5)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusStatus)
}

func TestMouseClick_IssuesTitleBarTabClick(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := appWithPanelDims(t, 120)
	app.keymap = DefaultKeymap()
	app.client = fake
	app.projectKey = testProject
	app.issuesList.SetTabs([]config.IssueTabConfig{
		{Name: "All", JQL: "project = X"},
		{Name: "Mine", JQL: "assignee = currentUser()"},
	})

	_, _ = app.mouseClick(panelIssues, 0, 15)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusIssues)
	testkit.AssertEqual(t, "active tab index", app.issuesList.GetTabIndex(), 1)
}

func TestMouseClick_InfoFocusesInfo(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideRight

	_, _ = app.mouseClick(panelInfo, 1, 5)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusInfo)
}

func TestMouseClick_InfoTitleBarClicksTab(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.infoPanel.SetIssue(&jira.Issue{Key: testKey})

	_, _ = app.mouseClick(panelInfo, 0, 12)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusInfo)
	testkit.AssertEqual(t, "active info tab", app.infoPanel.ActiveTab(), views.InfoTabLinks)
}

func TestMouseClick_ProjectsFocusesProjects(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideRight

	_, _ = app.mouseClick(panelProjects, 1, 5)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusProjects)
}

func TestMouseClick_DetailFocusesDetail(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideLeft

	_, _ = app.mouseClick(panelDetail, 1, 40)

	testkit.AssertEqual(t, "side", app.side, sideRight)
}

func TestMouseClick_DetailTitleBarClicksTab(t *testing.T) {
	t.Parallel()
	app := mouseApp(t)
	app.side = sideLeft
	app.detailView.SetIssue(&jira.Issue{Key: testKey, Comments: []jira.Comment{{ID: "1", Body: "hi"}}})
	app.layoutPanels()

	separatorWidth := 3
	tabsStart := len("[0] "+testKey) + separatorWidth
	commentsTabX := app.panelSideW + tabsStart + len("Body") + separatorWidth

	_, _ = app.mouseClick(panelDetail, 0, commentsTabX)

	testkit.AssertEqual(t, "side", app.side, sideRight)
	testkit.AssertEqual(t, "active detail tab", app.detailView.ActiveTab(), views.TabComments)
}
