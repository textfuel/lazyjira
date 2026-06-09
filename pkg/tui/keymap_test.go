package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/config"
	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/tui/components"
)

func TestKeymapFromConfig_OverridesAndMatches(t *testing.T) {
	t.Parallel()

	var cfg config.KeybindingConfig
	cfg.Universal.Quit = "Q"
	cfg.Navigation.Down = "n"

	km := KeymapFromConfig(cfg)

	testkit.AssertSliceEqual(t, "quit binding overridden", km[ActQuit], []string{"Q"})
	testkit.AssertEqual(t, "Match resolves override", km.Match("Q"), ActQuit)
	testkit.AssertEqual(t, "MatchNav resolves override", km.MatchNav("n"), components.NavDown)
}

func TestKeymapFromConfig_EmptyKeepsDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultKeymap()
	km := KeymapFromConfig(config.KeybindingConfig{})

	testkit.AssertSliceEqual(t, "quit default preserved", km[ActQuit], defaults[ActQuit])
}

func TestKeymap_MatchUnknownReturnsEmpty(t *testing.T) {
	t.Parallel()

	km := DefaultKeymap()

	testkit.AssertEqual(t, "unknown key", km.Match("this-key-is-unbound"), Action(""))
	testkit.AssertEqual(t, "unknown nav key", km.MatchNav("this-key-is-unbound"), components.NavNone)
}
