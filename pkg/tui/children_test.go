package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

// TestChildrenRequestMsg_Cloud_FiresGetChildren pins the happy path: a
// ChildrenRequestMsg on a Cloud app dispatches GetChildren with the right
// key and routes the loaded children into InfoPanel.
func TestChildrenRequestMsg_Cloud_FiresGetChildren(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	want := []jira.Issue{{Key: "C-1", Summary: "first"}, {Key: "C-2", Summary: "second"}}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return want, nil
	}

	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	_, cmd := a.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd == nil {
		t.Fatal("expected fetch Cmd, got nil")
	}
	loadedMsg := cmd()

	if len(fake.GetChildrenCalls) != 1 {
		t.Fatalf("expected 1 GetChildren call, got %d", len(fake.GetChildrenCalls))
	}
	if got := fake.GetChildrenCalls[0].ParentKey; got != "EPIC-1" {
		t.Errorf("GetChildren ParentKey = %q, want EPIC-1", got)
	}

	_, _ = a.Update(loadedMsg)

	got := a.infoPanel.Children()
	if len(got) != 2 || got[0].Key != "C-1" || got[1].Key != "C-2" {
		t.Errorf("InfoPanel children = %+v, want %+v", got, want)
	}
}

// TestChildrenRequestMsg_ServerDC_NoCall pins the Server/DC fast-path: the
// app shortcircuits before touching the client when isCloud=false.
func TestChildrenRequestMsg_ServerDC_NoCall(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	// No GetChildrenFunc — any call would t.Fatalf via fake.fatal.

	a := newAppWithFake(t, fake)
	a.isCloud = false
	a.infoPanel.SetCloud(false)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	_, cmd := a.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd != nil {
		t.Errorf("Server/DC: expected nil cmd, got non-nil")
	}
	if len(fake.GetChildrenCalls) != 0 {
		t.Errorf("Server/DC: expected 0 GetChildren calls, got %d", len(fake.GetChildrenCalls))
	}
}

// TestChildrenLoadedMsg_StaleEpochDropped pins the stale-drop invariant: a
// childrenLoadedMsg with a stale epoch (because a newer ChildrenRequestMsg
// has bumped the counter) is ignored.
func TestChildrenLoadedMsg_StaleEpochDropped(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	a.childrenEpoch = 5

	stale := childrenLoadedMsg{
		key:    "EPIC-1",
		issues: []jira.Issue{{Key: "STALE-CHILD"}},
		epoch:  3,
	}
	_, _ = a.Update(stale)

	if got := a.infoPanel.Children(); got != nil {
		t.Errorf("stale response: expected nil children, got %+v", got)
	}
}

// TestChildrenLoadedMsg_FetchError_SetsStatusPanelError pins the error path:
// a non-nil err lands as a StatusPanel error message and propagates to
// InfoPanel for the error row.
func TestChildrenLoadedMsg_FetchError_SetsStatusPanelError(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	a.childrenEpoch = 1
	_, _ = a.Update(childrenLoadedMsg{
		key:   "EPIC-1",
		err:   errors.New("network down"),
		epoch: 1,
	})

	if a.statusPanel.ErrorMessage() == "" {
		t.Error("StatusPanel error should be set on fetch failure")
	}
}

func TestIssueSelectedMsg_OnSubTab_DispatchesChildrenRequest(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)

	a.infoPanel.SetIssue(&jira.Issue{Key: "OLD"})
	for a.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		a.infoPanel.NextTab()
	}

	_, cmd := a.Update(views.IssueSelectedMsg{Issue: &jira.Issue{Key: "EPIC-1"}})
	if cmd == nil {
		t.Fatal("expected batch cmd, got nil")
	}

	if !batchContainsChildrenRequest(cmd, "EPIC-1") {
		t.Error("expected ChildrenRequestMsg{Key: EPIC-1} in cmd batch")
	}
}

func TestIssueSelectedMsg_OnFieldsTab_NoChildrenRequest(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "OLD"})

	_, cmd := a.Update(views.IssueSelectedMsg{Issue: &jira.Issue{Key: "EPIC-1"}})
	if batchContainsChildrenRequest(cmd, "EPIC-1") {
		t.Error("Fields tab should not dispatch ChildrenRequestMsg")
	}
}

func TestChildrenRequestMsg_CacheHit_NoClientCall(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	// No GetChildrenFunc — any call would t.Fatalf via fake.fatal.

	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})
	a.childrenCache["EPIC-1"] = []jira.Issue{{Key: "C-1", Summary: "cached"}}

	_, cmd := a.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd != nil {
		t.Errorf("cache hit: expected nil cmd, got non-nil")
	}
	if len(fake.GetChildrenCalls) != 0 {
		t.Errorf("cache hit: expected 0 GetChildren calls, got %d", len(fake.GetChildrenCalls))
	}
	got := a.infoPanel.Children()
	if len(got) != 1 || got[0].Key != "C-1" {
		t.Errorf("cache hit: InfoPanel children = %+v, want one C-1", got)
	}
}

