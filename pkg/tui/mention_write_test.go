package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

// identityConverter passes markdown straight through so tests can assert on the
// resolved markdown that reaches the write call, independent of ADF shape.
type identityConverter struct{}

func (identityConverter) ToMarkdown(adf any) (string, any, error) {
	s, _ := adf.(string)
	return s, nil, nil
}

func (identityConverter) FromMarkdown(md string, _ any) (any, error) {
	return md, nil
}

func soloUser() jira.User { return jira.User{DisplayName: "Solo One", AccountID: "s1"} }

func TestMentionColdPath_Comment(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{
		GetUsersFunc: func(_ context.Context, _ string) ([]jira.User, error) {
			return []jira.User{soloUser()}, nil
		},
		AddCommentFunc: func(_ context.Context, _ string, _ any) (*jira.Comment, error) {
			return &jira.Comment{}, nil
		},
	}
	app := newAppWithFake(t, fake)
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject
	app.editContext = editCtx{kind: editCommentNew, issueKey: testKey}

	cmd := app.applyEdit("ping @Solo_One")
	if cmd == nil {
		t.Fatal("applyEdit returned nil cmd on cold cache")
	}
	if app.pendingMention == nil {
		t.Fatal("pendingMention not set on cold path")
	}
	if len(fake.AddCommentCalls) != 0 {
		t.Fatal("AddComment must not run before users are loaded")
	}

	msg := cmd()
	loaded, ok := msg.(mentionUsersLoadedMsg)
	if !ok {
		t.Fatalf("expected mentionUsersLoadedMsg, got %T", msg)
	}
	_, follow := app.handleMentionUsersLoaded(loaded)
	if follow == nil {
		t.Fatal("handleMentionUsersLoaded returned nil cmd")
	}
	follow()

	if len(fake.AddCommentCalls) != 1 {
		t.Fatalf("AddComment called %d times, want 1", len(fake.AddCommentCalls))
	}
	body, _ := fake.AddCommentCalls[0].Body.(string)
	if !strings.Contains(body, "accountid:s1") {
		t.Errorf("comment body = %q, want it to contain accountid:s1", body)
	}
	if app.pendingMention != nil {
		t.Error("pendingMention should be cleared after completion")
	}
}

func TestMentionWarmPath_Comment(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{
		T: t,
		AddCommentFunc: func(_ context.Context, _ string, _ any) (*jira.Comment, error) {
			return &jira.Comment{}, nil
		},
	}
	app := newAppWithFake(t, fake)
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject
	app.usersCache[testProject] = []jira.User{soloUser()}
	app.editContext = editCtx{kind: editCommentNew, issueKey: testKey}

	cmd := app.applyEdit("ping @Solo_One")
	if cmd == nil {
		t.Fatal("applyEdit returned nil cmd")
	}
	if app.pendingMention != nil {
		t.Error("warm cache must not defer the write")
	}

	msg := cmd()
	if _, ok := msg.(mentionUsersLoadedMsg); ok {
		t.Fatal("warm path must not emit mentionUsersLoadedMsg")
	}
	if len(fake.AddCommentCalls) != 1 {
		t.Fatalf("AddComment called %d times, want 1", len(fake.AddCommentCalls))
	}
	body, _ := fake.AddCommentCalls[0].Body.(string)
	if !strings.Contains(body, "accountid:s1") {
		t.Errorf("comment body = %q, want it to contain accountid:s1", body)
	}
}

func TestMentionSkippedForPlainTextField(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{
		T: t,
		UpdateIssueFunc: func(_ context.Context, _ string, _ map[string]any) error {
			return nil
		},
	}
	app := newAppWithFake(t, fake)
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject
	app.usersCache[testProject] = []jira.User{soloUser()} // warm: would resolve if attempted
	app.editContext = editCtx{kind: editFieldText, issueKey: testKey, fieldID: "customfield_10010"}

	cmd := app.applyEdit("ping @Solo_One")
	if cmd == nil {
		t.Fatal("applyEdit returned nil cmd")
	}
	if app.pendingMention != nil {
		t.Error("plain-text field must not defer for mentions")
	}
	cmd()

	if len(fake.UpdateIssueCalls) != 1 {
		t.Fatalf("UpdateIssue called %d times, want 1", len(fake.UpdateIssueCalls))
	}
	got, _ := fake.UpdateIssueCalls[0].Fields["customfield_10010"].(string)
	if got != "ping @Solo_One" {
		t.Errorf("field value = %q, want %q", got, "ping @Solo_One")
	}
	if strings.Contains(got, "accountid") {
		t.Errorf("plain-text field must not get a resolved mention: %q", got)
	}
}

func TestMentionColdPath_CreateDesc(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject
	app.createForm = formWithFields([]components.CreateFormField{
		{FieldID: "summary"},
		{FieldID: "description"},
	})
	app.editContext = editCtx{kind: editCreateDesc, fieldIndex: 1}
	path := writeTempFile(t, "see @Solo_One")

	if _, cmd := app.handleEditorFinished(editorFinishedMsg{original: "old", tempPath: path}); cmd == nil {
		t.Fatal("editor finish on cold cache should return a fetch cmd")
	}
	if app.pendingMention == nil || !app.pendingMention.createDesc {
		t.Fatal("create-desc cold path must defer via pendingMention")
	}

	app.handleMentionUsersLoaded(mentionUsersLoadedMsg{users: []jira.User{soloUser()}})

	got, _ := app.createForm.FieldAt(1).Value.(string)
	if !strings.Contains(got, "accountid:s1") {
		t.Errorf("description value = %q, want it to contain accountid:s1", got)
	}
}

