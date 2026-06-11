package jira

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

const relationsIssueJSON = `{
	"id": "10002",
	"key": "PLAT-50",
	"fields": {
		"summary": "Parent linked issue",
		"reporter": {"accountId": "acc-2", "displayName": "Grace Hopper"},
		"parent": {"id": "10000", "key": "PLAT-10", "fields": {"summary": "Epic"}},
		"subtasks": [{"id": "10003", "key": "PLAT-51", "fields": {"summary": "Subtask one"}}],
		"issuelinks": [
			{"id": "1", "type": {"name": "Blocks", "inward": "is blocked by", "outward": "blocks"}, "inwardIssue": {"id": "2", "key": "PLAT-2", "fields": {"summary": "Blocker"}}},
			{"id": "2", "type": {"name": "Relates", "inward": "relates to", "outward": "relates to"}, "outwardIssue": {"id": "3", "key": "PLAT-3", "fields": {"summary": "Related"}}}
		],
		"sprint": {"id": 5, "name": "Sprint 5", "state": "active"}
	}
}`

func TestClient_GetIssue_MapsRelations(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: relationsIssueJSON})

	issue, err := client.GetIssue(t.Context(), "PLAT-50")
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}

	if issue.Reporter == nil {
		t.Fatal("Reporter is nil")
	}
	testkit.AssertEqual(t, "Reporter.DisplayName", issue.Reporter.DisplayName, "Grace Hopper")

	if issue.Parent == nil {
		t.Fatal("Parent is nil")
	}
	testkit.AssertEqual(t, "Parent.Key", issue.Parent.Key, "PLAT-10")

	if len(issue.Subtasks) != 1 {
		t.Fatalf("len(Subtasks) = %d, want 1", len(issue.Subtasks))
	}
	testkit.AssertEqual(t, "Subtasks[0].Key", issue.Subtasks[0].Key, "PLAT-51")

	if len(issue.IssueLinks) != 2 {
		t.Fatalf("len(IssueLinks) = %d, want 2", len(issue.IssueLinks))
	}
	if issue.IssueLinks[0].Type == nil || issue.IssueLinks[0].InwardIssue == nil {
		t.Fatalf("IssueLinks[0] = %+v, want type and inward issue", issue.IssueLinks[0])
	}
	testkit.AssertEqual(t, "IssueLinks[0].Type.Name", issue.IssueLinks[0].Type.Name, "Blocks")
	testkit.AssertEqual(t, "IssueLinks[0].InwardIssue.Key", issue.IssueLinks[0].InwardIssue.Key, "PLAT-2")
	if issue.IssueLinks[0].OutwardIssue != nil {
		t.Errorf("IssueLinks[0].OutwardIssue = %+v, want nil", issue.IssueLinks[0].OutwardIssue)
	}
	if issue.IssueLinks[1].OutwardIssue == nil {
		t.Fatal("IssueLinks[1].OutwardIssue is nil")
	}
	testkit.AssertEqual(t, "IssueLinks[1].OutwardIssue.Key", issue.IssueLinks[1].OutwardIssue.Key, "PLAT-3")

	if issue.Sprint == nil {
		t.Fatal("Sprint is nil")
	}
	testkit.AssertEqual(t, "Sprint.Name from sprint field", issue.Sprint.Name, "Sprint 5")
}

const sprintDiscoveryFieldsJSON = `[{"id": "customfield_10020", "name": "Sprint", "schema": {"custom": "com.pyxis.greenhopper.jira:gh-sprint"}}]`

func TestClient_GetIssue_SprintFromDiscoveredCustomField(t *testing.T) {
	t.Parallel()

	const issueJSON = `{
		"id": "1", "key": "PLAT-60",
		"fields": {
			"summary": "Sprint via discovered field",
			"customfield_10020": [{"id": 7, "name": "Sprint 7", "state": "active"}]
		}
	}`

	client, _ := newSequenceClient(t, cloudOpts(),
		testkit.StubResponse{Status: http.StatusOK, Body: sprintDiscoveryFieldsJSON},
		testkit.StubResponse{Status: http.StatusOK, Body: issueJSON},
	)

	if err := client.DiscoverFields(t.Context()); err != nil {
		t.Fatalf("DiscoverFields: %v", err)
	}

	issue, err := client.GetIssue(t.Context(), "PLAT-60")
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if issue.Sprint == nil {
		t.Fatal("Sprint is nil")
	}
	testkit.AssertEqual(t, "Sprint.ID", issue.Sprint.ID, 7)
	testkit.AssertEqual(t, "Sprint.Name", issue.Sprint.Name, "Sprint 7")
}

