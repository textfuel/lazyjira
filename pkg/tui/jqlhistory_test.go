package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestAddToHistory_PrependsDeduplicated(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		existing []string
		newQuery string
		want     []string
	}{
		{
			name:     "empty history gains one entry",
			existing: nil,
			newQuery: "project = X",
			want:     []string{"project = X"},
		},
		{
			name:     "new query prepended",
			existing: []string{"project = A", "project = B"},
			newQuery: "project = C",
			want:     []string{"project = C", "project = A", "project = B"},
		},
		{
			name:     "duplicate moved to front",
			existing: []string{"project = A", "project = B"},
			newQuery: "project = B",
			want:     []string{"project = B", "project = A"},
		},
		{
			name:     "blank query is no-op",
			existing: []string{"project = A"},
			newQuery: "   ",
			want:     []string{"project = A"},
		},
		{
			name:     "whitespace trimmed before dedup",
			existing: []string{"project = X"},
			newQuery: "  project = X  ",
			want:     []string{"project = X"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := AddToHistory(tc.existing, tc.newQuery)
			testkit.AssertSliceEqual(t, "history", got, tc.want)
		})
	}
}

func TestAddToHistory_CapsAt50(t *testing.T) {
	t.Parallel()
	existing := make([]string, 50)
	for i := range existing {
		existing[i] = "query"
	}
	existing[0] = "first"
	existing[49] = "last"

	result := AddToHistory(existing, "brand-new")

	if len(result) != 50 {
		t.Errorf("len = %d, want 50", len(result))
	}
	if result[0] != "brand-new" {
		t.Errorf("result[0] = %q, want brand-new", result[0])
	}
}

func TestSaveAndLoadJQLHistory_RoundTrip(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())

	queries := []string{"project = X ORDER BY updated DESC", "assignee = currentUser()"}

	if err := SaveJQLHistory(queries); err != nil {
		t.Fatalf("SaveJQLHistory: %v", err)
	}

	loaded := LoadJQLHistory()
	testkit.AssertSliceEqual(t, "loaded queries", loaded, queries)
}

func TestLoadJQLHistory_MissingFileReturnsNil(t *testing.T) {
	t.Setenv("LAZYJIRA_CONFIG_DIR", t.TempDir())

	result := LoadJQLHistory()
	if result != nil {
		t.Errorf("expected nil for missing file, got %v", result)
	}
}
