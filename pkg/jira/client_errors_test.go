package jira

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestClient_AllMethods_PropagateHTTPErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    ClientOpts
		invoke  func(ctx context.Context, client *Client) error
		wantErr string
	}{
		{
			name: "SearchIssues",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.SearchIssues(ctx, "project = PLAT", 0, 10)
				return err
			},
			wantErr: "search issues",
		},
		{
			name: "GetChildren cloud",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetChildren(ctx, errorTestIssueKey)
				return err
			},
			wantErr: "get children of " + errorTestIssueKey,
		},
		{
			name: "GetMyIssues",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetMyIssues(ctx)
				return err
			},
			wantErr: "get my issues",
		},
		{
			name: "GetTransitions",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetTransitions(ctx, errorTestIssueKey)
				return err
			},
			wantErr: "get transitions for " + errorTestIssueKey,
		},
		{
			name: "DoTransition",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.DoTransition(ctx, errorTestIssueKey, "31")
			},
			wantErr: "do transition 31 on " + errorTestIssueKey,
		},
		{
			name: "AddComment",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.AddComment(ctx, errorTestIssueKey, "hello")
				return err
			},
			wantErr: "add comment to " + errorTestIssueKey,
		},
		{
			name: "UpdateComment",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.UpdateComment(ctx, errorTestIssueKey, "5", "hello")
			},
			wantErr: "update comment 5 on " + errorTestIssueKey,
		},
		{
			name: "AssignIssue",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.AssignIssue(ctx, errorTestIssueKey, "acc-1")
			},
			wantErr: "assign issue " + errorTestIssueKey,
		},
		{
			name: "GetProjects cloud",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetProjects(ctx)
				return err
			},
			wantErr: "get projects",
		},
		{
			name: "GetProjects server",
			opts: serverOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetProjects(ctx)
				return err
			},
			wantErr: "get projects",
		},
		{
			name: "GetBoards",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetBoards(ctx)
				return err
			},
			wantErr: "get boards",
		},
		{
			name: "GetBoardIssues",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetBoardIssues(ctx, 4, "")
				return err
			},
			wantErr: "get board 4 issues",
		},
		{
			name: "UpdateIssue",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.UpdateIssue(ctx, errorTestIssueKey, map[string]any{"summary": "x"})
			},
			wantErr: "update issue " + errorTestIssueKey,
		},
		{
			name: "RemoveIssueParent",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.RemoveIssueParent(ctx, errorTestIssueKey)
			},
			wantErr: "remove parent of " + errorTestIssueKey,
		},
		{
			name: "GetPriorities",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetPriorities(ctx)
				return err
			},
			wantErr: "get priorities",
		},
		{
			name: "CreateIssue",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.CreateIssue(ctx, map[string]any{"summary": "x"})
				return err
			},
			wantErr: "create issue",
		},
		{
			name: "GetCreateMeta cloud",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetCreateMeta(ctx, errorTestProjectKey, "10001")
				return err
			},
			wantErr: "get create meta",
		},
		{
			name: "GetCreateMeta server",
			opts: serverOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetCreateMeta(ctx, errorTestProjectKey, "10001")
				return err
			},
			wantErr: "get create meta",
		},
		{
			name: "GetComments",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetComments(ctx, errorTestIssueKey)
				return err
			},
			wantErr: "get comments for " + errorTestIssueKey,
		},
		{
			name: "GetChangelog cloud",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetChangelog(ctx, errorTestIssueKey)
				return err
			},
			wantErr: "get changelog for " + errorTestIssueKey,
		},
		{
			name: "GetChangelog server",
			opts: serverOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetChangelog(ctx, errorTestIssueKey)
				return err
			},
			wantErr: "get changelog for " + errorTestIssueKey,
		},
		{
			name: "GetMyself",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetMyself(ctx)
				return err
			},
			wantErr: "get myself",
		},
		{
			name: "GetUsers",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetUsers(ctx, errorTestProjectKey)
				return err
			},
			wantErr: "get users for project " + errorTestProjectKey,
		},
		{
			name: "GetSprints",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetSprints(ctx, 4)
				return err
			},
			wantErr: "get sprints for board 4",
		},
		{
			name: "MoveToSprint",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.MoveToSprint(ctx, 7, errorTestIssueKey)
			},
			wantErr: "move " + errorTestIssueKey + " to sprint 7",
		},
		{
			name: "GetLabels",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetLabels(ctx)
				return err
			},
			wantErr: "get labels",
		},
		{
			name: "GetComponents",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetComponents(ctx, errorTestProjectKey)
				return err
			},
			wantErr: "get components for project " + errorTestProjectKey,
		},
		{
			name: "GetIssueTypes cloud",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetIssueTypes(ctx, "10000")
				return err
			},
			wantErr: "get issue types for project 10000",
		},
		{
			name: "GetIssueTypes server",
			opts: serverOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetIssueTypes(ctx, "10000")
				return err
			},
			wantErr: "get issue types",
		},
		{
			name: "DiscoverFields",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				return client.DiscoverFields(ctx)
			},
			wantErr: "discover fields",
		},
		{
			name: "GetJQLAutocompleteData",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetJQLAutocompleteData(ctx)
				return err
			},
			wantErr: "/jql/autocompletedata",
		},
		{
			name: "GetJQLAutocompleteSuggestions",
			opts: cloudOpts(),
			invoke: func(ctx context.Context, client *Client) error {
				_, err := client.GetJQLAutocompleteSuggestions(ctx, "status", "In")
				return err
			},
			wantErr: "/jql/autocompletedata/suggestions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, _ := newRecordingClient(t, tt.opts, testkit.StubResponse{Status: http.StatusInternalServerError, Body: "boom"})

			err := tt.invoke(t.Context(), client)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
			if !strings.Contains(err.Error(), "status 500") {
				t.Errorf("error %q does not report status 500", err.Error())
			}
		})
	}
}
