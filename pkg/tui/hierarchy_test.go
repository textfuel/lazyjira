package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
	"github.com/textfuel/lazyjira/v2/pkg/tui/navstack"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

const hierarchyTitleChildren = "Children"

// Enter on issue with Subtasks creates hierarchy tab populated
// from the in-memory Subtasks slice; no API call.
func TestHierarchy_EnterWithSubtasks_CreatesHierarchyTab(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)

	parent := jira.Issue{
		Key: "PARENT-1",
		Subtasks: []jira.Issue{
			{Key: "SUB-1", Summary: "first"},
			{Key: "SUB-2", Summary: "second"},
		},
	}
	a.issuesList.SetIssues([]jira.Issue{parent})

	_, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false, want true (subtasks present)")
	}
	if !a.issuesList.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = false after showChildren")
	}
	if got := a.issuesList.HierarchyTitle(); got != hierarchyTitleChildren {
		t.Errorf("HierarchyTitle() = %q, want %q", got, hierarchyTitleChildren)
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "SUB-1" {
		t.Errorf("SelectedIssue() = %+v, want SUB-1", sel)
	}
	if d := a.issuesList.HierarchyStack().Depth(); d != 1 {
		t.Errorf("HierarchyStack.Depth() = %d, want 1", d)
	}
	if len(fake.SearchIssuesCalls) != 0 {
		t.Errorf("SearchIssues called %d times, want 0 (Subtasks path)", len(fake.SearchIssuesCalls))
	}
}

// Enter on leaf issue (no Subtasks) does not create a hierarchy tab;
// handleActionSelect falls through to openIssueDetail which focuses the
// right side. SearchIssues is never called (lookup is sync via
// Issue.Subtasks only).
func TestHierarchy_EnterWithoutChildren_FocusesDetail(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.issuesList.SetIssues([]jira.Issue{{Key: "LEAF-1"}})

	_, _ = a.handleActionSelect()

	if a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = true, want false (no children)")
	}
	if a.side != sideRight {
		t.Errorf("a.side = %v after Enter on leaf, want sideRight", a.side)
	}
	if got := len(fake.SearchIssuesCalls); got != 0 {
		t.Errorf("SearchIssues calls = %d, want 0 (no API for children lookup)", got)
	}
}

// Enter from InfoPanel Subtasks tab creates a hierarchy tab
// containing the single selected subtask issue (title hierarchyTitleChildren).
func TestHierarchy_EnterFromInfoPanelSub_OneElementList(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.leftFocus = focusInfo

	issue := &jira.Issue{
		Key:      "MAIN-1",
		Subtasks: []jira.Issue{{Key: "SUB-7", Summary: "x"}},
	}
	a.infoPanel.SetIssue(issue)
	// SetIssue resets to InfoTabFields; advance to Subtasks.
	a.infoPanel.NextTab() // Links
	a.infoPanel.NextTab() // Subtasks
	if a.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		t.Fatalf("precondition: ActiveTab = %v, want InfoTabSubtasks", a.infoPanel.ActiveTab())
	}

	_, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}
	if got := a.issuesList.HierarchyTitle(); got != hierarchyTitleChildren {
		t.Errorf("HierarchyTitle() = %q, want %q", got, hierarchyTitleChildren)
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "SUB-7" {
		t.Errorf("SelectedIssue() = %+v, want SUB-7", sel)
	}
	if d := a.issuesList.HierarchyStack().Depth(); d != 1 {
		t.Errorf("HierarchyStack.Depth() = %d, want 1", d)
	}
}

// Enter from InfoPanel Links tab creates a hierarchy tab with
// the single linked issue (title "Link").
func TestHierarchy_EnterFromInfoPanelLink_OneElementList(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.leftFocus = focusInfo

	issue := &jira.Issue{
		Key: "MAIN-1",
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "Blocks"},
				OutwardIssue: &jira.Issue{Key: "LNK-9"},
			},
		},
	}
	a.infoPanel.SetIssue(issue)
	a.infoPanel.NextTab() // Links
	if a.infoPanel.ActiveTab() != views.InfoTabLinks {
		t.Fatalf("precondition: ActiveTab = %v, want InfoTabLinks", a.infoPanel.ActiveTab())
	}

	_, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}
	if got := a.issuesList.HierarchyTitle(); got != "Link" {
		t.Errorf("HierarchyTitle() = %q, want %q", got, "Link")
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "LNK-9" {
		t.Errorf("SelectedIssue() = %+v, want LNK-9", sel)
	}
}

