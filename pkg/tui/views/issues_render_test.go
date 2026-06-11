package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func navResolverForIssues(key string) components.NavAction {
	switch key {
	case "j":
		return components.NavDown
	case "k":
		return components.NavUp
	}
	return components.NavNone
}

func makeFocusedIssuesList(issues []jira.Issue) *IssuesList {
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues(issues)
	list.SetSize(80, 24)
	list.SetFocused(true)
	list.ResolveNav = navResolverForIssues
	return list
}

func TestIssuesList_SetFields_UpdatesFieldList(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetFields([]string{"key", "summary"})
	list.SetTabs([]config.IssueTabConfig{{Name: "All", JQL: "x"}})
	list.SetIssues([]jira.Issue{{Key: testKey, Summary: "first"}})
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, testKey) {
		t.Errorf("View() = %q, want to contain key %s", output, testKey)
	}
}

func TestIssuesList_SetTypeIcons_StoresMaxWidth(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetTypeIcons(map[string]string{"Story": "S", "Bug": "B"})
	testkit.AssertEqual(t, "icon cols set", list.typeIconCols, 1)
}

func TestIssuesList_SetStatusIcons_StoresMaxWidth(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetStatusIcons(map[string]string{"Done": "✓"})
	if list.statusIconCols == 0 {
		t.Error("statusIconCols should be > 0 after SetStatusIcons")
	}
}

func TestIssuesList_SetPriorityIcons_StoresMaxWidth(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetPriorityIcons(map[string]string{"High": "H", "Low": "L"})
	testkit.AssertEqual(t, "priority icon cols set", list.priorityIconCols, 1)
}

func TestIssuesList_SetUserEmail_SetsEmail(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetUserEmail(testEmail)
	testkit.AssertEqual(t, "user email", list.userEmail, testEmail)
}

func TestIssuesList_HasCachedTab_TrueAfterSetIssues(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues([]jira.Issue{{Key: testKey}})
	if !list.HasCachedTab() {
		t.Error("HasCachedTab() should be true after SetIssues")
	}
}

func TestIssuesList_HasCachedTab_FalseWhenNoCacheYet(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	if list.HasCachedTab() {
		t.Error("HasCachedTab() should be false before SetIssues")
	}
}

func TestIssuesList_ContentHeight_MinimumSeven(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	testkit.AssertEqual(t, "empty content height", list.ContentHeight(), 7)
}

