package views

import (
	"encoding/json"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func mustADF(t *testing.T, jsonStr string) any {
	t.Helper()
	var doc any
	if err := json.Unmarshal([]byte(jsonStr), &doc); err != nil {
		t.Fatalf("unmarshal adf: %v", err)
	}
	return doc
}

func TestADFToMarkdown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		adf  string
		want string
	}{
		{
			"paragraph",
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"hello"}]}]}`,
			"hello",
		},
		{
			"heading level two",
			`{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Title"}]}]}`,
			"## Title",
		},
		{
			"strong mark",
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"bold","marks":[{"type":"strong"}]}]}]}`,
			"**bold**",
		},
		{
			"link mark",
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"site","marks":[{"type":"link","attrs":{"href":"http://x"}}]}]}]}`,
			"[site](http://x)",
		},
		{
			"mention",
			`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"mention","attrs":{"text":"@Ann","id":"123"}}]}]}`,
			"[@Ann](accountid:123)",
		},
		{
			"bullet list",
			`{"type":"doc","content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"a"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"b"}]}]}]}]}`,
			"- a\n- b",
		},
		{
			"ordered list",
			`{"type":"doc","content":[{"type":"orderedList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"a"}]}]},{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"b"}]}]}]}]}`,
			"1. a\n2. b",
		},
		{
			"code block",
			`{"type":"doc","content":[{"type":"codeBlock","attrs":{"language":"go"},"content":[{"type":"text","text":"x := 1"}]}]}`,
			"```go\nx := 1\n```",
		},
		{
			"blockquote",
			`{"type":"doc","content":[{"type":"blockquote","content":[{"type":"paragraph","content":[{"type":"text","text":"quote"}]}]}]}`,
			"> quote",
		},
		{
			"rule",
			`{"type":"doc","content":[{"type":"rule"}]}`,
			"---",
		},
		{
			"table",
			`{"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"A"}]}]},{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"B"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"1"}]}]},{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"2"}]}]}]}]}]}`,
			"| A | B |\n| --- | --- |\n| 1 | 2 |",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "markdown", ADFToMarkdown(mustADF(t, tt.adf)), tt.want)
		})
	}
}

func TestADFToMarkdown_NonDocumentIsEmpty(t *testing.T) {
	t.Parallel()
	testkit.AssertEqual(t, "plain string", ADFToMarkdown("not a doc"), "")
	testkit.AssertEqual(t, "no content key", ADFToMarkdown(map[string]any{"type": "doc"}), "")
}

func mdBlocks(t *testing.T, md string) []any {
	t.Helper()
	doc, ok := MarkdownToADF(md).(map[string]any)
	if !ok {
		t.Fatalf("MarkdownToADF did not return a document: %T", MarkdownToADF(md))
	}
	content, ok := doc["content"].([]any)
	if !ok {
		t.Fatalf("document has no content array")
	}
	return content
}

func blockType(t *testing.T, block any) string {
	t.Helper()
	m, ok := block.(map[string]any)
	if !ok {
		t.Fatalf("block is not a map: %T", block)
	}
	typ, _ := m["type"].(string)
	return typ
}

func TestMarkdownToADF_BlockTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		md   string
		want string
	}{
		{"heading", "# Title", adfHeading},
		{"bullet list", "- item", adfBulletList},
		{"ordered list", "1. item", adfOrderedList},
		{"code block", "```go\ncode\n```", adfCodeBlock},
		{"blockquote", "> quote", adfBlockquote},
		{"rule", "---", adfRule},
		{"paragraph", "just text", adfParagraph},
		{"table", "| A | B |\n| --- | --- |\n| 1 | 2 |", adfTable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blocks := mdBlocks(t, tt.md)
			if len(blocks) == 0 {
				t.Fatal("expected at least one block")
			}
			testkit.AssertEqual(t, "block type", blockType(t, blocks[0]), tt.want)
		})
	}
}

func TestMarkdownToADF_HeadingLevel(t *testing.T) {
	t.Parallel()
	blocks := mdBlocks(t, "### Deep")
	heading, ok := blocks[0].(map[string]any)
	if !ok {
		t.Fatal("first block is not a map")
	}
	attrs, ok := heading["attrs"].(map[string]any)
	if !ok {
		t.Fatal("heading has no attrs")
	}
	level, ok := attrs["level"].(float64)
	if !ok {
		t.Fatalf("level is not numeric: %T", attrs["level"])
	}
	testkit.AssertEqual(t, "level", level, float64(3))
}

func TestMarkdownADFRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		md   string
	}{
		{"heading", "## Title"},
		{"bold", "**bold**"},
		{"italic", "*slanted*"},
		{"code span", "`snippet`"},
		{"link", "[site](http://x)"},
		{"mention", "[@Ann](accountid:123)"},
		{"bullet list", "- a\n- b"},
		{"ordered list", "1. a\n2. b"},
		{"rule", "---"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			roundTripped := ADFToMarkdown(MarkdownToADF(tt.md))
			testkit.AssertEqual(t, "round trip", roundTripped, tt.md)
		})
	}
}

func TestMarkdownToADF_OpaqueMarkerRestored(t *testing.T) {
	t.Parallel()
	original := mustADF(t, `{"type":"doc","content":[{"type":"panel","attrs":{"panelType":"info"}}]}`)

	markdown := ADFToMarkdown(original)
	restored := mdBlocks(t, markdown)

	if len(restored) == 0 {
		t.Fatal("expected the opaque node to survive the round trip")
	}
	testkit.AssertEqual(t, "restored type", blockType(t, restored[0]), "panel")
}
