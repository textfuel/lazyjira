package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

const sampleTheme = `
name: sample
palette:
  green: "#a6e3a1"
  red: "#f38ba8"
  yellow: "#f9e2af"
  blue: "#89b4fa"
  gray: "#6c7086"
  bg: "#585b70"
  teal: "#94e2d5"

styles:
  title:          green
  subtitle:       gray
  accent:         green
  muted:          gray
  errorText:      red
  successText:    green
  warningText:    yellow
  selectedItemBg: bg
  issueKey:       teal

types:
  Bug:       red
  Story:     green
  _fallback: gray

priorities:
  High: red
  Low:  green

statuses:
  done:          green
  indeterminate: yellow
  new:           blue

authorPalette:
  - red
  - green
  - blue
`

func TestLoadYAMLThemeRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.yml")
	if err := os.WriteFile(path, []byte(sampleTheme), 0o644); err != nil {
		t.Fatal(err)
	}
	th, err := loadYAMLTheme(path)
	if err != nil {
		t.Fatalf("loadYAMLTheme: %v", err)
	}
	if got := th.TypeStyle("Bug").GetForeground(); got != lipgloss.Color("#f38ba8") {
		t.Errorf("Bug type color = %q, want #f38ba8", got)
	}
	if got := th.TypeStyle("Unknown").GetForeground(); got != lipgloss.Color("#6c7086") {
		t.Errorf("Unknown type fallback = %q, want gray (#6c7086)", got)
	}
	if got := th.PriorityStyle("High").GetForeground(); got != lipgloss.Color("#f38ba8") {
		t.Errorf("Priority High = %q, want red", got)
	}
	if got := th.StatusStyle("done").GetForeground(); got != lipgloss.Color("#a6e3a1") {
		t.Errorf("Status done = %q, want green", got)
	}
	if got := th.IssueKey.GetForeground(); got != lipgloss.Color("#94e2d5") {
		t.Errorf("IssueKey = %q, want teal", got)
	}
	if len(th.AuthorPalette) != 3 {
		t.Errorf("AuthorPalette length = %d, want 3", len(th.AuthorPalette))
	}
}

func TestLoadYAMLThemeHexLiteralValues(t *testing.T) {
	yamlSrc := `
palette:
  one: "#111111"
styles:
  title: "#222222"
  subtitle: one
  accent: "#333333"
  muted: one
  errorText: "#f00"
  successText: "#0f0"
  warningText: "#ff0"
  selectedItemBg: one
  issueKey: "#abcdef"
`
	dir := t.TempDir()
	path := filepath.Join(dir, "hex.yml")
	if err := os.WriteFile(path, []byte(yamlSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	th, err := loadYAMLTheme(path)
	if err != nil {
		t.Fatalf("loadYAMLTheme: %v", err)
	}
	if got := th.IssueKey.GetForeground(); got != lipgloss.Color("#abcdef") {
		t.Errorf("IssueKey = %q, want #abcdef", got)
	}
	if got := th.Subtitle.GetForeground(); got != lipgloss.Color("#111111") {
		t.Errorf("Subtitle = %q, want #111111", got)
	}
}

func TestLoadYAMLThemeMissingRequiredStyle(t *testing.T) {
	yamlSrc := `
palette:
  red: "#f00"
styles:
  title: red
`
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")
	if err := os.WriteFile(path, []byte(yamlSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadYAMLTheme(path); err == nil {
		t.Fatal("expected error for missing required style")
	}
}

func TestLoadYAMLThemeMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "malformed.yml")
	if err := os.WriteFile(path, []byte("palette: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadYAMLTheme(path); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestLoadYAMLThemeOptionalSectionsMissing(t *testing.T) {
	yamlSrc := `
palette:
  red: "#f00"
  green: "#0f0"
  blue: "#00f"
  yellow: "#ff0"
  gray: "#888"
  bg: "#111"
styles:
  title: green
  subtitle: gray
  accent: green
  muted: gray
  errorText: red
  successText: green
  warningText: yellow
  selectedItemBg: bg
  issueKey: blue
`
	dir := t.TempDir()
	path := filepath.Join(dir, "minimal.yml")
	if err := os.WriteFile(path, []byte(yamlSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	th, err := loadYAMLTheme(path)
	if err != nil {
		t.Fatalf("loadYAMLTheme: %v", err)
	}
	// Missing types map — fallback to mutedC (gray).
	if got := th.TypeStyle("Anything").GetForeground(); got != lipgloss.Color("#888") {
		t.Errorf("missing types fallback = %q, want gray", got)
	}
	if got := th.PriorityStyle("High").GetForeground(); got != lipgloss.Color("#888") {
		t.Errorf("missing priorities fallback to muted = %q, want gray", got)
	}
}

func TestSetThemeYAMLFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "myth.yml"), []byte(sampleTheme), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LAZYJIRA_THEMES_DIR", dir)
	if err := SetTheme("myth"); err != nil {
		t.Fatalf("SetTheme(myth): %v", err)
	}
	if Default.IssueKey.GetForeground() != lipgloss.Color("#94e2d5") {
		t.Errorf("after SetTheme(myth), IssueKey = %q, want #94e2d5",
			Default.IssueKey.GetForeground())
	}
	_ = SetTheme("default")
}

func TestSetThemeYAMLMissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LAZYJIRA_THEMES_DIR", dir)
	if err := SetTheme("doesnotexist"); err == nil {
		t.Fatal("expected error for missing theme file")
	}
	_ = SetTheme("default")
}