// The hierarchy row inherits the summary already attached to the link
// payload, so the user sees the issue title and not just the key while
// the detail fetch is still in flight.
func TestHierarchy_LinkRowCarriesSummaryFromInfoPanel(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.leftFocus = focusInfo

	issue := &jira.Issue{
		Key: "MAIN-1",
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "Blocks"},
				OutwardIssue: &jira.Issue{Key: "LNK-9", Summary: "linked work"},
			},
		},
	}
	a.infoPanel.SetIssue(issue)
	a.infoPanel.NextTab() // Links

	if _, handled := a.showChildren(); !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}
	sel := a.issuesList.SelectedIssue()
	if sel == nil {
		t.Fatalf("SelectedIssue() = nil")
	}
	if sel.Summary != "linked work" {
		t.Errorf("Summary = %q, want %q (info-panel payload should be preserved)", sel.Summary, "linked work")
	}
}

// A pre-cached version of the targeted issue wins over the InfoPanel
// stub: the hierarchy row reflects whatever fields the cache holds,
// even when the in-flight fetch hasn't returned yet.
func TestHierarchy_LinkRowPrefersCacheOverStub(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.leftFocus = focusInfo

	const cachedSummary = "rich cached summary"
	a.issueCache["LNK-7"] = &jira.Issue{Key: "LNK-7", Summary: cachedSummary}

	issue := &jira.Issue{
		Key: "MAIN-2",
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "Relates"},
				OutwardIssue: &jira.Issue{Key: "LNK-7", Summary: "stale summary"},
			},
		},
	}
	a.infoPanel.SetIssue(issue)
	a.infoPanel.NextTab() // Links

	if _, handled := a.showChildren(); !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}
	sel := a.issuesList.SelectedIssue()
	if sel == nil {
		t.Fatalf("SelectedIssue() = nil")
	}
	if sel.Summary != cachedSummary {
		t.Errorf("Summary = %q, want %q (cache should win over info-panel stub)", sel.Summary, cachedSummary)
	}
}

// Backspace on issue with Parent dispatches an async
// fetchParent cmd. Executing the cmd yields a parentLoadedMsg.
// Dispatching that msg through Update creates the hierarchy tab.
func TestHierarchy_BackspaceWithParent_ShowsParent(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, key string) (*jira.Issue, error) {
		return &jira.Issue{Key: key, Summary: "parent issue"}, nil
	}
	a := newAppWithFake(t, fake)
	a.issuesList.SetIssues([]jira.Issue{
		{Key: "CHILD-1", Parent: &jira.Issue{Key: "PARENT-1"}},
	})

	cmd, handled := a.showParent()
	if !handled {
		t.Fatalf("showParent() handled = false, want true")
	}
	if cmd == nil {
		t.Fatalf("showParent() cmd = nil, want async fetch cmd")
	}
	if a.issuesList.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = true before msg dispatch, want false")
	}

	msg := cmd()
	loaded, ok := msg.(parentLoadedMsg)
	if !ok {
		t.Fatalf("cmd produced %T, want parentLoadedMsg", msg)
	}
	if loaded.parent == nil || loaded.parent.Key != "PARENT-1" {
		t.Fatalf("loaded.parent = %+v, want PARENT-1", loaded.parent)
	}

	_, _ = a.Update(loaded)

	if !a.issuesList.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = false after dispatching loaded msg")
	}
	if got := a.issuesList.HierarchyTitle(); got != "Parent" {
		t.Errorf("HierarchyTitle() = %q, want %q", got, "Parent")
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "PARENT-1" {
		t.Errorf("SelectedIssue() = %+v, want PARENT-1", sel)
	}
	if got := len(fake.GetIssueCalls); got != 1 || fake.GetIssueCalls[0].Key != "PARENT-1" {
		t.Errorf("GetIssue calls = %+v, want 1 call for PARENT-1", fake.GetIssueCalls)
	}
}

// Backspace on issue without Parent is a no-op.
func TestHierarchy_BackspaceWithoutParent_NoOp(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.issuesList.SetIssues([]jira.Issue{{Key: "ORPHAN-1"}})

	_, handled := a.showParent()
	if handled {
		t.Errorf("showParent() handled = true, want false (no parent)")
	}
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = true, want false")
	}
	if len(fake.GetIssueCalls) != 0 {
		t.Errorf("GetIssue called %d times, want 0", len(fake.GetIssueCalls))
	}
}

