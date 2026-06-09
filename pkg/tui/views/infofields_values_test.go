package views

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestFormatCustomFieldValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil is none", nil, noneLabelUpper},
		{"plain string", "hello", "hello"},
		{"integer float drops decimals", float64(8), "8"},
		{"fractional float keeps two decimals", 3.5, "3.50"},
		{"map with displayName", map[string]any{"displayName": "Ada"}, "Ada"},
		{"map with value", map[string]any{"value": "High"}, "High"},
		{"map with name", map[string]any{"name": "Bug"}, "Bug"},
		{"list joins with comma", []any{"a", "b", "c"}, "a, b, c"},
		{"list of options", []any{map[string]any{"value": "A"}, map[string]any{"value": "B"}}, "A, B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "formatted", formatCustomFieldValue(tt.input), tt.want)
		})
	}
}

func TestResolveCustomFieldType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configType string
		raw        any
		want       InfoFieldType
	}{
		{"select", "select", nil, FieldSingleSelect},
		{"multiselect", "multiselect", nil, FieldMultiSelect},
		{"user", "user", nil, FieldPerson},
		{"textarea", "textarea", nil, FieldMultiText},
		{"text", "text", nil, FieldSingleText},
		{"empty falls back to value detection", "", map[string]any{"displayName": "Ada"}, FieldPerson},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "type", resolveCustomFieldType(tt.configType, tt.raw), tt.want)
		})
	}
}

func TestDetectFieldTypeFromValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  InfoFieldType
	}{
		{"nil is single text", nil, FieldSingleText},
		{"person map", map[string]any{"displayName": "Ada"}, FieldPerson},
		{"option with value", map[string]any{"value": "High"}, FieldSingleSelect},
		{"option with name", map[string]any{"name": "Bug"}, FieldSingleSelect},
		{"list is multiselect", []any{"a", "b"}, FieldMultiSelect},
		{"plain string is single text", "hello", FieldSingleText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "type", detectFieldTypeFromValue(tt.input), tt.want)
		})
	}
}

func TestEditValueForInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"none placeholder becomes empty", noneLabelUpper, ""},
		{"unknown placeholder becomes empty", unknownLabel, ""},
		{"real value passes through", "Story Points", "Story Points"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "edit value", EditValueForInput(tt.input), tt.want)
		})
	}
}

func TestPatchIssueFields_CopiesFromSource(t *testing.T) {
	t.Parallel()

	source := &jira.Issue{
		Summary:      "Patched summary",
		Status:       &jira.Status{Name: "Done"},
		Priority:     &jira.Priority{Name: "High"},
		Labels:       []string{"backend"},
		CustomFields: map[string]any{"customfield_10015": float64(8)},
	}
	target := &jira.Issue{Summary: "old", Key: "PLAT-1"}

	PatchIssueFields(target, source)

	testkit.AssertEqual(t, "Summary", target.Summary, "Patched summary")
	testkit.AssertEqual(t, "Key is untouched", target.Key, "PLAT-1")
	if target.Status == nil || target.Status.Name != "Done" {
		t.Errorf("Status = %#v, want Done", target.Status)
	}
	testkit.AssertSliceEqual(t, "Labels", target.Labels, []string{"backend"})
	if target.CustomFields["customfield_10015"] != float64(8) {
		t.Errorf("CustomFields not patched: %#v", target.CustomFields)
	}
}

func TestSetBuiltinFieldValue(t *testing.T) {
	t.Parallel()

	issue := &jira.Issue{}

	if !SetBuiltinFieldValue(issue, "priority", &jira.Priority{Name: "High"}) {
		t.Fatal("SetBuiltinFieldValue(priority) returned false")
	}
	if issue.Priority == nil || issue.Priority.Name != "High" {
		t.Errorf("Priority = %#v, want High", issue.Priority)
	}

	if SetBuiltinFieldValue(issue, "not-a-field", nil) {
		t.Error("SetBuiltinFieldValue(unknown) returned true")
	}
}
