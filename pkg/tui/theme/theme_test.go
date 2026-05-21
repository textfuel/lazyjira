package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestSetThemeDefault(t *testing.T) {
	// Ensure SetTheme("default") restores ANSI palette.
	if err := SetTheme("default"); err != nil {
		t.Fatalf("SetTheme(default): %v", err)
	}
	if ColorGreen != lipgloss.Color("2") {
		t.Errorf("ColorGreen = %q, want %q", ColorGreen, "2")
	}
	if ColorBlue != lipgloss.Color("4") {
		t.Errorf("ColorBlue = %q, want %q", ColorBlue, "4")
	}
}

func TestSetThemeEmpty(t *testing.T) {
	// Empty preset must resolve to the legacy ANSI 16 default so that
	// upgrading users who never set gui.theme see no visual change.
	if err := SetTheme(""); err != nil {
		t.Fatalf("SetTheme(''): %v", err)
	}
	if ColorGreen != lipgloss.Color("2") {
		t.Errorf("ColorGreen = %q, want ANSI 2", ColorGreen)
	}
	if ColorBlue != lipgloss.Color("4") {
		t.Errorf("ColorBlue = %q, want ANSI 4", ColorBlue)
	}
	_ = SetTheme("default")
}

func TestSetThemeAuto(t *testing.T) {
	// "auto" picks a Catppuccin default based on terminal background. We
	// can't predict which lands, but it must populate the palette and the
	// chosen preset must be one of the registered defaults.
	if err := SetTheme("auto"); err != nil {
		t.Fatalf("SetTheme(auto): %v", err)
	}
	if Default.Colors.Green == "" {
		t.Error("auto-detected palette has empty Green")
	}
	got := string(Default.Colors.Green)
	if got != "#a6e3a1" && got != "#40a02b" {
		t.Errorf("auto Green = %q, want mocha or latte default", got)
	}
	_ = SetTheme("default")
}

func TestSetThemeCatppuccinMocha(t *testing.T) {
	if err := SetTheme("catppuccin-mocha"); err != nil {
		t.Fatalf("SetTheme(catppuccin-mocha): %v", err)
	}
	if ColorGreen != lipgloss.Color("#a6e3a1") {
		t.Errorf("ColorGreen = %q, want %q", ColorGreen, "#a6e3a1")
	}
	if ColorBlue != lipgloss.Color("#89b4fa") {
		t.Errorf("ColorBlue = %q, want %q", ColorBlue, "#89b4fa")
	}
	if Default.Colors.Red != lipgloss.Color("#f38ba8") {
		t.Errorf("Default.Colors.Red = %q, want %q", Default.Colors.Red, "#f38ba8")
	}

	// Restore default for other tests.
	_ = SetTheme("default")
}

func TestSetThemeAllFlavors(t *testing.T) {
	flavors := []string{
		"catppuccin-latte",
		"catppuccin-frappe",
		"catppuccin-macchiato",
		"catppuccin-mocha",
	}
	for _, name := range flavors {
		t.Run(name, func(t *testing.T) {
			if err := SetTheme(name); err != nil {
				t.Fatalf("SetTheme(%s): %v", name, err)
			}
			// Verify palette is populated (not empty).
			if Default.Colors.Green == "" {
				t.Error("Colors.Green is empty")
			}
			if len(Default.AuthorPalette) != 12 {
				t.Errorf("AuthorPalette has %d entries, want 12", len(Default.AuthorPalette))
			}
		})
	}
	_ = SetTheme("default")
}

func TestSetThemeUnknown(t *testing.T) {
	err := SetTheme("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown theme")
	}
}

func TestSetThemeSyncsColors(t *testing.T) {
	_ = SetTheme("catppuccin-mocha")
	// Package-level vars must match Default.Colors.
	if ColorGreen != Default.Colors.Green {
		t.Errorf("ColorGreen not synced: %q != %q", ColorGreen, Default.Colors.Green)
	}
	if ColorBlue != Default.Colors.Blue {
		t.Errorf("ColorBlue not synced: %q != %q", ColorBlue, Default.Colors.Blue)
	}
	if ColorOrange != Default.Colors.Orange {
		t.Errorf("ColorOrange not synced: %q != %q", ColorOrange, Default.Colors.Orange)
	}
	if ColorMagenta != Default.Colors.Magenta {
		t.Errorf("ColorMagenta not synced: %q != %q", ColorMagenta, Default.Colors.Magenta)
	}
	_ = SetTheme("default")
}

