package jira

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestParseErrorMessages_ErrorMessagesArray(t *testing.T) {
	t.Parallel()

	body := `{"errorMessages":["You cannot create issues in this project."],"errors":{}}`
	got := parseErrorMessages(body)

	want := []string{"You cannot create issues in this project."}
	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("parseErrorMessages(%q) = %v, want %v", body, got, want)
	}
}

func TestParseErrorMessages_FieldErrorsMap(t *testing.T) {
	t.Parallel()

	body := `{"errorMessages":[],"errors":{"parent":"Issue does not exist or you do not have permission to see it."}}`
	got := parseErrorMessages(body)

	want := []string{"Issue does not exist or you do not have permission to see it."}
	if len(got) != len(want) || got[0] != want[0] {
		t.Errorf("parseErrorMessages(%q) = %v, want %v", body, got, want)
	}
}

func TestParseErrorMessages_NonEnvelopeReturnsNil(t *testing.T) {
	t.Parallel()

	if got := parseErrorMessages("plain text, not json"); got != nil {
		t.Errorf("parseErrorMessages(non-json) = %v, want nil", got)
	}
}

func TestClient_HTTPError_ExposesAPIErrorWithMessages(t *testing.T) {
	t.Parallel()

	body := `{"errorMessages":["You cannot create issues in this project."]}`
	client, _ := newRecordingClient(t, cloudOpts(),
		testkit.StubResponse{Status: http.StatusNotFound, Body: body})

	_, err := client.GetCreateMeta(t.Context(), errorTestProjectKey, "10001")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error %v is not unwrappable to *APIError", err)
	}
	if len(apiErr.Messages) != 1 ||
		apiErr.Messages[0] != "You cannot create issues in this project." {
		t.Errorf("apiErr.Messages = %v", apiErr.Messages)
	}

	// Backward compatibility: the Error() string keeps status and wrapper.
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("Error() %q missing 'status 404'", err.Error())
	}
	if !strings.Contains(err.Error(), "get create meta") {
		t.Errorf("Error() %q missing wrapper prefix", err.Error())
	}
}
