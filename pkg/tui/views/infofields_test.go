package views

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestBuildInfoFields_SystemExtra_FixVersions(t *testing.T) {
	issue := &jira.Issue{
		Key: "PROJ-1",
		CustomFields: map[string]any{
			"fixVersions": []any{
				map[string]any{"name": "Version 1"},
				map[string]any{"name": "Version 2"},
			},
		},
	}
	cfg := []config.FieldConfig{{ID: "fixVersions", Name: "Fix Version/s"}}

	got := buildInfoFields(issue, cfg)
	if len(got) != 1 {
		t.Fatalf("expected 1 InfoField, got %d (%+v)", len(got), got)
	}
	if got[0].Name != "Fix Version/s" {
		t.Errorf("Name = %q, want %q", got[0].Name, "Fix Version/s")
	}
	if got[0].Value != "Version 1, Version 2" {
		t.Errorf("Value = %q, want %q", got[0].Value, "Version 1, Version 2")
	}
}

func TestBuildInfoFields_SystemExtra_Resolution(t *testing.T) {
	issue := &jira.Issue{
		Key: "PROJ-1",
		CustomFields: map[string]any{
			"resolution": map[string]any{"name": "Done"},
		},
	}
	cfg := []config.FieldConfig{{ID: "resolution", Name: "Resolution"}}

	got := buildInfoFields(issue, cfg)
	if len(got) != 1 || got[0].Value != "Done" {
		t.Errorf("buildInfoFields = %+v, want one row with Value=Done", got)
	}
}

func TestBuildInfoFields_SystemExtra_AbsentRendersNone(t *testing.T) {
	issue := &jira.Issue{Key: "PROJ-1"}
	cfg := []config.FieldConfig{{ID: "duedate", Name: "Due"}}

	got := buildInfoFields(issue, cfg)
	if len(got) != 1 {
		t.Fatalf("expected 1 InfoField, got %d", len(got))
	}
	if got[0].Value != noneLabelUpper {
		t.Errorf("absent system field: Value = %q, want %q", got[0].Value, noneLabelUpper)
	}
}

func TestBuildInfoFields_Duedate_HasDateType(t *testing.T) {
	issue := &jira.Issue{
		Key:          "PROJ-1",
		CustomFields: map[string]any{"duedate": "2026-05-15"},
	}
	cfg := []config.FieldConfig{{ID: "duedate", Name: "Due"}}

	got := buildInfoFields(issue, cfg)
	if len(got) != 1 || got[0].Type != FieldDate {
		t.Errorf("duedate Type = %v, want FieldDate; got rows %+v", got[0].Type, got)
	}
}
