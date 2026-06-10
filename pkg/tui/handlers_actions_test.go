package tui

import (
	"context"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func actionApp(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	return app
}

func TestHandleActionEdit_IssuePanelShowsInputModal(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	issue := &jira.Issue{Key: testKey, Summary: testSummary}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue

	_, _ = app.handleActionEdit()

	if !app.inputModal.IsVisible() {
		t.Error("input modal should be visible after edit on issues panel")
	}
	testkit.AssertEqual(t, "editContext kind", app.editContext.kind, editSummary)
	testkit.AssertEqual(t, "editContext issueKey", app.editContext.issueKey, testKey)
}

func TestHandleActionEdit_NoIssueIsNoop(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, cmd := app.handleActionEdit()

	if cmd != nil {
		t.Error("expected nil cmd when no issue selected")
	}
	if app.inputModal.IsVisible() {
		t.Error("input modal should not be visible when no issue selected")
	}
}

func TestHandleActionEdit_InfoPanelLinksTabEditsPreviewedSummary(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusInfo
	issue := &jira.Issue{Key: testKey, Summary: testSummary}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue
	app.infoPanel.SetIssue(issue)
	app.infoPanel.SetActiveTab(views.InfoTabLinks)

	_, _ = app.handleActionEdit()

	if !app.inputModal.IsVisible() {
		t.Error("input modal should be visible on links tab edit")
	}
}

func TestHandleActionEdit_DetailCommentsTabLaunchesEditor(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideRight
	issue := &jira.Issue{
		Key:      testKey,
		Summary:  testSummary,
		Comments: []jira.Comment{{ID: "99", Body: "comment text"}},
	}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue
	app.detailView.SetIssue(issue)
	app.detailView.SetActiveTab(views.TabComments)

	_, cmd := app.handleActionEdit()

	if cmd == nil {
		t.Error("expected editor launch cmd on comments tab edit")
	}
	testkit.AssertEqual(t, "editContext kind", app.editContext.kind, editCommentMod)
}

func TestHandleActionEdit_DetailDescriptionLaunchesEditor(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideRight
	issue := &jira.Issue{Key: testKey, Summary: testSummary, Description: "some desc"}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue
	app.detailView.SetIssue(issue)
	app.detailView.SetActiveTab(views.TabDetails)

	_, cmd := app.handleActionEdit()

	if cmd == nil {
		t.Error("expected editor launch cmd for description edit")
	}
	testkit.AssertEqual(t, "editContext kind", app.editContext.kind, editDesc)
}

func TestHandleActionSelect_IssuesSwitchesToDetailSide(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: testKey, Summary: testSummary})
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: testSummary}})

	_, _ = app.handleActionSelect()

	testkit.AssertEqual(t, "side after select", app.side, sideRight)
}

func TestHandleActionSelect_ProjectsOpensProject(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusProjects
	app.issuesList.SetTabs([]config.IssueTabConfig{})
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "10000", Name: "Test Project"}})

	_, _ = app.handleActionSelect()

	testkit.AssertEqual(t, "projectKey after select", app.projectKey, testProject)
	testkit.AssertEqual(t, "leftFocus after select", app.leftFocus, focusIssues)
}

func TestHandleActionOpen_IssuesSwitchesToDetail(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: testKey, Summary: testSummary})
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey, Summary: testSummary}})

	_, _ = app.handleActionOpen()

	testkit.AssertEqual(t, "side after open", app.side, sideRight)
}

func TestHandleActionOpen_ProjectsOpensProject(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusProjects
	app.issuesList.SetTabs([]config.IssueTabConfig{})
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "10000", Name: "Test Project"}})

	_, _ = app.handleActionOpen()

	testkit.AssertEqual(t, "projectKey after open", app.projectKey, testProject)
}

func TestHandleActionURLPicker_ShowsModalWhenURLsExist(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	issue := &jira.Issue{
		Key:         testKey,
		Summary:     testSummary,
		Description: "see https://example.atlassian.net/browse/PROJ-1 for details",
	}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue

	_, _ = app.handleActionURLPicker()
}

