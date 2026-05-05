package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/navstack"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

type parentLoadedMsg struct {
	childKey string
	parent   *jira.Issue
	epoch    int
	err      error
}

// childrenWalkRequestMsg triggers a children fetch and asks the
// handler to push a hierarchy frame once the response arrives.
type childrenWalkRequestMsg struct{ key string }

// pendingWalk binds a walk to one specific children fetch via its
// epoch; a later bump implicitly invalidates the walk.
type pendingWalk struct {
	key   string
	epoch int
}

func fetchParent(client jira.ClientInterface, childKey, parentKey string, epoch int) tea.Cmd {
	return func() tea.Msg {
		iss, err := client.GetIssue(context.Background(), parentKey)
		return parentLoadedMsg{childKey: childKey, parent: iss, epoch: epoch, err: err}
	}
}

// Returns handled=false to fall through to the caller's default Enter behavior.
func (a *App) showChildren() (tea.Cmd, bool) {
	if a.side != sideLeft {
		return nil, false
	}
	switch a.leftFocus { //nolint:exhaustive
	case focusIssues:
		return a.showChildrenFromList()
	case focusInfo:
		return a.showFromInfoPanel()
	}
	return nil, false
}

func (a *App) showChildrenFromList() (tea.Cmd, bool) {
	sel := a.issuesList.SelectedIssue()
	if sel == nil {
		return nil, false
	}
	children, resolved := a.childrenForList(sel)
	if !resolved {
		return func() tea.Msg { return childrenWalkRequestMsg{key: sel.Key} }, true
	}
	if len(children) == 0 {
		return nil, false
	}
	a.pushNav("Children", sel.Key, navstack.SourceFromList, children)
	return a.previewAfterNav(), true
}

// childrenForList returns the children to walk into and a flag
// indicating whether the answer is known synchronously. Cloud cache
// miss returns (nil, false) — caller is expected to dispatch an async
// fetch.
func (a *App) childrenForList(sel *jira.Issue) ([]jira.Issue, bool) {
	if a.isCloud {
		cached, ok := a.childrenCache[sel.Key]
		return cached, ok
	}
	return sel.Subtasks, true
}

func (a *App) showFromInfoPanel() (tea.Cmd, bool) {
	var picked *jira.Issue
	var title string
	var src navstack.Source
	switch a.infoPanel.ActiveTab() {
	case views.InfoTabSubtasks:
		picked = a.infoPanel.SelectedSubtaskIssue()
		title = "Children"
		src = navstack.SourceFromInfoSub
	case views.InfoTabLinks:
		picked = a.infoPanel.SelectedLinkIssue()
		title = "Link"
		src = navstack.SourceFromInfoLink
	default:
		return nil, false
	}
	if picked == nil || picked.Key == "" {
		return nil, false
	}
	// Prefer a fully-loaded cache entry over the InfoPanel stub so the
	// hierarchy row carries summary/status etc. immediately.
	entry := *picked
	if cached, ok := a.issueCache[picked.Key]; ok && cached != nil {
		entry = *cached
	}
	a.pushNav(title, picked.Key, src, []jira.Issue{entry})
	return a.previewAfterNav(), true
}

func (a *App) handleChildrenLoaded(msg childrenLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.epoch != a.childrenEpoch {
		return a, nil
	}
	walkPending := a.pendingWalk.key == msg.key && a.pendingWalk.epoch == msg.epoch
	a.pendingWalk = pendingWalk{}

	if msg.err != nil {
		a.statusPanel.SetError("Failed to load children: " + msg.err.Error())
		a.infoPanel.SetChildrenError(msg.key, msg.err.Error())
		return a, nil
	}
	a.childrenCache[msg.key] = msg.issues
	a.infoPanel.SetChildren(msg.key, msg.issues)
	cmds := []tea.Cmd{a.prefetchChildrenDetails(msg.issues)}
	if walkPending {
		if len(msg.issues) == 0 {
			_, openCmd := a.openIssueDetail()
			cmds = append(cmds, openCmd)
		} else {
			a.pushNav("Children", msg.key, navstack.SourceFromList, msg.issues)
			cmds = append(cmds, a.previewAfterNav())
		}
	}
	return a, tea.Batch(cmds...)
}

