package jira

import (
	"net/http"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestClient_GetComments_ParsesAuthorAndBody(t *testing.T) {
	t.Parallel()

	const commentsJSON = `{"comments": [
		{
			"id": "100",
			"author": {"accountId": "u1", "displayName": "Ann"},
			"body": "Looks good to me",
			"created": "2024-01-02T03:04:05.000+0000",
			"updated": "2024-01-02T03:04:05.000+0000"
		}
	]}`

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: commentsJSON})

	comments, err := client.GetComments(t.Context(), "PLAT-1")
	if err != nil {
		t.Fatalf("GetComments: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/comment")
	if len(comments) != 1 {
		t.Fatalf("len(comments) = %d, want 1", len(comments))
	}
	testkit.AssertEqual(t, "comments[0].ID", comments[0].ID, "100")
	testkit.AssertEqual(t, "comments[0].Body", comments[0].Body, "Looks good to me")
	if comments[0].Author == nil {
		t.Fatal("comments[0].Author is nil")
	}
	testkit.AssertEqual(t, "comments[0].Author.DisplayName", comments[0].Author.DisplayName, "Ann")
}

func TestClient_AddComment_PostsBodyAndParses(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{
		Status: http.StatusCreated,
		Body:   `{"id": "201", "body": "Shipped"}`,
	})

	comment, err := client.AddComment(t.Context(), "PLAT-1", "Shipped")
	if err != nil {
		t.Fatalf("AddComment: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodPost)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/comment")

	body := decodeBody(t, recorded.Body)
	if body["body"] != "Shipped" {
		t.Errorf("request body.body = %v, want Shipped", body["body"])
	}
	testkit.AssertEqual(t, "comment.ID", comment.ID, "201")
}

func TestClient_UpdateComment_PutsToCommentPath(t *testing.T) {
	t.Parallel()

	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: "{}"})

	if err := client.UpdateComment(t.Context(), "PLAT-1", "201", "Edited"); err != nil {
		t.Fatalf("UpdateComment: %v", err)
	}

	testkit.AssertEqual(t, "method", recorded.Method, http.MethodPut)
	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/comment/201")

	body := decodeBody(t, recorded.Body)
	if body["body"] != "Edited" {
		t.Errorf("request body.body = %v, want Edited", body["body"])
	}
}

const changelogEntryJSON = `{
	"author": {"displayName": "Bob"},
	"created": "2024-01-02T03:04:05.000+0000",
	"items": [{"field": "status", "fromString": "To Do", "toString": "Done"}]
}`

func TestClient_GetChangelog_CloudParsesValues(t *testing.T) {
	t.Parallel()

	body := `{"values": [` + changelogEntryJSON + `], "total": 1}`
	client, recorded := newRecordingClient(t, cloudOpts(), testkit.StubResponse{Status: http.StatusOK, Body: body})

	entries, err := client.GetChangelog(t.Context(), "PLAT-1")
	if err != nil {
		t.Fatalf("GetChangelog: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/3/issue/PLAT-1/changelog")
	if len(entries) != 1 || len(entries[0].Items) != 1 {
		t.Fatalf("entries = %#v, want one entry with one item", entries)
	}
	testkit.AssertEqual(t, "item.Field", entries[0].Items[0].Field, "status")
	testkit.AssertEqual(t, "item.FromString", entries[0].Items[0].FromString, "To Do")
	testkit.AssertEqual(t, "item.ToString", entries[0].Items[0].ToString, "Done")
}

func TestClient_GetChangelog_ServerUsesExpand(t *testing.T) {
	t.Parallel()

	body := `{"changelog": {"histories": [` + changelogEntryJSON + `]}}`
	client, recorded := newRecordingClient(t, serverOpts(), testkit.StubResponse{Status: http.StatusOK, Body: body})

	entries, err := client.GetChangelog(t.Context(), "OPS-7")
	if err != nil {
		t.Fatalf("GetChangelog: %v", err)
	}

	testkit.AssertEqual(t, "path", recorded.Path, "/rest/api/2/issue/OPS-7")
	testkit.AssertEqual(t, "expand query", recorded.Query.Get("expand"), "changelog")
	testkit.AssertEqual(t, "fields query", recorded.Query.Get("fields"), "none")
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
	testkit.AssertEqual(t, "entry.Author.DisplayName", entries[0].Author.DisplayName, "Bob")
}