// On Cloud, showChildrenFromList ignores Issue.Subtasks: a cache hit
// pushes immediately from a.childrenCache.
func TestHierarchy_Cloud_ChildrenWalk_CacheHit_PushesImmediately(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.issuesList.SetIssues([]jira.Issue{{Key: "EPIC-1"}})
	a.childrenCache["EPIC-1"] = []jira.Issue{{Key: "WALK-1", Summary: "first"}}

	cmd, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}
	if cmd == nil {
		t.Fatalf("expected previewAfterNav cmd, got nil")
	}
	if !a.issuesList.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = false")
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "WALK-1" {
		t.Errorf("SelectedIssue() = %+v, want C-1", sel)
	}
	if len(fake.GetChildrenCalls) != 0 {
		t.Errorf("GetChildren called %d times, want 0 (cache hit)", len(fake.GetChildrenCalls))
	}
}

// On Cloud, a cache miss dispatches a childrenWalkRequestMsg whose
// handler tags pendingWalk with the active childrenEpoch; the eventual
// childrenLoadedMsg pushes the hierarchy frame and primes the cache.
func TestHierarchy_Cloud_ChildrenWalk_CacheMiss_FetchesThenPushes(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetChildrenFunc = func(_ context.Context, parentKey string) ([]jira.Issue, error) {
		return []jira.Issue{{Key: "C-9", Summary: "nine"}}, nil
	}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.issuesList.SetIssues([]jira.Issue{{Key: "EPIC-1"}})

	cmd, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}
	if cmd == nil {
		t.Fatalf("showChildren() cmd = nil, want childrenWalkRequestMsg cmd")
	}

	req, ok := cmd().(childrenWalkRequestMsg)
	if !ok {
		t.Fatalf("cmd produced %T, want childrenWalkRequestMsg", cmd())
	}
	_, fetchCmd := a.Update(req)
	if fetchCmd == nil {
		t.Fatalf("childrenWalkRequestMsg produced nil cmd, want fetch cmd")
	}
	if a.pendingWalk.key != "EPIC-1" || a.pendingWalk.epoch != a.childrenEpoch {
		t.Errorf("pendingWalk = %+v, want {EPIC-1, %d}", a.pendingWalk, a.childrenEpoch)
	}

	loaded, ok := fetchCmd().(childrenLoadedMsg)
	if !ok {
		t.Fatalf("fetch produced %T, want childrenLoadedMsg", fetchCmd())
	}
	_, _ = a.Update(loaded)

	if !a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = false after childrenLoadedMsg")
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "C-9" {
		t.Errorf("SelectedIssue() = %+v, want C-9", sel)
	}
	if cached, ok := a.childrenCache["EPIC-1"]; !ok || len(cached) != 1 {
		t.Errorf("childrenCache[EPIC-1] = %+v, want 1 entry", cached)
	}
	if a.pendingWalk != (pendingWalk{}) {
		t.Errorf("pendingWalk = %+v, want zero after consume", a.pendingWalk)
	}
}

// On Cloud, an empty GetChildren result opens the detail view instead
// of pushing a hierarchy frame — matching the Server/DC and cache-hit
// empty paths.
func TestHierarchy_Cloud_ChildrenWalk_EmptyResult_OpensDetail(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return nil, nil
	}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.issuesList.SetIssues([]jira.Issue{{Key: "EPIC-1"}})

	cmd, _ := a.showChildren()
	req := cmd().(childrenWalkRequestMsg)
	_, fetchCmd := a.Update(req)
	loaded := fetchCmd().(childrenLoadedMsg)
	_, _ = a.Update(loaded)

	if a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = true, want false on empty result")
	}
	if a.side != sideRight {
		t.Errorf("side = %v, want sideRight (detail opened)", a.side)
	}
	if a.pendingWalk != (pendingWalk{}) {
		t.Errorf("pendingWalk = %+v, want zero after consume", a.pendingWalk)
	}
}

