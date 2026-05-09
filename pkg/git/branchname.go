package git

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"github.com/textfuel/lazyjira/v2/pkg/ascii"
)

// BranchTemplateData holds data available in branch name templates
type BranchTemplateData struct {
	Key        string
	ProjectKey string
	Number     string
	Summary    string
	Type       string
	ParentKey  string
}

const (
	defaultTemplate = "{{.Key}}-{{.Summary}}"
	maxBranchLen    = 60
)

var (
	issueKeyRe   = regexp.MustCompile(`(?i)([A-Z][A-Z0-9]+-\d+)`)
	skipBranches = map[string]bool{"main": true, "master": true, "develop": true, "dev": true}
	invalidChars = regexp.MustCompile(`[~^:?*\[\]\\` + "\x00-\x1f\x7f" + `]`)
	multiHyphens = regexp.MustCompile(`-{2,}`)
	multiDots    = regexp.MustCompile(`\.{2,}`)
	atBrace      = regexp.MustCompile(`@\{`)
)

// GenerateBranchName creates a branch name from issue data using the given template.
// Pass an empty string for tmplStr to use the default template. ASCII
// reduction is the caller's responsibility (per-field, before passing
// data in); see pkg/tui/handlers_keys.go.
func GenerateBranchName(data BranchTemplateData, tmplStr string) string {
	if tmplStr == "" {
		tmplStr = defaultTemplate
	}

	tmpl, err := template.New("branch").Parse(tmplStr)
	if err != nil {
		tmpl, _ = template.New("branch").Parse(defaultTemplate)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		tmpl, _ = template.New("branch").Parse(defaultTemplate)
		buf.Reset()
		_ = tmpl.Execute(&buf, data)
	}

	return Sanitize(buf.String())
}

// Sanitize cleans a branch name to be a valid git ref.
func Sanitize(name string) string {
	name = strings.TrimLeft(name, "/")
	name = strings.ReplaceAll(name, " ", "-")
	name = invalidChars.ReplaceAllString(name, "")
	name = strings.ReplaceAll(name, "..", ".")
	name = atBrace.ReplaceAllString(name, "")

	name = multiHyphens.ReplaceAllString(name, "-")
	name = multiDots.ReplaceAllString(name, ".")
	name = strings.TrimRight(name, "./")
	name = strings.TrimSuffix(name, ".lock")

	if len(name) > maxBranchLen {
		name = name[:maxBranchLen]
		name = strings.TrimRight(name, "-./")
	}

	return name
}

// SanitizeSummary converts an issue summary to a branch-name-friendly slug
func SanitizeSummary(summary string, asciiOnly bool) string {
	var s string
	if asciiOnly {
		s = ascii.Convert(summary)
	} else {
		s = strings.ToLower(summary)
	}
	s = strings.ReplaceAll(s, " ", "-")
	s = invalidChars.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		if r == '/' || r == '(' || r == ')' || r == '\'' || r == '"' || r == ',' || r == ';' || r == '!' || r == '&' || r == '=' || r == '+' || r == '#' || r == '{' || r == '}' || r == '@' {
			return -1
		}
		return r
	}, s)

	s = multiHyphens.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-.")

	return s
}

// ExtractIssueKey extracts the first Jira issue key from a branch name.
// Returns an empty string for main/master/develop/dev or if no key is found.
func ExtractIssueKey(branchName string) string {
	if skipBranches[strings.ToLower(branchName)] {
		return ""
	}

	match := issueKeyRe.FindStringSubmatch(branchName)
	if match == nil {
		return ""
	}
	return strings.ToUpper(match[1])
}
