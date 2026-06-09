package views

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestProjectList_SetAndSelect(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA", Name: "Alpha"}, {Key: "BBB", Name: "Beta"}})

	testkit.AssertEqual(t, "all projects", len(list.AllProjects()), 2)
	if sel := list.SelectedProject(); sel == nil || sel.Key != "AAA" {
		t.Errorf("selected = %v, want AAA", sel)
	}
}

func TestProjectList_PinActiveMovesToTop(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA"}, {Key: "BBB"}, {Key: "CCC"}})

	list.SetActiveKey("CCC")

	if sel := list.SelectedProject(); sel == nil || sel.Key != "CCC" {
		t.Errorf("active project should be pinned to top, got %v", sel)
	}
}

func TestProjectList_Filter(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	list.SetProjects([]jira.Project{{Key: "AAA", Name: "Alpha"}, {Key: "BBB", Name: "Beta"}})

	list.SetFilter("beta")

	if sel := list.SelectedProject(); sel == nil || sel.Key != "BBB" {
		t.Errorf("filtered selection = %v, want BBB", sel)
	}

	list.SetFilter("")
	testkit.AssertEqual(t, "restored count", len(list.AllProjects()), 2)
}

func TestProjectList_SelectedProjectNilWhenEmpty(t *testing.T) {
	t.Parallel()
	list := NewProjectList()
	if sel := list.SelectedProject(); sel != nil {
		t.Errorf("expected nil selection, got %v", sel)
	}
}
