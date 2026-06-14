package tui

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

// formatCreateError turns a create-flow failure into a user-facing message,
// surfacing Jira's own errorMessages without the internal request wrapper. It
// falls back to the raw error when the failure is not a Jira API error.
func formatCreateError(err error, projectKey string, subtask bool) string {
	var apiErr *jira.APIError
	if !errors.As(err, &apiErr) || len(apiErr.Messages) == 0 {
		return err.Error()
	}

	noun := "issue"
	if subtask {
		noun = "subtask"
	}

	detail := lowerFirst(strings.TrimRight(strings.Join(apiErr.Messages, "; "), "."))
	return fmt.Sprintf("Cannot create %s in %s: %s", noun, projectKey, detail)
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
