package views

import (
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func makeIssueWithStatus(key string) *jira.Issue {
	return &jira.Issue{
		Key:     key,
		Summary: "test summary",
		Status:  &jira.Status{Name: "Open"},
	}
}

func TestInfoPanel_IssueKey_ReturnsKey(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	testkit.AssertEqual(t, "issue key", panel.IssueKey(), testKey)
}

func TestInfoPanel_IssueKey_EmptyWhenNilIssue(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	testkit.AssertEqual(t, "nil issue key", panel.IssueKey(), "")
}

func TestInfoPanel_SetFields_UpdatesFields(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetFields([]config.FieldConfig{{ID: "status"}})
	fields := panel.Fields()
	if len(fields) == 0 {
		t.Error("SetFields should produce visible fields")
	}
}

func TestInfoPanel_SetFilter_FiltersRows(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key:     testKey,
		Status:  &jira.Status{Name: "Open"},
		Summary: "test summary",
	})
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	panel.SetFilter("status")
	output := stripANSI(panel.View())
	if !strings.Contains(output, "Status") {
		t.Errorf("filtered view = %q, want to contain Status", output)
	}
}

func TestInfoPanel_ClearFilter_RemovesFilter(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	panel.SetFilter("status")
	panel.ClearFilter()
	testkit.AssertEqual(t, "filter cleared", panel.filter, "")
}

func TestInfoPanel_Issue_ReturnsCurrentIssue(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	issue := makeIssueWithStatus(testKey)
	panel.SetIssue(issue)
	if panel.Issue() != issue {
		t.Error("Issue() should return the set issue pointer")
	}
}

func TestInfoPanel_SetActiveTab_SetsTab(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetActiveTab(InfoTabLinks)
	testkit.AssertEqual(t, "active tab", panel.ActiveTab(), InfoTabLinks)
}

func TestInfoPanel_Fields_ReturnsBuiltFields(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key:    testKey,
		Status: &jira.Status{Name: "Open"},
	})
	fields := panel.Fields()
	if len(fields) == 0 {
		t.Error("Fields() should return non-empty list for issue with status")
	}
}

func TestInfoPanel_SelectedInfoField_ReturnsFieldOnFieldsTab(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key:    testKey,
		Status: &jira.Status{Name: "Open"},
	})
	panel.SetActiveTab(InfoTabFields)
	panel.SetSize(80, 24)
	panel.Focused = true
	field := panel.SelectedInfoField()
	if field == nil {
		t.Error("SelectedInfoField() should return a field when on Fields tab with issue")
	}
}

func TestInfoPanel_SelectedInfoField_NilOnOtherTabs(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetActiveTab(InfoTabLinks)
	if panel.SelectedInfoField() != nil {
		t.Error("SelectedInfoField() should return nil when not on Fields tab")
	}
}

func TestInfoPanel_ContentHeight_MinimumThree(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	testkit.AssertEqual(t, "empty content height", panel.ContentHeight(), 3)
}

func TestInfoPanel_PrevTab_CyclesBack(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	testkit.AssertEqual(t, "starts at fields", panel.ActiveTab(), InfoTabFields)
	panel.PrevTab()
	testkit.AssertEqual(t, "prev wraps to sub", panel.ActiveTab(), InfoTabSubtasks)
}

func TestInfoPanel_Init_ReturnsNil(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	if cmd := panel.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestInfoPanel_View_CollapsedBar_HeightOne(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetSize(80, 1)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "Info") {
		t.Errorf("collapsed bar = %q, want 'Info' label", output)
	}
}

func TestInfoPanel_View_NilIssue_ShowsPlaceholder(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "No issue selected") {
		t.Errorf("nil issue View() = %q, want 'No issue selected'", output)
	}
}

func TestInfoPanel_View_FieldsTab_ShowsStatus(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key:    testKey,
		Status: &jira.Status{Name: "In Progress"},
	})
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	panel.SetActiveTab(InfoTabFields)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "In Progress") {
		t.Errorf("fields tab View() = %q, want to contain status 'In Progress'", output)
	}
}

