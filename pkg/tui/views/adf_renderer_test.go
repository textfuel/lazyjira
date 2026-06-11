package views

import (
	"regexp"
	"strings"
	"testing"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

func miniADF() map[string]any {
	return map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type":  "heading",
				"attrs": map[string]any{"level": 2.0},
				"content": []any{
					map[string]any{"type": "text", "text": "Heading"},
				},
			},
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": "hello "},
					map[string]any{
						"type": "text",
						"text": "world",
						"marks": []any{
							map[string]any{"type": "strong"},
						},
					},
				},
			},
			map[string]any{
				"type":  "codeBlock",
				"attrs": map[string]any{"language": "go"},
				"content": []any{
					map[string]any{"type": "text", "text": "x := 1"},
				},
			},
		},
	}
}

func TestADFRenderers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		renderer ADFRenderer
	}{
		{"builtin", BuiltinRenderer{}},
		{"glamour-notty", GlamourRenderer{Style: "notty"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			lines := tc.renderer.Render(miniADF(), 60)
			if len(lines) == 0 {
				t.Fatalf("%s renderer returned no lines", tc.name)
			}
			joined := stripANSI(strings.Join(lines, "\n"))
			for _, want := range []string{"Heading", "hello", "world", "x := 1"} {
				if !strings.Contains(joined, want) {
					t.Errorf("%s renderer output missing %q\n--- output ---\n%s",
						tc.name, want, joined)
				}
			}
		})
	}
}
