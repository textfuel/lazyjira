package views

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

// TestInfoPanel_RenderSubtaskRowPairs_FallbackToSubtasks pins the Server/DC
// path: cloud=false → Sub tab uses issue.Subtasks even if SetChildren has
// never been called.
func TestInfoPanel_RenderSubtaskRowPairs_FallbackToSubtasks(t *testing.T) {
	p := makeInfoPanelFocused()
	issue := &jira.Issue{Key: "MAIN-1", Subtasks: []jira.Issue{{Key: "SUB-1", Summary: "s1"}}}
	p.SetIssue(issue)
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	if got := p.tabItemCount(); got != 1 {
		t.Errorf("Server/DC fallback: tabItemCount = %d, want 1 (issue.Subtasks)", got)
	}
	if got := p.SelectedSubtaskKey(); got != "SUB-1" {
		t.Errorf("Server/DC fallback: SelectedSubtaskKey = %q, want SUB-1", got)
	}
}

// TestInfoPanel_RenderSubtaskRowPairs_UsesChildrenSlice pins the Cloud path:
// after SetChildren the Sub tab renders the children, ignoring Subtasks.
func TestInfoPanel_RenderSubtaskRowPairs_UsesChildrenSlice(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	issue := &jira.Issue{
		Key:      "EPIC-1",
		Subtasks: []jira.Issue{{Key: "OLD-SUB", Summary: "should not show"}},
	}
	p.SetIssue(issue)
	p.SetChildren("EPIC-1", []jira.Issue{{Key: "CHILD-1", Summary: "first"}, {Key: "CHILD-2", Summary: "second"}})
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	if got := p.tabItemCount(); got != 2 {
		t.Errorf("Cloud children: tabItemCount = %d, want 2", got)
	}
	if got := p.SelectedSubtaskKey(); got != "CHILD-1" {
		t.Errorf("Cloud children: SelectedSubtaskKey = %q, want CHILD-1", got)
	}
}

// TestInfoPanel_MaybeChildrenRequest_CloudFiresOnSubTab verifies the request
// trigger: cloud + Sub tab + no children loaded → emits ChildrenRequestMsg.
func TestInfoPanel_MaybeChildrenRequest_CloudFiresOnSubTab(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	cmd := p.MaybeChildrenRequest()
	if cmd == nil {
		t.Fatal("Cloud + SubTab + no children: expected non-nil Cmd")
	}
	msg := cmd()
	req, ok := msg.(ChildrenRequestMsg)
	if !ok {
		t.Fatalf("expected ChildrenRequestMsg, got %T", msg)
	}
	if req.Key != "EPIC-1" {
		t.Errorf("ChildrenRequestMsg.Key = %q, want EPIC-1", req.Key)
	}
}

// TestInfoPanel_MaybeChildrenRequest_ServerDCNoFire pins the Server/DC path:
// cloud=false → never emits ChildrenRequestMsg, even on Sub tab.
func TestInfoPanel_MaybeChildrenRequest_ServerDCNoFire(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	if cmd := p.MaybeChildrenRequest(); cmd != nil {
		t.Errorf("Server/DC: expected nil Cmd, got non-nil")
	}
}

// TestInfoPanel_MaybeChildrenRequest_NotOnFieldsTab pins that the request
// only fires on the Sub tab, not on Fields/Links.
func TestInfoPanel_MaybeChildrenRequest_NotOnFieldsTab(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	// activeTab defaults to InfoTabFields after SetIssue.

	if cmd := p.MaybeChildrenRequest(); cmd != nil {
		t.Errorf("Fields tab: expected nil Cmd, got non-nil")
	}
}

// TestInfoPanel_MaybeChildrenRequest_AlreadyLoadedNoFire pins that once
// children are loaded for the current key, no further request fires.
func TestInfoPanel_MaybeChildrenRequest_AlreadyLoadedNoFire(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	p.SetChildren("EPIC-1", []jira.Issue{{Key: "C-1"}})
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	if cmd := p.MaybeChildrenRequest(); cmd != nil {
		t.Errorf("Already-loaded: expected nil Cmd, got non-nil")
	}
}

// TestInfoPanel_SetChildren_StaleKeyDropped pins the stale-drop invariant at
// the panel boundary: a SetChildren call for a key other than the currently
// displayed issue is ignored.
func TestInfoPanel_SetChildren_StaleKeyDropped(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "NEW-EPIC"})

	p.SetChildren("OLD-EPIC", []jira.Issue{{Key: "STALE-CHILD"}})

	if got := p.Children(); got != nil {
		t.Errorf("Stale SetChildren: expected nil children, got %+v", got)
	}
}

// TestInfoPanel_SetChildrenError_RendersErrorRow pins that fetch errors
// surface as a single error row in the Sub tab.
func TestInfoPanel_SetChildrenError_RendersErrorRow(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	p.SetChildrenError("EPIC-1", "boom")
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	_, plain := p.renderSubtaskRowPairs(40)
	if len(plain) != 1 {
		t.Fatalf("error path: expected 1 row, got %d (%v)", len(plain), plain)
	}
	if plain[0] == "" || plain[0][0:5] != " Fail" {
		t.Errorf("error row content = %q, want prefix ' Fail'", plain[0])
	}
}

// TestInfoPanel_EmptyChildren_RendersEmptyState pins the 0-children empty
// state for the Cloud path.
func TestInfoPanel_EmptyChildren_RendersEmptyState(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	p.SetChildren("EPIC-1", []jira.Issue{}) // empty but loaded
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	_, plain := p.renderSubtaskRowPairs(40)
	if len(plain) != 1 {
		t.Fatalf("empty state: expected 1 placeholder row, got %d (%v)", len(plain), plain)
	}
	if plain[0] != " No children" {
		t.Errorf("empty row = %q, want %q", plain[0], " No children")
	}
}

// TestInfoPanel_SetIssue_ResetsChildrenState pins that switching to a new
// issue clears any previously loaded children — so MaybeChildrenRequest
// fires again for the new key.
func TestInfoPanel_SetIssue_ResetsChildrenState(t *testing.T) {
	p := makeInfoPanelFocused()
	p.SetCloud(true)
	p.SetIssue(&jira.Issue{Key: "EPIC-1"})
	p.SetChildren("EPIC-1", []jira.Issue{{Key: "C-1"}})

	p.SetIssue(&jira.Issue{Key: "EPIC-2"})
	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	if got := p.Children(); got != nil {
		t.Errorf("after issue switch: expected nil children, got %+v", got)
	}
	cmd := p.MaybeChildrenRequest()
	if cmd == nil {
		t.Fatal("after issue switch: expected MaybeChildrenRequest to re-fire")
	}
	msg := cmd().(ChildrenRequestMsg)
	if msg.Key != "EPIC-2" {
		t.Errorf("re-fire key = %q, want EPIC-2", msg.Key)
	}
}
