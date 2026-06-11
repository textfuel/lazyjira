package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func marshalPage(t *testing.T, page any) string {
	t.Helper()
	data, err := json.Marshal(page)
	if err != nil {
		t.Fatalf("marshal page: %v", err)
	}
	return string(data)
}

func numberedObjects(count int, build func(index int) map[string]any) []map[string]any {
	objects := make([]map[string]any, count)
	for index := range count {
		objects[index] = build(index)
	}
	return objects
}

func TestClient_PaginatedMethods_FetchAllPages(t *testing.T) {
	t.Parallel()

	boardObject := func(index int) map[string]any {
		return map[string]any{"id": index + 1, "name": fmt.Sprintf("Board %d", index+1)}
	}
	issueObject := func(index int) map[string]any {
		return map[string]any{"id": strconv.Itoa(index + 1), "key": fmt.Sprintf("PAGE-%d", index+1), "fields": map[string]any{"summary": "s"}}
	}
	changelogObject := func(index int) map[string]any {
		return map[string]any{"items": []map[string]any{{"field": "status", "fromString": fmt.Sprintf("State%d", index), "toString": fmt.Sprintf("State%d", index+1)}}}
	}
	userObject := func(index int) map[string]any {
		return map[string]any{"accountId": fmt.Sprintf("acc-%d", index+1), "displayName": fmt.Sprintf("User %d", index+1)}
	}
	sprintObject := func(index int) map[string]any {
		return map[string]any{"id": index + 1, "name": fmt.Sprintf("Sprint %d", index+1)}
	}
	createMetaFieldObject := func(index int) map[string]any {
		return map[string]any{"fieldId": fmt.Sprintf("customfield_%d", index+1), "name": fmt.Sprintf("Field %d", index+1)}
	}

	tests := []struct {
		name              string
		buildPages        func(t *testing.T) []testkit.StubResponse
		invoke            func(ctx context.Context, client *Client) (int, error)
		wantCount         int
		wantSecondStartAt string
	}{
		{
			name: "GetBoards stops on isLast",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": numberedObjects(100, boardObject), "isLast": false})},
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": numberedObjects(1, boardObject), "isLast": true})},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				boards, err := client.GetBoards(ctx)
				return len(boards), err
			},
			wantCount:         101,
			wantSecondStartAt: "100",
		},
		{
			name: "GetBoardIssues stops on short page",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"issues": numberedObjects(50, issueObject)})},
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"issues": numberedObjects(1, issueObject)})},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				issues, err := client.GetBoardIssues(ctx, 4, "")
				return len(issues), err
			},
			wantCount:         51,
			wantSecondStartAt: "50",
		},
		{
			name: "GetChangelog cloud stops at total",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": numberedObjects(100, changelogObject), "total": 101})},
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": numberedObjects(1, changelogObject), "total": 101})},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				entries, err := client.GetChangelog(ctx, "PAGE-1")
				return len(entries), err
			},
			wantCount:         101,
			wantSecondStartAt: "100",
		},
		{
			name: "GetUsers stops on short page",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, numberedObjects(100, userObject))},
					{Status: http.StatusOK, Body: marshalPage(t, numberedObjects(1, userObject))},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				users, err := client.GetUsers(ctx, errorTestProjectKey)
				return len(users), err
			},
			wantCount:         101,
			wantSecondStartAt: "100",
		},
		{
			name: "GetSprints stops on isLast",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": numberedObjects(50, sprintObject), "isLast": false})},
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": numberedObjects(1, sprintObject), "isLast": true})},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				sprints, err := client.GetSprints(ctx, 4)
				return len(sprints), err
			},
			wantCount:         51,
			wantSecondStartAt: "50",
		},
		{
			name: "GetLabels stops at total",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				labels := make([]string, 1000)
				for index := range labels {
					labels[index] = fmt.Sprintf("label-%d", index+1)
				}
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": labels, "total": 1001})},
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"values": []string{"last"}, "total": 1001})},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				labels, err := client.GetLabels(ctx)
				return len(labels), err
			},
			wantCount:         1001,
			wantSecondStartAt: "1000",
		},
		{
			name: "GetCreateMeta cloud stops at total",
			buildPages: func(t *testing.T) []testkit.StubResponse {
				t.Helper()
				return []testkit.StubResponse{
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"startAt": 0, "total": 3, "fields": numberedObjects(2, createMetaFieldObject)})},
					{Status: http.StatusOK, Body: marshalPage(t, map[string]any{"startAt": 2, "total": 3, "fields": numberedObjects(1, createMetaFieldObject)})},
				}
			},
			invoke: func(ctx context.Context, client *Client) (int, error) {
				fields, err := client.GetCreateMeta(ctx, errorTestProjectKey, "10001")
				return len(fields), err
			},
			wantCount:         3,
			wantSecondStartAt: "2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, recorded := newSequenceClient(t, cloudOpts(), tt.buildPages(t)...)

			count, err := tt.invoke(t.Context(), client)
			if err != nil {
				t.Fatalf("invoke: %v", err)
			}

			testkit.AssertEqual(t, "item count", count, tt.wantCount)
			if len(*recorded) != 2 {
				t.Fatalf("made %d requests, want 2", len(*recorded))
			}
			testkit.AssertEqual(t, "first startAt", (*recorded)[0].Query.Get("startAt"), "0")
			testkit.AssertEqual(t, "second startAt", (*recorded)[1].Query.Get("startAt"), tt.wantSecondStartAt)
		})
	}
}
