package components

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestAdjustOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cursor  int
		offset  int
		visible int
		total   int
		want    int
	}{
		{"all items fit", 5, 3, 10, 8, 0},
		{"cursor at top scrolls up", 0, 5, 5, 20, 0},
		{"cursor at bottom scrolls down", 19, 0, 5, 20, 15},
		{"cursor in middle keeps offset", 10, 8, 5, 20, 8},
		{"tiny viewport no margin keeps top", 2, 0, 3, 10, 0},
		{"tiny viewport no margin scrolls", 5, 0, 3, 10, 3},
		{"offset clamped to last page", 18, 18, 5, 20, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := AdjustOffset(tt.cursor, tt.offset, tt.visible, tt.total)
			testkit.AssertEqual(t, "offset", got, tt.want)
		})
	}
}
