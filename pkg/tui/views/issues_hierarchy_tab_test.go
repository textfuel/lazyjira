package views

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestAddHierarchyTab_SetsTabAndFocus(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	issues := []jira.Issue{{Key: "CHILD-1"}}

	idx := m.AddHierarchyTab("Child", issues)

	if !m.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = false, want true")
	}
	if !m.IsHierarchyTab() {
		t.Fatalf("IsHierarchyTab() = false, want true after AddHierarchyTab")
	}
	if m.GetTabIndex() != idx {
		t.Fatalf("GetTabIndex() = %d, want %d", m.GetTabIndex(), idx)
	}
	if got := m.HierarchyTitle(); got != "Child" {
		t.Fatalf("HierarchyTitle() = %q, want %q", got, "Child")
	}
	if m.HierarchyStack() == nil {
		t.Fatalf("HierarchyStack() = nil, want non-nil after AddHierarchyTab")
	}
}

func TestAddHierarchyTab_AppendedAfterJQLTab(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddJQLTab("project = FOO")
	jqlIdx := m.GetTabIndex()

	hierarchyIdx := m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	if hierarchyIdx <= jqlIdx {
		t.Fatalf("hierarchyIdx=%d must be greater than jqlIdx=%d (hierarchy after JQL)", hierarchyIdx, jqlIdx)
	}
	if !m.HasJQLTab() {
		t.Fatalf("JQL-Tab must survive AddHierarchyTab")
	}
}

func TestReplaceHierarchyTabContent_UpdatesTitleAndIssues(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	idx := m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	m.ReplaceHierarchyTabContent("Parent", []jira.Issue{{Key: "P-1"}})

	if m.GetTabIndex() != idx {
		t.Fatalf("GetTabIndex() = %d, want %d (stable across replace)", m.GetTabIndex(), idx)
	}
	if got := m.HierarchyTitle(); got != "Parent" {
		t.Fatalf("HierarchyTitle() = %q, want Parent", got)
	}
	sel := m.SelectedIssue()
	if sel == nil || sel.Key != "P-1" {
		t.Fatalf("SelectedIssue() after replace = %+v, want P-1", sel)
	}
}

func TestReplaceHierarchyTabContent_NoHierarchyTab_NoOp(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})

	m.ReplaceHierarchyTabContent("Parent", []jira.Issue{{Key: "P-1"}})

	if m.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = true, want false (no-op when no hierarchy tab)")
	}
}

func TestRemoveHierarchyTab_RemovesTabAndStack(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	m.RemoveHierarchyTab()

	if m.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = true, want false after RemoveHierarchyTab")
	}
	if m.HierarchyStack() != nil {
		t.Fatalf("HierarchyStack() = non-nil, want nil after RemoveHierarchyTab")
	}
	if m.GetTabIndex() != 0 {
		t.Fatalf("GetTabIndex() = %d, want 0 after RemoveHierarchyTab", m.GetTabIndex())
	}
}

func TestRemoveHierarchyTab_NoHierarchyTab_NoOp(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})

	m.RemoveHierarchyTab()

	if m.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = true, want false")
	}
}

func TestRemoveHierarchyTab_WithJQLTabPresent_JQLIdxStable(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddJQLTab("project = FOO")
	jqlIdxBefore := len(m.tabs) - 1
	m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	m.RemoveHierarchyTab()

	if !m.HasJQLTab() {
		t.Fatalf("JQL-Tab must survive RemoveHierarchyTab")
	}
	if got := len(m.tabs) - 1; got != jqlIdxBefore {
		t.Fatalf("JQL-Tab index shifted after RemoveHierarchyTab: got %d, want %d", got, jqlIdxBefore)
	}
}

func TestHierarchyTab_SurvivesJQLTabSwitch(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddJQLTab("project = FOO")
	jqlIdx := m.GetTabIndex()
	hierarchyIdx := m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})
	stackBefore := m.HierarchyStack()

	m.SetTabIndex(jqlIdx)
	m.SetTabIndex(hierarchyIdx)

	if !m.HasHierarchyTab() {
		t.Fatalf("hierarchy tab lost after JQL round-trip")
	}
	if !m.IsHierarchyTab() {
		t.Fatalf("IsHierarchyTab() = false after switching back")
	}
	if m.HierarchyStack() != stackBefore {
		t.Fatalf("HierarchyStack() identity changed after switch (want same pointer)")
	}
	if m.HierarchyTitle() != "Child" {
		t.Fatalf("HierarchyTitle() = %q, want Child", m.HierarchyTitle())
	}
}

func TestHierarchyTab_StackAccessibleAfterSwitch(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddJQLTab("project = FOO")
	jqlIdx := m.GetTabIndex()
	hierarchyIdx := m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	depthBefore := m.HierarchyStack().Depth()
	m.SetTabIndex(jqlIdx)
	m.SetTabIndex(hierarchyIdx)

	if got := m.HierarchyStack().Depth(); got != depthBefore {
		t.Fatalf("HierarchyStack Depth after switch = %d, want %d", got, depthBefore)
	}
}

func TestInvalidateTabCache_RemovesHierarchyTab(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	m.InvalidateTabCache()

	if m.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = true, want false after InvalidateTabCache")
	}
	if m.HierarchyStack() != nil {
		t.Fatalf("HierarchyStack() = non-nil, want nil after InvalidateTabCache")
	}
}

func TestInvalidateTabCache_WithBothTabs_RemovesBoth(t *testing.T) {
	m := NewIssuesList()
	m.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	m.AddJQLTab("project = FOO")
	m.AddHierarchyTab("Child", []jira.Issue{{Key: "C-1"}})

	m.InvalidateTabCache()

	if m.HasHierarchyTab() {
		t.Fatalf("hierarchy tab survived InvalidateTabCache")
	}
	if m.HasJQLTab() {
		t.Fatalf("JQL-Tab survived InvalidateTabCache")
	}
}