func TestChildrenRequestMsg_CacheMiss_PopulatesCacheOnLoad(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	want := []jira.Issue{{Key: "C-1"}, {Key: "C-2"}}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return want, nil
	}

	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	_, cmd := a.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd == nil {
		t.Fatal("cache miss: expected fetch cmd, got nil")
	}
	if _, ok := a.childrenCache["EPIC-1"]; ok {
		t.Error("cache miss: cache should still be empty before response")
	}
	_, _ = a.Update(cmd())

	cached, ok := a.childrenCache["EPIC-1"]
	if !ok {
		t.Fatal("after load: expected cache entry for EPIC-1")
	}
	if len(cached) != 2 || cached[0].Key != "C-1" {
		t.Errorf("cached entry = %+v, want %+v", cached, want)
	}
}

func TestChildrenLoadedMsg_PrefetchesChildDetails(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	var seenJQL string
	fake.SearchIssuesFunc = func(_ context.Context, jql string, _, _ int) (*jira.SearchResult, error) {
		seenJQL = jql
		return &jira.SearchResult{}, nil
	}

	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})
	a.childrenEpoch = 1

	_, cmd := a.Update(childrenLoadedMsg{
		key:    "EPIC-1",
		issues: []jira.Issue{{Key: "C-1"}, {Key: "C-2"}},
		epoch:  1,
	})
	if cmd == nil {
		t.Fatal("expected prefetch cmd, got nil")
	}
	cmd()

	if seenJQL == "" {
		t.Fatal("expected SearchIssues call for prefetch")
	}
	if !strings.Contains(seenJQL, "C-1") || !strings.Contains(seenJQL, "C-2") {
		t.Errorf("prefetch JQL %q must include both child keys", seenJQL)
	}
}

func TestChildrenLoadedMsg_PrefetchSkipsAlreadyCached(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})
	a.issueCache["C-1"] = &jira.Issue{Key: "C-1"}
	a.issueCache["C-2"] = &jira.Issue{Key: "C-2"}
	a.childrenEpoch = 1

	_, cmd := a.Update(childrenLoadedMsg{
		key:    "EPIC-1",
		issues: []jira.Issue{{Key: "C-1"}, {Key: "C-2"}},
		epoch:  1,
	})
	if cmd != nil {
		t.Errorf("all children cached: expected nil cmd, got non-nil")
	}
}

func TestActRefresh_ClearsChildrenCacheForPreviewKey(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, _ string) (*jira.Issue, error) {
		return &jira.Issue{Key: "EPIC-1"}, nil
	}
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.previewKey = "EPIC-1"
	a.childrenCache["EPIC-1"] = []jira.Issue{{Key: "STALE"}}
	a.childrenCache["OTHER"] = []jira.Issue{{Key: "OTHER-CHILD"}}

	_, _, _ = a.handleIssueAction(ActRefresh)

	if _, ok := a.childrenCache["EPIC-1"]; ok {
		t.Error("refresh: expected EPIC-1 entry to be cleared")
	}
	if _, ok := a.childrenCache["OTHER"]; !ok {
		t.Error("refresh: unrelated cache entry must survive")
	}
}

// ActRefresh on a Cloud issue refetches children.
func TestActRefresh_OnCloud_RefetchesChildren(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "REFRESH-1"})
	a := newAppWithFake(t, fake)
	a.isCloud = true
	a.infoPanel.SetCloud(true)
	a.infoPanel.SetIssue(&jira.Issue{Key: "REFRESH-1"})
	a.previewKey = "REFRESH-1"
	a.childrenCache["REFRESH-1"] = []jira.Issue{{Key: "STALE"}}

	_, cmd, handled := a.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if !batchContainsChildrenRequest(cmd, "REFRESH-1") {
		t.Error("expected ChildrenRequestMsg{Key: REFRESH-1} in cmd batch")
	}
}

// ActRefresh on a Server/DC issue does not refetch children.
func TestActRefresh_OnServerDC_NoChildrenRequest(t *testing.T) {
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "REFRESH-2"})
	a := newAppWithFake(t, fake)
	a.isCloud = false
	a.infoPanel.SetCloud(false)
	a.infoPanel.SetIssue(&jira.Issue{Key: "REFRESH-2"})
	a.previewKey = "REFRESH-2"

	_, cmd, handled := a.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if batchContainsChildrenRequest(cmd, "REFRESH-2") {
		t.Error("Server/DC: ActRefresh must not dispatch ChildrenRequestMsg")
	}
}

func batchContainsChildrenRequest(cmd tea.Cmd, key string) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	if m, ok := msg.(views.ChildrenRequestMsg); ok && m.Key == key {
		return true
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			if batchContainsChildrenRequest(sub, key) {
				return true
			}
		}
	}
	return false
}
