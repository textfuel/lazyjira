package config

// KeybindingsConfig holds user-customizable key bindings grouped by context.
type KeybindingsConfig struct {
	Global map[string]string `yaml:"global"`
	Issues map[string]string `yaml:"issues"`
	Detail map[string]string `yaml:"detail"`
}

// DefaultKeymap returns the default keybinding configuration.
func DefaultKeymap() KeybindingsConfig {
	return KeybindingsConfig{
		Global: map[string]string{
			"quit":          "q",
			"help":          "?",
			"search":        "/",
			"command":       ":",
			"refresh":       "r",
			"copyKey":       "y",
			"switchProject": "P",
			"focusToggle":   "Tab",
			"up":            "k",
			"down":          "j",
			"left":          "h",
			"right":         "l",
			"top":           "g",
			"bottom":        "G",
			"halfPageDown":  "ctrl+d",
			"halfPageUp":    "ctrl+u",
		},
		Issues: map[string]string{
			"open":       "enter",
			"transition": "t",
			"assign":     "a",
			"comment":    "c",
			"edit":       "e",
			"priority":   "p",
			"labels":     "l",
			"sprint":     "s",
			"watch":      "w",
			"vote":       "v",
			"new":        "n",
			"filter":     "f",
			"select":     "space",
			"bulk":       "B",
			"browser":    "o",
		},
		Detail: map[string]string{
			"transition": "t",
			"comment":    "c",
			"edit":       "e",
			"browser":    "o",
			"back":       "esc",
			"links":      "L",
		},
	}
}