// A passive Sub-tab fetch for the same key as an in-flight walk
// must not be treated as a walk when its response arrives.
func TestHierarchy_Cloud_ChildrenWalk_NoLeakIntoPassiveFetchSameKey(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return []jira.Issue{{Key: "C-1"}}, nil
	}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.issuesList.SetIssues([]jira.Issue{{Key: "X"}})

	cmd, _ := a.showChildren()
	req := cmd().(childrenWalkRequestMsg)
	_, fetch1 := a.Update(req)
	loaded1 := fetch1().(childrenLoadedMsg)

	_, fetch2 := a.Update(views.ChildrenRequestMsg{Key: "X"})
	loaded2 := fetch2().(childrenLoadedMsg)

	_, _ = a.Update(loaded1)
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("stale walk response pushed a hierarchy tab")
	}

	_, _ = a.Update(loaded2)
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("passive response for same key was treated as walk (leak)")
	}
}

// Project switch invalidates an in-flight walk; its stale response
// must not push a hierarchy frame on the new project.
func TestHierarchy_Cloud_SelectProject_InvalidatesInFlightWalk(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return []jira.Issue{{Key: "C-1"}}, nil
	}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.demoMode = true
	a.usersCache = map[string][]jira.User{"OLD": nil, "NEW": nil}
	a.issuesList.SetIssues([]jira.Issue{{Key: "EPIC-1"}})

	cmd, _ := a.showChildren()
	req := cmd().(childrenWalkRequestMsg)
	_, fetchCmd := a.Update(req)
	loaded := fetchCmd().(childrenLoadedMsg)

	_ = a.selectProject(&jira.Project{ID: "2", Key: "NEW"})

	if a.pendingWalk != (pendingWalk{}) {
		t.Fatalf("selectProject left pendingWalk = %+v", a.pendingWalk)
	}

	_, _ = a.Update(loaded)
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("stale walk response after project switch pushed a hierarchy tab")
	}
}

// goBack only fires when the Issues panel itself has focus; on the
// other left panels h/Esc/Left fall through to their natural behavior.
func TestHierarchy_GoBack_SkipsWhenLeftFocusNotIssues(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.issuesList.AddHierarchyTab(hierarchyTitleChildren, []jira.Issue{{Key: "CHILD-1"}})
	a.issuesList.HierarchyStack().Push(navstack.NavFrame{ParentKey: "P-1"})

	for _, focus := range []focusPanel{focusInfo, focusStatus, focusProjects} {
		a.leftFocus = focus
		if _, handled := a.goBack(); handled {
			t.Errorf("leftFocus=%v: goBack() handled=true, want false", focus)
		}
		if !a.issuesList.HasHierarchyTab() {
			t.Errorf("leftFocus=%v: hierarchy tab disappeared", focus)
		}
	}

	a.leftFocus = focusIssues
	if _, handled := a.goBack(); !handled {
		t.Errorf("leftFocus=focusIssues: goBack() handled=false, want true")
	}
}

func TestHierarchy_Backspace_PushesEvenWhenTopMatches(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, key string) (*jira.Issue, error) {
		return &jira.Issue{Key: key}, nil
	}
	a := newAppWithFake(t, fake)

	a.issuesList.AddHierarchyTab(hierarchyTitleChildren, []jira.Issue{
		{Key: "CHILD-1", Parent: &jira.Issue{Key: "FOO-1"}},
	})
	a.issuesList.HierarchyStack().Push(navstack.NavFrame{ParentKey: "FOO-1"})
	if d := a.issuesList.HierarchyStack().Depth(); d != 1 {
		t.Fatalf("precondition: stack depth = %d, want 1", d)
	}

	cmd, handled := a.showParent()
	if !handled {
		t.Fatalf("showParent() handled = false, want true")
	}
	if cmd == nil {
		t.Fatalf("showParent() cmd = nil, want async fetch cmd")
	}

	_, _ = a.Update(cmd().(parentLoadedMsg))

	if !a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = false, want true (stack should not collapse)")
	}
	if d := a.issuesList.HierarchyStack().Depth(); d != 2 {
		t.Errorf("HierarchyStack.Depth() = %d, want 2 (existing + new push)", d)
	}
}