func TestClient_GetIssue_SprintScansOtherCustomFieldsWhenDiscoveredFieldEmpty(t *testing.T) {
	t.Parallel()

	const issueJSON = `{
		"id": "1", "key": "PLAT-61",
		"fields": {
			"summary": "Sprint via scan",
			"customfield_10020": [],
			"customfield_10333": [{"id": 9, "name": "Sprint 9", "state": "active"}]
		}
	}`

	client, _ := newSequenceClient(t, cloudOpts(),
		testkit.StubResponse{Status: http.StatusOK, Body: sprintDiscoveryFieldsJSON},
		testkit.StubResponse{Status: http.StatusOK, Body: issueJSON},
	)

	if err := client.DiscoverFields(t.Context()); err != nil {
		t.Fatalf("DiscoverFields: %v", err)
	}

	issue, err := client.GetIssue(t.Context(), "PLAT-61")
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if issue.Sprint == nil {
		t.Fatal("Sprint is nil")
	}
	testkit.AssertEqual(t, "Sprint.ID", issue.Sprint.ID, 9)
}

func TestClient_GetMyself_ExtractsAvatar(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `{"accountId": "acc-1", "displayName": "Ada", "avatarUrls": {"48x48": "https://avatars.example.com/ada-48.png"}}`,
	})

	user, err := client.GetMyself(t.Context())
	if err != nil {
		t.Fatalf("GetMyself: %v", err)
	}
	testkit.AssertEqual(t, "AvatarURL", user.AvatarURL, "https://avatars.example.com/ada-48.png")
}

func TestClient_GetProjects_ExtractsAvatar(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, serverOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body:   `[{"id": "1", "key": "OPS", "name": "Operations", "avatarUrls": {"48x48": "https://avatars.example.com/ops-48.png"}}]`,
	})

	projects, err := client.GetProjects(t.Context())
	if err != nil {
		t.Fatalf("GetProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}
	testkit.AssertEqual(t, "AvatarURL", projects[0].AvatarURL, "https://avatars.example.com/ops-48.png")
}

func TestClient_GetComments_ADFBodyKeepsRawDocument(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusOK,
		Body: `{"comments": [{
			"id": "100",
			"body": {"type": "doc", "content": [{"type": "paragraph", "content": [{"type": "text", "text": "rich comment"}]}]}
		}]}`,
	})

	comments, err := client.GetComments(t.Context(), "PLAT-50")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("len(comments) = %d, want 1", len(comments))
	}
	testkit.AssertEqual(t, "Body", comments[0].Body, "rich comment\n")
	adfDoc, isMap := comments[0].BodyADF.(map[string]any)
	if !isMap {
		t.Fatalf("BodyADF type = %T, want map[string]any", comments[0].BodyADF)
	}
	if adfDoc["type"] != "doc" {
		t.Errorf("BodyADF[\"type\"] = %v, want \"doc\"", adfDoc["type"])
	}
}

func TestIssueFieldsResponse_UnmarshalRejectsNonObjectFields(t *testing.T) {
	t.Parallel()

	var fields issueFieldsResponse
	if err := json.Unmarshal([]byte(`[1, 2]`), &fields); err == nil {
		t.Fatal("expected error for non-object fields payload, got nil")
	}
}

func TestClient_GetCreateMeta_AllowedValueNameFallsBackToValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts ClientOpts
		body string
	}{
		{
			name: "cloud",
			opts: cloudOpts(),
			body: `{"startAt": 0, "total": 1, "fields": [
				{"fieldId": "customfield_77", "name": "Flavor", "schema": {"type": "option"}, "allowedValues": [{"id": "9", "value": "Vanilla"}]}
			]}`,
		},
		{
			name: "server",
			opts: serverOpts(),
			body: `{"projects": [{"issuetypes": [{"fields": {
				"customfield_77": {"name": "Flavor", "schema": {"type": "option"}, "allowedValues": [{"id": "9", "value": "Vanilla"}]}
			}}]}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, _ := newRecordingClient(t, tt.opts, testkit.StubResponse{Status: http.StatusOK, Body: tt.body})

			fields, err := client.GetCreateMeta(t.Context(), "OPS", "10002")
			if err != nil {
				t.Fatalf("GetCreateMeta: %v", err)
			}
			if len(fields) != 1 || len(fields[0].AllowedValues) != 1 {
				t.Fatalf("fields = %+v, want one field with one allowed value", fields)
			}
			testkit.AssertEqual(t, "allowed value id", fields[0].AllowedValues[0].ID, "9")
			testkit.AssertEqual(t, "allowed value name falls back", fields[0].AllowedValues[0].Name, "Vanilla")
		})
	}
}

func TestClient_GetCreateMeta_ServerWithoutMatchReturnsNil(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, serverOpts(), testkit.StubResponse{Status: http.StatusOK, Body: `{"projects": []}`})

	fields, err := client.GetCreateMeta(t.Context(), "OPS", "10002")
	if err != nil {
		t.Fatalf("GetCreateMeta: %v", err)
	}
	if fields != nil {
		t.Errorf("fields = %+v, want nil when no project matches", fields)
	}
}

func TestClient_DecodeFailureSurfacesPathContext(t *testing.T) {
	t.Parallel()

	client, _ := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "{ broken"})

	_, err := client.GetPriorities(t.Context())
	if err == nil || !strings.Contains(err.Error(), "decode response for GET /priority") {
		t.Errorf("error = %v, want decode failure naming the request", err)
	}
}
