package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func TestApplyParentEdit_SetSendsKey(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.UpdateIssueFunc = func(_ context.Context, _ string, _ map[string]any) error { return nil }

	app := newAppWithFake(t, fake)
	app.issueCache[mainKey] = &jira.Issue{Key: mainKey}

	cmd := app.applyParentEdit(mainKey, "NEW-1")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	_ = cmd()

	if got := app.issueCache[mainKey].Parent; got == nil || got.Key != "NEW-1" {
		t.Errorf("optimistic Parent = %+v, want Key=NEW-1", got)
	}
	if len(fake.UpdateIssueCalls) != 1 {
		t.Fatalf("UpdateIssue calls = %d, want 1", len(fake.UpdateIssueCalls))
	}
	parent, _ := fake.UpdateIssueCalls[0].Fields["parent"].(map[string]string)
	if parent["key"] != "NEW-1" {
		t.Errorf("UpdateIssue fields.parent.key = %v, want NEW-1", parent)
	}
}

func TestApplyParentEdit_EmptyCallsRemove(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	fake.RemoveIssueParentFunc = func(_ context.Context, _ string) error { return nil }

	app := newAppWithFake(t, fake)
	app.issueCache[mainKey] = &jira.Issue{Key: mainKey, Parent: &jira.Issue{Key: "OLD-1"}}

	cmd := app.applyParentEdit(mainKey, "")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	_ = cmd()

	if got := app.issueCache[mainKey].Parent; got != nil {
		t.Errorf("optimistic Parent = %+v, want nil", got)
	}
	if len(fake.RemoveIssueParentCalls) != 1 {
		t.Errorf("RemoveIssueParent calls = %d, want 1", len(fake.RemoveIssueParentCalls))
	}
	if len(fake.UpdateIssueCalls) != 0 {
		t.Errorf("UpdateIssue should not be called for unset; got %d calls", len(fake.UpdateIssueCalls))
	}
}

func TestApplyParentEdit_InvalidKeyShortCircuits(t *testing.T) {
	t.Parallel()
	fake := &jiratest.FakeClient{T: t}
	app := newAppWithFake(t, fake)
	app.issueCache[mainKey] = &jira.Issue{Key: mainKey}

	cmd := app.applyParentEdit(mainKey, "not-a-key")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	emsg, ok := msg.(errorMsg)
	if !ok {
		t.Fatalf("msg = %T, want errorMsg", msg)
	}
	if emsg.err == nil || !strings.Contains(emsg.err.Error(), "invalid parent key") {
		t.Errorf("err = %v, want 'invalid parent key ...'", emsg.err)
	}
	if len(fake.UpdateIssueCalls)+len(fake.RemoveIssueParentCalls) != 0 {
		t.Errorf("no client call expected on validation failure")
	}
}
