package jira

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestNewClientWithOpts_CloudVsServer(t *testing.T) {
	t.Parallel()

	cloud := NewClientWithOpts(ClientOpts{
		Host:    "https://test.atlassian.net",
		Email:   "user@test.com",
		Token:   "tok123",
		IsCloud: true,
	})
	if !strings.HasSuffix(cloud.baseURL, "/rest/api/3") {
		t.Errorf("Cloud: expected API v3, got %s", cloud.baseURL)
	}
	if !strings.HasPrefix(cloud.authHeader, "Basic ") {
		t.Errorf("Cloud: expected Basic auth, got %s", cloud.authHeader)
	}

	server := NewClientWithOpts(ClientOpts{
		Host:    "https://jira.corp.com",
		Token:   "pat-token",
		IsCloud: false,
	})
	if !strings.HasSuffix(server.baseURL, "/rest/api/2") {
		t.Errorf("Server: expected API v2, got %s", server.baseURL)
	}
	if server.authHeader != "Bearer pat-token" {
		t.Errorf("Server: expected Bearer auth, got %s", server.authHeader)
	}
}

func TestNewClientWithOpts_HostNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"https://jira.com", "https://jira.com"},
		{"http://jira.com", "http://jira.com"},
		{"jira.com", "https://jira.com"},
		{"https://jira.com/", "https://jira.com"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			client := NewClientWithOpts(ClientOpts{Host: tt.input, Token: "x", IsCloud: false})
			testkit.AssertEqual(t, "hostURL", client.hostURL, tt.want)
		})
	}
}

func TestClient_GetChildren_CloudFiresJQL(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"issues":[],"total":0,"maxResults":100,"startAt":0}`,
	})

	if _, err := client.GetChildren(t.Context(), "PROJ-123"); err != nil {
		t.Fatalf("GetChildren returned error: %v", err)
	}

	if !strings.HasSuffix(recorded.Path, "/search/jql") {
		t.Errorf("Cloud: expected /search/jql, got %s", recorded.Path)
	}
	testkit.AssertEqual(t, "jql query", recorded.Query.Get("jql"), "parent = PROJ-123")
}

type countingRoundTripper struct {
	calls int
}

func (transport *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	transport.calls++
	return nil, http.ErrUseLastResponse
}

func TestClient_GetChildren_ServerDCNoCall(t *testing.T) {
	t.Parallel()

	transport := &countingRoundTripper{}
	client := NewClientWithOpts(ClientOpts{
		Host:       "https://jira.corp.example",
		Token:      "pat-token",
		IsCloud:    false,
		HTTPClient: &http.Client{Transport: transport},
	})

	issues, err := client.GetChildren(t.Context(), "PROJ-123")
	if err != nil {
		t.Fatalf("Server/DC GetChildren: unexpected error %v", err)
	}
	if issues != nil {
		t.Errorf("Server/DC GetChildren: expected nil slice, got %v", issues)
	}
	testkit.AssertEqual(t, "HTTP call count", transport.calls, 0)
}

func TestClient_UpdateIssue_ParentSet(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	err := client.UpdateIssue(t.Context(), "PROJ-2", map[string]any{
		"parent": map[string]string{"key": "PROJ-1"},
	})
	if err != nil {
		t.Fatalf("UpdateIssue: %v", err)
	}

	if !strings.HasSuffix(recorded.Path, "/issue/PROJ-2") {
		t.Errorf("path = %q", recorded.Path)
	}

	body := decodeBody(t, recorded.Body)
	fields, _ := body["fields"].(map[string]any)
	parent, _ := fields["parent"].(map[string]any)
	if parent["key"] != "PROJ-1" {
		t.Errorf("body fields.parent.key = %v, want PROJ-1", parent["key"])
	}
}

func TestClient_RemoveIssueParent_Cloud(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	if err := client.RemoveIssueParent(t.Context(), "PROJ-2"); err != nil {
		t.Fatalf("RemoveIssueParent: %v", err)
	}

	body := decodeBody(t, recorded.Body)
	fields, ok := body["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields wrapper, got %#v", body)
	}
	if value, exists := fields["parent"]; !exists || value != nil {
		t.Errorf("fields.parent = %v (exists=%v), want nil literal", value, exists)
	}
	if _, exists := body["update"]; exists {
		t.Errorf("Cloud body should not contain 'update', got %#v", body)
	}
}

func TestClient_RemoveIssueParent_DC(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{Status: http.StatusNoContent})

	if err := client.RemoveIssueParent(t.Context(), "PROJ-2"); err != nil {
		t.Fatalf("RemoveIssueParent: %v", err)
	}

	body := decodeBody(t, recorded.Body)
	update, ok := body["update"].(map[string]any)
	if !ok {
		t.Fatalf("expected update wrapper, got %#v", body)
	}
	operations, ok := update["parent"].([]any)
	if !ok || len(operations) != 1 {
		t.Fatalf("update.parent = %#v, want [{remove:{}}]", update["parent"])
	}
	operation, _ := operations[0].(map[string]any)
	if _, exists := operation["remove"]; !exists {
		t.Errorf("update.parent[0] = %#v, want remove op", operation)
	}
	if _, exists := body["fields"]; exists {
		t.Errorf("DC body should not contain 'fields', got %#v", body)
	}
}

func TestUserResponse_ToUser_FallbackToName(t *testing.T) {
	t.Parallel()

	cloud := &userResponse{AccountID: "abc123", Name: "jsmith", DisplayName: "Alice"}
	testkit.AssertEqual(t, "cloud AccountID", cloud.toUser().AccountID, "abc123")

	server := &userResponse{Name: "jsmith", DisplayName: "John"}
	testkit.AssertEqual(t, "server AccountID fallback", server.toUser().AccountID, "jsmith")
}

func decodeBody(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal request body %q: %v", raw, err)
	}
	return body
}
