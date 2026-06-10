package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func TestHandleResize_SetsWidthAndHeight(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 140, Height: 50})

	testkit.AssertEqual(t, "width", app.width, 140)
	testkit.AssertEqual(t, "height", app.height, 50)
}

func TestHandleResize_PanelsLaidOut(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 120, Height: 40})

	if app.panelSideW == 0 {
		t.Error("panelSideW should be set after resize")
	}
	if app.panelDetailH == 0 {
		t.Error("panelDetailH should be set after resize")
	}
}

func TestHandleResize_VerticalLayoutOnNarrowTerminal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleResize(tea.WindowSizeMsg{Width: 60, Height: 40})

	if !app.isVerticalLayout() {
		t.Error("terminal narrower than 80 cols should use vertical layout")
	}
}
