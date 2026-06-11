package jira

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestNewClient_DefaultsToCloudV3(t *testing.T) {
	t.Parallel()

	client := NewClient("example.atlassian.net", "user@example.com", "api-token")

	testkit.AssertEqual(t, "BaseURL", client.BaseURL(), "https://example.atlassian.net/rest/api/3")
	testkit.AssertEqual(t, "IsCloud", client.IsCloud(), true)

	wantCredentials := base64.StdEncoding.EncodeToString([]byte("user@example.com:api-token"))
	testkit.AssertEqual(t, "AuthHeader", client.AuthHeader(), "Basic "+wantCredentials)

	if client.HTTPClient() == nil {
		t.Fatal("HTTPClient is nil")
	}
	testkit.AssertEqual(t, "HTTPClient timeout", client.HTTPClient().Timeout, 30*time.Second)
}

func TestNewOAuthClient_UsesAtlassianGateway(t *testing.T) {
	t.Parallel()

	client := NewOAuthClient("cloud-id-123", "access-token")

	testkit.AssertEqual(t, "BaseURL", client.BaseURL(), "https://api.atlassian.com/ex/jira/cloud-id-123/rest/api/3")
	testkit.AssertEqual(t, "AuthHeader", client.AuthHeader(), "Bearer access-token")
	testkit.AssertEqual(t, "IsCloud", client.IsCloud(), true)
	if client.HTTPClient() == nil {
		t.Fatal("HTTPClient is nil")
	}
}

func TestClient_IsCloud_FalseForServer(t *testing.T) {
	t.Parallel()

	client := NewClientWithOpts(serverOpts())
	testkit.AssertEqual(t, "IsCloud", client.IsCloud(), false)
}

func TestClient_SetDryRun_SkipsWrite(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "[]"})
	var logBuffer bytes.Buffer
	client.SetLogger(&logBuffer)
	client.SetDryRun(true)

	if err := client.DoTransition(t.Context(), "PLAT-1", "31"); err != nil {
		t.Fatalf("DoTransition in dry-run: %v", err)
	}
	testkit.AssertEqual(t, "no request dispatched for write", recorded.Method, "")
	if !strings.Contains(logBuffer.String(), "[DRY-RUN] skipped write operation") {
		t.Errorf("log %q does not mention the dry-run skip", logBuffer.String())
	}
}

func TestClient_SetDryRun_AllowsRead(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "[]"})
	client.SetDryRun(true)

	if _, err := client.GetPriorities(t.Context()); err != nil {
		t.Fatalf("GetPriorities in dry-run: %v", err)
	}
	testkit.AssertEqual(t, "read request dispatched", recorded.Method, http.MethodGet)
}

func TestClient_SetLogger_WritesRequestAndResponseLines(t *testing.T) {
	t.Parallel()

	t.Run("success logs status and size", func(t *testing.T) {
		t.Parallel()

		client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "[]"})
		var logBuffer bytes.Buffer
		client.SetLogger(&logBuffer)

		if _, err := client.GetPriorities(t.Context()); err != nil {
			t.Fatalf("GetPriorities: %v", err)
		}

		logged := logBuffer.String()
		if !strings.Contains(logged, "GET") || !strings.Contains(logged, "/rest/api/3/priority") {
			t.Errorf("log %q does not record the request line", logged)
		}
		if !strings.Contains(logged, "-> 200") {
			t.Errorf("log %q does not record the response status", logged)
		}
	})

	t.Run("failure logs response body", func(t *testing.T) {
		t.Parallel()

		client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusInternalServerError, Body: "boom"})
		var logBuffer bytes.Buffer
		client.SetLogger(&logBuffer)

		if _, err := client.GetPriorities(t.Context()); err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(logBuffer.String(), "BODY: boom") {
			t.Errorf("log %q does not record the error body", logBuffer.String())
		}
	})
}

func TestClient_SetOnRequest_ReceivesRequestLog(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "[]"})
	var received []RequestLog
	client.SetOnRequest(func(entry RequestLog) { received = append(received, entry) })

	if _, err := client.GetPriorities(t.Context()); err != nil {
		t.Fatalf("GetPriorities: %v", err)
	}

	if len(received) != 1 {
		t.Fatalf("received %d request logs, want 1", len(received))
	}
	testkit.AssertEqual(t, "Method", received[0].Method, http.MethodGet)
	testkit.AssertEqual(t, "Path", received[0].Path, "/priority")
	testkit.AssertEqual(t, "Status", received[0].Status, http.StatusOK)
	if received[0].Elapsed < 0 {
		t.Errorf("Elapsed = %v, want non-negative", received[0].Elapsed)
	}
}

func TestClient_RequestEdgeFailures(t *testing.T) {
	t.Parallel()

	t.Run("unmarshalable body", func(t *testing.T) {
		t.Parallel()

		client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "{}"})

		_, err := client.AddComment(t.Context(), "PLAT-1", make(chan int))
		if err == nil || !strings.Contains(err.Error(), "marshal request body") {
			t.Errorf("error = %v, want marshal request body failure", err)
		}
	})

	t.Run("invalid host fails request creation", func(t *testing.T) {
		t.Parallel()

		opts := cloudOpts()
		opts.Host = "http://invalid host"
		client := NewClientWithOpts(opts)

		_, err := client.GetPriorities(t.Context())
		if err == nil || !strings.Contains(err.Error(), "create request") {
			t.Errorf("error = %v, want create request failure", err)
		}
	})

	t.Run("unreachable server fails transport", func(t *testing.T) {
		t.Parallel()

		server, _ := testkit.RecordingServer(t, testkit.StubResponse{Status: http.StatusOK, Body: "[]"})
		opts := cloudOpts()
		opts.Host = server.URL
		server.Close()
		client := NewClientWithOpts(opts)

		_, err := client.GetPriorities(t.Context())
		if err == nil || !strings.Contains(err.Error(), "execute request GET /priority") {
			t.Errorf("error = %v, want execute request failure", err)
		}
	})

	t.Run("truncated response body fails read", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Length", "1000")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "short")
		}))
		t.Cleanup(server.Close)
		opts := cloudOpts()
		opts.Host = server.URL
		client := NewClientWithOpts(opts)

		_, err := client.GetPriorities(t.Context())
		if err == nil || !strings.Contains(err.Error(), "read response body") {
			t.Errorf("error = %v, want read response body failure", err)
		}
	})
}
