package git

import (
	"slices"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestCreateBranch_CreatesAndChecksOut(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	if err := CreateBranch(dir, "feature/PROJ-1-foo"); err != nil {
		t.Fatalf("CreateBranch: %v", err)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	testkit.AssertEqual(t, "current branch", branch, "feature/PROJ-1-foo")
	testkit.AssertEqual(t, "BranchExists", BranchExists(dir, "feature/PROJ-1-foo"), true)
}

func TestCreateBranch_DuplicateReturnsError(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	createBranch(t, dir, "dup")

	if err := CreateBranch(dir, "dup"); err == nil {
		t.Error("expected error creating an existing branch, got nil")
	}
}

func TestCheckout_SwitchesBranch(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	createBranch(t, dir, "topic")

	if err := Checkout(dir, "topic"); err != nil {
		t.Fatalf("Checkout: %v", err)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	testkit.AssertEqual(t, "current branch", branch, "topic")
}

func TestCheckout_UnknownBranchReturnsError(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	if err := Checkout(dir, "ghost"); err == nil {
		t.Error("expected error checking out an unknown branch, got nil")
	}
}

func TestCheckoutTracking_CreatesLocalTrackingBranch(t *testing.T) {
	t.Parallel()
	upstream := initRepo(t)
	createBranch(t, upstream, "feature-x")

	dir := initRepo(t)
	gitRun(t, dir, "remote", "add", "origin", upstream)
	gitRun(t, dir, "fetch", "-q", "origin")

	if err := CheckoutTracking(dir, "origin/feature-x"); err != nil {
		t.Fatalf("CheckoutTracking: %v", err)
	}

	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	testkit.AssertEqual(t, "current branch", branch, "feature-x")
}

func TestCheckoutTracking_InvalidFormatReturnsError(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	if err := CheckoutTracking(dir, "nosuchslash"); err == nil {
		t.Error("expected error for remote branch without a slash, got nil")
	}
}

func TestLocalBranches_ListsAllLocal(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	createBranch(t, dir, "topic-a")
	createBranch(t, dir, "topic-b")

	branches, err := LocalBranches(dir)
	if err != nil {
		t.Fatalf("LocalBranches: %v", err)
	}

	for _, want := range []string{"main", "topic-a", "topic-b"} {
		if !slices.Contains(branches, want) {
			t.Errorf("LocalBranches %v missing %q", branches, want)
		}
	}
}

func TestRemoteBranches_ReturnsRemotesWithoutSymbolicHead(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)
	addRemoteRef(t, dir, "origin/main")
	addRemoteRef(t, dir, "origin/feature-x")

	branches, err := RemoteBranches(dir)
	if err != nil {
		t.Fatalf("RemoteBranches: %v", err)
	}

	testkit.AssertSliceEqual(t, "remote branches", branches, []string{"origin/feature-x", "origin/main"})
}

func TestBranchExists_TrueForMainFalseForUnknown(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	testkit.AssertEqual(t, "main exists", BranchExists(dir, "main"), true)
	testkit.AssertEqual(t, "ghost exists", BranchExists(dir, "ghost"), false)
}

func TestCheckoutTracking_UnknownRemoteBranchReturnsError(t *testing.T) {
	t.Parallel()
	dir := initRepo(t)

	if err := CheckoutTracking(dir, "origin/does-not-exist"); err == nil {
		t.Fatal("CheckoutTracking with unknown remote branch should error")
	}
}
