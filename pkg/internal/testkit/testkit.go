package testkit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"
)

type RecordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   []byte
}

type StubResponse struct {
	Status int
	Body   string
}

func RecordingServer(t *testing.T, response StubResponse) (*httptest.Server, *RecordedRequest) {
	t.Helper()

	recorded := &RecordedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		recorded.Method = r.Method
		recorded.Path = r.URL.Path
		recorded.Query = r.URL.Query()
		recorded.Header = r.Header.Clone()
		recorded.Body = body

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.Status)
		_, _ = io.WriteString(w, response.Body)
	}))
	t.Cleanup(server.Close)

	return server, recorded
}

func AssertEqual[T comparable](t *testing.T, label string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

func AssertSliceEqual[T comparable](t *testing.T, label string, got, want []T) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}