func TestMentionWarmPath_CrossProject(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{
		T: t,
		AddCommentFunc: func(_ context.Context, _ string, _ any) (*jira.Comment, error) {
			return &jira.Comment{}, nil
		},
	}
	app := newAppWithFake(t, fake)
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject // board project differs from the edited issue
	app.usersCache["OTHER"] = []jira.User{soloUser()}
	app.editContext = editCtx{kind: editCommentNew, issueKey: "OTHER-7"}

	cmd := app.applyEdit("ping @Solo_One")
	if cmd == nil {
		t.Fatal("applyEdit returned nil cmd")
	}
	if app.pendingMention != nil {
		t.Error("warm cross-project cache must not defer the write")
	}

	msg := cmd()
	if _, ok := msg.(mentionUsersLoadedMsg); ok {
		t.Fatal("warm path must not emit mentionUsersLoadedMsg")
	}
	if len(fake.AddCommentCalls) != 1 {
		t.Fatalf("AddComment called %d times, want 1", len(fake.AddCommentCalls))
	}
	body, _ := fake.AddCommentCalls[0].Body.(string)
	if !strings.Contains(body, "accountid:s1") {
		t.Errorf("comment body = %q, want accountid:s1 from the issue's project users", body)
	}
}

func TestMentionColdPath_CrossProject(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{
		GetUsersFunc: func(_ context.Context, _ string) ([]jira.User, error) {
			return []jira.User{soloUser()}, nil
		},
		AddCommentFunc: func(_ context.Context, _ string, _ any) (*jira.Comment, error) {
			return &jira.Comment{}, nil
		},
	}
	app := newAppWithFake(t, fake)
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject
	app.editContext = editCtx{kind: editCommentNew, issueKey: "OTHER-7"}

	cmd := app.applyEdit("ping @Solo_One")
	if cmd == nil {
		t.Fatal("applyEdit returned nil cmd on cold cache")
	}

	msg := cmd()
	loaded, ok := msg.(mentionUsersLoadedMsg)
	if !ok {
		t.Fatalf("expected mentionUsersLoadedMsg, got %T", msg)
	}
	if len(fake.GetUsersCalls) != 1 || fake.GetUsersCalls[0].ProjectKey != "OTHER" {
		t.Fatalf("GetUsers should be called once for OTHER, got %+v", fake.GetUsersCalls)
	}

	_, follow := app.handleMentionUsersLoaded(loaded)
	if follow == nil {
		t.Fatal("handleMentionUsersLoaded returned nil cmd")
	}
	follow()

	if _, ok := app.usersCache["OTHER"]; !ok {
		t.Error("loaded users should be cached under the issue's project key OTHER")
	}
	if len(fake.AddCommentCalls) != 1 {
		t.Fatalf("AddComment called %d times, want 1", len(fake.AddCommentCalls))
	}
	body, _ := fake.AddCommentCalls[0].Body.(string)
	if !strings.Contains(body, "accountid:s1") {
		t.Errorf("comment body = %q, want accountid:s1", body)
	}
}

func TestMentionColdPath_GetUsersError(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{
		T: t,
		GetUsersFunc: func(_ context.Context, _ string) ([]jira.User, error) {
			return nil, errors.New("boom")
		},
		AddCommentFunc: func(_ context.Context, _ string, _ any) (*jira.Comment, error) {
			return &jira.Comment{}, nil
		},
	}
	app := newAppWithFake(t, fake)
	app.converter = identityConverter{}
	app.isCloud = true
	app.projectKey = testProject
	app.editContext = editCtx{kind: editCommentNew, issueKey: testKey}

	cmd := app.applyEdit("ping @Solo_One")
	msg := cmd()
	loaded, ok := msg.(mentionUsersLoadedMsg)
	if !ok {
		t.Fatalf("expected mentionUsersLoadedMsg even on fetch error, got %T", msg)
	}
	if len(loaded.users) != 0 {
		t.Fatalf("expected empty user list on error, got %d", len(loaded.users))
	}
	_, follow := app.handleMentionUsersLoaded(loaded)
	if follow == nil {
		t.Fatal("write must still proceed after a fetch error")
	}
	follow()

	if len(fake.AddCommentCalls) != 1 {
		t.Fatalf("AddComment called %d times, want 1", len(fake.AddCommentCalls))
	}
	body, _ := fake.AddCommentCalls[0].Body.(string)
	if !strings.Contains(body, "@Solo_One") {
		t.Errorf("comment body = %q, want the literal @Solo_One on fetch error", body)
	}
	if strings.Contains(body, "accountid") {
		t.Errorf("comment body = %q, must not resolve when users could not be fetched", body)
	}
}
