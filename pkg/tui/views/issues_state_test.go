package views

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func listWithTabs(tabs ...config.IssueTabConfig) *IssuesList {
	list := NewIssuesList()
	list.SetTabs(tabs)
	return list
}

func TestIssuesList_JQLTabLifecycle(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})

	if list.IsJQLTab() {
		t.Error("no JQL tab should exist initially")
	}

	list.AddJQLTab("project = X")

	if !list.HasJQLTab() || !list.IsJQLTab() {
		t.Error("JQL tab should be active after AddJQLTab")
	}
	testkit.AssertEqual(t, "jql query", list.JQLQuery(), "project = X")

	list.AddJQLTab("project = Y")
	testkit.AssertEqual(t, "replaced jql", list.JQLQuery(), "project = Y")

	list.RemoveJQLTab()

	if list.HasJQLTab() || list.IsJQLTab() {
		t.Error("JQL tab should be gone after RemoveJQLTab")
	}
	testkit.AssertEqual(t, "cleared query", list.JQLQuery(), "")
	testkit.AssertEqual(t, "back to tab zero", list.GetTabIndex(), 0)
}

func TestIssuesList_SetIssuesAndSelect(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues([]jira.Issue{{Key: testKey}, {Key: testKey2}})

	testkit.AssertEqual(t, "issue count", len(list.CurrentIssues()), 2)

	if !list.SelectByKey(testKey2) {
		t.Fatal("SelectByKey should find PLAT-2")
	}
	if sel := list.SelectedIssue(); sel == nil || sel.Key != testKey2 {
		t.Errorf("selected = %v, want %s", sel, testKey2)
	}
	if list.SelectByKey("MISSING-1") {
		t.Error("SelectByKey should report false for an unknown key")
	}
}

func TestIssuesList_PatchIssue(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues([]jira.Issue{{Key: testKey, Summary: "old"}})

	list.PatchIssue(&jira.Issue{Key: testKey, Summary: "new"})

	if got := list.CurrentIssues()[0].Summary; got != "new" {
		t.Errorf("summary = %q, want new", got)
	}
}

func TestIssuesList_FindInAnyTab(t *testing.T) {
	t.Parallel()
	list := listWithTabs(
		config.IssueTabConfig{Name: "All", JQL: "x"},
		config.IssueTabConfig{Name: "Mine", JQL: "y"},
	)
	list.SetIssues([]jira.Issue{{Key: testKey}})
	list.SetIssuesForTab(1, []jira.Issue{{Key: "PLAT-9"}})

	if tab, ok := list.FindInAnyTab(testKey); !ok || tab != 0 {
		t.Errorf("FindInAnyTab(current) = (%d,%v), want (0,true)", tab, ok)
	}
	if tab, ok := list.FindInAnyTab("PLAT-9"); !ok || tab != 1 {
		t.Errorf("FindInAnyTab(other tab) = (%d,%v), want (1,true)", tab, ok)
	}
	if _, ok := list.FindInAnyTab("MISSING-1"); ok {
		t.Error("FindInAnyTab should report false for an unknown key")
	}
}

func TestIssuesList_InjectIssue(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})

	list.InjectIssue(jira.Issue{Key: testKey})
	list.InjectIssue(jira.Issue{Key: testKey})
	list.SetTabIndex(0)

	if got := len(list.CurrentIssues()); got != 1 {
		t.Errorf("injected count = %d, want 1 (no duplicates)", got)
	}
}

func TestIssuesList_FilterSelectsMatch(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues([]jira.Issue{{Key: testKey, Summary: "alpha"}, {Key: testKey2, Summary: "beta"}})

	list.SetFilter("beta")
	if sel := list.SelectedIssue(); sel == nil || sel.Key != testKey2 {
		t.Errorf("filtered selection = %v, want %s", sel, testKey2)
	}

	list.ClearFilter()
	if sel := list.SelectedIssue(); sel == nil {
		t.Error("clearing filter should keep a selection")
	}
}

func TestIssuesList_TabNavigation(t *testing.T) {
	t.Parallel()
	list := listWithTabs(
		config.IssueTabConfig{Name: "All", JQL: "x"},
		config.IssueTabConfig{Name: "Mine", JQL: "y"},
	)

	list.NextTab()
	testkit.AssertEqual(t, "next tab", list.GetTabIndex(), 1)

	list.NextTab()
	testkit.AssertEqual(t, "wraps to zero", list.GetTabIndex(), 0)

	list.PrevTab()
	testkit.AssertEqual(t, "prev wraps to last", list.GetTabIndex(), 1)
}
