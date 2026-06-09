package git

import (
	"strings"
	"testing"
)

func TestGenerateBranchName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data BranchTemplateData
		tmpl string
		want string
	}{
		{
			name: "default template",
			data: BranchTemplateData{
				Key:     "PROJ-123",
				Summary: "fix-login",
			},
			want: "PROJ-123-fix-login",
		},
		{
			name: "with parent key",
			data: BranchTemplateData{
				Key:       "PROJ-142",
				ParentKey: "PROJ-100",
				Summary:   "fix-login-validation",
			},
			tmpl: "{{.ParentKey}}/{{.Key}}_{{.Summary}}",
			want: "PROJ-100/PROJ-142_fix-login-validation",
		},
		{
			name: "empty parent key strips leading slash",
			data: BranchTemplateData{
				Key:     "PROJ-142",
				Summary: "fix-login",
			},
			tmpl: "{{.ParentKey}}/{{.Key}}_{{.Summary}}",
			want: "PROJ-142_fix-login",
		},
		{
			name: "all fields",
			data: BranchTemplateData{
				Key:        "PROJ-42",
				ProjectKey: "PROJ",
				Number:     "42",
				Summary:    "add-feature",
				Type:       "Story",
				ParentKey:  "PROJ-10",
			},
			tmpl: "{{.Type}}/{{.ParentKey}}/{{.Key}}-{{.Summary}}",
			want: "Story/PROJ-10/PROJ-42-add-feature",
		},
		{
			// GenerateBranchName does not transliterate by itself; ASCII
			// reduction is a per-field caller responsibility. Non-ASCII
			// in raw fields survives.
			name: "non-ASCII in raw fields survives",
			data: BranchTemplateData{
				Key:     "PROJ-1",
				Type:    "L\u00f6sung",
				Summary: "fix-bug",
			},
			tmpl: "{{.Type}}/{{.Key}}-{{.Summary}}",
			want: "L\u00f6sung/PROJ-1-fix-bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GenerateBranchName(tt.data, tt.tmpl)
			if got != tt.want {
				t.Errorf("GenerateBranchName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"spaces to hyphens", "hello world", "hello-world"},
		{"multiple hyphens", "a---b", "a-b"},
		{"trailing dot", "branch.", "branch"},
		{"trailing slash", "branch/", "branch"},
		{"max length truncation", "a-" + strings.Repeat("b", 100), "a-" + strings.Repeat("b", 58)},
		{"slash preserved", "parent/child", "parent/child"},
		{"leading slash stripped", "/child", "child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Sanitize(tt.input)
			if got != tt.want {
				t.Errorf("Sanitize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanitizeSummary(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		input     string
		asciiOnly bool
		want      string
	}{
		{"basic", "Fix Login Bug", false, "fix-login-bug"},
		{"special chars", "Add feature (v2) & test!", false, "add-feature-v2-test"},
		{"ascii only", "Umlaut aeoeue", true, "umlaut-aeoeue"},
		{"umlauts transliterated", "Größe der Straße", true, "groesse-der-strasse"},
		{"accents stripped", "café piñata", true, "cafe-pinata"},
		{"umlauts with droppable special chars", "Fehler in Größe (v2)", true, "fehler-in-groesse-v2"},
		{"umlauts kept when not asciiOnly", "Größe", false, "größe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizeSummary(tt.input, tt.asciiOnly)
			if got != tt.want {
				t.Errorf("SanitizeSummary(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