// Two rapid Backspaces bump parentEpoch to 2. A
// parentLoadedMsg carrying epoch=1 (stale) must be dropped — no
// hierarchy tab is created.
func TestHierarchy_StaleParent_Dropped(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, key string) (*jira.Issue, error) {
		return &jira.Issue{Key: key}, nil
	}
	a := newAppWithFake(t, fake)
	a.issuesList.SetIssues([]jira.Issue{
		{Key: "CHILD-1", Parent: &jira.Issue{Key: "PARENT-1"}},
	})

	// Two rapid showParent invocations advance the epoch to 2.
	if _, ok := a.showParent(); !ok {
		t.Fatal("first showParent() handled = false")
	}
	if _, ok := a.showParent(); !ok {
		t.Fatal("second showParent() handled = false")
	}
	if a.parentEpoch != 2 {
		t.Fatalf("parentEpoch = %d after two showParent(), want 2", a.parentEpoch)
	}

	// Stale msg from the first request (epoch=1) arrives after the second
	// request has already bumped the epoch.
	stale := parentLoadedMsg{
		childKey: "CHILD-1",
		parent:   &jira.Issue{Key: "PARENT-1", Summary: "stale"},
		epoch:    1,
	}
	_, _ = a.Update(stale)

	if a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = true after stale msg, want false (dropped)")
	}
}

// seedHierarchyTab prepares an active hierarchy tab plus the tab 0 origin so tests
// that exercise goBack / hard-replace have a stable baseline.
func seedHierarchyTab(t *testing.T, a *App) {
	t.Helper()
	a.issuesList.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	a.issuesList.AddHierarchyTab(hierarchyTitleChildren, []jira.Issue{{Key: "DRILL-0"}})
}

// Esc in hierarchy tab with Depth>1 pops the top frame (the
// pre-hierarchy snapshot of level 1) and restores its UI snapshot. The new
// top's Source determines the hierarchy tab title.
func TestHierarchy_Esc_PopsFrame(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	seedHierarchyTab(t, a)
	stack := a.issuesList.HierarchyStack()

	// originFrame: snapshot of origin JQL tab (bottom of stack). Its
	// Source drives the hierarchy tab title after the level-1 pop.
	stack.Push(navstack.NavFrame{
		ParentKey: "P1",
		Source:    navstack.SourceFromList,
		Issues:    []jira.Issue{{Key: "ORIG-A"}},
	})
	// level1Frame: pre-hierarchy snapshot of level 1 (gets popped on Esc and
	// its data is restored into the hierarchy tab).
	stack.Push(navstack.NavFrame{
		ParentKey:   "P2",
		Source:      navstack.SourceFromList,
		Issues:      []jira.Issue{{Key: "L1-A"}, {Key: "L1-B"}},
		SelectedIdx: 1,
		FocusPanel:  navstack.FocusPanel(focusInfo),
		InfoTab:     int(views.InfoTabSubtasks),
		InfoCursor:  3,
	})
	if d := stack.Depth(); d != 2 {
		t.Fatalf("setup: stack depth = %d, want 2", d)
	}

	cmd, handled := a.goBack()
	if !handled {
		t.Fatalf("goBack() handled = false, want true")
	}
	_ = cmd

	if !a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = false, want true after pop with depth>1")
	}
	if d := a.issuesList.HierarchyStack().Depth(); d != 1 {
		t.Errorf("stack depth = %d, want 1", d)
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "L1-B" {
		t.Errorf("SelectedIssue() = %+v, want L1-B", sel)
	}
	if a.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		t.Errorf("info tab = %v, want InfoTabSubtasks", a.infoPanel.ActiveTab())
	}
	if a.infoPanel.Cursor != 3 {
		t.Errorf("info cursor = %d, want 3", a.infoPanel.Cursor)
	}
	if a.leftFocus != focusInfo {
		t.Errorf("leftFocus = %v, want focusInfo", a.leftFocus)
	}
}

// Esc with Depth==1 pops the last frame (origin snapshot),
// closes the hierarchy tab and restores the origin cursor/focus.
func TestHierarchy_Esc_LastFrame_ClosesTab(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	seedHierarchyTab(t, a)
	a.issuesList.HierarchyStack().Push(navstack.NavFrame{
		ParentKey:   "P1",
		Source:      navstack.SourceFromList,
		Issues:      []jira.Issue{{Key: "ORIG-A"}, {Key: "ORIG-B"}},
		SelectedIdx: 1,
		FocusPanel:  navstack.FocusPanel(focusIssues),
	})

	_, handled := a.goBack()
	if !handled {
		t.Fatalf("goBack() handled = false, want true")
	}
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = true, want false after last pop")
	}
	if a.leftFocus != focusIssues {
		t.Errorf("leftFocus = %v, want focusIssues", a.leftFocus)
	}
	if a.issuesList.Cursor != 1 {
		t.Errorf("Cursor = %d, want 1", a.issuesList.Cursor)
	}
}