func TestIssuesList_Init_ReturnsNil(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	if cmd := list.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestIssuesList_Update_UnfocusedNoCmd(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues([]jira.Issue{{Key: testKey}, {Key: testKey2}})
	list.SetFocused(false)
	list.ResolveNav = navResolverForIssues
	_, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if cmd != nil {
		t.Error("unfocused Update should return nil cmd")
	}
}

func TestIssuesList_Update_FocusedNavEmitsSelected(t *testing.T) {
	t.Parallel()
	list := makeFocusedIssuesList([]jira.Issue{{Key: testKey}, {Key: testKey2}})
	_, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if cmd == nil {
		t.Fatal("focused nav should emit a Cmd")
	}
	msg := cmd()
	sel, ok := msg.(IssueSelectedMsg)
	if !ok {
		t.Fatalf("expected IssueSelectedMsg, got %T", msg)
	}
	testkit.AssertEqual(t, "selected key", sel.Issue.Key, testKey2)
}

func TestIssuesList_Update_NonNavKeyNoCmd(t *testing.T) {
	t.Parallel()
	list := makeFocusedIssuesList([]jira.Issue{{Key: testKey}})
	_, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd != nil {
		t.Error("non-nav key should return nil cmd")
	}
}

func TestIssuesList_ClickTabAt_SwitchesTab(t *testing.T) {
	t.Parallel()
	list := listWithTabs(
		config.IssueTabConfig{Name: "All", JQL: "x"},
		config.IssueTabConfig{Name: "Mine", JQL: "y"},
	)
	list.SetIssues([]jira.Issue{{Key: testKey}})
	list.SetSize(80, 24)
	switched := list.ClickTabAt(10)
	if !switched {
		t.Error("ClickTabAt on second tab should switch")
	}
	testkit.AssertEqual(t, "tab index", list.GetTabIndex(), 1)
}

func TestIssuesList_ClickTabAt_SameTabNoSwitch(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetSize(80, 24)
	switched := list.ClickTabAt(4)
	if switched {
		t.Error("ClickTabAt same tab should not switch")
	}
}

func TestIssuesList_ClickTabAt_EmptyTabsNoSwitch(t *testing.T) {
	t.Parallel()
	list := NewIssuesList()
	list.SetSize(80, 24)
	switched := list.ClickTabAt(5)
	if switched {
		t.Error("ClickTabAt with no tabs should not switch")
	}
}

func TestIssuesList_View_CollapsedBar_HeightOne(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetIssues([]jira.Issue{{Key: testKey}})
	list.SetSize(80, 1)
	output := stripANSI(list.View())
	if !strings.Contains(output, "All") {
		t.Errorf("collapsed bar = %q, want tab label All", output)
	}
	if !strings.Contains(output, "1 of 1") {
		t.Errorf("collapsed bar = %q, want '1 of 1' footer", output)
	}
}

func TestIssuesList_View_ShowsIssueKeys(t *testing.T) {
	t.Parallel()
	list := makeFocusedIssuesList([]jira.Issue{
		{Key: testKey, Summary: "first issue"},
		{Key: testKey2, Summary: "second issue"},
	})
	output := stripANSI(list.View())
	if !strings.Contains(output, testKey) {
		t.Errorf("View() = %q, want to contain %s", output, testKey)
	}
	if !strings.Contains(output, testKey2) {
		t.Errorf("View() = %q, want to contain %s", output, testKey2)
	}
}

func TestIssuesList_View_ShowsFooterCount(t *testing.T) {
	t.Parallel()
	list := makeFocusedIssuesList([]jira.Issue{
		{Key: testKey},
		{Key: testKey2},
	})
	output := stripANSI(list.View())
	if !strings.Contains(output, "1 of 2") {
		t.Errorf("View() = %q, want '1 of 2' footer", output)
	}
}

func TestIssuesList_View_EmptyNoFooter(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if strings.Contains(output, "of") {
		t.Errorf("empty issues View() = %q, should not contain footer", output)
	}
}

func TestIssuesList_RenderIssueRow_AllFields(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetFields([]string{"key", fieldStatus, "summary", "assignee", "priority", "type", "updated"})
	list.SetIssues([]jira.Issue{
		{
			Key:       testKey,
			Summary:   "full row issue",
			Status:    &jira.Status{Name: "Open", CategoryKey: "new"},
			Priority:  &jira.Priority{Name: "High"},
			Assignee:  &jira.User{DisplayName: "Alice"},
			IssueType: &jira.IssueType{Name: "Story"},
			Updated:   time.Now().Add(-2 * time.Hour),
		},
	})
	list.SetSize(120, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, testKey) {
		t.Errorf("View() = %q, want key %s", output, testKey)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("View() = %q, want assignee Alice", output)
	}
}

func TestIssuesList_RenderIssueRow_WithIcons(t *testing.T) {
	t.Parallel()
	list := listWithTabs(config.IssueTabConfig{Name: "All", JQL: "x"})
	list.SetFields([]string{"key", fieldStatus, "summary", "priority", "type"})
	list.SetStatusIcons(map[string]string{"Open": "○"})
	list.SetPriorityIcons(map[string]string{"High": "!"})
	list.SetTypeIcons(map[string]string{"Story": "S"})
	list.SetIssues([]jira.Issue{
		{
			Key:       testKey,
			Summary:   "icon issue",
			Status:    &jira.Status{Name: "Open"},
			Priority:  &jira.Priority{Name: "High"},
			IssueType: &jira.IssueType{Name: "Story"},
		},
	})
	list.SetSize(120, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, testKey) {
		t.Errorf("View() with icons = %q, want key %s", output, testKey)
	}
	if !strings.Contains(output, "○") {
		t.Errorf("View() with icons = %q, want status icon ○", output)
	}
	if !strings.Contains(output, "!") {
		t.Errorf("View() with icons = %q, want priority icon !", output)
	}
	if !strings.Contains(output, "S") {
		t.Errorf("View() with icons = %q, want type icon S", output)
	}
}

func TestPadRight_PadsToWidth(t *testing.T) {
	t.Parallel()
	result := padRight("hi", 5)
	testkit.AssertEqual(t, "padded length", len(result), 5)
	if !strings.HasPrefix(result, "hi") {
		t.Errorf("padRight = %q, want 'hi   '", result)
	}
}

func TestPadRight_NoOpWhenAlreadyWide(t *testing.T) {
	t.Parallel()
	result := padRight("hello", 3)
	testkit.AssertEqual(t, "no padding needed", result, "hello")
}

func TestIssueTimeAgo_Zero(t *testing.T) {
	t.Parallel()
	result := issueTimeAgo(time.Time{})
	testkit.AssertEqual(t, "zero time", result, "")
}

func TestIssueTimeAgo_Minutes(t *testing.T) {
	t.Parallel()
	result := issueTimeAgo(time.Now().Add(-30 * time.Minute))
	if !strings.HasSuffix(result, "m") {
		t.Errorf("issueTimeAgo for 30min = %q, want suffix 'm'", result)
	}
}

func TestIssueTimeAgo_Hours(t *testing.T) {
	t.Parallel()
	result := issueTimeAgo(time.Now().Add(-5 * time.Hour))
	if !strings.HasSuffix(result, "h") {
		t.Errorf("issueTimeAgo for 5h = %q, want suffix 'h'", result)
	}
}

func TestIssueTimeAgo_Days(t *testing.T) {
	t.Parallel()
	result := issueTimeAgo(time.Now().Add(-3 * 24 * time.Hour))
	if !strings.HasSuffix(result, "d") {
		t.Errorf("issueTimeAgo for 3d = %q, want suffix 'd'", result)
	}
}

func TestIssueTimeAgo_Months(t *testing.T) {
	t.Parallel()
	result := issueTimeAgo(time.Now().Add(-60 * 24 * time.Hour))
	if !strings.HasSuffix(result, "mo") {
		t.Errorf("issueTimeAgo for 60d = %q, want suffix 'mo'", result)
	}
}

func TestStatusIcon_NilStatusReturnsEmpty(t *testing.T) {
	t.Parallel()
	result := statusIcon(map[string]string{"Open": "○"}, nil)
	testkit.AssertEqual(t, "nil status icon", result, "")
}

func TestStatusIcon_ConfiguredIconReturned(t *testing.T) {
	t.Parallel()
	result := statusIcon(map[string]string{"Open": "○"}, &jira.Status{Name: "Open"})
	testkit.AssertEqual(t, "icon value", result, "○")
}

func TestStatusIcon_MissingKeyReturnsEmpty(t *testing.T) {
	t.Parallel()
	result := statusIcon(map[string]string{"Done": "✓"}, &jira.Status{Name: "Open"})
	testkit.AssertEqual(t, "missing icon", result, "")
}

func TestPriorityIcon_NilPriorityReturnsEmpty(t *testing.T) {
	t.Parallel()
	result := priorityIcon(map[string]string{"High": "H"}, nil)
	testkit.AssertEqual(t, "nil priority icon", result, "")
}

func TestPriorityIcon_ConfiguredIconReturned(t *testing.T) {
	t.Parallel()
	result := priorityIcon(map[string]string{"High": "H"}, &jira.Priority{Name: "High"})
	testkit.AssertEqual(t, "icon value", result, "H")
}
