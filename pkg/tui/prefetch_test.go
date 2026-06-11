package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func runAll(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, subCmd := range batch {
			runAll(subCmd)
		}
	}
}

func TestPrefetchRelated_IncludesParent(t *testing.T) {
	t.Parallel()
	const parentKey = "PARENT-1"

	var got string
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, jql string, _, _ int) (*jira.SearchResult, error) {
		got = jql
		return &jira.SearchResult{}, nil
	}
	app := newAppWithFake(t, fake)

	issue := &jira.Issue{
		Key:      mainKey,
		Parent:   &jira.Issue{Key: parentKey},
		Subtasks: []jira.Issue{{Key: subKey1}},
	}

	cmd := app.prefetchRelated(issue)
	if cmd == nil {
		t.Fatal("expected non-nil prefetch cmd")
	}
	cmd()

	if !strings.Contains(got, parentKey) {
		t.Errorf("SearchIssues JQL %q does not contain parent %q", got, parentKey)
	}
	if !strings.Contains(got, subKey1) {
		t.Errorf("SearchIssues JQL %q does not contain subtask %q", got, subKey1)
	}
}

func TestIssueSelectedMsg_PrefetchesRelated(t *testing.T) {
	t.Parallel()
	var got string
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, jql string, _, _ int) (*jira.SearchResult, error) {
		got = jql
		return &jira.SearchResult{}, nil
	}
	app := newAppWithFake(t, fake)

	issue := &jira.Issue{
		Key:      mainKey,
		Subtasks: []jira.Issue{{Key: subKey1}},
	}

	_, cmd := app.Update(views.IssueSelectedMsg{Issue: issue})
	if cmd == nil {
		t.Fatal("expected a prefetch cmd from IssueSelectedMsg, got nil")
	}
	runAll(cmd)

	if !strings.Contains(got, subKey1) {
		t.Errorf("SearchIssues JQL %q does not contain %q", got, subKey1)
	}
}

func TestPreviewSelectedIssue_ReturnsPrefetchCmd(t *testing.T) {
	t.Parallel()
	var got string
	fake := &jiratest.FakeClient{T: t}
	fake.SearchIssuesFunc = func(_ context.Context, jql string, _, _ int) (*jira.SearchResult, error) {
		got = jql
		return &jira.SearchResult{}, nil
	}
	app := newAppWithFake(t, fake)
	app.issuesList.SetIssues([]jira.Issue{{
		Key:      mainKey,
		Subtasks: []jira.Issue{{Key: subKey1}},
	}})

	cmd := app.previewSelectedIssue()
	if cmd == nil {
		t.Fatal("expected a prefetch cmd from previewSelectedIssue, got nil")
	}
	cmd()

	if !strings.Contains(got, subKey1) {
		t.Errorf("SearchIssues JQL %q does not contain %q", got, subKey1)
	}
}
