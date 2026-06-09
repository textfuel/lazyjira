package jira

import (
	"net/http"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

const fullIssueJSON = `{
	"id": "10001",
	"key": "PLAT-42",
	"fields": {
		"summary": "Fix login redirect loop",
		"status": {"id": "3", "name": "In Progress", "statusCategory": {"key": "indeterminate"}},
		"priority": {"id": "2", "name": "High"},
		"assignee": {"accountId": "acc-1", "displayName": "Ada Lovelace", "emailAddress": "ada@example.com"},
		"labels": ["backend", "auth"],
		"description": {
			"type": "doc",
			"content": [{"type": "paragraph", "content": [{"type": "text", "text": "Repro steps inside"}]}]
		},
		"customfield_10015": 8
	}
}`

func TestClient_GetIssue_RequestShape(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: fullIssueJSON})

	if _, err := client.GetIssue(t.Context(), "PLAT-42"); err != nil {
		t.Fatalf("GetIssue returned error: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodGet)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-42")
	testkit.AssertEqual(t, "Accept header", recorded.Header.Get("Accept"), "application/json")

	if auth := recorded.Header.Get("Authorization"); !strings.HasPrefix(auth, "Basic ") {
		t.Errorf("Authorization = %q, want Basic scheme", auth)
	}
}

func TestClient_GetIssue_MapsFields(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: fullIssueJSON})

	issue, err := client.GetIssue(t.Context(), "PLAT-42")
	if err != nil {
		t.Fatalf("GetIssue returned error: %v", err)
	}

	testkit.AssertEqual(t, "Key", issue.Key, "PLAT-42")
	testkit.AssertEqual(t, "ID", issue.ID, "10001")
	testkit.AssertEqual(t, "Summary", issue.Summary, "Fix login redirect loop")
	testkit.AssertEqual(t, "Description", issue.Description, "Repro steps inside\n")

	if issue.Status == nil {
		t.Fatal("Status is nil, want In Progress")
	}
	testkit.AssertEqual(t, "Status.Name", issue.Status.Name, "In Progress")
	testkit.AssertEqual(t, "Status.CategoryKey", issue.Status.CategoryKey, "indeterminate")

	if issue.Priority == nil {
		t.Fatal("Priority is nil, want High")
	}
	testkit.AssertEqual(t, "Priority.Name", issue.Priority.Name, "High")

	if issue.Assignee == nil {
		t.Fatal("Assignee is nil, want Ada Lovelace")
	}
	testkit.AssertEqual(t, "Assignee.AccountID", issue.Assignee.AccountID, "acc-1")
	testkit.AssertEqual(t, "Assignee.DisplayName", issue.Assignee.DisplayName, "Ada Lovelace")

	testkit.AssertSliceEqual(t, "Labels", issue.Labels, []string{"backend", "auth"})

	storyPoints, ok := issue.CustomFields["customfield_10015"].(float64)
	if !ok {
		t.Fatalf("CustomFields[customfield_10015] = %v, want a float64", issue.CustomFields["customfield_10015"])
	}
	testkit.AssertEqual(t, "CustomFields[customfield_10015]", storyPoints, float64(8))
}

func TestClient_GetIssue_ServerNameFallback(t *testing.T) {
	t.Parallel()

	const serverIssueJSON = `{
		"id": "5",
		"key": "OPS-7",
		"fields": {
			"summary": "Rotate certs",
			"assignee": {"name": "jsmith", "displayName": "John Smith"}
		}
	}`

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{Status: http.StatusOK, Body: serverIssueJSON})

	issue, err := client.GetIssue(t.Context(), "OPS-7")
	if err != nil {
		t.Fatalf("GetIssue returned error: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/2/issue/OPS-7")
	if auth := recorded.Header.Get("Authorization"); auth != "Bearer pat-token" {
		t.Errorf("Authorization = %q, want Bearer pat-token", auth)
	}
	if issue.Assignee == nil {
		t.Fatal("Assignee is nil, want John Smith")
	}
	testkit.AssertEqual(t, "Assignee.AccountID falls back to name", issue.Assignee.AccountID, "jsmith")
}

func TestClient_GetIssue_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		response testkit.StubResponse
		wantErr  string
	}{
		{
			name:     "not found wraps issue key",
			response: testkit.StubResponse{Status: http.StatusNotFound, Body: `{"errorMessages":["Issue does not exist"]}`},
			wantErr:  "PLAT-404",
		},
		{
			name:     "server error reports status",
			response: testkit.StubResponse{Status: http.StatusInternalServerError, Body: "upstream exploded"},
			wantErr:  "status 500",
		},
		{
			name:     "malformed body surfaces decode failure",
			response: testkit.StubResponse{Status: http.StatusOK, Body: "{ not json"},
			wantErr:  "decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, _ := newRecordingClient(t, cloudOpts(), tt.response)

			issue, err := client.GetIssue(t.Context(), "PLAT-404")
			if err == nil {
				t.Fatalf("expected error, got nil (issue=%v)", issue)
			}
			if issue != nil {
				t.Errorf("issue = %v, want nil on error", issue)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
