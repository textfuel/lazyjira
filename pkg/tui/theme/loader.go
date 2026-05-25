package theme

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// FallbackKey is the reserved key used in value-keyed maps (types,
// priorities, statuses) to set the entry consulted when a name is missing.
const FallbackKey = "_fallback"

// yamlTheme mirrors the on-disk YAML schema.
type yamlTheme struct {
	Name          string            `yaml:"name"`
	Palette       map[string]string `yaml:"palette"`
	Styles        yamlStyles        `yaml:"styles"`
	Types         map[string]string `yaml:"types"`
	Priorities    map[string]string `yaml:"priorities"`
	Statuses      map[string]string `yaml:"statuses"`
	AuthorPalette []string          `yaml:"authorPalette"`
}

type yamlStyles struct {
	Title          string `yaml:"title"`
	Subtitle       string `yaml:"subtitle"`
	HintBar        string `yaml:"hintBar"`
	Accent         string `yaml:"accent"`
	Muted          string `yaml:"muted"`
	ErrorText      string `yaml:"errorText"`
	SuccessText    string `yaml:"successText"`
	WarningText    string `yaml:"warningText"`
	SelectedItemBg string `yaml:"selectedItemBg"`
	IssueKey       string `yaml:"issueKey"`
	ActiveBorder   string `yaml:"activeBorder"`
	InactiveBorder string `yaml:"inactiveBorder"`
}

// loadYAMLTheme parses, validates, and resolves a YAML theme file into a
// fully populated *Theme.
func loadYAMLTheme(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw yamlTheme
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	return resolveYAMLTheme(&raw)
}

func resolveYAMLTheme(raw *yamlTheme) (*Theme, error) {
	if len(raw.Palette) == 0 {
		return nil, fmt.Errorf("palette: must have at least one entry")
	}

	palette := make(map[string]lipgloss.Color, len(raw.Palette))
	for name, value := range raw.Palette {
		if value == "" {
			return nil, fmt.Errorf("palette.%s: empty value", name)
		}
		// Palette entries cannot reference other palette entries (no chaining).
		palette[name] = lipgloss.Color(value)
	}

	requireStyle := func(key, value string) (lipgloss.Color, error) {
		if value == "" {
			return "", fmt.Errorf("styles.%s: required", key)
		}
		return ResolveColor(value, palette), nil
	}

	// Required style slots.
	titleC, err := requireStyle("title", raw.Styles.Title)
	if err != nil {
		return nil, err
	}
	subtitleC, err := requireStyle("subtitle", raw.Styles.Subtitle)
	if err != nil {
		return nil, err
	}
	accentC, err := requireStyle("accent", raw.Styles.Accent)
	if err != nil {
		return nil, err
	}
	mutedC, err := requireStyle("muted", raw.Styles.Muted)
	if err != nil {
		return nil, err
	}
	errorC, err := requireStyle("errorText", raw.Styles.ErrorText)
	if err != nil {
		return nil, err
	}
	successC, err := requireStyle("successText", raw.Styles.SuccessText)
	if err != nil {
		return nil, err
	}
	warningC, err := requireStyle("warningText", raw.Styles.WarningText)
	if err != nil {
		return nil, err
	}
	selBgC, err := requireStyle("selectedItemBg", raw.Styles.SelectedItemBg)
	if err != nil {
		return nil, err
	}
	issueKeyC, err := requireStyle("issueKey", raw.Styles.IssueKey)
	if err != nil {
		return nil, err
	}

	hintBarC := mutedC
	if raw.Styles.HintBar != "" {
		hintBarC = ResolveColor(raw.Styles.HintBar, palette)
	}
	activeBorderC := accentC
	if raw.Styles.ActiveBorder != "" {
		activeBorderC = ResolveColor(raw.Styles.ActiveBorder, palette)
	}
	inactiveBorderC := lipgloss.Color("-1")
	if raw.Styles.InactiveBorder != "" {
		inactiveBorderC = ResolveColor(raw.Styles.InactiveBorder, palette)
	}

	resolveMap := func(section string, in map[string]string) (map[string]lipgloss.Style, lipgloss.Style, error) {
		out := make(map[string]lipgloss.Style, len(in))
		var fallback lipgloss.Style
		fallbackSet := false
		for k, v := range in {
			if v == "" {
				return nil, fallback, fmt.Errorf("%s.%s: empty value", section, k)
			}
			c := ResolveColor(v, palette)
			s := lipgloss.NewStyle().Foreground(c)
			if k == FallbackKey {
				fallback = s
				fallbackSet = true
				continue
			}
			out[k] = s
		}
		if !fallbackSet {
			fallback = lipgloss.NewStyle().Foreground(mutedC)
		}
		return out, fallback, nil
	}

	typeColors, typeFallback, err := resolveMap("types", raw.Types)
	if err != nil {
		return nil, err
	}
	priorityColors, _, err := resolveMap("priorities", raw.Priorities)
	if err != nil {
		return nil, err
	}
	statusColors, _, err := resolveMap("statuses", raw.Statuses)
	if err != nil {
		return nil, err
	}

	authors := make([]lipgloss.Color, 0, len(raw.AuthorPalette))
	for i, v := range raw.AuthorPalette {
		if v == "" {
			return nil, fmt.Errorf("authorPalette[%d]: empty value", i)
		}
		authors = append(authors, ResolveColor(v, palette))
	}
	if len(authors) == 0 {
		authors = defaultAuthorPalette()
	}

	// Populate the ColorPalette compatibility shim with sensible mappings
	// from styles. This keeps the legacy package-level ColorX vars working
	// for code that has not yet moved to semantic accessors.
	pal := ColorPalette{
		Green:     accentC,
		Blue:      issueKeyC,
		Red:       errorC,
		Yellow:    warningC,
		Cyan:      issueKeyC,
		Magenta:   accentC,
		White:     titleC,
		Gray:      mutedC,
		Orange:    accentC,
		None:      inactiveBorderC,
		Highlight: selBgC,
	}

	th := &Theme{
		Title:          lipgloss.NewStyle().Bold(true).Foreground(titleC),
		Subtitle:       lipgloss.NewStyle().Foreground(subtitleC),
		HintBar:        lipgloss.NewStyle().Foreground(hintBarC),
		SelectedItem:   lipgloss.NewStyle().Bold(true).Background(selBgC),
		NormalItem:     lipgloss.NewStyle(),
		ActiveBorder:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(activeBorderC),
		InactiveBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(inactiveBorderC),
		ErrorText:      lipgloss.NewStyle().Foreground(errorC).Bold(true),
		SuccessText:    lipgloss.NewStyle().Foreground(successC),
		WarningText:    lipgloss.NewStyle().Foreground(warningC),
		KeyStyle:       lipgloss.NewStyle().Foreground(accentC),
		ValueStyle:     lipgloss.NewStyle(),

		Accent:   lipgloss.NewStyle().Foreground(accentC),
		Muted:    lipgloss.NewStyle().Foreground(mutedC),
		IssueKey: lipgloss.NewStyle().Foreground(issueKeyC),

		TypeColors:     typeColors,
		TypeFallback:   typeFallback,
		PriorityColors: priorityColors,
		StatusColors:   statusColors,

		Colors:        pal,
		AuthorPalette: authors,
	}
	return th, nil
}
