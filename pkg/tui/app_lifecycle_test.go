package tui

import (
	"context"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestNewApp_ReturnsApp(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SetOnRequestFunc = func(_ func(jira.RequestLog)) {}
	fake.DiscoverFieldsFunc = func(_ context.Context) error { return nil }
	cfg := &config.Config{
		Jira: config.JiraConfig{
			Host:  "example.atlassian.net",
			Email: "test@example.com",
		},
	}

	app := NewApp(cfg, fake)

	if app == nil {
		t.Fatal("expected non-nil App")
	}
	testkit.AssertEqual(t, "default side", app.side, sideLeft)
	testkit.AssertEqual(t, "default leftFocus", app.leftFocus, focusIssues)
}

func TestNewAppWithAuth_SetsAuthMethod(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SetOnRequestFunc = func(_ func(jira.RequestLog)) {}
	fake.DiscoverFieldsFunc = func(_ context.Context) error { return nil }
	cfg := &config.Config{
		Jira: config.JiraConfig{
			Host:  "example.atlassian.net",
			Email: "test@example.com",
		},
	}

	app := NewAppWithAuth(cfg, fake, AuthWizard)

	if app == nil {
		t.Fatal("expected non-nil App")
	}
	if app.splashInfo.AuthMethod != string(AuthWizard) {
		t.Errorf("authMethod = %q, want %q", app.splashInfo.AuthMethod, string(AuthWizard))
	}
}

func TestNewADFRenderer_BuiltinByDefault(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{}

	renderer := newADFRenderer(cfg)

	if renderer == nil {
		t.Fatal("expected non-nil renderer")
	}
	if _, ok := renderer.(views.BuiltinRenderer); !ok {
		t.Errorf("expected BuiltinRenderer, got %T", renderer)
	}
}

func TestNewADFRenderer_GlamourWhenConfigured(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Renderer: config.RendererGlamour,
	}

	renderer := newADFRenderer(cfg)

	if renderer == nil {
		t.Fatal("expected non-nil renderer")
	}
	if _, ok := renderer.(views.GlamourRenderer); !ok {
		t.Errorf("expected GlamourRenderer, got %T", renderer)
	}
}

func TestShutdown_DoesNotBlock(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SetOnRequestFunc = func(_ func(jira.RequestLog)) {}
	fake.DiscoverFieldsFunc = func(_ context.Context) error { return nil }
	cfg := &config.Config{
		Jira: config.JiraConfig{
			Host:  "example.atlassian.net",
			Email: "test@example.com",
		},
	}
	app := NewApp(cfg, fake)

	app.Shutdown()
	app.Shutdown()
}

func TestInit_ReturnsBatchCmd(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.SetOnRequestFunc = func(_ func(jira.RequestLog)) {}
	fake.DiscoverFieldsFunc = func(_ context.Context) error { return nil }
	cfg := &config.Config{
		Jira: config.JiraConfig{
			Host:  "example.atlassian.net",
			Email: "test@example.com",
		},
		IssueTabs: []config.IssueTabConfig{
			{Name: "All", JQL: "project = {{.ProjectKey}} ORDER BY updated DESC"},
		},
	}
	app := NewApp(cfg, fake)
	app.projectKey = testProject

	cmd := app.Init()

	if cmd == nil {
		t.Error("Init() should return a non-nil batch cmd")
	}
}

func TestOpenLinkedIssueDetail_NavigatesToLinkedIssue(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: mainKey, Summary: "linked issue"})
	app := newAppWithFake(t, fake)
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()
	app.infoPanel.SetIssue(&jira.Issue{
		Key:     testKey,
		Summary: testSummary,
		IssueLinks: []jira.IssueLink{
			{
				Type:         &jira.IssueLinkType{Name: "relates to"},
				OutwardIssue: &jira.Issue{Key: mainKey},
			},
		},
	})
	app.infoPanel.SetActiveTab(views.InfoTabLinks)

	_, cmd := app.openLinkedIssueDetail()

	if cmd == nil {
		t.Error("expected fetch cmd for linked issue detail")
	}
	testkit.AssertEqual(t, "side after open", app.side, sideRight)
}

func TestOpenLinkedIssueDetail_NoopWithNoSelection(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()

	_, cmd := app.openLinkedIssueDetail()

	if cmd != nil {
		t.Error("expected nil cmd with no link selected")
	}
}