func (a *App) previewAfterNav() tea.Cmd {
	sel := a.issuesList.SelectedIssue()
	if sel == nil {
		return nil
	}
	key := sel.Key
	return func() tea.Msg {
		return views.PreviewRequestMsg{Key: key}
	}
}

// Returns (cmd, true) on async parent fetch; the hierarchy tab is only
// pushed when the resulting parentLoadedMsg is dispatched.
func (a *App) showParent() (tea.Cmd, bool) {
	sel := a.currentIssue()
	if sel == nil || sel.Parent == nil || sel.Parent.Key == "" {
		return nil, false
	}
	a.parentEpoch++
	return fetchParent(a.client, sel.Key, sel.Parent.Key, a.parentEpoch), true
}

func (a *App) pushNav(title, parentKey string, src navstack.Source, issues []jira.Issue) {
	frame := navstack.NavFrame{
		Issues:       a.issuesList.CurrentIssues(),
		SelectedIdx:  a.issuesList.Cursor,
		FocusPanel:   navstack.FocusPanel(a.leftFocus),
		InfoTab:      int(a.infoPanel.ActiveTab()),
		InfoCursor:   a.infoPanel.Cursor,
		Source:       src,
		ParentKey:    parentKey,
		OriginTabIdx: a.issuesList.GetTabIndex(),
	}

	idx := a.issuesList.AddHierarchyTab(title, issues)
	stack := a.issuesList.HierarchyStack()
	if stack == nil {
		a.leftFocus = focusIssues
		a.updateFocusState()
		return
	}

	stack.Push(frame)
	a.issuesList.SetTabIndex(idx)
	a.leftFocus = focusIssues
	a.updateFocusState()
}

// Pops the top NavFrame and restores it; removes the hierarchy tab when the stack empties.
func (a *App) goBack() (tea.Cmd, bool) {
	if a.side != sideLeft || !a.issuesList.IsHierarchyTab() {
		return nil, false
	}
	stack := a.issuesList.HierarchyStack()
	if stack == nil || stack.Depth() == 0 {
		return nil, false
	}
	popped := stack.Pop()
	a.restoreFromFrame(popped, stack)
	return nil, true
}

// Empty stack closes the hierarchy tab; otherwise the tab is rewritten
// from the new top's Source. The InfoPanel is rehydrated synchronously
// so the snapshot's active tab survives — the async PreviewRequestMsg
// path would let SetIssue reset it to InfoTabFields.
func (a *App) restoreFromFrame(popped navstack.NavFrame, stack *navstack.NavStack) {
	if stack.Depth() == 0 {
		a.issuesList.RemoveHierarchyTab()
		a.issuesList.SetTabIndex(popped.OriginTabIdx)
	} else {
		newTopTitle := titleForNavSource(stack.Peek().Source)
		a.issuesList.ReplaceHierarchyTabContent(newTopTitle, popped.Issues)
	}
	a.issuesList.Cursor = popped.SelectedIdx
	if sel := a.issuesList.SelectedIssue(); sel != nil {
		a.previewKey = sel.Key
		issue := sel
		if cached, ok := a.issueCache[sel.Key]; ok && cached != nil {
			issue = cached
		}
		a.infoPanel.SetIssue(issue)
		a.detailView.SetIssue(issue)
	}
	a.infoPanel.SetActiveTab(views.InfoPanelTab(popped.InfoTab))
	a.infoPanel.Cursor = popped.InfoCursor
	a.leftFocus = focusPanel(popped.FocusPanel)
	a.updateFocusState()
}

// Avoids storing a title in each NavFrame.
func titleForNavSource(src navstack.Source) string {
	switch src {
	case navstack.SourceParent:
		return "Parent"
	case navstack.SourceFromInfoLink:
		return "Link"
	case navstack.SourceFromList, navstack.SourceFromInfoSub:
		return "Children"
	}
	return ""
}
