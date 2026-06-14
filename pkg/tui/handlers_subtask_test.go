package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestCanCreateSubtask(t *testing.T) {
	t.Parallel()

	standard := &jira.IssueType{ID: "1", Name: "Story", HierarchyLevel: 0}
	subtask := &jira.IssueType{ID: "2", Name: "Sub-task", Subtask: true}
	epic := &jira.IssueType{ID: "3", Name: "Epic", HierarchyLevel: 1}

	cases := []struct {
		name  string
		setup func(*App)
		want  bool
	}{
		{"standard issue", func(a *App) {
			a.issuesList.SetIssues([]jira.Issue{{Key: testKey, IssueType: standard}})
		}, true},
		{"unknown type allowed", func(a *App) {
			a.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
		}, true},
		{"subtask excluded", func(a *App) {
			a.issuesList.SetIssues([]jira.Issue{{Key: testKey, IssueType: subtask}})
		}, false},
		{"epic excluded", func(a *App) {
			a.issuesList.SetIssues([]jira.Issue{{Key: testKey, IssueType: epic}})
		}, false},
		{"board project irrelevant", func(a *App) {
			// C4: the gate no longer depends on the board project; the parent
			// carries its own project.
			a.projectKey = ""
			a.issuesList.SetIssues([]jira.Issue{{Key: testKey, IssueType: standard}})
		}, true},
		{"no selection", func(a *App) {}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := focusApp(t)
			app.side = sideLeft
			app.leftFocus = focusIssues
			app.projectKey = testProject
			tc.setup(app)

			if got := app.canCreateSubtask(); got != tc.want {
				t.Errorf("canCreateSubtask() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanCreateSubtask_WrongContext(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.side = sideRight
	app.leftFocus = focusIssues
	app.projectKey = testProject
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	if app.canCreateSubtask() {
		t.Error("canCreateSubtask should be false when issues panel is not focused")
	}
}

func TestStartCreateSubtask(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.projectKey = testProject
	app.projectID = "10000"
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "10000"}})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	_, cmd := app.startCreateSubtask()

	if cmd == nil {
		t.Fatal("expected issue-types fetch command")
	}
	if app.createCtx.parentKey != testKey {
		t.Errorf("parentKey = %q, want %q", app.createCtx.parentKey, testKey)
	}
	if !app.createCtx.intent {
		t.Error("intent should be set")
	}
}

func TestStartCreateSubtask_RequiresSelection(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.projectKey = testProject

	if _, cmd := app.startCreateSubtask(); cmd != nil {
		t.Error("expected nil cmd without a selected parent issue")
	}
}

func TestProjectKeyFromIssueKey(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"PLAT-1":      "PLAT",
		"DSOTEST-123": "DSOTEST",
		"X-1":         "X",
		"NODASH":      "NODASH",
	}
	for in, want := range cases {
		if got := projectKeyFromIssueKey(in); got != want {
			t.Errorf("projectKeyFromIssueKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStartCreateSubtask_ProjectFromParent(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	// Board sits on PLAT, but the selected parent lives in DSOTEST (cross-project
	// JQL tab). The subtask must target the parent's project, not the board's.
	app.projectKey = "PLAT"
	app.projectID = "10000"
	app.projectList.SetProjects([]jira.Project{
		{Key: "PLAT", ID: "10000"},
		{Key: "DSOTEST", ID: "20000"},
	})
	app.issuesList.SetIssues([]jira.Issue{{Key: "DSOTEST-5"}})

	_, cmd := app.startCreateSubtask()

	if cmd == nil {
		t.Fatal("expected issue-types fetch command")
	}
	if app.createCtx.projectKey != "DSOTEST" {
		t.Errorf("projectKey = %q, want DSOTEST (parent's project)", app.createCtx.projectKey)
	}
	if app.createCtx.projectID != "20000" {
		t.Errorf("projectID = %q, want 20000 (resolved from parent project)", app.createCtx.projectID)
	}
	if app.createCtx.parentKey != "DSOTEST-5" {
		t.Errorf("parentKey = %q, want DSOTEST-5", app.createCtx.parentKey)
	}
}

func TestStartCreateSubtask_UnresolvableProjectAborts(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.projectKey = "PLAT"
	app.projectList.SetProjects([]jira.Project{{Key: "PLAT", ID: "10000"}})
	app.issuesList.SetIssues([]jira.Issue{{Key: "OTHER-9"}}) // OTHER not in project list

	_, cmd := app.startCreateSubtask()

	if cmd != nil {
		t.Error("expected abort (nil cmd) when the parent project cannot be resolved")
	}
	if app.statusPanel.ErrorMessage() == "" {
		t.Error("expected a status error on unresolvable parent project")
	}
	if app.createCtx.intent {
		t.Error("createCtx should not be armed on abort")
	}
}

func TestCanCreateSubtask_SubTabSurface(t *testing.T) {
	t.Parallel()

	standard := &jira.IssueType{ID: "1", Name: "Story", HierarchyLevel: 0}
	subtask := &jira.IssueType{ID: "2", Name: "Sub-task", Subtask: true}

	cases := []struct {
		name  string
		setup func(*App)
		want  bool
	}{
		{"epic child is a valid parent", func(a *App) {
			a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1",
				Subtasks: []jira.Issue{{Key: "STORY-2", IssueType: standard}}})
			a.infoPanel.SetActiveTab(views.InfoTabSubtasks)
		}, true},
		{"subtask child excluded", func(a *App) {
			a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1",
				Subtasks: []jira.Issue{{Key: "SUB-2", IssueType: subtask}}})
			a.infoPanel.SetActiveTab(views.InfoTabSubtasks)
		}, false},
		{"unknown child type allowed", func(a *App) {
			a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1",
				Subtasks: []jira.Issue{{Key: "X-2"}}})
			a.infoPanel.SetActiveTab(views.InfoTabSubtasks)
		}, true},
		{"non-subtask tab excluded", func(a *App) {
			a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1",
				Subtasks: []jira.Issue{{Key: "STORY-2", IssueType: standard}}})
			a.infoPanel.SetActiveTab(views.InfoTabLinks)
		}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			app := focusApp(t)
			app.side = sideLeft
			app.leftFocus = focusInfo
			tc.setup(app)

			if got := app.canCreateSubtask(); got != tc.want {
				t.Errorf("canCreateSubtask() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStartCreateSubtask_FromSubTab(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.side = sideLeft
	app.leftFocus = focusInfo
	app.projectList.SetProjects([]jira.Project{{Key: "DSOTEST", ID: "20000"}})
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1",
		Subtasks: []jira.Issue{{Key: "DSOTEST-7"}}})
	app.infoPanel.SetActiveTab(views.InfoTabSubtasks)

	_, cmd := app.startCreateSubtask()

	if cmd == nil {
		t.Fatal("expected issue-types fetch command")
	}
	if app.createCtx.parentKey != "DSOTEST-7" {
		t.Errorf("parentKey = %q, want DSOTEST-7 (selected sub-tab child)", app.createCtx.parentKey)
	}
	if app.createCtx.projectKey != "DSOTEST" {
		t.Errorf("projectKey = %q, want DSOTEST", app.createCtx.projectKey)
	}
}

