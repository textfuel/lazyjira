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

func TestChildrenRequestMsg_Cloud_FiresGetChildren(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	want := []jira.Issue{{Key: "C-1", Summary: "first"}, {Key: "C-2", Summary: "second"}}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return want, nil
	}

	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	_, cmd := app.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
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

	_, _ = app.Update(loadedMsg)

	got := app.infoPanel.Children()
	if len(got) != 2 || got[0].Key != "C-1" || got[1].Key != "C-2" {
		t.Errorf("InfoPanel children = %+v, want %+v", got, want)
	}
}

func TestChildrenRequestMsg_ServerDC_NoCall(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}

	app := newAppWithFake(t, fake)
	app.isCloud = false
	app.infoPanel.SetCloud(false)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	_, cmd := app.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd != nil {
		t.Errorf("Server/DC: expected nil cmd, got non-nil")
	}
	if len(fake.GetChildrenCalls) != 0 {
		t.Errorf("Server/DC: expected 0 GetChildren calls, got %d", len(fake.GetChildrenCalls))
	}
}

func TestChildrenLoadedMsg_StaleEpochDropped(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	app.childrenEpoch = 5

	stale := childrenLoadedMsg{
		key:    "EPIC-1",
		issues: []jira.Issue{{Key: "STALE-CHILD"}},
		epoch:  3,
	}
	_, _ = app.Update(stale)

	if got := app.infoPanel.Children(); got != nil {
		t.Errorf("stale response: expected nil children, got %+v", got)
	}
}

func TestChildrenLoadedMsg_FetchError_SetsStatusPanelError(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	app.childrenEpoch = 1
	_, _ = app.Update(childrenLoadedMsg{
		key:   "EPIC-1",
		err:   errors.New("network down"),
		epoch: 1,
	})

	if app.statusPanel.ErrorMessage() == "" {
		t.Error("StatusPanel error should be set on fetch failure")
	}
}

func TestIssueSelectedMsg_OnSubTab_DispatchesChildrenRequest(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)

	app.infoPanel.SetIssue(&jira.Issue{Key: "OLD"})
	for app.infoPanel.ActiveTab() != views.InfoTabSubtasks {
		app.infoPanel.NextTab()
	}

	_, cmd := app.Update(views.IssueSelectedMsg{Issue: &jira.Issue{Key: "EPIC-1"}})
	if cmd == nil {
		t.Fatal("expected batch cmd, got nil")
	}

	if !batchContainsChildrenRequest(cmd, "EPIC-1") {
		t.Error("expected ChildrenRequestMsg{Key: EPIC-1} in cmd batch")
	}
}

func TestIssueSelectedMsg_OnFieldsTab_NoChildrenRequest(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "OLD"})

	_, cmd := app.Update(views.IssueSelectedMsg{Issue: &jira.Issue{Key: "EPIC-1"}})
	if batchContainsChildrenRequest(cmd, "EPIC-1") {
		t.Error("Fields tab should not dispatch ChildrenRequestMsg")
	}
}

func TestChildrenRequestMsg_CacheHit_NoClientCall(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}

	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})
	app.childrenCache["EPIC-1"] = []jira.Issue{{Key: "C-1", Summary: "cached"}}

	_, cmd := app.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd != nil {
		t.Errorf("cache hit: expected nil cmd, got non-nil")
	}
	if len(fake.GetChildrenCalls) != 0 {
		t.Errorf("cache hit: expected 0 GetChildren calls, got %d", len(fake.GetChildrenCalls))
	}
	got := app.infoPanel.Children()
	if len(got) != 1 || got[0].Key != "C-1" {
		t.Errorf("cache hit: InfoPanel children = %+v, want one C-1", got)
	}
}

func TestChildrenRequestMsg_CacheMiss_PopulatesCacheOnLoad(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	want := []jira.Issue{{Key: "C-1"}, {Key: "C-2"}}
	fake.GetChildrenFunc = func(_ context.Context, _ string) ([]jira.Issue, error) {
		return want, nil
	}

	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})

	_, cmd := app.Update(views.ChildrenRequestMsg{Key: "EPIC-1"})
	if cmd == nil {
		t.Fatal("cache miss: expected fetch cmd, got nil")
	}
	if _, ok := app.childrenCache["EPIC-1"]; ok {
		t.Error("cache miss: cache should still be empty before response")
	}
	_, _ = app.Update(cmd())

	cached, ok := app.childrenCache["EPIC-1"]
	if !ok {
		t.Fatal("after load: expected cache entry for EPIC-1")
	}
	if len(cached) != 2 || cached[0].Key != "C-1" {
		t.Errorf("cached entry = %+v, want %+v", cached, want)
	}
}

func TestChildrenLoadedMsg_PrefetchesChildDetails(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	var seenJQL string
	fake.SearchIssuesFunc = func(_ context.Context, jql string, _, _ int) (*jira.SearchResult, error) {
		seenJQL = jql
		return &jira.SearchResult{}, nil
	}

	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})
	app.childrenEpoch = 1

	_, cmd := app.Update(childrenLoadedMsg{
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
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "EPIC-1"})
	app.issueCache["C-1"] = &jira.Issue{Key: "C-1"}
	app.issueCache["C-2"] = &jira.Issue{Key: "C-2"}
	app.childrenEpoch = 1

	_, cmd := app.Update(childrenLoadedMsg{
		key:    "EPIC-1",
		issues: []jira.Issue{{Key: "C-1"}, {Key: "C-2"}},
		epoch:  1,
	})
	if cmd != nil {
		t.Errorf("all children cached: expected nil cmd, got non-nil")
	}
}

func TestActRefresh_ClearsChildrenCacheForPreviewKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.GetIssueFunc = func(_ context.Context, _ string) (*jira.Issue, error) {
		return &jira.Issue{Key: "EPIC-1"}, nil
	}
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.previewKey = "EPIC-1"
	app.childrenCache["EPIC-1"] = []jira.Issue{{Key: "STALE"}}
	app.childrenCache["OTHER"] = []jira.Issue{{Key: "OTHER-CHILD"}}

	_, _, _ = app.handleIssueAction(ActRefresh)

	if _, ok := app.childrenCache["EPIC-1"]; ok {
		t.Error("refresh: expected EPIC-1 entry to be cleared")
	}
	if _, ok := app.childrenCache["OTHER"]; !ok {
		t.Error("refresh: unrelated cache entry must survive")
	}
}

func TestActRefresh_OnCloud_RefetchesChildren(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "REFRESH-1"})
	app := newAppWithFake(t, fake)
	app.isCloud = true
	app.infoPanel.SetCloud(true)
	app.infoPanel.SetIssue(&jira.Issue{Key: "REFRESH-1"})
	app.previewKey = "REFRESH-1"
	app.childrenCache["REFRESH-1"] = []jira.Issue{{Key: "STALE"}}

	_, cmd, handled := app.handleIssueAction(ActRefresh)
	if !handled {
		t.Fatal("ActRefresh was not handled")
	}
	if !batchContainsChildrenRequest(cmd, "REFRESH-1") {
		t.Error("expected ChildrenRequestMsg{Key: REFRESH-1} in cmd batch")
	}
}

func TestActRefresh_OnServerDC_NoChildrenRequest(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: "REFRESH-2"})
	app := newAppWithFake(t, fake)
	app.isCloud = false
	app.infoPanel.SetCloud(false)
	app.infoPanel.SetIssue(&jira.Issue{Key: "REFRESH-2"})
	app.previewKey = "REFRESH-2"

	_, cmd, handled := app.handleIssueAction(ActRefresh)
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
