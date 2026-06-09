package jira

import (
	"net/http"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

const singleIssueSearchJSON = `{
	"issues": [{"id": "1", "key": "PLAT-1", "fields": {"summary": "First"}}],
	"total": 1,
	"maxResults": 50,
	"startAt": 0
}`

func TestClient_SearchIssues_CloudRequestAndParse(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: singleIssueSearchJSON})

	result, err := client.SearchIssues(t.Context(), "project = PLAT", 0, 50)
	if err != nil {
		t.Fatalf("SearchIssues: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/search/jql")
	testkit.AssertEqual(t, "jql query", recorded.Query.Get("jql"), "project = PLAT")
	testkit.AssertEqual(t, "startAt query", recorded.Query.Get("startAt"), "0")
	testkit.AssertEqual(t, "maxResults query", recorded.Query.Get("maxResults"), "50")

	fields := recorded.Query.Get("fields")
	for _, want := range []string{"summary", "status", "assignee", "issuetype"} {
		if !strings.Contains(fields, want) {
			t.Errorf("fields query %q missing %q", fields, want)
		}
	}

	testkit.AssertEqual(t, "Total", result.Total, 1)
	if len(result.Issues) != 1 {
		t.Fatalf("len(Issues) = %d, want 1", len(result.Issues))
	}
	testkit.AssertEqual(t, "Issues[0].Key", result.Issues[0].Key, "PLAT-1")
}

func TestClient_SearchIssues_ServerUsesV2Path(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{Status: http.StatusOK, Body: singleIssueSearchJSON})

	if _, err := client.SearchIssues(t.Context(), "project = OPS", 0, 25); err != nil {
		t.Fatalf("SearchIssues: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/2/search")
}

func TestClient_SearchIssues_CustomFieldsAppended(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: singleIssueSearchJSON})
	client.SetCustomFields([]string{"customfield_10015"})

	if _, err := client.SearchIssues(t.Context(), "project = PLAT", 0, 50); err != nil {
		t.Fatalf("SearchIssues: %v", err)
	}

	if fields := recorded.Query.Get("fields"); !strings.Contains(fields, "customfield_10015") {
		t.Errorf("fields query %q missing custom field", fields)
	}
}

func TestClient_GetMyIssues_FiresCurrentUserJQL(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: singleIssueSearchJSON})

	if _, err := client.GetMyIssues(t.Context()); err != nil {
		t.Fatalf("GetMyIssues: %v", err)
	}

	if jql := recorded.Query.Get("jql"); !strings.Contains(jql, "assignee=currentUser()") {
		t.Errorf("jql %q missing assignee=currentUser()", jql)
	}
}

func TestClient_GetTransitions_ParsesList(t *testing.T) {
	t.Parallel()

	const transitionsJSON = `{"transitions": [
		{"id": "11", "name": "To Do", "to": {"id": "1", "name": "Open"}},
		{"id": "21", "name": "Done", "to": {"id": "5", "name": "Closed"}}
	]}`

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: transitionsJSON})

	transitions, err := client.GetTransitions(t.Context(), "PLAT-1")
	if err != nil {
		t.Fatalf("GetTransitions: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/transitions")
	if len(transitions) != 2 {
		t.Fatalf("len(transitions) = %d, want 2", len(transitions))
	}
	testkit.AssertEqual(t, "transitions[0].Name", transitions[0].Name, "To Do")
	if transitions[0].To == nil {
		t.Fatal("transitions[0].To is nil")
	}
	testkit.AssertEqual(t, "transitions[0].To.Name", transitions[0].To.Name, "Open")
}

func TestClient_DoTransition_PostsTransitionID(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	if err := client.DoTransition(t.Context(), "PLAT-1", "21"); err != nil {
		t.Fatalf("DoTransition: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodPost)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/transitions")

	body := decodeBody(t, recorded.Body)
	transition, _ := body["transition"].(map[string]any)
	if transition["id"] != "21" {
		t.Errorf("body transition.id = %v, want 21", transition["id"])
	}
}

func TestClient_CreateIssue_WrapsFieldsAndParses(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusCreated,
		Body:   `{"id": "10010", "key": "PLAT-99"}`,
	})

	issue, err := client.CreateIssue(t.Context(), map[string]any{"summary": "New bug"})
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodPost)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue")

	body := decodeBody(t, recorded.Body)
	fields, ok := body["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields wrapper, got %#v", body)
	}
	if fields["summary"] != "New bug" {
		t.Errorf("body fields.summary = %v, want New bug", fields["summary"])
	}
	testkit.AssertEqual(t, "created issue Key", issue.Key, "PLAT-99")
}
