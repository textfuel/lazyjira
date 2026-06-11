package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
	"github.com/textfuel/lazyjira/v2/pkg/tui/views"
)

func TestPreviewRequestMsg_SetsPreviewKeyAndFetches(t *testing.T) {
	t.Parallel()
	subKey := subKey1

	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: subKey, Summary: "sub issue"})

	app := newAppWithFake(t, fake)

	_, _ = app.Update(views.PreviewRequestMsg{Key: subKey})

	if got := app.previewKey; got != subKey {
		t.Errorf("previewKey = %q, want %q", got, subKey)
	}
}

func TestPreviewRequestMsg_CmdEventuallyCallsGetIssue(t *testing.T) {
	t.Parallel()
	subKey := subKey1

	fake := &jiratest.FakeClient{T: t}
	stubFullIssueFetch(fake, &jira.Issue{Key: subKey, Summary: "sub issue"})

	app := newAppWithFake(t, fake)

	_, tickCmd := app.Update(views.PreviewRequestMsg{Key: subKey})
	if tickCmd == nil {
		t.Fatal("expected non-nil tea.Cmd from PreviewRequestMsg handler, got nil")
	}

	_, fetchCmd := app.Update(previewDebounceMsg{key: subKey, epoch: app.previewEpoch})
	if fetchCmd == nil {
		t.Fatal("expected fetch cmd from debounce tick, got nil")
	}

	fetchCmd()

	if len(fake.GetIssueCalls) != 1 {
		t.Fatalf("expected 1 GetIssue call, got %d: %+v", len(fake.GetIssueCalls), fake.GetIssueCalls)
	}
	if got := fake.GetIssueCalls[0].Key; got != subKey {
		t.Errorf("GetIssue called with key %q, want %q", got, subKey)
	}
}
