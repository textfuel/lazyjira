package git

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestSearchBranches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		localBranches []string
		remoteRefs    []string
		query         string
		wantLocal     []string
		wantRemote    []string
	}{
		{
			name:          "local match respects digit boundary",
			localBranches: []string{"PLAT-3-fix", "PLAT-30-other", "unrelated"},
			query:         "PLAT-3",
			wantLocal:     []string{"PLAT-3-fix"},
		},
		{
			name:          "match is case insensitive",
			localBranches: []string{"plat-3-low"},
			query:         "PLAT-3",
			wantLocal:     []string{"plat-3-low"},
		},
		{
			name:       "remote only match",
			remoteRefs: []string{"origin/PLAT-5-z"},
			query:      "PLAT-5",
			wantRemote: []string{"origin/PLAT-5-z"},
		},
		{
			name:          "remote deduped against local",
			localBranches: []string{"PLAT-7-x"},
			remoteRefs:    []string{"origin/PLAT-7-x"},
			query:         "PLAT-7",
			wantLocal:     []string{"PLAT-7-x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := initRepo(t)
			for _, branch := range tt.localBranches {
				createBranch(t, dir, branch)
			}
			for _, ref := range tt.remoteRefs {
				addRemoteRef(t, dir, ref)
			}

			result, err := SearchBranches(dir, tt.query)
			if err != nil {
				t.Fatalf("SearchBranches: %v", err)
			}
			testkit.AssertSliceEqual(t, "local", result.Local, tt.wantLocal)
			testkit.AssertSliceEqual(t, "remote", result.Remote, tt.wantRemote)
		})
	}
}

func TestSearchBranches_NonGitDirReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if _, err := SearchBranches(dir, "PLAT-1"); err == nil {
		t.Error("expected error outside a git repo, got nil")
	}
}
