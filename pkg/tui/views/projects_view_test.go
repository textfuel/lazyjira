package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func navDownResolverProjects(key string) components.NavAction {
	if key == "j" {
		return components.NavDown
	}
	return components.NavNone
}

func makeFocusedProjectList(projects []jira.Project) *ProjectList {
	list := NewProjectList()
	list.SetProjects(projects)
	list.SetSize(80, 24)
	list.SetFocused(true)
	list.ResolveNav = navDownResolverProjects
	return list
}

func TestProjectList_ContentHeight_MinimumFive(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	testkit.AssertEqual(t, "empty content height", list.ContentHeight(), 5)
}

func TestProjectList_ContentHeight_GrowsWithItems(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{
		{Key: "AAA"},
		{Key: "BBB"},
		{Key: "CCC"},
		{Key: "DDD"},
		{Key: "EEE"},
		{Key: "FFF"},
	})
	if list.ContentHeight() < 8 {
		t.Errorf("ContentHeight() = %d, want >= 8 for 6 items", list.ContentHeight())
	}
}

func TestProjectList_Init_ReturnsNil(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	if cmd := list.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestProjectList_Update_UnfocusedDoesNothing(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA"}, {Key: "BBB"}})
	list.SetFocused(false)
	list.ResolveNav = navDownResolverProjects
	result, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "cursor stays at 0", result.Cursor, 0)
	if cmd != nil {
		t.Error("unfocused Update should return nil cmd")
	}
}

func TestProjectList_Update_FocusedNavEmitsHovered(t *testing.T) {
	t.Parallel()
	list := makeFocusedProjectList([]jira.Project{{Key: "AAA"}, {Key: "BBB"}})
	result, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	testkit.AssertEqual(t, "cursor moved to 1", result.Cursor, 1)
	if cmd == nil {
		t.Fatal("focused nav should emit a Cmd")
	}
	msg := cmd()
	hovered, ok := msg.(ProjectHoveredMsg)
	if !ok {
		t.Fatalf("expected ProjectHoveredMsg, got %T", msg)
	}
	testkit.AssertEqual(t, "hovered key", hovered.Project.Key, "BBB")
}

func TestProjectList_Update_NonNavKey_NoCmd(t *testing.T) {
	t.Parallel()
	list := makeFocusedProjectList([]jira.Project{{Key: "AAA"}})
	_, cmd := list.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if cmd != nil {
		t.Error("non-nav key should return nil cmd")
	}
}

func TestProjectList_View_CollapsedBar_HeightOne(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA"}, {Key: "BBB"}})
	list.SetSize(80, 1)
	output := stripANSI(list.View())
	if !strings.Contains(output, "Projects") {
		t.Errorf("collapsed bar = %q, want 'Projects'", output)
	}
	if !strings.Contains(output, "1 of 2") {
		t.Errorf("collapsed bar = %q, want '1 of 2' footer", output)
	}
}

func TestProjectList_View_ShowsProjectKeys(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{
		{Key: "AAA", Name: "Alpha"},
		{Key: "BBB", Name: "Beta"},
	})
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, "AAA") {
		t.Errorf("View() = %q, want to contain AAA", output)
	}
	if !strings.Contains(output, "Beta") {
		t.Errorf("View() = %q, want to contain project name Beta", output)
	}
}

func TestProjectList_View_ShowsActiveMarker(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA"}, {Key: "BBB"}})
	list.SetActiveKey("AAA")
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, "*") {
		t.Errorf("View() = %q, want * active marker", output)
	}
}

func TestProjectList_View_ShowsLead(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{
		{Key: "AAA", Name: "Alpha", Lead: &jira.User{DisplayName: "Alice"}},
	})
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, "Alice") {
		t.Errorf("View() = %q, want to contain lead name Alice", output)
	}
}

func TestProjectList_View_FooterShowsCount(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA"}, {Key: "BBB"}})
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if !strings.Contains(output, "1 of 2") {
		t.Errorf("View() = %q, want '1 of 2' footer", output)
	}
}

func TestProjectList_View_EmptyNoFooter(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetSize(80, 24)
	output := stripANSI(list.View())
	if strings.Contains(output, "of") {
		t.Errorf("empty list View() = %q, should not contain footer", output)
	}
}

func TestProjectList_View_LongNameTruncated(t *testing.T) {
	t.Parallel()
	longName := strings.Repeat("X", 200)
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA", Name: longName}})
	list.SetSize(40, 24)
	output := stripANSI(list.View())
	if strings.Contains(output, longName) {
		t.Errorf("expected long name to be truncated in View()")
	}
}
