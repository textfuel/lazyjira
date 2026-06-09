package jira

import (
	"net/http"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestClient_GetMyself_ParsesUser(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"accountId": "me-1", "displayName": "Current User", "emailAddress": "me@example.com"}`,
	})

	user, err := client.GetMyself(t.Context())
	if err != nil {
		t.Fatalf("GetMyself: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/myself")
	testkit.AssertEqual(t, "AccountID", user.AccountID, "me-1")
	testkit.AssertEqual(t, "DisplayName", user.DisplayName, "Current User")
}

func TestClient_GetUsers_UsesAssignableSearch(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"accountId": "u1", "displayName": "Ann"}, {"accountId": "u2", "displayName": "Bob"}]`,
	})

	users, err := client.GetUsers(t.Context(), "PLAT")
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/user/assignable/search")
	testkit.AssertEqual(t, "project query", recorded.Query.Get("project"), "PLAT")
	if len(users) != 2 {
		t.Fatalf("len(users) = %d, want 2", len(users))
	}
	testkit.AssertEqual(t, "users[0].AccountID", users[0].AccountID, "u1")
}

func TestClient_AssignIssue_CloudUsesAccountID(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	if err := client.AssignIssue(t.Context(), "PLAT-1", "acc-1"); err != nil {
		t.Fatalf("AssignIssue: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodPut)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/assignee")

	body := decodeBody(t, recorded.Body)
	if body["accountId"] != "acc-1" {
		t.Errorf("body = %#v, want accountId=acc-1", body)
	}
}

func TestClient_AssignIssue_ServerUsesName(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	if err := client.AssignIssue(t.Context(), "OPS-7", "jsmith"); err != nil {
		t.Fatalf("AssignIssue: %v", err)
	}

	body := decodeBody(t, recorded.Body)
	if body["name"] != "jsmith" {
		t.Errorf("body = %#v, want name=jsmith", body)
	}
	if _, hasAccountID := body["accountId"]; hasAccountID {
		t.Errorf("server body should not contain accountId, got %#v", body)
	}
}

func TestClient_GetSprints_UsesAgileBoardPath(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"values": [{"id": 42, "name": "Sprint 1", "state": "active"}], "isLast": true}`,
	})

	sprints, err := client.GetSprints(t.Context(), 5)
	if err != nil {
		t.Fatalf("GetSprints: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/agile/1.0/board/5/sprint")
	if len(sprints) != 1 {
		t.Fatalf("len(sprints) = %d, want 1", len(sprints))
	}
	testkit.AssertEqual(t, "sprints[0].Name", sprints[0].Name, "Sprint 1")
}

func TestClient_MoveToSprint_PostsIssueList(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	if err := client.MoveToSprint(t.Context(), 9, "PLAT-1"); err != nil {
		t.Fatalf("MoveToSprint: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodPost)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/agile/1.0/sprint/9/issue")

	body := decodeBody(t, recorded.Body)
	issues, ok := body["issues"].([]any)
	if !ok || len(issues) != 1 || issues[0] != "PLAT-1" {
		t.Errorf("body issues = %#v, want [PLAT-1]", body["issues"])
	}
}
