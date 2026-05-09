package views

import (
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func findField(fields []InfoField, id string) (InfoField, bool) {
	for _, f := range fields {
		if f.FieldID == id {
			return f, true
		}
	}
	return InfoField{}, false
}

func TestParentField_Present(t *testing.T) {
	issue := &jira.Issue{
		Key: "PROJ-2",
		Parent: &jira.Issue{
			Key:     "PROJ-1",
			Summary: "epic summary",
		},
	}

	fields := buildInfoFields(issue, nil)
	f, ok := findField(fields, "parent")
	if !ok {
		t.Fatalf("expected 'parent' field in default info fields, got: %+v", fields)
	}
	if f.Name != "Parent" {
		t.Errorf("Name = %q, want %q", f.Name, "Parent")
	}
	if f.Value != "[PROJ-1] epic summary" {
		t.Errorf("Value = %q, want %q", f.Value, "[PROJ-1] epic summary")
	}
	if f.Type != FieldSingleText {
		t.Errorf("Type = %v, want FieldSingleText", f.Type)
	}
}

func TestParentField_AbsentWhenNil(t *testing.T) {
	issue := &jira.Issue{Key: "PROJ-2"}

	fields := buildInfoFields(issue, nil)
	if _, ok := findField(fields, "parent"); ok {
		t.Errorf("expected no 'parent' field when Issue.Parent and IssueType are nil, got one")
	}
}

func TestParentField_NoneForSubtaskWithoutParent(t *testing.T) {
	issue := &jira.Issue{
		Key:       "PROJ-2",
		IssueType: &jira.IssueType{Name: "Sub-task", Subtask: true},
	}
	fields := buildInfoFields(issue, nil)
	f, ok := findField(fields, "parent")
	if !ok {
		t.Fatalf("expected 'parent' field for subtask without parent, got: %+v", fields)
	}
	if f.Value != noneLabelUpper {
		t.Errorf("Value = %q, want %q", f.Value, noneLabelUpper)
	}
}

func TestParentField_NoneForStandardWithoutParent(t *testing.T) {
	issue := &jira.Issue{
		Key:       "PROJ-2",
		IssueType: &jira.IssueType{Name: "Story", HierarchyLevel: 0},
	}
	fields := buildInfoFields(issue, nil)
	f, ok := findField(fields, "parent")
	if !ok {
		t.Fatalf("expected 'parent' field for standard issue without parent, got: %+v", fields)
	}
	if f.Value != noneLabelUpper {
		t.Errorf("Value = %q, want %q", f.Value, noneLabelUpper)
	}
}

func TestParentField_HiddenForEpicWithoutParent(t *testing.T) {
	issue := &jira.Issue{
		Key:       "PROJ-2",
		IssueType: &jira.IssueType{Name: "Epic", HierarchyLevel: 1},
	}
	fields := buildInfoFields(issue, nil)
	if _, ok := findField(fields, "parent"); ok {
		t.Errorf("expected no 'parent' field for epic (level 1) without parent, got one")
	}
}

func TestParentField_ReadOnly(t *testing.T) {
	issue := &jira.Issue{
		Key:    "PROJ-2",
		Parent: &jira.Issue{Key: "PROJ-1", Summary: "epic"},
	}
	// SetBuiltinFieldValue must be a no-op for "parent" because no setValue is wired.
	if SetBuiltinFieldValue(issue, "parent", &jira.Issue{Key: "OTHER-1"}) {
		t.Errorf("SetBuiltinFieldValue('parent', ...) returned true; expected false (read-only)")
	}
	if issue.Parent.Key != "PROJ-1" {
		t.Errorf("Parent was mutated: Key=%q", issue.Parent.Key)
	}
}

func TestParentField_LongSummaryTruncated(t *testing.T) {
	longSummary := strings.Repeat("x", 200)
	issue := &jira.Issue{
		Key:    "PROJ-2",
		Parent: &jira.Issue{Key: "PROJ-1", Summary: longSummary},
	}

	// Render with a narrow width so the value cannot fit.
	styled, plain := renderInfoRowPairs(issue, nil, nil, 30)
	_ = styled

	var parentRow string
	for _, row := range plain {
		if strings.Contains(row, "Parent:") {
			parentRow = row
			break
		}
	}
	if parentRow == "" {
		t.Fatalf("expected Parent row in rendered output, got: %v", plain)
	}
	if !strings.HasSuffix(parentRow, "…") {
		t.Errorf("expected truncated Parent row to end with '…', got %q", parentRow)
	}
	if strings.Contains(parentRow, longSummary) {
		t.Errorf("expected long summary to be truncated, but full summary appears in row: %q", parentRow)
	}
}
