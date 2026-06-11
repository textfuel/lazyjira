package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func navResolverJK(key string) components.NavAction {
	switch key {
	case "j":
		return components.NavDown
	case "k":
		return components.NavUp
	}
	return components.NavNone
}

func TestPreviewFollowsCursor_IssuesList_Down(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	issues := []jira.Issue{{Key: mainKey}, {Key: "ABC-2"}}
	app.issuesList.SetIssues(issues)

	_, cmd := app.Update(views.IssueSelectedMsg{Issue: &issues[1]})

	if app.previewKey != "ABC-2" {
		t.Errorf("previewKey = %q, want %q", app.previewKey, "ABC-2")
	}
	if app.previewEpoch != 1 {
		t.Errorf("previewEpoch = %d, want 1 (IssueSelectedMsg must delegate to PreviewRequestMsg)", app.previewEpoch)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd (debounce tick), got nil")
	}
}

func TestPreviewFollowsCursor_IssuesList_Up(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	issues := []jira.Issue{{Key: mainKey}, {Key: "ABC-2"}}
	app.issuesList.SetIssues(issues)

	_, _ = app.Update(views.IssueSelectedMsg{Issue: &issues[1]})
	_, _ = app.Update(views.IssueSelectedMsg{Issue: &issues[0]})

	if app.previewKey != mainKey {
		t.Errorf("previewKey = %q, want %q", app.previewKey, mainKey)
	}
	if app.previewEpoch != 2 {
		t.Errorf("previewEpoch = %d, want 2", app.previewEpoch)
	}
}

func TestPreviewFollowsCursor_InfoSubtasks_ExistingPath(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{
		Key:      mainKey,
		Subtasks: []jira.Issue{{Key: "SUB-1"}, {Key: "SUB-2"}},
	}
	app.infoPanel.SetIssue(main)
	app.infoPanel.SetActiveTab(views.InfoTabSubtasks)
	app.infoPanel.SetFocused(true)
	app.infoPanel.ResolveNav = navResolverJK

	_, cmd := app.infoPanel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd == nil {
		t.Fatal("expected PreviewRequestMsg cmd from Subtasks cursor move, got nil")
	}
	msg := cmd()
	pr, ok := msg.(views.PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if pr.Key != "SUB-2" {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", pr.Key, "SUB-2")
	}
}

func TestPreviewFollowsCursor_InfoLinks_ExistingPath(t *testing.T) {
	t.Parallel()
	const linkKey1 = "LNK-1"
	const linkKey2 = "LNK-2"

	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{
		Key: mainKey,
		IssueLinks: []jira.IssueLink{
			{Type: &jira.IssueLinkType{Name: "relates to"}, OutwardIssue: &jira.Issue{Key: linkKey1}},
			{Type: &jira.IssueLinkType{Name: "relates to"}, OutwardIssue: &jira.Issue{Key: linkKey2}},
		},
	}
	app.infoPanel.SetIssue(main)
	app.infoPanel.SetActiveTab(views.InfoTabLinks)
	app.infoPanel.SetFocused(true)
	app.infoPanel.ResolveNav = navResolverJK

	_, cmd := app.infoPanel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd == nil {
		t.Fatal("expected PreviewRequestMsg cmd from Links cursor move, got nil")
	}
	msg := cmd()
	pr, ok := msg.(views.PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if pr.Key != linkKey2 {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", pr.Key, linkKey2)
	}
}

func TestPreviewFollowsCursor_RapidCursor_OnlyLastFetch(t *testing.T) {
	t.Parallel()
	const lastKey = "ABC-5"

	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, key string) (*jira.Issue, error) {
		if key != lastKey {
			t.Errorf("unexpected GetIssue(%q); only %q should reach the client", key, lastKey)
		}
		return &jira.Issue{Key: key}, nil
	}
	fake.GetCommentsFunc = func(_ context.Context, _ string) ([]jira.Comment, error) { return nil, nil }
	fake.GetChangelogFunc = func(_ context.Context, _ string) ([]jira.ChangelogEntry, error) { return nil, nil }

	app := newAppWithFake(t, fake)
	issues := []jira.Issue{
		{Key: "ABC-1"}, {Key: "ABC-2"}, {Key: "ABC-3"}, {Key: "ABC-4"}, {Key: lastKey},
	}
	app.issuesList.SetIssues(issues)

	for i := range issues {
		_, _ = app.Update(views.IssueSelectedMsg{Issue: &issues[i]})
	}
	if app.previewEpoch != 5 {
		t.Fatalf("previewEpoch = %d after 5 moves, want 5", app.previewEpoch)
	}

	for epoch := 1; epoch <= 4; epoch++ {
		_, stale := app.Update(previewDebounceMsg{key: issues[epoch-1].Key, epoch: epoch})
		if stale != nil {
			stale()
		}
	}
	if len(fake.GetIssueCalls) != 0 {
		t.Errorf("stale debounce ticks caused %d GetIssue call(s), want 0", len(fake.GetIssueCalls))
	}

	_, fetchCmd := app.Update(previewDebounceMsg{key: lastKey, epoch: 5})
	if fetchCmd == nil {
		t.Fatal("expected fetch cmd from fresh debounce tick, got nil")
	}
	fetchCmd()
	if len(fake.GetIssueCalls) != 1 {
		t.Fatalf("expected 1 GetIssue call after fresh tick, got %d", len(fake.GetIssueCalls))
	}
	if got := fake.GetIssueCalls[0].Key; got != lastKey {
		t.Errorf("GetIssue called with key %q, want %q", got, lastKey)
	}
}

func TestPreviewFollowsCursor_Projects(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	projects := []jira.Project{{Key: "P1", Name: "Project One"}, {Key: "P2", Name: "Project Two"}}

	_, _ = app.Update(views.ProjectHoveredMsg{Project: &projects[1]})

	if got := app.detailView.Mode(); got != views.ModeProject {
		t.Errorf("detailView.Mode = %v, want ModeProject", got)
	}
	_, _ = app.Update(views.ProjectHoveredMsg{Project: nil})
	if got := app.detailView.Mode(); got != views.ModeProject {
		t.Errorf("nil hover changed mode away from ModeProject (got %v)", got)
	}
}

func TestPreviewFollowsCursor_UnknownKey_FallsBackToContext(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	main := &jira.Issue{Key: mainKey, Summary: "main issue"}
	app.issuesList.SetIssues([]jira.Issue{*main})
	app.detailView.SetIssue(main)

	_, _ = app.Update(views.PreviewRequestMsg{Key: "UNKNOWN-99"})

	if got := app.detailView.IssueKey(); got != mainKey {
		t.Errorf("DetailView.IssueKey = %q, want context fallback %q", got, mainKey)
	}
}
