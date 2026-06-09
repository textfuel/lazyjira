package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestResolveTabJQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		jql        string
		projectKey string
		email      string
		want       string
	}{
		{
			name:       "project placeholder is quoted",
			jql:        "project = {{.ProjectKey}} ORDER BY updated DESC",
			projectKey: testProject,
			want:       `project = "PLAT" ORDER BY updated DESC`,
		},
		{
			name:       "user email placeholder is raw",
			jql:        "assignee = {{.UserEmail}}",
			projectKey: testProject,
			email:      "user@example.com",
			want:       "assignee = user@example.com",
		},
		{
			name:       "invalid template falls back to default",
			jql:        "project = {{.Bad",
			projectKey: testProject,
			want:       `project = "PLAT" ORDER BY updated DESC`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolveTabJQL(config.IssueTabConfig{JQL: tt.jql}, tt.projectKey, tt.email)
			testkit.AssertEqual(t, "jql", got, tt.want)
		})
	}
}
