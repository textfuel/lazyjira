package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func navDownResolver(key string) components.NavAction {
	if key == "j" {
		return components.NavDown
	}
	return components.NavNone
}

func makeInfoPanelFocused() *InfoPanel {
	p := NewInfoPanel()
	p.ResolveNav = navDownResolver
	p.Focused = true
	return p
}

func pressJ() tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
}

func TestInfoPanel_SubTab_CursorMove_DispatchesPreviewRequestMsg(t *testing.T) {
	t.Parallel()
	p := makeInfoPanelFocused()

	issue := &jira.Issue{
		Key: "MAIN-1",
		Subtasks: []jira.Issue{
			{Key: "SUB-1", Summary: "first subtask"},
			{Key: "SUB-2", Summary: "second subtask"},
		},
	}
	p.SetIssue(issue)

	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	_, cmd := p.Update(pressJ())
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd after cursor move in Sub tab, got nil")
	}

	msg := cmd()
	prm, ok := msg.(PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if prm.Key != "SUB-2" {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", prm.Key, "SUB-2")
	}
}

func TestInfoPanel_LnkTab_CursorMove_OutwardLink(t *testing.T) {
	t.Parallel()
	p := makeInfoPanelFocused()

	issue := &jira.Issue{
		Key: "MAIN-1",
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "Blocks", Outward: "blocks", Inward: "is blocked by"},
				OutwardIssue: &jira.Issue{Key: "OUT-1", Summary: "outward issue"},
			},
			{
				Type:         &jira.IssueLinkType{Name: "Blocks", Outward: "blocks", Inward: "is blocked by"},
				OutwardIssue: &jira.Issue{Key: "OUT-2", Summary: "second outward"},
			},
		},
	}
	p.SetIssue(issue)

	for p.activeTab != InfoTabLinks {
		p.NextTab()
	}

	_, cmd := p.Update(pressJ())
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd after cursor move in Lnk tab, got nil")
	}

	msg := cmd()
	prm, ok := msg.(PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if prm.Key != "OUT-2" {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", prm.Key, "OUT-2")
	}
}

func TestInfoPanel_LnkTab_CursorMove_InwardLink(t *testing.T) {
	t.Parallel()
	p := makeInfoPanelFocused()

	issue := &jira.Issue{
		Key: "MAIN-1",
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "Blocks", Outward: "blocks", Inward: "is blocked by"},
				OutwardIssue: &jira.Issue{Key: "OUT-1", Summary: "outward"},
				InwardIssue:  &jira.Issue{Key: "IN-1", Summary: "inward"},
			},
		},
	}
	p.SetIssue(issue)

	for p.activeTab != InfoTabLinks {
		p.NextTab()
	}

	_, cmd := p.Update(pressJ())
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd after cursor move in Lnk tab (inward), got nil")
	}

	msg := cmd()
	prm, ok := msg.(PreviewRequestMsg)
	if !ok {
		t.Fatalf("expected PreviewRequestMsg, got %T", msg)
	}
	if prm.Key != "IN-1" {
		t.Errorf("PreviewRequestMsg.Key = %q, want %q", prm.Key, "IN-1")
	}
}

func TestInfoPanel_SetIssue_PreservesActiveTab(t *testing.T) {
	t.Parallel()
	p := makeInfoPanelFocused()

	issueA := &jira.Issue{
		Key: "MAIN-1",
		Subtasks: []jira.Issue{
			{Key: "SUB-1", Summary: "first subtask"},
			{Key: "SUB-2", Summary: "second subtask"},
		},
	}
	p.SetIssue(issueA)

	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	p.Update(pressJ())
	if p.Cursor == 0 {
		t.Fatalf("setup: expected cursor > 0 after nav, got 0")
	}

	issueB := &jira.Issue{
		Key: "MAIN-2",
		Subtasks: []jira.Issue{
			{Key: "OTHER-1", Summary: "different subtask"},
		},
	}
	p.SetIssue(issueB)

	if p.ActiveTab() != InfoTabSubtasks {
		t.Errorf("ActiveTab = %v, want InfoTabSubtasks", p.ActiveTab())
	}
	if p.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0", p.Cursor)
	}
	if p.Offset != 0 {
		t.Errorf("Offset = %d, want 0", p.Offset)
	}
}

func TestInfoPanel_SetIssue_SameKey_PreservesCursor(t *testing.T) {
	t.Parallel()
	p := makeInfoPanelFocused()

	issue := &jira.Issue{
		Key: "MAIN-1",
		Subtasks: []jira.Issue{
			{Key: "SUB-1", Summary: "first subtask"},
			{Key: "SUB-2", Summary: "second subtask"},
		},
	}
	p.SetIssue(issue)

	for p.activeTab != InfoTabSubtasks {
		p.NextTab()
	}

	p.Update(pressJ())
	cursorBefore := p.Cursor
	if cursorBefore == 0 {
		t.Fatalf("setup: expected cursor > 0 after nav, got 0")
	}

	refreshed := &jira.Issue{
		Key: "MAIN-1",
		Subtasks: []jira.Issue{
			{Key: "SUB-1", Summary: "first subtask"},
			{Key: "SUB-2", Summary: "second subtask"},
		},
	}
	p.SetIssue(refreshed)

	if p.Cursor != cursorBefore {
		t.Errorf("Cursor = %d, want %d", p.Cursor, cursorBefore)
	}
	if p.ActiveTab() != InfoTabSubtasks {
		t.Errorf("ActiveTab = %v, want InfoTabSubtasks", p.ActiveTab())
	}
}

func TestInfoPanel_FieldsTab_CursorMove_NoPreviewRequestMsg(t *testing.T) {
	t.Parallel()
	p := makeInfoPanelFocused()

	issue := &jira.Issue{
		Key:     "MAIN-1",
		Summary: "something",
	}
	p.SetIssue(issue)

	if p.activeTab != InfoTabFields {
		t.Fatal("expected InfoTabFields as default tab")
	}

	_, cmd := p.Update(pressJ())

	if cmd == nil {
		return // nil is acceptable: no preview dispatch
	}

	msg := cmd()
	if _, ok := msg.(PreviewRequestMsg); ok {
		t.Error("Fields tab must not dispatch PreviewRequestMsg on cursor move")
	}
}
