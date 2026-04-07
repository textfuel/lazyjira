package theme

import "github.com/charmbracelet/lipgloss"

// Catppuccin color palettes.
// Spec: https://github.com/catppuccin/catppuccin

func catppuccinLatte() *Theme {
	return buildTheme(
		ColorPalette{
			Green:     lipgloss.Color("#40a02b"),
			Blue:      lipgloss.Color("#1e66f5"),
			Red:       lipgloss.Color("#d20f39"),
			Yellow:    lipgloss.Color("#df8e1d"),
			Cyan:      lipgloss.Color("#179299"),
			White:     lipgloss.Color("#4c4f69"), // Text
			Gray:      lipgloss.Color("#9ca0b0"), // Overlay0
			Orange:    lipgloss.Color("#fe640b"), // Peach
			None:      lipgloss.Color("-1"),
			Highlight: lipgloss.Color("#acb0be"), // Surface2
		},
		[]lipgloss.Color{
			lipgloss.Color("#dc8a78"), // Rosewater
			lipgloss.Color("#dd7878"), // Flamingo
			lipgloss.Color("#ea76cb"), // Pink
			lipgloss.Color("#8839ef"), // Mauve
			lipgloss.Color("#d20f39"), // Red
			lipgloss.Color("#e64553"), // Maroon
			lipgloss.Color("#fe640b"), // Peach
			lipgloss.Color("#df8e1d"), // Yellow
			lipgloss.Color("#40a02b"), // Green
			lipgloss.Color("#179299"), // Teal
			lipgloss.Color("#209fb5"), // Sapphire
			lipgloss.Color("#7287fd"), // Lavender
		},
	)
}

func catppuccinFrappe() *Theme {
	return buildTheme(
		ColorPalette{
			Green:     lipgloss.Color("#a6d189"),
			Blue:      lipgloss.Color("#8caaee"),
			Red:       lipgloss.Color("#e78284"),
			Yellow:    lipgloss.Color("#e5c890"),
			Cyan:      lipgloss.Color("#81c8be"),
			White:     lipgloss.Color("#c6d0f5"), // Text
			Gray:      lipgloss.Color("#737994"), // Overlay0
			Orange:    lipgloss.Color("#ef9f76"), // Peach
			None:      lipgloss.Color("-1"),
			Highlight: lipgloss.Color("#626880"), // Surface2
		},
		[]lipgloss.Color{
			lipgloss.Color("#f2d5cf"), // Rosewater
			lipgloss.Color("#eebebe"), // Flamingo
			lipgloss.Color("#f4b8e4"), // Pink
			lipgloss.Color("#ca9ee6"), // Mauve
			lipgloss.Color("#e78284"), // Red
			lipgloss.Color("#ea999c"), // Maroon
			lipgloss.Color("#ef9f76"), // Peach
			lipgloss.Color("#e5c890"), // Yellow
			lipgloss.Color("#a6d189"), // Green
			lipgloss.Color("#81c8be"), // Teal
			lipgloss.Color("#85c1dc"), // Sapphire
			lipgloss.Color("#babbf1"), // Lavender
		},
	)
}

func catppuccinMacchiato() *Theme {
	return buildTheme(
		ColorPalette{
			Green:     lipgloss.Color("#a6da95"),
			Blue:      lipgloss.Color("#8aadf4"),
			Red:       lipgloss.Color("#ed8796"),
			Yellow:    lipgloss.Color("#eed49f"),
			Cyan:      lipgloss.Color("#8bd5ca"),
			White:     lipgloss.Color("#cad3f5"), // Text
			Gray:      lipgloss.Color("#6e738d"), // Overlay0
			Orange:    lipgloss.Color("#f5a97f"), // Peach
			None:      lipgloss.Color("-1"),
			Highlight: lipgloss.Color("#5b6078"), // Surface2
		},
		[]lipgloss.Color{
			lipgloss.Color("#f4dbd6"), // Rosewater
			lipgloss.Color("#f0c6c6"), // Flamingo
			lipgloss.Color("#f5bde6"), // Pink
			lipgloss.Color("#c6a0f6"), // Mauve
			lipgloss.Color("#ed8796"), // Red
			lipgloss.Color("#ee99a0"), // Maroon
			lipgloss.Color("#f5a97f"), // Peach
			lipgloss.Color("#eed49f"), // Yellow
			lipgloss.Color("#a6da95"), // Green
			lipgloss.Color("#8bd5ca"), // Teal
			lipgloss.Color("#7dc4e4"), // Sapphire
			lipgloss.Color("#b7bdf8"), // Lavender
		},
	)
}

func catppuccinMocha() *Theme {
	return buildTheme(
		ColorPalette{
			Green:     lipgloss.Color("#a6e3a1"),
			Blue:      lipgloss.Color("#89b4fa"),
			Red:       lipgloss.Color("#f38ba8"),
			Yellow:    lipgloss.Color("#f9e2af"),
			Cyan:      lipgloss.Color("#94e2d5"),
			White:     lipgloss.Color("#cdd6f4"), // Text
			Gray:      lipgloss.Color("#6c7086"), // Overlay0
			Orange:    lipgloss.Color("#fab387"), // Peach
			None:      lipgloss.Color("-1"),
			Highlight: lipgloss.Color("#585b70"), // Surface2
		},
		[]lipgloss.Color{
			lipgloss.Color("#f5e0dc"), // Rosewater
			lipgloss.Color("#f2cdcd"), // Flamingo
			lipgloss.Color("#f5c2e7"), // Pink
			lipgloss.Color("#cba6f7"), // Mauve
			lipgloss.Color("#f38ba8"), // Red
			lipgloss.Color("#eba0ac"), // Maroon
			lipgloss.Color("#fab387"), // Peach
			lipgloss.Color("#f9e2af"), // Yellow
			lipgloss.Color("#a6e3a1"), // Green
			lipgloss.Color("#94e2d5"), // Teal
			lipgloss.Color("#74c7ec"), // Sapphire
			lipgloss.Color("#b4befe"), // Lavender
		},
	)
}
