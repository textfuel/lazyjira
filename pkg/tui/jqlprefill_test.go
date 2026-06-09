package tui

import (
	"maps"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func TestParseJQLPrefill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		jql  string
		want map[string]string
	}{
		{"empty", "", map[string]string{}},
		{"single equality", "project = PLAT", map[string]string{"project": "PLAT"}},
		{
			name: "currentUser marker",
			jql:  "project = PLAT AND assignee = currentUser()",
			want: map[string]string{"project": "PLAT", "assignee": currentUserMarker},
		},
		{"order by stripped", "project = PLAT ORDER BY updated DESC", map[string]string{"project": "PLAT"}},
		{"quoted value unquoted", `project = "My Proj"`, map[string]string{"project": "My Proj"}},
		{"or clause skipped entirely", "project = PLAT OR status = Done", map[string]string{}},
		{"other function skipped", "assignee = membersOf(grp)", map[string]string{}},
		{
			name: "valid clause survives alongside a skipped function",
			jql:  "project = PLAT AND assignee = membersOf(grp)",
			want: map[string]string{"project": "PLAT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseJQLPrefill(tt.jql)
			if !maps.Equal(got, tt.want) {
				t.Errorf("ParseJQLPrefill(%q) = %v, want %v", tt.jql, got, tt.want)
			}
		})
	}
}

func TestApplyPrefill_TextAndCurrentUser(t *testing.T) {
	t.Parallel()

	fields := []components.CreateFormField{
		{FieldID: "summary"},
		{FieldID: "assignee"},
	}
	prefill := map[string]string{
		"summary":  "Fix the bug",
		"assignee": currentUserMarker,
	}
	currentUser := &jira.User{AccountID: "acc-1", DisplayName: "Ada"}

	ApplyPrefill(fields, prefill, currentUser, true)

	testkit.AssertEqual(t, "summary display", fields[0].DisplayValue, "Fix the bug")
	testkit.AssertEqual(t, "summary value", fields[0].Value, any("Fix the bug"))
	testkit.AssertEqual(t, "assignee display", fields[1].DisplayValue, "Ada")

	assigneeValue, ok := fields[1].Value.(map[string]string)
	if !ok {
		t.Fatalf("assignee value = %#v, want map", fields[1].Value)
	}
	testkit.AssertEqual(t, "assignee accountId", assigneeValue[fldAccountID], "acc-1")
}
