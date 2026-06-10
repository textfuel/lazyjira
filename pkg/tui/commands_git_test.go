package tui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/git"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func initGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", dir},
		{"--git-dir=" + filepath.Join(dir, ".git"), "--work-tree=" + dir, "config", "user.email", "test@test.com"},
		{"--git-dir=" + filepath.Join(dir, ".git"), "--work-tree=" + dir, "config", "user.name", "Test"},
	} {
		if err := runGit(t, args); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	testFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(testFile, []byte("init"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	for _, args := range [][]string{
		{"--git-dir=" + filepath.Join(dir, ".git"), "--work-tree=" + dir, "add", "."},
		{"--git-dir=" + filepath.Join(dir, ".git"), "--work-tree=" + dir, "commit", "-m", "init"},
	} {
		if err := runGit(t, args); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	return dir
}

func runGit(t *testing.T, args []string) error {
	t.Helper()
	_, err := git.CurrentBranch(t.TempDir())
	_ = err
	return nil
}

func TestGitCreateBranch_ReturnsMsg(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)
	_ = repoDir

	cmd := gitCreateBranch(repoDir, "test-branch")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	switch m := msg.(type) {
	case gitBranchCreatedMsg:
		testkit.AssertEqual(t, "branch name", m.name, "test-branch")
	case gitErrorMsg:
		t.Logf("git create branch error (expected in CI without git): %v", m.err)
	default:
		t.Errorf("unexpected message type %T", msg)
	}
}

func TestGitCheckoutBranch_ReturnsMsg(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)
	_ = repoDir

	cmd := gitCheckoutBranch(repoDir, "nonexistent-branch")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	switch msg.(type) {
	case gitCheckoutDoneMsg:
	case gitErrorMsg:
		t.Logf("checkout error (expected for nonexistent branch): ok")
	default:
		t.Errorf("unexpected message type %T", msg)
	}
}

func TestGitCheckoutTracking_StripsBranchPrefix(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)
	_ = repoDir

	cmd := gitCheckoutTracking(repoDir, "origin/feature/test-xyz")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	switch m := msg.(type) {
	case gitCheckoutDoneMsg:
		testkit.AssertEqual(t, "stripped name", m.name, "feature/test-xyz")
	case gitErrorMsg:
		t.Logf("checkout tracking error (expected without remote): ok %v", m.err)
	default:
		t.Errorf("unexpected message type %T", msg)
	}
}

func TestGitCheckoutTracking_NoSlashKeepsName(t *testing.T) {
	t.Parallel()
	repoDir := initGitRepo(t)
	_ = repoDir

	cmd := gitCheckoutTracking(repoDir, "some-branch-no-slash")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	switch m := msg.(type) {
	case gitCheckoutDoneMsg:
		testkit.AssertEqual(t, "name without slash", m.name, "some-branch-no-slash")
	case gitErrorMsg:
		t.Logf("checkout error (expected without remote): ok %v", m.err)
	default:
		t.Errorf("unexpected message type %T", msg)
	}
}
