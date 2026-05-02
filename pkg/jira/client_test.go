package jira

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// countingRoundTripper records every HTTP call without performing one.
type countingRoundTripper struct {
	calls int
}

func (c *countingRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, http.ErrUseLastResponse // any error suffices; we only count
}

func TestClient_GetChildren_CloudFiresJQL(t *testing.T) {
	var (
		gotPath string
		gotJQL  string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotJQL = r.URL.Query().Get("jql")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issues":[],"total":0,"maxResults":100,"startAt":0}`))
	}))
	t.Cleanup(srv.Close)

	c := NewClientWithOpts(ClientOpts{
		Host:    srv.URL,
		Email:   "u",
		Token:   "t",
		IsCloud: true,
	})

	if _, err := c.GetChildren(context.Background(), "PROJ-123"); err != nil {
		t.Fatalf("GetChildren returned error: %v", err)
	}

	if !strings.HasSuffix(gotPath, "/search/jql") {
		t.Errorf("Cloud: expected /search/jql, got %s", gotPath)
	}
	if gotJQL != "parent = PROJ-123" {
		t.Errorf("Cloud JQL: got %q, want %q", gotJQL, "parent = PROJ-123")
	}
}

func TestClient_GetChildren_ServerDCNoCall(t *testing.T) {
	rt := &countingRoundTripper{}
	c := NewClientWithOpts(ClientOpts{
		Host:       "https://jira.corp.example",
		Token:      "pat",
		IsCloud:    false,
		HTTPClient: &http.Client{Transport: rt},
	})

	issues, err := c.GetChildren(context.Background(), "PROJ-123")
	if err != nil {
		t.Fatalf("Server/DC GetChildren: unexpected error %v", err)
	}
	if issues != nil {
		t.Errorf("Server/DC GetChildren: expected nil slice, got %v", issues)
	}
	if rt.calls != 0 {
		t.Errorf("Server/DC GetChildren: expected 0 HTTP calls, got %d", rt.calls)
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
