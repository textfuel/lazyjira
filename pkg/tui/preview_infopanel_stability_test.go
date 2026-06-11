package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestPreviewDetailLoaded_DoesNotMutateInfoPanel(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	main := &jira.Issue{Key: mainKey, Summary: "main"}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)

	app.previewKey = subKey1
	app.previewEpoch = 1

	_, _ = app.Update(previewDetailLoadedMsg{
		issue: &jira.Issue{Key: subKey1, Summary: "sub"},
		epoch: 1,
	})

	if got := app.infoPanel.IssueKey(); got != mainKey {
		t.Errorf("infoPanel.IssueKey() = %q after preview response, want %q (InfoPanel must stay on main issue)", got, mainKey)
	}
}

func TestIssueDetailLoaded_DoesNotMutateInfoPanelForSubPreview(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	main := &jira.Issue{Key: mainKey, Summary: "main"}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)

	app.previewKey = subKey1

	_, _ = app.handleIssueDetailLoaded(issueDetailLoadedMsg{
		issue: &jira.Issue{Key: subKey1, Summary: "sub-fresh"},
	})

	if got := app.infoPanel.IssueKey(); got != mainKey {
		t.Errorf("infoPanel.IssueKey() = %q after sub-issue refresh, want %q", got, mainKey)
	}
}

func TestShowCachedIssue_DoesNotMutateInfoPanelForForeignKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{Key: mainKey, Summary: "main"}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)
	app.issueCache[subKey1] = &jira.Issue{Key: subKey1, Summary: "sub"}

	app.showCachedIssue(subKey1)

	if got := app.infoPanel.IssueKey(); got != mainKey {
		t.Errorf("infoPanel.IssueKey() = %q after showCachedIssue(%q), want %q",
			got, subKey1, mainKey)
	}
}

func TestPreviewRequestMsg_CacheHit_UpdatesDetailViewImmediately(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	cached := &jira.Issue{Key: subKey1, Summary: "cached sub"}
	app.issueCache[subKey1] = cached

	_, _ = app.Update(views.PreviewRequestMsg{Key: subKey1})

	if got := app.detailView.IssueKey(); got != subKey1 {
		t.Errorf("detailView.IssueKey() = %q, want %q (cache hit should update synchronously)", got, subKey1)
	}
	if app.previewKey != subKey1 {
		t.Errorf("previewKey = %q, want %q", app.previewKey, subKey1)
	}
}

func TestTabSwitchToSubtasks_DispatchesPreviewRequest(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{
		Key:      mainKey,
		Subtasks: []jira.Issue{{Key: subKey1}},
	}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)
	app.side = sideLeft
	app.leftFocus = focusInfo

	_, _, _ = app.handleTabAction(ActNextTab)
	_, cmd, handled := app.handleTabAction(ActNextTab)
	if !handled {
		t.Fatal("ActNextTab not handled")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd on sub-tab entry, got nil")
	}
	msg := cmd()
	pr, ok := msg.(views.PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if pr.Key != subKey1 {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", pr.Key, subKey1)
	}
}

func TestTabSwitchToLinks_DispatchesPreviewRequest(t *testing.T) {
	t.Parallel()
	const linkKey = "LNK-1"

	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{
		Key: mainKey,
		IssueLinks: []jira.IssueLink{{
			Type:         &jira.IssueLinkType{Name: "relates to"},
			OutwardIssue: &jira.Issue{Key: linkKey},
		}},
	}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)
	app.side = sideLeft
	app.leftFocus = focusInfo

	_, cmd, handled := app.handleTabAction(ActNextTab)
	if !handled {
		t.Fatal("ActNextTab not handled")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd on link-tab entry, got nil")
	}
	msg := cmd()
	pr, ok := msg.(views.PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if pr.Key != linkKey {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", pr.Key, linkKey)
	}
}

func TestTabSwitchToSubtasks_EmptyListNoDispatch(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{Key: mainKey}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)
	app.side = sideLeft
	app.leftFocus = focusInfo

	_, cmd1, _ := app.handleTabAction(ActNextTab)
	if cmd1 != nil {
		if msg := cmd1(); msg != nil {
			if _, ok := msg.(views.PreviewRequestMsg); ok {
				t.Errorf("empty Links tab dispatched PreviewRequestMsg, want no dispatch")
			}
		}
	}
	_, cmd2, _ := app.handleTabAction(ActNextTab)
	if cmd2 != nil {
		if msg := cmd2(); msg != nil {
			if _, ok := msg.(views.PreviewRequestMsg); ok {
				t.Errorf("empty Subtasks tab dispatched PreviewRequestMsg, want no dispatch")
			}
		}
	}
}
