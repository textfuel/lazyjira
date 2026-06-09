package jira

import (
	"net/http"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestClient_GetProjects_CloudPaginates(t *testing.T) {
	t.Parallel()

	client, recorded := newSequenceClient(t, cloudOpts(),
		testkit.StubResponse{Status: http.StatusOK, Body: `{"values": [{"id": "1", "key": "PLAT", "name": "Platform", "lead": {"displayName": "Lead One"}}], "total": 2}`},
		testkit.StubResponse{Status: http.StatusOK, Body: `{"values": [{"id": "2", "key": "OPS", "name": "Operations"}], "total": 2}`},
	)

	projects, err := client.GetProjects(t.Context())
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	if len(*recorded) != 2 {
		t.Fatalf("made %d requests, want 2", len(*recorded))
	}
	testkit.AssertEqual(t, "first path", (*recorded)[0].Path, "/rest/api/3/project/search")
	testkit.AssertEqual(t, "second startAt", (*recorded)[1].Query.Get("startAt"), "1")

	if len(projects) != 2 {
		t.Fatalf("len(projects) = %d, want 2", len(projects))
	}
	testkit.AssertEqual(t, "projects[0].Key", projects[0].Key, "PLAT")
	testkit.AssertEqual(t, "projects[1].Key", projects[1].Key, "OPS")
	if projects[0].Lead == nil {
		t.Fatal("projects[0].Lead is nil")
	}
	testkit.AssertEqual(t, "projects[0].Lead.DisplayName", projects[0].Lead.DisplayName, "Lead One")
}

func TestClient_GetProjects_ServerUsesProjectList(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "1", "key": "OPS", "name": "Operations"}]`,
	})

	projects, err := client.GetProjects(t.Context())
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/2/project")
	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}
	testkit.AssertEqual(t, "projects[0].Key", projects[0].Key, "OPS")
}

func TestClient_GetBoards_ParsesLocation(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"values": [{"id": 7, "name": "Board A", "type": "scrum", "location": {"projectKey": "PLAT"}}], "isLast": true}`,
	})

	boards, err := client.GetBoards(t.Context())
	if err != nil {
		t.Fatalf("GetBoards: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/agile/1.0/board")
	if len(boards) != 1 {
		t.Fatalf("len(boards) = %d, want 1", len(boards))
	}
	testkit.AssertEqual(t, "boards[0].ID", boards[0].ID, 7)
	testkit.AssertEqual(t, "boards[0].ProjectKey", boards[0].ProjectKey, "PLAT")
}

func TestClient_GetBoardIssues_UsesAgilePath(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"issues": [{"id": "1", "key": "PLAT-1", "fields": {"summary": "x"}}], "total": 1, "maxResults": 50, "startAt": 0}`,
	})

	issues, err := client.GetBoardIssues(t.Context(), 5, "status = Done")
	if err != nil {
		t.Fatalf("GetBoardIssues: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/agile/1.0/board/5/issue")
	testkit.AssertEqual(t, "jql query", recorded.Query.Get("jql"), "status = Done")
	if len(issues) != 1 {
		t.Fatalf("len(issues) = %d, want 1", len(issues))
	}
	testkit.AssertEqual(t, "issues[0].Key", issues[0].Key, "PLAT-1")
}

func TestClient_GetPriorities_ParsesList(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "1", "name": "High"}, {"id": "2", "name": "Low"}]`,
	})

	priorities, err := client.GetPriorities(t.Context())
	if err != nil {
		t.Fatalf("GetPriorities: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/priority")
	if len(priorities) != 2 {
		t.Fatalf("len(priorities) = %d, want 2", len(priorities))
	}
	testkit.AssertEqual(t, "priorities[0].Name", priorities[0].Name, "High")
}

func TestClient_GetComponents_UsesProjectPath(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "10", "name": "backend"}]`,
	})

	components, err := client.GetComponents(t.Context(), "PLAT")
	if err != nil {
		t.Fatalf("GetComponents: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/project/PLAT/components")
	if len(components) != 1 {
		t.Fatalf("len(components) = %d, want 1", len(components))
	}
	testkit.AssertEqual(t, "components[0].Name", components[0].Name, "backend")
}

func TestClient_GetIssueTypes_CloudUsesProjectQuery(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "1", "name": "Story", "subtask": false, "hierarchyLevel": 0}]`,
	})

	issueTypes, err := client.GetIssueTypes(t.Context(), "10000")
	if err != nil {
		t.Fatalf("GetIssueTypes: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issuetype/project")
	testkit.AssertEqual(t, "projectId query", recorded.Query.Get("projectId"), "10000")
	if len(issueTypes) != 1 {
		t.Fatalf("len(issueTypes) = %d, want 1", len(issueTypes))
	}
	testkit.AssertEqual(t, "issueTypes[0].Name", issueTypes[0].Name, "Story")
}

func TestClient_GetIssueTypes_ServerListsAll(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "1", "name": "Bug"}]`,
	})

	if _, err := client.GetIssueTypes(t.Context(), "10000"); err != nil {
		t.Fatalf("GetIssueTypes: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/2/issuetype")
}

func TestClient_GetLabels_ParsesValues(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"values": ["backend", "frontend"], "total": 2}`,
	})

	labels, err := client.GetLabels(t.Context())
	if err != nil {
		t.Fatalf("GetLabels: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/label")
	testkit.AssertSliceEqual(t, "labels", labels, []string{"backend", "frontend"})
}

func TestClient_GetCreateMeta_CloudParsesFields(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body: `{
			"startAt": 0, "maxResults": 200, "total": 2,
			"fields": [
				{"fieldId": "summary", "name": "Summary", "required": true, "schema": {"type": "string", "system": "summary"}},
				{"fieldId": "priority", "name": "Priority", "required": false, "schema": {"type": "priority", "system": "priority"}, "allowedValues": [{"id": "2", "name": "High"}]}
			]
		}`,
	})

	fields, err := client.GetCreateMeta(t.Context(), "PLAT", "10001")
	if err != nil {
		t.Fatalf("GetCreateMeta: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/createmeta/PLAT/issuetypes/10001")
	if len(fields) != 2 {
		t.Fatalf("len(fields) = %d, want 2", len(fields))
	}
	testkit.AssertEqual(t, "fields[0].FieldID", fields[0].FieldID, "summary")
	testkit.AssertEqual(t, "fields[0].Required", fields[0].Required, true)
	if len(fields[1].AllowedValues) != 1 {
		t.Fatalf("priority AllowedValues = %#v, want one", fields[1].AllowedValues)
	}
	testkit.AssertEqual(t, "priority allowed value", fields[1].AllowedValues[0].Name, "High")
}

func TestClient_GetCreateMeta_ServerParsesNestedFields(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body: `{"projects": [{"issuetypes": [{"fields": {
			"summary": {"name": "Summary", "required": true, "schema": {"type": "string", "system": "summary"}}
		}}]}]}`,
	})

	fields, err := client.GetCreateMeta(t.Context(), "OPS", "10002")
	if err != nil {
		t.Fatalf("GetCreateMeta: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/2/issue/createmeta")
	testkit.AssertEqual(t, "projectKeys query", recorded.Query.Get("projectKeys"), "OPS")
	if len(fields) != 1 {
		t.Fatalf("len(fields) = %d, want 1", len(fields))
	}
	testkit.AssertEqual(t, "fields[0].FieldID", fields[0].FieldID, "summary")
}
