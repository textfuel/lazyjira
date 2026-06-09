package tui

import (
	"context"
	"errors"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func TestFetchProjects(t *testing.T) {
	t.Parallel()

	t.Run("success returns projectsLoadedMsg", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetProjectsFunc = func(context.Context) ([]jira.Project, error) {
			return []jira.Project{{Key: testProject}}, nil
		}

		msg := fetchProjects(fake)()

		loaded, ok := msg.(projectsLoadedMsg)
		if !ok {
			t.Fatalf("msg = %T, want projectsLoadedMsg", msg)
		}
		if len(loaded.projects) != 1 || loaded.projects[0].Key != testProject {
			t.Errorf("projects = %+v", loaded.projects)
		}
		if len(fake.GetProjectsCalls) != 1 {
			t.Errorf("GetProjects called %d times, want 1", len(fake.GetProjectsCalls))
		}
	})

	t.Run("error returns errorMsg", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.GetProjectsFunc = func(context.Context) ([]jira.Project, error) {
			return nil, errors.New("network down")
		}

		msg := fetchProjects(fake)()
		if _, ok := msg.(errorMsg); !ok {
			t.Fatalf("msg = %T, want errorMsg", msg)
		}
	})
}

func TestFetchTransitions_PassesIssueKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetTransitionsFunc = func(_ context.Context, _ string) ([]jira.Transition, error) {
		return []jira.Transition{{ID: "11", Name: "To Do"}}, nil
	}

	msg := fetchTransitions(fake, testKey)()

	loaded, ok := msg.(transitionsLoadedMsg)
	if !ok {
		t.Fatalf("msg = %T, want transitionsLoadedMsg", msg)
	}
	if loaded.issueKey != testKey {
		t.Errorf("issueKey = %q, want PLAT-1", loaded.issueKey)
	}
	if len(fake.GetTransitionsCalls) != 1 || fake.GetTransitionsCalls[0].Key != testKey {
		t.Errorf("GetTransitions calls = %+v", fake.GetTransitionsCalls)
	}
}

func TestFetchPriorities_Success(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetPrioritiesFunc = func(context.Context) ([]jira.Priority, error) {
		return []jira.Priority{{ID: "2", Name: "High"}}, nil
	}

	msg := fetchPriorities(fake)()

	loaded, ok := msg.(prioritiesLoadedMsg)
	if !ok {
		t.Fatalf("msg = %T, want prioritiesLoadedMsg", msg)
	}
	if len(loaded.priorities) != 1 || loaded.priorities[0].Name != "High" {
		t.Errorf("priorities = %+v", loaded.priorities)
	}
}

func TestDoTransition(t *testing.T) {
	t.Parallel()

	t.Run("success returns transitionDoneMsg and records call", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.DoTransitionFunc = func(_ context.Context, _, _ string) error { return nil }

		msg := doTransition(fake, testKey, "21")()

		if _, ok := msg.(transitionDoneMsg); !ok {
			t.Fatalf("msg = %T, want transitionDoneMsg", msg)
		}
		if len(fake.DoTransitionCalls) != 1 {
			t.Fatalf("DoTransition called %d times, want 1", len(fake.DoTransitionCalls))
		}
		call := fake.DoTransitionCalls[0]
		if call.Key != testKey || call.TransitionID != "21" {
			t.Errorf("call = %+v, want key=PLAT-1 id=21", call)
		}
	})

	t.Run("error returns errorMsg", func(t *testing.T) {
		t.Parallel()
		fake := &jiratest.FakeClient{T: t}
		fake.DoTransitionFunc = func(_ context.Context, _, _ string) error { return errors.New("forbidden") }

		if _, ok := doTransition(fake, testKey, "21")().(errorMsg); !ok {
			t.Error("want errorMsg on failed transition")
		}
	})
}

func TestFetchLabels_Success(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetLabelsFunc = func(context.Context) ([]string, error) {
		return []string{"backend", "frontend"}, nil
	}

	msg := fetchLabels(fake)()

	loaded, ok := msg.(labelsLoadedMsg)
	if !ok {
		t.Fatalf("msg = %T, want labelsLoadedMsg", msg)
	}
	if len(loaded.labels) != 2 {
		t.Errorf("labels = %+v", loaded.labels)
	}
}

func TestHandlePrioritiesLoaded_ShowsModalAndSetsHandler(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	_, cmd := app.handlePrioritiesLoaded(prioritiesLoadedMsg{priorities: []jira.Priority{{ID: "2", Name: "High"}}})

	if cmd != nil {
		t.Errorf("expected nil cmd, got %T", cmd)
	}
	if !app.modal.IsVisible() {
		t.Error("priority modal should be visible")
	}
	if app.onSelect == nil {
		t.Error("onSelect handler should be set")
	}
}

func TestHandlePrioritiesLoaded_EmptyIsNoop(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	_, _ = app.handlePrioritiesLoaded(prioritiesLoadedMsg{})

	if app.modal.IsVisible() {
		t.Error("modal should stay hidden for empty priorities")
	}
}

func TestHandleProjectsLoaded_SelectsFirstWhenNoneActive(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.demoMode = true
	app.issuesList.SetTabs([]config.IssueTabConfig{{Name: "All", JQL: "project = {{.ProjectKey}} ORDER BY updated DESC"}})

	_, cmd := app.handleProjectsLoaded(projectsLoadedMsg{projects: []jira.Project{{Key: testProject, ID: "1"}}})

	if app.projectKey != testProject {
		t.Errorf("projectKey = %q, want PLAT", app.projectKey)
	}
	if cmd == nil {
		t.Error("expected a fetchActiveTab command after auto-selecting a project")
	}
	if len(app.projectList.AllProjects()) != 1 {
		t.Errorf("projects not stored: %+v", app.projectList.AllProjects())
	}
}

func TestUpdate_MyselfLoadedSetsCurrentUser(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)

	_, _ = app.Update(myselfLoadedMsg{user: &jira.User{AccountID: "me-1"}})

	if app.currentUser == nil || app.currentUser.AccountID != "me-1" {
		t.Errorf("currentUser = %+v, want AccountID=me-1", app.currentUser)
	}
}
