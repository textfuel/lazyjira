package components

import (
	"testing"

	"github.com/charmbracelet/x/ansi"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestComputeLCS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    []string
		b    []string
		want []string
	}{
		{"both empty", nil, nil, []string{}},
		{"identical", []string{"a", "b"}, []string{"a", "b"}, []string{"a", "b"}},
		{"no common", []string{"a", "b"}, []string{"x", "y"}, []string{}},
		{"subsequence", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"a", "c"}},
		{"interleaved", []string{"a", "b", "c", "d"}, []string{"b", "d"}, []string{"b", "d"}},
		{"single common in middle", []string{"a", "b", "c"}, []string{"x", "b", "y"}, []string{"b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := computeLCS(tt.a, tt.b)
			testkit.AssertSliceEqual(t, "lcs", got, tt.want)
		})
	}
}

func TestComputeUnifiedDiff_MarksAddsRemovesAndContext(t *testing.T) {
	t.Parallel()

	got := computeUnifiedDiff("a\nb\nc", "a\nx\nc")

	stripped := make([]string, len(got))
	for i, line := range got {
		stripped[i] = ansi.Strip(line)
	}

	testkit.AssertSliceEqual(t, "diff lines", stripped, []string{"  a", "- b", "+ x", "  c"})
}
