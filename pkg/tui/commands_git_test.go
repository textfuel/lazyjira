package tui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitArgs := [][]string{
		{"init", dir},
		{"-C", dir, "config", "user.email", "test@test.com"},
		{"-C", dir, "config", "user.name", "Test"},
	}
	for _, args := range gitArgs {
		if err := runGit(t, args); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	testFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(testFile, []byte("init"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	for _, args := range [][]string{
		{"-C", dir, "add", "."},
		{"-C", dir, "commit", "-m", "init"},
	} {
		if err := runGit(t, args); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	return dir
}

func runGit(t *testing.T, args []string) error {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("git %v output: %s", args, out)
	}
	return err
}

func TestGitCreateBranch_ReturnsMsg(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)

	cmd := gitCreateBranch(repoDir, "test-branch")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	created, ok := msg.(gitBranchCreatedMsg)
	if !ok {
		t.Fatalf("expected gitBranchCreatedMsg, got %T: %v", msg, msg)
	}
	testkit.AssertEqual(t, "branch name", created.name, "test-branch")
}

func TestGitCheckoutBranch_ReturnsMsg(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)

	if err := runGit(t, []string{"-C", repoDir, "branch", "existing-branch"}); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	cmd := gitCheckoutBranch(repoDir, "existing-branch")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	checked, ok := msg.(gitCheckoutDoneMsg)
	if !ok {
		t.Fatalf("expected gitCheckoutDoneMsg, got %T: %v", msg, msg)
	}
	testkit.AssertEqual(t, "branch name", checked.name, "existing-branch")
}

func TestGitCheckoutTracking_StripsBranchPrefix(t *testing.T) {
	t.Parallel()
	remotePath := initGitRepo(t)
	if err := runGit(t, []string{"-C", remotePath, "branch", "tracked-feature"}); err != nil {
		t.Fatalf("create remote branch: %v", err)
	}

	repoDir := initGitRepo(t)
	if err := runGit(t, []string{"-C", repoDir, "remote", "add", "origin", remotePath}); err != nil {
		t.Fatalf("add remote: %v", err)
	}
	if err := runGit(t, []string{"-C", repoDir, "fetch", "origin"}); err != nil {
		t.Fatalf("fetch: %v", err)
	}

	cmd := gitCheckoutTracking(repoDir, "origin/tracked-feature")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	checked, ok := msg.(gitCheckoutDoneMsg)
	if !ok {
		t.Fatalf("expected gitCheckoutDoneMsg, got %T: %v", msg, msg)
	}
	testkit.AssertEqual(t, "stripped name", checked.name, "tracked-feature")
}

func TestGitCheckoutTracking_NoSlashReturnsError(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)

	cmd := gitCheckoutTracking(repoDir, "some-branch-no-slash")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	if _, ok := msg.(gitErrorMsg); !ok {
		t.Errorf("expected gitErrorMsg for no-slash remote name, got %T", msg)
	}
}