func TestHandleIssueTypesLoaded_SubtaskFilter(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.modal.SetSize(120, 40)
	app.createCtx = createCtx{intent: true, parentKey: testKey}

	_, _ = app.handleIssueTypesLoaded(issueTypesLoadedMsg{issueTypes: []jira.IssueType{
		{ID: "1", Name: "Story"},
		{ID: "2", Name: "Sub-task", Subtask: true},
	}})

	view := app.modal.View()
	if !strings.Contains(view, "Sub-task") {
		t.Errorf("subtask type should be offered, view:\n%s", view)
	}
	if strings.Contains(view, "Story") {
		t.Errorf("standard type should be filtered out, view:\n%s", view)
	}
	if !strings.Contains(view, "Select subtask type") {
		t.Errorf("title should mark subtask selection, view:\n%s", view)
	}
}

func TestHandleIssueTypesLoaded_NonSubtaskFilter(t *testing.T) {
	t.Parallel()
	app := focusApp(t)
	app.modal.SetSize(120, 40)
	app.createCtx = createCtx{intent: true}

	_, _ = app.handleIssueTypesLoaded(issueTypesLoadedMsg{issueTypes: []jira.IssueType{
		{ID: "1", Name: "Story"},
		{ID: "2", Name: "Sub-task", Subtask: true},
	}})

	view := app.modal.View()
	if !strings.Contains(view, "Story") {
		t.Errorf("standard type should be offered, view:\n%s", view)
	}
	if strings.Contains(view, "Sub-task") {
		t.Errorf("subtask type should be filtered out without a parent, view:\n%s", view)
	}
	if !strings.Contains(view, "Select issue type") {
		t.Errorf("title should mark plain issue selection, view:\n%s", view)
	}
}

func TestHandleCreateFormSubmit_InjectsParent(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.CreateIssueFunc = func(context.Context, map[string]any) (*jira.Issue, error) {
		return &jira.Issue{Key: "PLAT-99"}, nil
	}
	app := newAppWithFake(t, fake)
	app.createCtx = createCtx{projectKey: testProject, issueTypeID: "10001", parentKey: testKey}

	fields := map[string]any{"summary": "hi"}
	_, _ = app.handleCreateFormSubmit(components.CreateFormSubmitMsg{Fields: fields})

	parent, ok := fields["parent"].(map[string]string)
	if !ok || parent["key"] != testKey {
		t.Errorf("parent field not injected: %v", fields["parent"])
	}
}
