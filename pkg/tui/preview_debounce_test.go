package tui

import (
	"context"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestPreviewDebounce_RapidMovement(t *testing.T) {
	t.Parallel()
	const key2 = "SUB-2"
	key1 := subKey1

	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, key string) (*jira.Issue, error) {
		if key != key2 {
			t.Errorf("unexpected GetIssue call for key %q (expected %q only)", key, key2)
		}
		return &jira.Issue{Key: key}, nil
	}
	fake.GetCommentsFunc = func(_ context.Context, _ string) ([]jira.Comment, error) {
		return nil, nil
	}
	fake.GetChangelogFunc = func(_ context.Context, _ string) ([]jira.ChangelogEntry, error) {
		return nil, nil
	}

	app := newAppWithFake(t, fake)

	_, _ = app.Update(views.PreviewRequestMsg{Key: key1})
	_, _ = app.Update(views.PreviewRequestMsg{Key: key2})

	if app.previewEpoch != 2 {
		t.Fatalf("previewEpoch = %d after two msgs, want 2", app.previewEpoch)
	}

	_, fetchCmd := app.Update(previewDebounceMsg{key: key1, epoch: 1})
	if fetchCmd != nil {
		fetchCmd()
		if len(fake.GetIssueCalls) > 0 {
			t.Errorf("stale debounce tick caused %d GetIssue call(s), want 0", len(fake.GetIssueCalls))
		}
	}

	before := len(fake.GetIssueCalls)

	_, fetchCmd2 := app.Update(previewDebounceMsg{key: key2, epoch: 2})
	if fetchCmd2 == nil {
		t.Fatal("expected fetch cmd from fresh debounce tick, got nil")
	}
	fetchCmd2()

	after := len(fake.GetIssueCalls)
	if got := after - before; got != 1 {
		t.Errorf("fresh debounce tick caused %d GetIssue call(s), want 1", got)
	}
	if after > 0 && fake.GetIssueCalls[after-1].Key != key2 {
		t.Errorf("GetIssue called with key %q, want %q", fake.GetIssueCalls[after-1].Key, key2)
	}
}

func TestPreviewDebounce_Lapse(t *testing.T) {
	t.Parallel()
	const key2 = "SUB-2"
	key1 := subKey1

	var issueCalls []string
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, key string) (*jira.Issue, error) {
		issueCalls = append(issueCalls, key)
		return &jira.Issue{Key: key}, nil
	}
	fake.GetCommentsFunc = func(_ context.Context, _ string) ([]jira.Comment, error) {
		return nil, nil
	}
	fake.GetChangelogFunc = func(_ context.Context, _ string) ([]jira.ChangelogEntry, error) {
		return nil, nil
	}

	app := newAppWithFake(t, fake)

	_, _ = app.Update(views.PreviewRequestMsg{Key: key1})

	_, fetchCmd1 := app.Update(previewDebounceMsg{key: key1, epoch: 1})
	if fetchCmd1 == nil {
		t.Fatal("expected fetch cmd from first debounce tick, got nil")
	}
	fetchCmd1()

	_, _ = app.Update(views.PreviewRequestMsg{Key: key2})

	_, fetchCmd2 := app.Update(previewDebounceMsg{key: key2, epoch: 2})
	if fetchCmd2 == nil {
		t.Fatal("expected fetch cmd from second debounce tick, got nil")
	}
	fetchCmd2()

	if len(issueCalls) != 2 {
		t.Fatalf("expected 2 GetIssue calls, got %d: %v", len(issueCalls), issueCalls)
	}
	if issueCalls[0] != key1 {
		t.Errorf("first GetIssue call key = %q, want %q", issueCalls[0], key1)
	}
	if issueCalls[1] != key2 {
		t.Errorf("second GetIssue call key = %q, want %q", issueCalls[1], key2)
	}
}

func TestPreviewStaleResponse_DroppedWhenEpochAdvanced(t *testing.T) {
	t.Parallel()
	const key2 = "SUB-2"
	key1 := subKey1

	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: key2, Summary: "current"})

	app := newAppWithFake(t, fake)
	main := &jira.Issue{Key: mainKey, Summary: "main"}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.infoPanel.SetIssue(main)

	app.previewKey = key2
	app.previewEpoch = 2

	_, _ = app.Update(previewDetailLoadedMsg{
		issue: &jira.Issue{Key: key1, Summary: "stale"},
		epoch: 1,
	})

	if got := app.infoPanel.IssueKey(); got != mainKey {
		t.Errorf("infoPanel.IssueKey() = %q, want %q (must stay with main)", got, mainKey)
	}
	if got := app.detailView.IssueKey(); got == key1 {
		t.Errorf("detailView updated with stale key %q, want no update", key1)
	}
	if _, ok := app.issueCache[key1]; ok {
		t.Errorf("issueCache populated with stale key %q, want absent", key1)
	}

	_, _ = app.Update(previewDetailLoadedMsg{
		issue: &jira.Issue{Key: key2, Summary: "current"},
		epoch: 2,
	})

	if got := app.detailView.IssueKey(); got != key2 {
		t.Errorf("detailView.IssueKey() = %q after fresh response, want %q", got, key2)
	}
	if _, ok := app.issueCache[key2]; !ok {
		t.Errorf("issueCache missing key %q after fresh response", key2)
	}
}

func TestNonPreviewFetch_NotAffectedByDebounce(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: mainKey, Summary: "main"})

	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{Key: mainKey}})
	app.previewKey = mainKey
	app.previewEpoch = 99

	_, cmd, handled := app.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if cmd == nil {
		t.Fatal("expected cmd from ActRefresh, got nil")
	}
	msg := cmd()

	loaded, ok := msg.(issueDetailLoadedMsg)
	if !ok {
		t.Fatalf("ActRefresh produced %T, want issueDetailLoadedMsg", msg)
	}
	if loaded.issue == nil || loaded.issue.Key != mainKey {
		t.Errorf("issueDetailLoadedMsg.issue.Key = %q, want %q", loaded.issue.Key, mainKey)
	}

	_, _ = app.handleIssueDetailLoaded(loaded)
	if got := app.infoPanel.IssueKey(); got != mainKey {
		t.Errorf("infoPanel.IssueKey() = %q after ActRefresh, want %q", got, mainKey)
	}
}