// Showing children from the issues list captures a pre-hierarchy
// snapshot; popping restores the origin cursor, info tab, info cursor
// and focus.
func TestHierarchy_SnapshotPreservesContext(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)

	parent := jira.Issue{
		Key:      "PARENT-3",
		Subtasks: []jira.Issue{{Key: "SUB-X"}},
	}
	a.issuesList.SetIssues([]jira.Issue{
		{Key: "A"}, {Key: "B"}, parent,
	})
	a.issuesList.Cursor = 2
	a.leftFocus = focusIssues

	a.infoPanel.SetIssue(&parent)
	a.infoPanel.NextTab() // Links
	a.infoPanel.NextTab() // Subtasks
	if a.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		t.Fatalf("setup: ActiveTab = %v, want Subtasks", a.infoPanel.ActiveTab())
	}
	a.infoPanel.Cursor = 1

	cmd, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false")
	}
	if !a.issuesList.IsHierarchyTab() {
		t.Fatalf("IsHierarchyTab() = false after showChildren")
	}
	// Drive the returned cmd through Update so PreviewRequestMsg fires
	// and SetIssue runs — that path resets the Info tab if the snapshot
	// is not rehydrated correctly on goBack.
	if cmd != nil {
		_, _ = a.Update(cmd())
	}

	cmd, handled = a.goBack()
	if !handled {
		t.Fatalf("goBack() handled = false")
	}
	if cmd != nil {
		_, _ = a.Update(cmd())
	}

	if a.issuesList.Cursor != 2 {
		t.Errorf("Cursor = %d, want 2", a.issuesList.Cursor)
	}
	if a.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		t.Errorf("ActiveTab = %v, want Subtasks", a.infoPanel.ActiveTab())
	}
	if a.infoPanel.Cursor != 1 {
		t.Errorf("infoPanel.Cursor = %d, want 1", a.infoPanel.Cursor)
	}
	if a.leftFocus != focusIssues {
		t.Errorf("leftFocus = %v, want focusIssues", a.leftFocus)
	}
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("HasHierarchyTab() = true, want false after last pop")
	}
}

// Esc outside the hierarchy tab falls through to the default
// ActFocusLeft mapping (sideRight -> sideLeft); no goBack side effects.
func TestHierarchy_Esc_OutsideHierarchyTab(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.keymap = DefaultKeymap()
	a.issuesList.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})
	a.side = sideRight

	_, _ = a.handleKeyMsg(tea.KeyMsg{Type: tea.KeyEsc})

	if a.side != sideLeft {
		t.Errorf("a.side = %v, want sideLeft (default Esc behavior)", a.side)
	}
	if a.issuesList.HasHierarchyTab() {
		t.Errorf("Esc outside hierarchy accidentally created hierarchy tab")
	}
}

// Triggering a hierarchy from a non-hierarchy JQL tab while a hierarchy tab
// already exists clears the old stack, replaces the tab content, and
// focuses the hierarchy tab (hard-replace).
// Pushing a new hierarchy view from a non-hierarchy tab adds onto the
// existing stack and refocuses the hierarchy tab.
func TestHierarchy_NavFromDifferentJQLTab_StacksUp(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.issuesList.SetTabs([]config.IssueTabConfig{{Name: "My", JQL: ""}})

	a.issuesList.AddHierarchyTab(hierarchyTitleChildren, []jira.Issue{{Key: "OLD-1"}})
	hierarchyIdx := a.issuesList.GetTabIndex()
	stack := a.issuesList.HierarchyStack()
	stack.Push(navstack.NavFrame{ParentKey: "P1", Issues: []jira.Issue{{Key: "OLD-1"}}})
	stack.Push(navstack.NavFrame{ParentKey: "P2", Issues: []jira.Issue{{Key: "OLD-1"}}})
	stack.Push(navstack.NavFrame{ParentKey: "P3", Issues: []jira.Issue{{Key: "OLD-1"}}})
	if d := stack.Depth(); d != 3 {
		t.Fatalf("setup: stack depth = %d, want 3", d)
	}

	a.issuesList.SetTabIndex(0)
	a.issuesList.SetIssues([]jira.Issue{
		{Key: "NEW-PARENT", Subtasks: []jira.Issue{{Key: "SUB-A", Summary: "a"}}},
	})

	_, handled := a.showChildren()
	if !handled {
		t.Fatalf("showChildren() handled = false, want true")
	}

	if a.issuesList.GetTabIndex() != hierarchyIdx {
		t.Errorf("GetTabIndex() = %d, want hierarchy tab %d", a.issuesList.GetTabIndex(), hierarchyIdx)
	}
	if d := a.issuesList.HierarchyStack().Depth(); d != 4 {
		t.Errorf("HierarchyStack.Depth() = %d, want 4 (3 existing + 1 new push)", d)
	}
	if sel := a.issuesList.SelectedIssue(); sel == nil || sel.Key != "SUB-A" {
		t.Errorf("SelectedIssue() = %+v, want SUB-A", sel)
	}
}

