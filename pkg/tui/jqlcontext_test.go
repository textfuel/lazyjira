package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestParseJQLContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		cursor int
		want   JQLContext
	}{
		{"empty input is none", "", 0, JQLContext{Mode: jqlCtxNone}},
		{"typing a field", "pro", 3, JQLContext{Mode: jqlCtxField, Partial: "pro", PartialLen: 3}},
		{"after operator wants value", "project = ", 10, JQLContext{Mode: jqlCtxValue, FieldName: "project"}},
		{"typing a value", "project = Foo", 13, JQLContext{Mode: jqlCtxValue, FieldName: "project", Partial: "Foo", PartialLen: 3}},
		{"inside IN list wants value", "status in (", 11, JQLContext{Mode: jqlCtxValue, FieldName: "status"}},
		{"after AND wants field", "project = x AND ", 16, JQLContext{Mode: jqlCtxField}},
		{"field then space is none", "project ", 8, JQLContext{Mode: jqlCtxNone}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "context", parseJQLContext(tt.input, tt.cursor), tt.want)
		})
	}
}

func TestTokenizeJQL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"fields operator and quoted value", `project = "Foo Bar"`, []string{"project", "=", `"Foo Bar"`}},
		{"parens and commas are separate tokens", "a in(b,c)", []string{"a", "in", "(", "b", ",", "c", ")"}},
		{"empty is no tokens", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertSliceEqual(t, "tokens", tokenizeJQL(tt.input), tt.want)
		})
	}
}

func TestMatchFieldSuggestions(t *testing.T) {
	t.Parallel()

	fields := []jira.AutocompleteField{
		{Value: "status"},
		{Value: "statusCategory"},
		{Value: "summary"},
	}

	tests := []struct {
		name    string
		partial string
		want    []string
	}{
		{"empty returns all in order", "", []string{"status", "statusCategory", "summary"}},
		{"prefix match ranks together", "stat", []string{"status", "statusCategory"}},
		{"exact match wins", "summary", []string{"summary"}},
		{"no match is empty", "xyz", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertSliceEqual(t, "suggestions", matchFieldSuggestions(fields, tt.partial), tt.want)
		})
	}
}