func TestSetThemeSyncsAuthorPalette(t *testing.T) {
	_ = SetTheme("default")
	defaultFirst := authorPalette[0]

	_ = SetTheme("catppuccin-mocha")
	if authorPalette[0] == defaultFirst {
		t.Error("authorPalette did not switch when theme changed")
	}
	if len(authorPalette) != len(Default.AuthorPalette) {
		t.Errorf("authorPalette length %d != Default.AuthorPalette length %d",
			len(authorPalette), len(Default.AuthorPalette))
	}
	_ = SetTheme("default")
}

func TestSetThemeResetsAuthorCache(t *testing.T) {
	_ = SetTheme("default")
	// Prime the author cache.
	_ = AuthorStyle("Alice")
	if len(authorCache) == 0 {
		t.Fatal("author cache should have an entry")
	}

	// Switching theme must clear the cache.
	_ = SetTheme("catppuccin-mocha")
	if len(authorCache) != 0 {
		t.Error("author cache should be empty after theme switch")
	}
	_ = SetTheme("default")
}

func TestInitAppliesSharedOverrides(t *testing.T) {
	err := Init(Options{
		Preset: "default",
		Colors: map[string]string{
			"green":     "#abcdef",
			"highlight": "#123456",
			"bogus":     "ignored",
			"red":       "", // empty value must be skipped
		},
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Default.Colors.Green != lipgloss.Color("#abcdef") {
		t.Errorf("Green = %q, want #abcdef", Default.Colors.Green)
	}
	if ColorGreen != lipgloss.Color("#abcdef") {
		t.Errorf("ColorGreen not synced: %q", ColorGreen)
	}
	if Default.Colors.Highlight != lipgloss.Color("#123456") {
		t.Errorf("Highlight = %q, want #123456", Default.Colors.Highlight)
	}
	// Empty value must leave Red at the preset's value.
	if Default.Colors.Red != lipgloss.Color("1") {
		t.Errorf("Red = %q, want preset default 1", Default.Colors.Red)
	}
	_ = SetTheme("default")
}

func TestInitAppliesDarkOverridesOnlyOnDarkPreset(t *testing.T) {
	// Mocha is a dark preset → ColorsDark applies, ColorsLight is ignored.
	err := Init(Options{
		Preset:      "catppuccin-mocha",
		ColorsDark:  map[string]string{"green": "#111111"},
		ColorsLight: map[string]string{"green": "#999999"},
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Default.Colors.Green != lipgloss.Color("#111111") {
		t.Errorf("dark override not applied: Green = %q", Default.Colors.Green)
	}

	// Latte is a light preset → ColorsLight applies instead.
	err = Init(Options{
		Preset:      "catppuccin-latte",
		ColorsDark:  map[string]string{"green": "#111111"},
		ColorsLight: map[string]string{"green": "#999999"},
	})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if Default.Colors.Green != lipgloss.Color("#999999") {
		t.Errorf("light override not applied: Green = %q", Default.Colors.Green)
	}
	_ = SetTheme("default")
}

func TestFindPresetCaseInsensitive(t *testing.T) {
	if FindPreset("Catppuccin-Mocha") == nil {
		t.Error("FindPreset should be case-insensitive")
	}
	if FindPreset("does-not-exist") != nil {
		t.Error("FindPreset should return nil for unknown names")
	}
}

func TestPresetsListed(t *testing.T) {
	got := Presets()
	if len(got) < 5 {
		t.Errorf("expected at least 5 presets, got %d", len(got))
	}
}

func TestInitToleratesInvalidColorValues(t *testing.T) {
	err := Init(Options{
		Preset: "default",
		Colors: map[string]string{
			"green": "not-a-color",
			"blue":  "#zzzzzz",
			"red":   "totally bogus value",
		},
	})
	if err != nil {
		t.Fatalf("Init must accept malformed color values: %v", err)
	}
	// Palette must still be populated; lipgloss.Color is a string type so
	// the override is stored verbatim and rendering must not panic on use.
	if Default.Colors.Green == "" {
		t.Error("Green became empty after invalid override")
	}
	// Sanity check: rendering with the invalid color does not crash.
	_ = Default.Title.Render("smoke test")
	_ = SetTheme("default")
}

func TestPresetsReturnsCopy(t *testing.T) {
	got := Presets()
	if len(got) == 0 {
		t.Fatal("Presets returned empty slice")
	}
	original := got[0].Name
	got[0].Name = "mutated"
	again := Presets()
	if again[0].Name != original {
		t.Errorf("Presets backing slice was mutated: got %q, want %q", again[0].Name, original)
	}
}