func TestHandleActionURLPicker_NoopWhenNoIssue(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues

	_, cmd := app.handleActionURLPicker()

	if cmd != nil {
		t.Error("expected nil cmd with no issue selected")
	}
}

func TestHandleActionCreateBranch_RequiresGitRepo(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.gitRepoPath = ""
	issue := &jira.Issue{Key: testKey, Summary: testSummary}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue

	_, _ = app.handleActionCreateBranch()

	if app.inputModal.IsVisible() {
		t.Error("input modal should not show when gitRepoPath is empty")
	}
}

func TestHandleActionCreateBranch_ShowsInputModalWithSuggestedName(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusIssues
	app.gitRepoPath = t.TempDir()
	issue := &jira.Issue{Key: testKey, Summary: testSummary}
	app.issuesList.SetIssues([]jira.Issue{*issue})
	app.previewKey = testKey
	app.issueCache[testKey] = issue

	_, _ = app.handleActionCreateBranch()

	if !app.inputModal.IsVisible() {
		t.Error("input modal should be visible for branch creation")
	}
	testkit.AssertEqual(t, "editContext kind", app.editContext.kind, editBranch)
}

func TestHandleActionCreateBranch_WrongFocusIsNoop(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.side = sideLeft
	app.leftFocus = focusInfo
	app.gitRepoPath = t.TempDir()

	_, cmd := app.handleActionCreateBranch()

	if cmd != nil {
		t.Error("expected nil cmd when not focused on issues")
	}
	if app.inputModal.IsVisible() {
		t.Error("input modal should not show when not on issues panel")
	}
}

func TestOpenProject_UpdatesProjectKey(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.issuesList.SetTabs([]config.IssueTabConfig{})
	app.projectList.SetProjects([]jira.Project{{Key: testProject, ID: "10000", Name: "Test"}})

	_, _ = app.openProject()

	testkit.AssertEqual(t, "projectKey", app.projectKey, testProject)
	testkit.AssertEqual(t, "leftFocus", app.leftFocus, focusIssues)
}

func TestOpenProject_NoopWithNoSelection(t *testing.T) {
	t.Parallel()
	app := actionApp(t)

	_, cmd := app.openProject()

	if cmd != nil {
		t.Error("expected nil cmd with no project selected")
	}
}

func TestInfoPanelSelectedKey_ReturnsEmptyWithNoSelection(t *testing.T) {
	t.Parallel()
	app := actionApp(t)

	key := app.infoPanelSelectedKey()

	if key != "" {
		t.Errorf("expected empty key with no selection, got %q", key)
	}
}

func TestNavigateToLinkedIssue_SelectsFromCacheWhenNotInList(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: mainKey, Summary: "linked issue"})
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	linkedIssue := &jira.Issue{Key: mainKey, Summary: "linked"}
	app.issueCache[mainKey] = linkedIssue
	app.infoPanel.SetIssue(&jira.Issue{
		Key:     testKey,
		Summary: testSummary,
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "relates to"},
				OutwardIssue: &jira.Issue{Key: mainKey},
			},
		},
	})
	app.infoPanel.SetActiveTab(views.InfoTabLinks)

	_, cmd := app.navigateToLinkedIssue()

	if cmd == nil {
		t.Error("expected fetch cmd for linked issue")
	}
}

func TestStartCreateIssue_RequiresProjectKey(t *testing.T) {
	t.Parallel()
	app := actionApp(t)
	app.projectKey = ""

	_, cmd := app.startCreateIssue()

	if cmd != nil {
		t.Error("expected nil cmd with no project key")
	}
}

func TestStartCreateIssue_WithProjectKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueTypesFunc = func(_ context.Context, _ string) ([]jira.IssueType, error) {
		return []jira.IssueType{{ID: "1", Name: "Story"}}, nil
	}
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.projectKey = testProject
	app.projectID = "10000"

	_, cmd := app.startCreateIssue()

	if cmd == nil {
		t.Error("expected a fetch issue types cmd")
	}
	testkit.AssertEqual(t, "createCtx.intent", app.createCtx.intent, true)
	testkit.AssertEqual(t, "createCtx.projectKey", app.createCtx.projectKey, testProject)
}
