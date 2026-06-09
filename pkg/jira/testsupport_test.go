package jira

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func cloudOpts() ClientOpts {
	return ClientOpts{Email: "ci@example.com", Token: "secret-token", IsCloud: true}
}

func serverOpts() ClientOpts {
	return ClientOpts{Token: "pat-token", IsCloud: false}
}

func newRecordingClient(t *testing.T, opts ClientOpts, response testkit.StubResponse) (*Client, *testkit.RecordedRequest) {
	t.Helper()

	server, recorded := testkit.RecordingServer(t, response)
	opts.Host = server.URL
	return NewClientWithOpts(opts), recorded
}

func newSequenceClient(t *testing.T, opts ClientOpts, responses ...testkit.StubResponse) (*Client, *[]testkit.RecordedRequest) {
	t.Helper()

	server, recorded := testkit.RecordingSequenceServer(t, responses)
	opts.Host = server.URL
	return NewClientWithOpts(opts), recorded
}
