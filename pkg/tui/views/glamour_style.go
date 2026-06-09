package views

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/config"
)

// Glamour style names accepted by adf-converter's display module.
const (
	glamourStyleDark  = "dark"
	glamourStyleLight = "light"
	glamourStyleNoTTY = "notty"
)

// hasDarkBackground is the terminal background probe used by ResolveGlamourStyle.
// Indirected through a package variable so tests can stub the result without a
// real TTY. lipgloss.HasDarkBackground itself relies on terminal queries that
// can fail under tmux/ssh; "auto" callers should expect a best-effort answer.
var hasDarkBackground = lipgloss.HasDarkBackground

// ResolveGlamourStyle maps a Config.RendererStyle value to the Glamour style
// name accepted by adf-converter's display module.
//
// Empty string and "auto" fall through to lipgloss's terminal background
// detection. Unknown values are treated as "auto" so a typo cannot leave the
// preview blank — config validation already rejects unknown values upstream.
func ResolveGlamourStyle(style string) string {
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
