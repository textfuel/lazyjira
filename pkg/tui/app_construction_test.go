package tui

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func fullConfig() *config.Config {
	return &config.Config{
		Jira:     config.JiraConfig{Host: "example.atlassian.net", Email: "me@example.com"},
		Projects: []config.ProjectConfig{{Key: testProject}},
		GUI: config.GUIConfig{
			IssueListFields: []string{"key", "summary"},
			TypeIcons:       map[string]string{"Story": "S"},
			StatusIcons:     map[string]string{"Done": "D"},
			PriorityIcons:   map[string]string{"High": "H"},
		},
		Fields: []config.FieldConfig{
			{ID: "status"},
			{ID: "customfield_10015", Name: "Story Points"},
		},
		Converter: config.ConverterAdfConverter,
	}
}

func TestNewAppWithAuth_FullConfig(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	var capturedRequestLogCallback func(jira.RequestLog)
	fake.SetOnRequestFunc = func(fn func(jira.RequestLog)) { capturedRequestLogCallback = fn }
	fake.SetCustomFieldsFunc = func([]string) {}

	app := NewAppWithAuth(fullConfig(), fake, AuthSaved)

	if app.projectKey != testProject {
		t.Errorf("projectKey = %q, want %s", app.projectKey, testProject)
	}
	if len(fake.SetCustomFieldsCalls) != 1 || !slices.Equal(fake.SetCustomFieldsCalls[0], []string{"customfield_10015"}) {
		t.Errorf("SetCustomFields calls = %v", fake.SetCustomFieldsCalls)
	}
	if _, ok := app.converter.(AdfConvConverter); !ok {
		t.Errorf("converter = %T, want AdfConvConverter", app.converter)
	}
	if capturedRequestLogCallback == nil {
		t.Fatal("SetOnRequest callback was not registered")
	}
	requestLog := jira.RequestLog{Method: "GET", Path: "/x", Status: 200, Elapsed: time.Millisecond}
	capturedRequestLogCallback(requestLog)
	*app.logFlag = true
	capturedRequestLogCallback(requestLog)
}

func TestNewAppWithAuth_WarnsWhenQuitShadowedEverywhere(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SetOnRequestFunc = func(func(jira.RequestLog)) {}
	cfg := &config.Config{
		Jira: config.JiraConfig{Host: "example.atlassian.net"},
		CustomCommands: []config.CustomCommandConfig{
			{Key: "q", Name: "shadow q", Command: "echo {{.Key}}", Contexts: []string{"issues", "info", "projects", "detail", "detail.comments"}},
			{Key: "ctrl+c", Name: "shadow ctrl c", Command: "echo {{.Key}}", Contexts: []string{"issues", "info", "projects", "detail", "detail.comments"}},
		},
	}

	app := NewAppWithAuth(cfg, fake, AuthEnv)

	if !strings.Contains(app.statusPanel.ErrorMessage(), "unreachable") {
		t.Errorf("status error = %q, want quit-unreachable warning", app.statusPanel.ErrorMessage())
	}
}
