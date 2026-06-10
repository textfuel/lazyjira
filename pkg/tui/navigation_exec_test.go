package tui

import (
	"slices"
	"testing"
)

func TestExternalCommandDispatch(t *testing.T) {
	type capture struct {
		input       string
		waitForExit bool
		name        string
		args        []string
	}
	var got capture
	original := runExternalCommand
	t.Cleanup(func() { runExternalCommand = original })
	runExternalCommand = func(input string, waitForExit bool, name string, args ...string) {
		got = capture{input: input, waitForExit: waitForExit, name: name, args: args}
	}

	t.Run("openBrowser launches without waiting", func(t *testing.T) {
		got = capture{}
		url := "https://x.atlassian.net/browse/" + testKey

		openBrowser(url)

		if got.name == "" {
			t.Fatal("openBrowser did not dispatch a command")
		}
		if got.waitForExit {
			t.Error("browser launch should be fire and forget")
		}
		if got.input != "" {
			t.Error("browser launch should not feed stdin")
		}
		if !slices.Contains(got.args, url) {
			t.Errorf("args %v should carry the url", got.args)
		}
	})

	t.Run("copyToClipboard feeds stdin and waits", func(t *testing.T) {
		got = capture{}

		copyToClipboard("payload")

		if got.name == "" {
			t.Fatal("copyToClipboard did not dispatch a command")
		}
		if !got.waitForExit {
			t.Error("clipboard write should wait for the pipe to drain")
		}
		if got.input != "payload" {
			t.Errorf("stdin = %q, want payload", got.input)
		}
	})
}
