package tui

import (
	"context"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

const mainKey = "ABC-1"
const subKey1 = "SUB-1"

func newAppWithFake(t *testing.T, fake *jiratest.FakeClient) *App {
	t.Helper()
	app := newTestApp()
	app.client = fake
	logFlag := false
	app.logFlag = &logFlag
	app.infoPanel = views.NewInfoPanel()
	app.statusPanel = views.NewStatusPanel("", "", "")
	app.logPanel = views.NewLogPanel()
	app.issueCache = map[string]*jira.Issue{}
	app.childrenCache = map[string][]jira.Issue{}
	app.usersCache = map[string][]jira.User{}
	app.createMetaCache = map[string][]jira.CreateMetaField{}
	return app
}

func stubFullIssueFetch(fake *jiratest.FakeClient, issue *jira.Issue) {
	fake.GetIssueFunc = func(_ context.Context, _ string) (*jira.Issue, error) {
		return issue, nil
	}
	fake.GetCommentsFunc = func(_ context.Context, _ string) ([]jira.Comment, error) {
		return nil, nil
	}
	fake.GetChangelogFunc = func(_ context.Context, _ string) ([]jira.ChangelogEntry, error) {
		return nil, nil
	}
}

func TestActRefresh_FetchesPreviewedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: mainKey, Summary: "updated"})

	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})
	app.previewKey = mainKey

	_, cmd, handled := app.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if cmd == nil {
		t.Fatal("expected tea.Cmd, got nil")
	}
	msg := cmd()

	if len(fake.GetIssueCalls) != 1 {
		t.Fatalf("expected 1 GetIssue call, got %d: %+v", len(fake.GetIssueCalls), fake.GetIssueCalls)
	}
	if got := fake.GetIssueCalls[0].Key; got != mainKey {
		t.Errorf("GetIssue called with key %q, want %q", got, mainKey)
	}

	loaded, ok := msg.(issueDetailLoadedMsg)
	if !ok {
		t.Fatalf("expected issueDetailLoadedMsg, got %T", msg)
	}
	if loaded.issue == nil || loaded.issue.Key != mainKey {
		t.Errorf("loaded.issue = %+v, want Key=ABC-1", loaded.issue)
	}
}

func TestIssueSelectedMsg_UpdatesPreviewKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	_, _ = app.Update(views.IssueSelectedMsg{Issue: &jira.Issue{Key: mainKey}})

	if got := app.previewKey; got != mainKey {
		t.Errorf("previewKey = %q, want %q", got, mainKey)
	}
}

func TestPreviewSelectedIssue_UpdatesPreviewKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: "XYZ-9"}})

	app.previewSelectedIssue()

	if got := app.previewKey; got != "XYZ-9" {
		t.Errorf("previewKey = %q, want %q", got, "XYZ-9")
	}
}

func TestHandleIssueDetailLoaded_RoutesByPreviewKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{Key: mainKey, Summary: "main"}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)
	app.previewKey = subKey1

	_, _ = app.handleIssueDetailLoaded(issueDetailLoadedMsg{
		issue: &jira.Issue{Key: subKey1, Summary: "fresh"},
	})

	if got := app.detailView.IssueKey(); got != subKey1 {
		t.Errorf("detailView.IssueKey() = %q, want %q (DetailView follows previewKey)", got, subKey1)
	}
	if got := app.infoPanel.IssueKey(); got != mainKey {
		t.Errorf("infoPanel.IssueKey() = %q, want %q (InfoPanel stays on main issue)", got, mainKey)
	}
}

func TestActRefresh_NoFetchWhenPreviewKeyEmpty(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})

	_, cmd, handled := app.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd (no fetch), got non-nil")
	}
	if len(fake.GetIssueCalls) != 0 {
		t.Errorf("expected 0 GetIssue calls, got %d: %+v", len(fake.GetIssueCalls), fake.GetIssueCalls)
	}
}

func TestActRefresh_UsesPreviewKey_WhenSet(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "ABC-2", Summary: "sub-item"})

	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})
	app.previewKey = "ABC-2"

	_, cmd, handled := app.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if cmd == nil {
		t.Fatal("expected tea.Cmd, got nil")
	}
	cmd()

	if len(fake.GetIssueCalls) != 1 {
		t.Fatalf("expected 1 GetIssue call, got %d: %+v", len(fake.GetIssueCalls), fake.GetIssueCalls)
	}
	if got := fake.GetIssueCalls[0].Key; got != "ABC-2" {
		t.Errorf("GetIssue called with key %q, want %q (preview key)", got, "ABC-2")
	}
}

func TestActRefresh_InvalidatesCacheBeforeFetch(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: mainKey, Summary: "fresh"})

	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})
	app.previewKey = mainKey
	stale := &jira.Issue{Key: mainKey, Summary: "stale"}
	app.issueCache[mainKey] = stale

	_, _, handled := app.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}

	if _, ok := app.issueCache[mainKey]; ok {
		t.Errorf("issueCache[%q] still present after ActRefresh; expected cache invalidation", mainKey)
	}
}
