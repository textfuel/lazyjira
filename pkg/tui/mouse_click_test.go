package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
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

	_, _ = app.handleMouse(tea.MouseMsg{
		Button: tea.MouseButtonWheelUp,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      3,
	})
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

	_, _ = app.mouseClick(panelIssues, 0, 5)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusIssues)
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

	_, _ = app.mouseClick(panelInfo, 0, 5)

	testkit.AssertEqual(t, "side", app.side, sideLeft)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusInfo)
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

	_, _ = app.mouseClick(panelDetail, 0, 50)

	testkit.AssertEqual(t, "side", app.side, sideRight)
}
