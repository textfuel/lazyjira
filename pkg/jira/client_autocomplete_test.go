package jira

import (
	"net/http"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestClient_GetJQLAutocompleteData_ParsesVisibleFields(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"visibleFieldNames": [{"value": "status", "displayName": "Status", "operators": ["=", "!="]}]}`,
	})

	fields, err := client.GetJQLAutocompleteData(t.Context())
	if err != nil {
		t.Fatalf("GetJQLAutocompleteData: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/jql/autocompletedata")
	if len(fields) != 1 {
		t.Fatalf("len(fields) = %d, want 1", len(fields))
	}
	testkit.AssertEqual(t, "fields[0].Value", fields[0].Value, "status")
	testkit.AssertEqual(t, "fields[0].DisplayName", fields[0].DisplayName, "Status")
	testkit.AssertSliceEqual(t, "fields[0].Operators", fields[0].Operators, []string{"=", "!="})
}

func TestClient_GetJQLAutocompleteSuggestions_PassesFieldNameAndValue(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"results": [{"value": "Done", "displayName": "Done"}]}`,
	})

	suggestions, err := client.GetJQLAutocompleteSuggestions(t.Context(), "status", "Do")
	if err != nil {
		t.Fatalf("GetJQLAutocompleteSuggestions: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/jql/autocompletedata/suggestions")
	testkit.AssertEqual(t, "fieldName query", recorded.Query.Get("fieldName"), "status")
	testkit.AssertEqual(t, "fieldValue query", recorded.Query.Get("fieldValue"), "Do")
	if len(suggestions) != 1 {
		t.Fatalf("len(suggestions) = %d, want 1", len(suggestions))
	}
	testkit.AssertEqual(t, "suggestions[0].Value", suggestions[0].Value, "Done")
}

func TestClient_DiscoverFields_ResolvesSprintFieldID(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body: `[
			{"id": "summary", "name": "Summary", "schema": {}},
			{"id": "customfield_10020", "name": "Sprint", "schema": {"custom": "com.pyxis.greenhopper.jira:gh-sprint"}}
		]`,
	})

	if err := client.DiscoverFields(t.Context()); err != nil {
		t.Fatalf("DiscoverFields: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/field")
	testkit.AssertEqual(t, "SprintFieldID", client.SprintFieldID(), "customfield_10020")
}

func TestClient_DiscoverFields_NoSprintFieldKeepsAlias(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "summary", "name": "Summary", "schema": {}}]`,
	})

	if err := client.DiscoverFields(t.Context()); err != nil {
		t.Fatalf("DiscoverFields: %v", err)
	}

	testkit.AssertEqual(t, "SprintFieldID fallback", client.SprintFieldID(), "sprint")
}
