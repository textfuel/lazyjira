package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func TestHandleTransitionsLoaded(t *testing.T) {
	t.Parallel()

	t.Run("shows modal and installs handler", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

		_, cmd := app.handleTransitionsLoaded(transitionsLoadedMsg{
			issueKey:    testKey,
			transitions: []jira.Transition{{ID: "11", Name: "Done", To: &jira.Status{Name: "Closed"}}},
		})

		if cmd != nil {
			t.Errorf("expected nil cmd")
		}
		if !app.modal.IsVisible() {
			t.Error("transition modal should be visible")
		}
		if app.onSelect == nil {
			t.Error("onSelect should be set")
		}
	})

	t.Run("empty is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, _ = app.handleTransitionsLoaded(transitionsLoadedMsg{issueKey: testKey})
		if app.modal.IsVisible() {
			t.Error("modal should stay hidden for no transitions")
		}
	})
}

func TestHandleBoardsLoaded_ResolvesBoardForProject(t *testing.T) {
	t.Parallel()

	t.Run("matching project sets board id", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.projectKey = testProject

		_, _ = app.handleBoardsLoaded(boardsLoadedMsg{boards: []jira.Board{{ID: 7, ProjectKey: testProject}}})

		testkit.AssertEqual(t, "boardID", app.boardID, 7)
	})

	t.Run("no matching project leaves board id zero", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.projectKey = testProject

		_, _ = app.handleBoardsLoaded(boardsLoadedMsg{boards: []jira.Board{{ID: 7, ProjectKey: "OPS"}}})

		testkit.AssertEqual(t, "boardID", app.boardID, 0)
	})
}

func TestHandleSprintsLoaded(t *testing.T) {
	t.Parallel()

	t.Run("shows modal for selected issue", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

		_, _ = app.handleSprintsLoaded(sprintsLoadedMsg{sprints: []jira.Sprint{
			{ID: 1, Name: "Sprint 1", State: "active"},
			{ID: 2, Name: "Old", State: "closed"},
		}})

		if !app.modal.IsVisible() {
			t.Error("sprint modal should be visible")
		}
		if app.onSelect == nil {
			t.Error("onSelect should be set")
		}
	})

	t.Run("no selected issue is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, _ = app.handleSprintsLoaded(sprintsLoadedMsg{sprints: []jira.Sprint{{ID: 1}}})
		if app.modal.IsVisible() {
			t.Error("modal should stay hidden without a selected issue")
		}
	})
}

func TestHandleLabelsLoaded_ShowsChecklist(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Labels: []string{"backend"}}})

	_, _ = app.handleLabelsLoaded(labelsLoadedMsg{labels: []string{"backend", "frontend"}})

	if !app.modal.IsVisible() {
		t.Error("labels checklist should be visible")
	}
}

func TestHandleComponentsLoaded_ShowsChecklist(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	_, _ = app.handleComponentsLoaded(componentsLoadedMsg{components: []jira.Component{{ID: "10", Name: "backend"}}})

	if !app.modal.IsVisible() {
		t.Error("components checklist should be visible")
	}
}

func TestHandleIssuePrefetched_CachesIssue(t *testing.T) {
	t.Parallel()

	t.Run("caches the issue", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})

		_, _ = app.handleIssuePrefetched(issuePrefetchedMsg{issue: &jira.Issue{Key: testKey}})

		if app.issueCache[testKey] == nil {
			t.Error("issue not cached")
		}
	})

	t.Run("nil issue is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, _ = app.handleIssuePrefetched(issuePrefetchedMsg{})
		if len(app.issueCache) != 0 {
			t.Error("nil issue should not populate cache")
		}
	})
}

func TestHandleBatchPrefetched_CachesAll(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})

	_, _ = app.handleBatchPrefetched(batchPrefetchedMsg{issues: []jira.Issue{{Key: "A-1"}, {Key: "B-2"}}})

	if app.issueCache["A-1"] == nil || app.issueCache["B-2"] == nil {
		t.Errorf("batch not fully cached: %v", app.issueCache)
	}
}

func TestHandleTransitionDone(t *testing.T) {
	t.Parallel()

	t.Run("refetches for selected issue", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

		_, cmd := app.handleTransitionDone()
		if cmd == nil {
			t.Error("expected refetch command after transition")
		}
	})

	t.Run("no selection is noop", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		_, cmd := app.handleTransitionDone()
		if cmd != nil {
			t.Error("expected nil cmd without a selected issue")
		}
	})
}

func TestHandleUsersLoaded_ShowsAssigneeModal(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.usersCache = map[string][]jira.User{}
	app.projectKey = testProject
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})

	_, _ = app.handleUsersLoaded(usersLoadedMsg{
		users:    []jira.User{{AccountID: "u1", DisplayName: "Ann"}},
		issueKey: testKey,
	})

	if !app.modal.IsVisible() {
		t.Error("assignee modal should be visible")
	}
	if len(app.usersCache[testProject]) != 1 {
		t.Errorf("users not cached: %v", app.usersCache)
	}
}

func TestBuildUserItems(t *testing.T) {
	t.Parallel()

	t.Run("prepends me and None and dedups self", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})
		app.currentUser = &jira.User{AccountID: "me", DisplayName: "Me"}

		items := app.buildUserItems([]jira.User{{AccountID: "me", DisplayName: "Me"}, {AccountID: "u1", DisplayName: "Ann"}})

		if len(items) != 3 {
			t.Fatalf("items = %d, want 3 (me, None, Ann)", len(items))
		}
		testkit.AssertEqual(t, "me label", items[0].Label, "Me (me)")
		testkit.AssertEqual(t, "None label", items[1].Label, "None")
		testkit.AssertEqual(t, "other label", items[2].Label, "Ann")
	})

	t.Run("without current user starts with None", func(t *testing.T) {
		t.Parallel()
		app := newAppWithFake(t, &jiratest.FakeClient{T: t})

		items := app.buildUserItems([]jira.User{{AccountID: "u1", DisplayName: "Ann"}})

		if len(items) != 2 {
			t.Fatalf("items = %d, want 2 (None, Ann)", len(items))
		}
		testkit.AssertEqual(t, "first label", items[0].Label, "None")
	})
}
