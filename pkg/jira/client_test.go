package jira

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewClientWithOpts_CloudVsServer(t *testing.T) {
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
		c := NewClientWithOpts(ClientOpts{Host: tt.input, Token: "x", IsCloud: false})
		if c.hostURL != tt.want {
			t.Errorf("Host %q: got %q, want %q", tt.input, c.hostURL, tt.want)
		}
	}
}

func TestIssueFieldsResponse_UnmarshalJSON_StashesSystemExtras(t *testing.T) {
	body := []byte(`{
		"summary": "the headline",
		"fixVersions": [{"name": "v1"}, {"name": "v2"}],
		"resolution": {"name": "Done"},
		"duedate": "2026-05-15",
		"customfield_10042": "anything"
	}`)

	var f issueFieldsResponse
	if err := json.Unmarshal(body, &f); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if f.Summary != "the headline" {
		t.Errorf("Summary lost: %q", f.Summary)
	}
	if _, ok := f.RawExtra["summary"]; ok {
		t.Error("summary leaked into RawExtra")
	}
	for _, key := range []string{"fixVersions", "resolution", "duedate", "customfield_10042"} {
		if _, ok := f.RawExtra[key]; !ok {
			t.Errorf("RawExtra missing %q", key)
		}
	}
}

func TestIsSystemExtraField(t *testing.T) {
	for _, id := range []string{"fixVersions", "versions", "duedate", "resolution", "environment"} {
		if !IsSystemExtraField(id) {
			t.Errorf("expected %q to be a system extra field", id)
		}
	}
	for _, id := range []string{"summary", "status", "customfield_10001", "fixversions", ""} {
		if IsSystemExtraField(id) {
			t.Errorf("expected %q NOT to be a system extra field", id)
		}
	}
}

func TestUserResponse_ToUser_FallbackToName(t *testing.T) {
	// Cloud: accountId set
	cloud := &userResponse{AccountID: "abc123", Name: "jsmith", DisplayName: "Alice"}
	u := cloud.toUser()
	if u.AccountID != "abc123" {
		t.Errorf("with accountId: got %q, want abc123", u.AccountID)
	}

	// Server: only name set
	server := &userResponse{Name: "jsmith", DisplayName: "John"}
	u = server.toUser()
	if u.AccountID != "jsmith" {
		t.Errorf("name fallback: got %q, want jsmith", u.AccountID)
	}
}
