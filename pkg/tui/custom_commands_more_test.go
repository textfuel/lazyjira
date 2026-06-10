package tui

import (
	"context"
	"errors"
	"testing"
	"text/template"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

var errCommandFailed = errors.New("command failed")

func TestInitCustomCommands_ValidConfig(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.cfg.CustomCommands = []config.CustomCommandConfig{
		{
			Key:      "x",
			Name:     "test-cmd",
			Command:  "echo hello",
			Contexts: []string{"issues"},
		},
	}

	app.initCustomCommands()

	if len(app.customCmds) != 1 {
		t.Errorf("customCmds len = %d, want 1", len(app.customCmds))
	}
}

func TestInitCustomCommands_InvalidConfig(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.cfg.CustomCommands = []config.CustomCommandConfig{
		{
			Key:      "x",
			Name:     "bad-cmd",
			Command:  "{{.Broken",
			Contexts: []string{"issues"},
		},
	}

	app.initCustomCommands()

	if len(app.customCmds) != 0 {
		t.Error("customCmds should be empty after invalid config")
	}
}

func TestScopeNoun_AllBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		scope config.ScopeMask
		want  string
	}{
		{config.ScopeIssue, "issue"},
		{config.ScopeProject, "project"},
		{config.ScopeIssue | config.ScopeComment, "comment"},
		{config.ScopeComment, "selection"},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "noun", scopeNoun(tc.scope), tc.want)
		})
	}
}

func TestLastNonEmptyLine_ReturnsLastLine(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"single line", "hello world", "hello world"},
		{"trailing newline", "line one\nline two\n", "line two"},
		{"empty input", "", ""},
		{"only newlines", "\n\n\n", ""},
		{"multiple lines", "line1\nline2\nline3", "line3"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "last line", lastNonEmptyLine(tc.input), tc.want)
		})
	}
}

func TestCustomCommandBindings_ReturnsMatchingContext(t *testing.T) {
	t.Parallel()
	app := newTestApp()
	tmpl, _ := template.New("t").Parse("echo hello")
	app.customCmds = []config.ResolvedCustomCommand{
		{
			Key:      "y",
			Name:     "my-cmd",
			Scopes:   config.ScopeIssue,
			Contexts: []config.Context{config.CtxIssues},
			Template: tmpl,
		},
		{
			Key:      "z",
			Name:     "other-cmd",
			Scopes:   config.ScopeProject,
			Contexts: []config.Context{config.CtxProjects},
			Template: tmpl,
		},
	}

	bindings := app.customCommandBindings(config.CtxIssues)

	if len(bindings) != 1 {
		t.Errorf("bindings len = %d, want 1", len(bindings))
	}
	testkit.AssertEqual(t, "binding key", bindings[0].Key, "y")
	testkit.AssertEqual(t, "binding description", bindings[0].Description, "my-cmd")
}

func TestHandleCustomCommandFinished_ErrorShowsInStatusPanel(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleCustomCommandFinished(customCommandFinishedMsg{
		err:    errCommandFailed,
		output: "something went wrong\ndetailed error",
	})
}

func TestHandleCustomCommandFinished_SuccessWithOutput(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _ = app.handleCustomCommandFinished(customCommandFinishedMsg{
		output: "operation succeeded",
	})
}

func TestHandleCustomCommandFinished_RefreshFetchesIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: testKey, Summary: "fresh"})
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.issuesList.SetIssues([]jira.Issue{{Key: testKey}})
	app.previewKey = testKey
	app.issueCache[testKey] = &jira.Issue{Key: testKey, Summary: "stale"}

	_, cmd := app.handleCustomCommandFinished(customCommandFinishedMsg{
		refresh: true,
	})

	if cmd == nil {
		t.Error("expected refresh cmd when refresh=true")
	}
}

func TestExecuteCustomCommand_BackgroundCapture(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.cfg.CustomCommands = nil
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	app.ctx = ctx
	tmpl, _ := template.New("t").Parse("echo test-output")
	suspendFalse := false
	rc := config.ResolvedCustomCommand{
		Key:      "x",
		Name:     "bg-cmd",
		Scopes:   config.ScopeIssue,
		Contexts: []config.Context{config.CtxIssues},
		Template: tmpl,
		Suspend:  &suspendFalse,
	}

	cmd := app.executeCustomCommand(rc, issueScopeData{Key: testKey})
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}

	msg := cmd()
	result, ok := msg.(customCommandFinishedMsg)
	if !ok {
		t.Fatalf("expected customCommandFinishedMsg, got %T", msg)
	}
	if result.err != nil {
		t.Errorf("unexpected error: %v", result.err)
	}
}
