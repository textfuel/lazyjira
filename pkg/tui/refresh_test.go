package tui

import (
	"context"
	"testing"

	"github.com/textfuel/lazyjira/pkg/jira"
	"github.com/textfuel/lazyjira/pkg/jira/jiratest"
)

// newAppWithFake augments newTestApp() with a FakeClient and a non-nil logFlag,
// the minimum needed for action handlers that fetch via the Jira client.
func newAppWithFake(t *testing.T, fake *jiratest.FakeClient) *App {
	t.Helper()
	a := newTestApp()
	a.client = fake
	logFlag := false
	a.logFlag = &logFlag
	return a
}

// stubFullIssueFetch wires the three methods that fetchFullIssue calls.
// Only GetIssue returns a real issue; Comments + Changelog are empty.
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

// TestActRefresh_FetchesSelectedIssueKey documents the current behavior:
// pressing the refresh action re-fetches the currently selected issue via
// GetIssue. This is a characterization test guarding the happy path.
func TestActRefresh_FetchesSelectedIssueKey(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "ABC-1", Summary: "updated"})

	a := newAppWithFake(t, fake)
	a.issuesList.SetIssues([]jira.Issue{{Key: "ABC-1"}})

	_, cmd, handled := a.handleIssueAction(ActRefresh)
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
	if got := fake.GetIssueCalls[0].Key; got != "ABC-1" {
		t.Errorf("GetIssue called with key %q, want %q", got, "ABC-1")
	}

	loaded, ok := msg.(issueDetailLoadedMsg)
	if !ok {
		t.Fatalf("expected issueDetailLoadedMsg, got %T", msg)
	}
	if loaded.issue == nil || loaded.issue.Key != "ABC-1" {
		t.Errorf("loaded.issue = %+v, want Key=ABC-1", loaded.issue)
	}
}

// TestActRefresh_UsesPreviewKey_WhenSet documents the desired behavior:
// when a preview (e.g. a sub-issue selected in the info panel) is active,
// the refresh action must re-fetch that preview's issue, not the main list
// selection.
func TestActRefresh_UsesPreviewKey_WhenSet(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "ABC-2", Summary: "sub-item"})

	a := newAppWithFake(t, fake)
	a.issuesList.SetIssues([]jira.Issue{{Key: "ABC-1"}})
	a.previewKey = "ABC-2"

	_, cmd, handled := a.handleIssueAction(ActRefresh)
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
