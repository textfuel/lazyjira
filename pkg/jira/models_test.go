package jira

import (
	"strings"
	"testing"
	"time"
)

func TestJiraTime_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr string
	}{
		{
			name:  "empty string stays zero",
			input: `""`,
			want:  time.Time{},
		},
		{
			name:  "null stays zero",
			input: `null`,
			want:  time.Time{},
		},
		{
			name:  "jira millisecond format with no colon offset",
			input: `"2024-03-05T10:11:12.000+0300"`,
			want:  time.Date(2024, time.March, 5, 10, 11, 12, 0, time.FixedZone("", 3*60*60)),
		},
		{
			name:  "second precision with no colon offset",
			input: `"2024-03-05T10:11:12+0300"`,
			want:  time.Date(2024, time.March, 5, 10, 11, 12, 0, time.FixedZone("", 3*60*60)),
		},
		{
			name:  "rfc3339 utc",
			input: `"2024-03-05T10:11:12Z"`,
			want:  time.Date(2024, time.March, 5, 10, 11, 12, 0, time.UTC),
		},
		{
			name:  "rfc3339 with colon offset",
			input: `"2024-03-05T10:11:12+03:00"`,
			want:  time.Date(2024, time.March, 5, 10, 11, 12, 0, time.FixedZone("", 3*60*60)),
		},
		{
			name:    "unparseable value errors",
			input:   `"yesterday"`,
			wantErr: "cannot parse Jira time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var parsed JiraTime
			err := parsed.UnmarshalJSON([]byte(tt.input))

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON(%s): %v", tt.input, err)
			}
			if !parsed.Equal(tt.want) {
				t.Errorf("parsed time = %v, want %v", parsed.Time, tt.want)
			}
		})
	}
}

func TestIssueType_CanHaveParent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		issueType *IssueType
		want      bool
	}{
		{
			name:      "nil issue type",
			issueType: nil,
			want:      false,
		},
		{
			name:      "subtask always allowed",
			issueType: &IssueType{Subtask: true, HierarchyLevel: 5},
			want:      true,
		},
		{
			name:      "standard level zero allowed",
			issueType: &IssueType{HierarchyLevel: 0},
			want:      true,
		},
		{
			name:      "negative level allowed",
			issueType: &IssueType{HierarchyLevel: -1},
			want:      true,
		},
		{
			name:      "epic level one rejected",
			issueType: &IssueType{HierarchyLevel: 1},
			want:      false,
		},
		{
			name:      "above epic rejected",
			issueType: &IssueType{HierarchyLevel: 2},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.issueType.CanHaveParent(); got != tt.want {
				t.Errorf("CanHaveParent() = %v, want %v", got, tt.want)
			}
		})
	}
}
