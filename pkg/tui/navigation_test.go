package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func TestExtractIssueKey(t *testing.T) {
	t.Parallel()
	app := newTestApp()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"plain browse url", "example.atlassian.net/browse/DR-13819", "DR-13819"},
		{"strips query params", "example.atlassian.net/browse/DR-13819?focusedId=1", "DR-13819"},
		{"strips fragment", "example.atlassian.net/browse/DR-13819#comment", "DR-13819"},
		{"foreign host returns empty", "other.example.com/browse/DR-1", ""},
		{"non browse url returns empty", "example.atlassian.net/projects/DR", ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "extracted key", app.extractIssueKey(testCase.url), testCase.want)
		})
	}
}

func TestPlatformCommand(t *testing.T) {
	t.Parallel()

	t.Run("open returns a launcher for the host os", func(t *testing.T) {
		t.Parallel()
		name, _ := platformCommand("open", "https://example.com")
		if name == "" {
			t.Error("open action should resolve to a command")
		}
	})

	t.Run("copy returns a clipboard command", func(t *testing.T) {
		t.Parallel()
		name, _ := platformCommand("copy", "")
		if name == "" {
			t.Error("copy action should resolve to a command")
		}
	})

	t.Run("unknown action returns empty", func(t *testing.T) {
		t.Parallel()
		name, args := platformCommand("teleport", "")
		if name != "" || args != nil {
			t.Errorf("unknown action = (%q, %v), want empty", name, args)
		}
	})
}

func TestNavigateToIssue_SelectsInList(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 80
	app.height = 24
	app.side = sideRight
	app.leftFocus = focusProjects
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	app.issueCache[testKey] = &jira.Issue{Key: testKey, Summary: "cached"}

	app.navigateToIssue(testKey)

	if app.side != sideLeft || app.leftFocus != focusIssues {
		t.Errorf("focus = (%v,%v), want left/issues", app.side, app.leftFocus)
	}
	if got := app.detailView.IssueKey(); got != testKey {
		t.Errorf("detailView issue = %q, want %s", got, testKey)
	}
}
