package views

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/config"
)

const (
	glamourStyleDark  = "dark"
	glamourStyleLight = "light"
	glamourStyleNoTTY = "notty"
)

func ResolveGlamourStyle(style string) string {
	return resolveGlamourStyle(style, lipgloss.HasDarkBackground)
}

func resolveGlamourStyle(style string, hasDarkBackground func() bool) string {
	switch style {
	case config.RendererStyleDark:
		return glamourStyleDark
	case config.RendererStyleLight:
		return glamourStyleLight
	case config.RendererStyleNoTTY:
		return glamourStyleNoTTY
	}
	if hasDarkBackground() {
		return glamourStyleDark
	}
	return glamourStyleLight
}
