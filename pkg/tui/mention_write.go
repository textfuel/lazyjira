package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

// pendingMention holds a write that was deferred because the user cache was
// cold. Once the users arrive it is completed via the createDesc or applyEdit
// path.
type pendingMention struct {
	content     string
	createDesc  bool    // true: create-form description field; false: applyEdit
	fieldIndex  int     // createDesc only
	convState   any     // createDesc only (applyEdit uses editContext.converterState)
	editContext editCtx // applyEdit only
	projectKey  string  // project whose users resolve this mention
}

// mentionUsersLoadedMsg carries the users fetched for a deferred mention write.
type mentionUsersLoadedMsg struct{ users []jira.User }

// mentionsApply reports whether an edit kind is rendered as ADF, where a
// resolved @-mention becomes a real mention. Plain-text custom fields
// (editFieldText) are written as-is without ADF conversion, so resolving there
// would store the link syntax as literal text; they are excluded.
func mentionsApply(kind editKind) bool {
	switch kind {
	case editDesc, editCommentNew, editCommentMod:
		return true
	default:
		return false
	}
}

// projectUsers returns the cached assignable users for the given project key
// and whether the cache holds an entry for it.
func (a *App) projectUsers(projectKey string) ([]jira.User, bool) {
	u, ok := a.usersCache[projectKey]
	return u, ok
}

// fetchUsersForMention loads the project users for a deferred write. Unlike
// fetchUsers it never emits errorMsg: on failure it returns an empty list so
// the pending write still completes (mentions fall back to literal text).
func fetchUsersForMention(client jira.ClientInterface, projectKey string) tea.Cmd {
	return func() tea.Msg {
		users, _ := client.GetUsers(context.Background(), projectKey)
		return mentionUsersLoadedMsg{users: users}
	}
}

// handleMentionUsersLoaded caches the freshly loaded users and resumes the
// deferred write.
func (a *App) handleMentionUsersLoaded(msg mentionUsersLoadedMsg) (tea.Model, tea.Cmd) {
	pm := a.pendingMention
	a.pendingMention = nil
	if pm == nil {
		return a, nil
	}
	if pm.projectKey != "" && len(msg.users) > 0 {
		a.usersCache[pm.projectKey] = msg.users
	}
	if pm.createDesc {
		return a, a.completeCreateDesc(*pm, msg.users)
	}
	return a, a.completeApplyEdit(*pm, msg.users)
}

// completeCreateDesc resolves mentions with the now-available users, converts
// to ADF and writes the create-form description field.
func (a *App) completeCreateDesc(pm pendingMention, users []jira.User) tea.Cmd {
	content := resolveMentions(pm.content, users)
	adf, err := a.converter.FromMarkdown(content, pm.convState)
	if err != nil {
		a.statusPanel.SetError("convert description: " + err.Error())
		return nil
	}
	a.createForm.SetFieldValue(pm.fieldIndex, adf, content)
	return nil
}

// completeApplyEdit resolves mentions with the now-available users, converts to
// ADF and submits the edit.
func (a *App) completeApplyEdit(pm pendingMention, users []jira.User) tea.Cmd {
	md := resolveMentions(pm.content, users)
	adf, err := a.converter.FromMarkdown(md, pm.editContext.converterState)
	if err != nil {
		a.statusPanel.SetError("convert markdown: " + err.Error())
		return nil
	}
	return a.submitEdit(pm.editContext, adf, md)
}
