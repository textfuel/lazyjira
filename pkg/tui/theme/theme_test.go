package theme

import (
	"regexp"
	"slices"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestSetThemeDefault(t *testing.T) {
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
	if err := SetTheme(""); err != nil {
		t.Fatalf("SetTheme(''): %v", err)
	}
	if ColorGreen != lipgloss.Color("2") {
		t.Errorf("ColorGreen = %q, want %q", ColorGreen, "2")
	}
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
	_ = AuthorStyle("Alice")
	if len(authorCache) == 0 {
		t.Fatal("author cache should have an entry")
	}

	_ = SetTheme("catppuccin-mocha")
	if len(authorCache) != 0 {
		t.Error("author cache should be empty after theme switch")
	}
	_ = SetTheme("default")
}

const authorFixtureName = "Ada Lovelace"

var ansiEscapeSequences = regexp.MustCompile("\x1b\\[[0-9;]*m")

func forceColorProfile(t *testing.T) {
	t.Helper()
	originalProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() { lipgloss.SetColorProfile(originalProfile) })
}

func TestDefaultTheme_ReturnsSingleton(t *testing.T) {
	_ = SetTheme("default")
	if DefaultTheme() != Default {
		t.Error("DefaultTheme() should return the Default singleton")
	}
}

func TestPriorityStyled_RoutesNamesToPriorityStyles(t *testing.T) {
	_ = SetTheme("default")
	forceColorProfile(t)

	if Default.PriorityHigh.Render("x") == Default.PriorityLow.Render("x") {
		t.Fatal("color profile did not produce distinct priority styles")
	}

	tests := []struct {
		name      string
		priority  string
		wantStyle lipgloss.Style
	}{
		{"highest", "Highest", Default.PriorityHigh},
		{"high lowercase", "high", Default.PriorityHigh},
		{"critical", "Critical", Default.PriorityHigh},
		{"blocker uppercase", "BLOCKER", Default.PriorityHigh},
		{"medium", "Medium", Default.PriorityMedium},
		{"low routes to low style", "Low", Default.PriorityLow},
		{"lowest routes to low style", "Lowest", Default.PriorityLow},
		{"unknown routes to low style", "Whatever", Default.PriorityLow},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.wantStyle.Render(tc.priority)
			if got := PriorityStyled(tc.priority); got != want {
				t.Errorf("PriorityStyled(%q) = %q, want %q", tc.priority, got, want)
			}
		})
	}
}

func TestStatusColor_MapsCategoryToThemeColor(t *testing.T) {
	_ = SetTheme("default")

	tests := []struct {
		name        string
		categoryKey string
		want        lipgloss.Color
	}{
		{"done is green", "done", lipgloss.Color("2")},
		{"indeterminate is yellow", "indeterminate", lipgloss.Color("3")},
		{"new is blue", "new", lipgloss.Color("4")},
		{"unknown is gray", "undefined", lipgloss.Color("8")},
		{"empty is gray", "", lipgloss.Color("8")},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := StatusColor(tc.categoryKey).GetForeground()
			if got != tc.want {
				t.Errorf("StatusColor(%q) foreground = %v, want %v", tc.categoryKey, got, tc.want)
			}
		})
	}
}

func TestAuthorStyle_DeterministicAndNormalized(t *testing.T) {
	_ = SetTheme("default")

	plain := AuthorStyle(authorFixtureName)
	prefixed := AuthorStyle("@" + authorFixtureName + " ")
	repeated := AuthorStyle(authorFixtureName)

	if plain.GetForeground() != prefixed.GetForeground() {
		t.Error("@-prefixed and plain names should share one color")
	}
	if plain.GetForeground() != repeated.GetForeground() {
		t.Error("repeated lookups should return the cached color")
	}

	foreground, ok := plain.GetForeground().(lipgloss.Color)
	if !ok {
		t.Fatalf("foreground = %T, want lipgloss.Color", plain.GetForeground())
	}
	if !slices.Contains(Default.AuthorPalette, foreground) {
		t.Errorf("foreground %v is not part of the author palette", foreground)
	}
}

func TestAuthorRender_AppliesAuthorColorToName(t *testing.T) {
	_ = SetTheme("default")
	forceColorProfile(t)

	rendered := AuthorRender(authorFixtureName)
	want := AuthorStyle(authorFixtureName).Render(authorFixtureName)
	if rendered != want {
		t.Errorf("AuthorRender = %q, want %q", rendered, want)
	}
	if plainText := ansiEscapeSequences.ReplaceAllString(rendered, ""); plainText != authorFixtureName {
		t.Errorf("plain text = %q, want %q", plainText, authorFixtureName)
	}
	if rendered == authorFixtureName {
		t.Error("rendered name should carry color escape codes under a color profile")
	}
}