func TestInfoPanel_View_LinksTab_ShowsLinks(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key: testKey,
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Outward: "blocks", Inward: "is blocked by"},
				OutwardIssue: &jira.Issue{Key: testKey2, Summary: "linked issue"},
			},
		},
	})
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	panel.SetActiveTab(InfoTabLinks)
	output := stripANSI(panel.View())
	if !strings.Contains(output, testKey2) {
		t.Errorf("links tab View() = %q, want to contain linked key %s", output, testKey2)
	}
}

func TestInfoPanel_View_SubtasksTab_ShowsSubtasks(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key: testKey,
		Subtasks: []jira.Issue{
			{Key: testKey2, Summary: "a subtask"},
		},
	})
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	panel.SetActiveTab(InfoTabSubtasks)
	output := stripANSI(panel.View())
	if !strings.Contains(output, testKey2) {
		t.Errorf("sub tab View() = %q, want to contain subtask key %s", output, testKey2)
	}
}

func TestInfoPanel_View_FooterCount(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key:    testKey,
		Status: &jira.Status{Name: "Open"},
	})
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "of") {
		t.Errorf("View() = %q, should contain footer count", output)
	}
}

func TestInfoPanel_BuildTitle_ContainsAllTabs(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "Info") {
		t.Errorf("View() = %q, want tab label 'Info'", output)
	}
	if !strings.Contains(output, "Lnk") {
		t.Errorf("View() = %q, want tab label 'Lnk'", output)
	}
	if !strings.Contains(output, "Sub") {
		t.Errorf("View() = %q, want tab label 'Sub'", output)
	}
}

func TestInfoPanel_ClickTabAt_SwitchesToLinks(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	testkit.AssertEqual(t, "starts on fields", panel.ActiveTab(), InfoTabFields)
	panel.ClickTabAt(11)
	testkit.AssertEqual(t, "clicked to links", panel.ActiveTab(), InfoTabLinks)
}

func TestInfoPanel_ClickTabAt_SameTabNoChange(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.ClickTabAt(4)
	testkit.AssertEqual(t, "stays on fields", panel.ActiveTab(), InfoTabFields)
}

func TestInfoPanel_InfoPanelTabLabel_AllTabs(t *testing.T) {
	t.Parallel()
	testkit.AssertEqual(t, "fields label", infoPanelTabLabel(InfoTabFields), "Info")
	testkit.AssertEqual(t, "links label", infoPanelTabLabel(InfoTabLinks), "Lnk")
	testkit.AssertEqual(t, "sub label", infoPanelTabLabel(InfoTabSubtasks), "Sub")
}

func TestInfoPanel_RenderLinkRowPairs_InwardAndOutward(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(&jira.Issue{
		Key: testKey,
		IssueLinks: []jira.IssueLink{
			{
				Type:        &jira.IssueLinkType{Outward: "blocks", Inward: "is blocked by"},
				InwardIssue: &jira.Issue{Key: "IN-1", Summary: "inward issue"},
			},
		},
	})
	panel.SetSize(80, 24)
	panel.SetActiveTab(InfoTabLinks)
	styled, plain := panel.renderLinkRowPairs(70)
	if len(styled) == 0 || len(plain) == 0 {
		t.Error("renderLinkRowPairs should return rows for inward links")
	}
	if !strings.Contains(plain[0], "IN-1") {
		t.Errorf("plain rows = %v, want to contain IN-1", plain)
	}
}

func TestInfoPanel_View_FilterWithNoMatch_Empty(t *testing.T) {
	t.Parallel()
	panel := NewInfoPanel()
	panel.SetIssue(makeIssueWithStatus(testKey))
	panel.SetSize(80, 24)
	panel.SetFocused(true)
	panel.SetFilter("zzznomatch")
	output := stripANSI(panel.View())
	if strings.Contains(output, "of") {
		t.Errorf("no-match filter view = %q, should not have footer count", output)
	}
}
