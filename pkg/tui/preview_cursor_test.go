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

// navResolverJK resolves a minimal j/k keymap used to drive list cursor
// movement in tests without loading the full app keymap.
func navResolverJK(key string) components.NavAction {
	switch key {
	case "j":
		return components.NavDown
	case "k":
		return components.NavUp
	}
	return components.NavNone
}

// TestPreviewFollowsCursor_IssuesList_Down verifies that IssueSelectedMsg
// (emitted by the main list on a down-cursor move) delegates to the preview
// pipeline: previewKey follows the new selection and previewEpoch bumps so
// the debounce+cancel mechanics engage.
func TestPreviewFollowsCursor_IssuesList_Down(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	issues := []jira.Issue{{Key: mainKey}, {Key: "ABC-2"}}
	a.issuesList.SetIssues(issues)

	_, cmd := a.Update(views.IssueSelectedMsg{Issue: &issues[1]})

	if a.previewKey != "ABC-2" {
		t.Errorf("previewKey = %q, want %q", a.previewKey, "ABC-2")
	}
	if a.previewEpoch != 1 {
		t.Errorf("previewEpoch = %d, want 1 (IssueSelectedMsg must delegate to PreviewRequestMsg)", a.previewEpoch)
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd (debounce tick), got nil")
	}
}

// TestPreviewFollowsCursor_IssuesList_Up covers the symmetric up-cursor path
// and also pins that the epoch advances once per move.
func TestPreviewFollowsCursor_IssuesList_Up(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	issues := []jira.Issue{{Key: mainKey}, {Key: "ABC-2"}}
	a.issuesList.SetIssues(issues)

	_, _ = a.Update(views.IssueSelectedMsg{Issue: &issues[1]})
	_, _ = a.Update(views.IssueSelectedMsg{Issue: &issues[0]})

	if a.previewKey != mainKey {
		t.Errorf("previewKey = %q, want %q", a.previewKey, mainKey)
	}
	if a.previewEpoch != 2 {
		t.Errorf("previewEpoch = %d, want 2", a.previewEpoch)
	}
}

// TestPreviewFollowsCursor_InfoSubtasks_ExistingPath verifies that cursor
// movement inside the InfoPanel Subtasks tab dispatches a PreviewRequestMsg
// carrying the newly-selected subtask key.
func TestPreviewFollowsCursor_InfoSubtasks_ExistingPath(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	main := &jira.Issue{
		Key:      mainKey,
		Subtasks: []jira.Issue{{Key: "SUB-1"}, {Key: "SUB-2"}},
	}
	a.infoPanel.SetIssue(main)
	a.infoPanel.SetActiveTab(views.InfoTabSubtasks)
	a.infoPanel.SetFocused(true)
	a.infoPanel.ResolveNav = navResolverJK

	_, cmd := a.infoPanel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
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

// TestPreviewFollowsCursor_InfoLinks_ExistingPath is the Links-tab analog of
// the Subtasks cursor test above.
func TestPreviewFollowsCursor_InfoLinks_ExistingPath(t *testing.T) {
	const linkKey1 = "LNK-1"
	const linkKey2 = "LNK-2"

	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	main := &jira.Issue{
		Key: mainKey,
		IssueLinks: []jira.IssueLink{
			{Type: &jira.IssueLinkType{Name: "relates to"}, OutwardIssue: &jira.Issue{Key: linkKey1}},
			{Type: &jira.IssueLinkType{Name: "relates to"}, OutwardIssue: &jira.Issue{Key: linkKey2}},
		},
	}
	a.infoPanel.SetIssue(main)
	a.infoPanel.SetActiveTab(views.InfoTabLinks)
	a.infoPanel.SetFocused(true)
	a.infoPanel.ResolveNav = navResolverJK

	_, cmd := a.infoPanel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
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

// TestPreviewFollowsCursor_RapidCursor_OnlyLastFetch pins the debounce
// guarantee through the IssueSelectedMsg path: five rapid moves advance the
// epoch to 5, stale debounce ticks from epochs 1-4 are dropped, only the
// fresh tick for epoch 5 issues a GetIssue call.
func TestPreviewFollowsCursor_RapidCursor_OnlyLastFetch(t *testing.T) {
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

	a := newAppWithFake(t, fake)
	issues := []jira.Issue{
		{Key: "ABC-1"}, {Key: "ABC-2"}, {Key: "ABC-3"}, {Key: "ABC-4"}, {Key: lastKey},
	}
	a.issuesList.SetIssues(issues)

	for i := range issues {
		_, _ = a.Update(views.IssueSelectedMsg{Issue: &issues[i]})
	}
	if a.previewEpoch != 5 {
		t.Fatalf("previewEpoch = %d after 5 moves, want 5", a.previewEpoch)
	}

	for epoch := 1; epoch <= 4; epoch++ {
		_, stale := a.Update(previewDebounceMsg{key: issues[epoch-1].Key, epoch: epoch})
		if stale != nil {
			stale()
		}
	}
	if len(fake.GetIssueCalls) != 0 {
		t.Errorf("stale debounce ticks caused %d GetIssue call(s), want 0", len(fake.GetIssueCalls))
	}

	_, fetchCmd := a.Update(previewDebounceMsg{key: lastKey, epoch: 5})
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

// TestPreviewFollowsCursor_Projects verifies that ProjectHoveredMsg routes
// the hovered project into DetailView in project mode. The hover message
// is emitted by ProjectList on every cursor move (see views.ProjectList.Update).
func TestPreviewFollowsCursor_Projects(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	projects := []jira.Project{{Key: "P1", Name: "Project One"}, {Key: "P2", Name: "Project Two"}}

	_, _ = a.Update(views.ProjectHoveredMsg{Project: &projects[1]})

	if got := a.detailView.Mode(); got != views.ModeProject {
		t.Errorf("detailView.Mode = %v, want ModeProject", got)
	}
	// The detailView keeps a copy of the project; verify via re-render path
	// would couple to View() output. Instead, hover with nil and confirm
	// no panic and mode stays.
	_, _ = a.Update(views.ProjectHoveredMsg{Project: nil})
	if got := a.detailView.Mode(); got != views.ModeProject {
		t.Errorf("nil hover changed mode away from ModeProject (got %v)", got)
	}
}

// TestPreviewFollowsCursor_UnknownKey_FallsBackToContext verifies the
// content-fallback chain: when a preview is requested for a key that is
// neither cached nor matches the main-list selection (e.g. a sub/link key
// without cache), the DetailView keeps showing the previously displayed
// context issue rather than blanking out. The spec calls for "letzter
// Content bleibt stehen" until the fetch resolves.
func TestPreviewFollowsCursor_UnknownKey_FallsBackToContext(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	main := &jira.Issue{Key: mainKey, Summary: "main issue"}
	a.issuesList.SetIssues([]jira.Issue{*main})
	a.detailView.SetIssue(main)

	_, _ = a.Update(views.PreviewRequestMsg{Key: "UNKNOWN-99"})

	if got := a.detailView.IssueKey(); got != mainKey {
		t.Errorf("DetailView.IssueKey = %q, want context fallback %q", got, mainKey)
	}
}

