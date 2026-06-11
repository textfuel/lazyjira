package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestApp_SideWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		configured int
		want       int
	}{
		{"vertical layout returns full width", 60, 40, 60},
		{"default when unset", 200, 0, 40},
		{"explicit width honored when wide", 200, 50, 50},
		{"clamped to half the terminal", 200, 150, 100},
		{"shrinks to 35 percent on narrow terminal", 100, 40, 35},
		{"floors at minimum width", 100, 10, 25},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := &App{width: testCase.width, cfg: &config.Config{}}
			app.cfg.GUI.SidePanelWidth = testCase.configured

			testkit.AssertEqual(t, "sideWidth", app.sideWidth(), testCase.want)
		})
	}
}
