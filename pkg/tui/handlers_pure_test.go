package tui

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestFormatCustomVal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"nil is empty", nil, ""},
		{"string", "hello", "hello"},
		{"integer float", float64(8), "8"},
		{"fractional float", 3.5, "3.5"},
		{"map displayName", map[string]any{"displayName": "Ada"}, "Ada"},
		{"map value", map[string]any{"value": "High"}, "High"},
		{"map name", map[string]any{"name": "Bug"}, "Bug"},
		{"list joins", []any{"a", "b"}, "a, b"},
		{"unknown type is empty", true, ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "formatted", formatCustomVal(testCase.input), testCase.want)
		})
	}
}

func TestStripADFMedia(t *testing.T) {
	t.Parallel()

	t.Run("non document passes through", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "passthrough", stripADFMedia("plain"), any("plain"))
	})

	t.Run("removes media nodes keeps the rest", func(t *testing.T) {
		t.Parallel()
		doc := map[string]any{
			"type": "doc",
			"content": []any{
				map[string]any{"type": "paragraph"},
				map[string]any{"type": "mediaSingle"},
				map[string]any{"type": "media"},
				map[string]any{"type": "mediaGroup"},
			},
		}

		result, ok := stripADFMedia(doc).(map[string]any)
		if !ok {
			t.Fatalf("result is not a map: %#v", stripADFMedia(doc))
		}
		content, ok := result["content"].([]any)
		if !ok || len(content) != 1 {
			t.Fatalf("content = %#v, want one paragraph", result["content"])
		}
		node, _ := content[0].(map[string]any)
		testkit.AssertEqual(t, "kept node type", node["type"], any("paragraph"))
	})
}