func findHelpItem(items []components.HelpItem, desc string) (components.HelpItem, bool) {
	for _, it := range items {
		if it.Description == desc {
			return it, true
		}
	}
	return components.HelpItem{}, false
}

// The help-bar advertises "parent" only when the current
// issue has a Parent.
func TestHelpBar_Backspace_OnlyWhenParent(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.keymap = DefaultKeymap()

	a.issuesList.SetIssues([]jira.Issue{{Key: "A1"}})
	if _, ok := findHelpItem(a.helpBarItems(), "parent"); ok {
		t.Errorf("help bar shows 'parent' without a parent issue")
	}

	a.issuesList.SetIssues([]jira.Issue{{Key: "A1", Parent: &jira.Issue{Key: "P1"}}})
	it, ok := findHelpItem(a.helpBarItems(), "parent")
	if !ok {
		t.Fatalf("help bar missing 'parent' entry when parent exists")
	}
	if it.Key != "backspace" {
		t.Errorf("parent help key = %q, want %q", it.Key, "backspace")
	}
}

// The help-bar advertises the ActSelect key as "children" when
// the current issue has Subtasks, and "detail" otherwise.
func TestHelpBar_Enter_ChildrenVsDetail(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.keymap = DefaultKeymap()

	a.issuesList.SetIssues([]jira.Issue{
		{Key: "A1", Subtasks: []jira.Issue{{Key: "S1"}}},
	})
	if _, ok := findHelpItem(a.helpBarItems(), "children"); !ok {
		t.Errorf("help bar missing 'children' entry when subtasks exist")
	}
	if _, ok := findHelpItem(a.helpBarItems(), "detail"); ok {
		t.Errorf("help bar shows 'detail' when subtasks exist")
	}

	a.issuesList.SetIssues([]jira.Issue{{Key: "A1"}})
	if _, ok := findHelpItem(a.helpBarItems(), "detail"); !ok {
		t.Errorf("help bar missing 'detail' entry for leaf issue")
	}
	if _, ok := findHelpItem(a.helpBarItems(), "children"); ok {
		t.Errorf("help bar shows 'children' for leaf issue")
	}
}

// Navigating from a non-first JQL tab and fully popping the
// hierarchy-stack must return the user to the origin tab.
func TestHierarchy_FullPop_RestoresOriginTab(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.issuesList.SetTabs([]config.IssueTabConfig{
		{Name: "My", JQL: ""},
		{Name: "Watched", JQL: ""},
	})
	a.issuesList.SetTabIndex(1)
	parent := jira.Issue{
		Key:      "PARENT-42",
		Subtasks: []jira.Issue{{Key: "SUB-1"}},
	}
	a.issuesList.SetIssues([]jira.Issue{parent})

	if _, handled := a.showChildren(); !handled {
		t.Fatalf("showChildren() handled = false")
	}
	if !a.issuesList.IsHierarchyTab() {
		t.Fatalf("IsHierarchyTab() = false after showChildren")
	}

	if _, handled := a.goBack(); !handled {
		t.Fatalf("goBack() handled = false")
	}
	if a.issuesList.HasHierarchyTab() {
		t.Fatalf("HasHierarchyTab() = true after full pop, want false")
	}
	if got := a.issuesList.GetTabIndex(); got != 1 {
		t.Errorf("GetTabIndex() = %d, want 1 (origin tab 'Watched')", got)
	}
}
