package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func setupInfoFocused(t *testing.T, tab views.InfoPanelTab) *App {
	t.Helper()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	mainIssue := jira.Issue{
		Key:     mainKey,
		Summary: "main issue",
		Subtasks: []jira.Issue{
			{Key: subKey1, Summary: "sub one"},
		},
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "Blocks", Outward: "blocks", Inward: "is blocked by"},
				OutwardIssue: &jira.Issue{Key: "LNK-1", Summary: "linked"},
			},
		},
	}
	app.issuesList.SetIssues([]jira.Issue{mainIssue})
	app.infoPanel.SetIssue(&mainIssue)

	for app.infoPanel.ActiveTab() != tab {
		app.infoPanel.NextTab()
	}

	app.previewKey = subKey1

	app.side = sideLeft
	app.leftFocus = focusInfo
	app.infoPanel.Focused = true
	app.keymap = DefaultKeymap()
	app.infoPanel.ResolveNav = DefaultKeymap().MatchNav
	return app
}

func TestEscFromSubTab_ResetsPreviewToMainIssue(t *testing.T) {
	t.Parallel()
	app := setupInfoFocused(t, views.InfoTabSubtasks)
	if app.previewKey != subKey1 {
		t.Fatalf("precondition: previewKey = %q, want SUB-1", app.previewKey)
	}

	app.handleFocusAction(ActFocusLeft)

	if got := app.previewKey; got != mainKey {
		t.Errorf("previewKey = %q after FocusLeft from Sub tab, want %q", got, mainKey)
	}
}

func TestEscFromLnkTab_ResetsPreviewToMainIssue(t *testing.T) {
	t.Parallel()
	app := setupInfoFocused(t, views.InfoTabLinks)
	app.previewKey = "LNK-1"

	app.handleFocusAction(ActFocusLeft)

	if got := app.previewKey; got != mainKey {
		t.Errorf("previewKey = %q after FocusLeft from Lnk tab, want %q", got, mainKey)
	}
}

func TestInfoPanelTabSwitchToFields_ResetsPreviewKey(t *testing.T) {
	t.Parallel()
	app := setupInfoFocused(t, views.InfoTabSubtasks)
	app.previewKey = subKey1

	app.handleTabAction(ActPrevTab)
	app.handleTabAction(ActPrevTab)

	if got := app.previewKey; got != mainKey {
		t.Errorf("previewKey = %q after switching to Fields tab, want %q", got, mainKey)
	}
}

func TestEmptySubList_NoPreviewDispatch(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	issueNoSubs := &jira.Issue{Key: mainKey, Summary: "no subs"}
	app.issuesList.SetIssues([]jira.Issue{*issueNoSubs})
	app.infoPanel.SetIssue(issueNoSubs)
	app.previewKey = mainKey
	app.side = sideLeft
	app.leftFocus = focusInfo
	app.infoPanel.Focused = true
	app.keymap = DefaultKeymap()

	if app.infoPanel.ActiveTab() != views.InfoTabFields {
		t.Fatal("precondition: expected InfoTabFields as default tab")
	}

	app.handleTabAction(ActNextTab)
	app.handleTabAction(ActNextTab)

	if app.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		t.Fatal("expected InfoTabSubtasks after two NextTab actions")
	}

	navKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	app.infoPanel.ResolveNav = DefaultKeymap().MatchNav
	_, cmd := app.infoPanel.Update(navKey)
	if cmd == nil {
		return
	}
	msg := cmd()
	if _, ok := msg.(views.PreviewRequestMsg); ok {
		t.Error("empty Sub list must not dispatch PreviewRequestMsg on cursor move")
	}
}
