package git

import (
	"testing"
)

func TestResolveBranchAction_SlashedNewName_IsCreate(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	addRemoteRef(t, dir, "origin/main")

	got := ResolveBranchAction(dir, "feature/PROJ-1-foo")
	if got != ActionCreate {
		t.Errorf("ResolveBranchAction = %v, want ActionCreate", got)
	}
}

func TestResolveBranchAction_existingLocal_isCheckout(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	got := ResolveBranchAction(dir, "main")
	if got != ActionCheckout {
		t.Errorf("ResolveBranchAction = %v, want ActionCheckout", got)
	}
}

func TestResolveBranchAction_existingRemote_isTracking(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	addRemoteRef(t, dir, "origin/feature-x")

	got := ResolveBranchAction(dir, "origin/feature-x")
	if got != ActionCheckoutTracking {
		t.Errorf("ResolveBranchAction = %v, want ActionCheckoutTracking", got)
	}
}

func TestResolveBranchAction_plainName_isCreate(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	got := ResolveBranchAction(dir, "PROJ-1-foo")
	if got != ActionCreate {
		t.Errorf("ResolveBranchAction = %v, want ActionCreate", got)
	}
}

func TestIsRemoteBranch_exactMatchRequired(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	addRemoteRef(t, dir, "origin/feature/x")

	if IsRemoteBranch(dir, "origin/feature/y") {
		t.Errorf("IsRemoteBranch matched non-existent sibling branch")
	}
	if !IsRemoteBranch(dir, "origin/feature/x") {
		t.Errorf("IsRemoteBranch did not match exact remote branch")
	}
}

func TestIsRemoteBranch_nonGitDir_returnsFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if IsRemoteBranch(dir, "origin/main") {
		t.Errorf("IsRemoteBranch returned true for non-git dir")
	}
}
