package git

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestGitAvailable_FindsGitInPath(t *testing.T) {
	t.Parallel()
	if !GitAvailable() {
		t.Skip("git binary not found in PATH")
	}
	testkit.AssertEqual(t, "GitAvailable", GitAvailable(), true)
}

func TestIsRepo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		insideRepo bool
		want       bool
	}{
		{name: "inside a repo", insideRepo: true, want: true},
		{name: "outside a repo", insideRepo: false, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if tt.insideRepo {
				dir = initRepo(t)
			}
			testkit.AssertEqual(t, "IsRepo", IsRepo(dir), tt.want)
		})
	}
}

func TestCurrentBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		nonRepo bool
		detach  bool
		want    string
	}{
		{name: "returns checked out branch", want: "main"},
		{name: "detached head returns empty", detach: true, want: ""},
		{name: "non repo returns empty", nonRepo: true, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			if !tt.nonRepo {
				dir = initRepo(t)
				if tt.detach {
					gitRun(t, dir, "checkout", "-q", "--detach", "HEAD")
				}
			}

			got, err := CurrentBranch(dir)
			if err != nil {
				t.Fatalf("CurrentBranch: %v", err)
			}
			testkit.AssertEqual(t, "branch", got, tt.want)
		})
	}
}

func TestCurrentBranch_GitMissingFromPathReturnsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", "")

	branch, err := CurrentBranch(dir)
	if err == nil {
		t.Fatal("CurrentBranch without git in PATH should error")
	}
	if branch != "" {
		t.Errorf("branch = %q, want empty on error", branch)
	}
}
