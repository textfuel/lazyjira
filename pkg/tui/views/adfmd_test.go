package views

import (
	"encoding/json"
	"strings"
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
		{
			"task list",
			`{"type":"doc","content":[{"type":"taskList","content":[{"type":"taskItem","attrs":{"state":"TODO"},"content":[{"type":"text","text":"a"}]},{"type":"taskItem","attrs":{"state":"DONE"},"content":[{"type":"text","text":"b"}]}]}]}`,
			"- [ ] a\n- [x] b",
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
		{"task list", "- [ ] item", adfTaskList},
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
		{"task list", "- [ ] a\n- [x] b"},
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

func taskItemState(t *testing.T, block any) string {
	t.Helper()
	m, ok := block.(map[string]any)
	if !ok {
		t.Fatalf("block is not a map: %T", block)
	}
	attrs, ok := m["attrs"].(map[string]any)
	if !ok {
		t.Fatal("taskItem has no attrs")
	}
	state, _ := attrs["state"].(string)
	return state
}

func TestMarkdownToADF_TaskItemState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		md   string
		want string
	}{
		{"unchecked", "- [ ] todo", "TODO"},
		{"checked lowercase", "- [x] done", "DONE"},
		{"checked uppercase", "- [X] done", "DONE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blocks := mdBlocks(t, tt.md)
			taskList, ok := blocks[0].(map[string]any)
			if !ok {
				t.Fatal("first block is not a map")
			}
			items, ok := taskList["content"].([]any)
			if !ok || len(items) == 0 {
				t.Fatal("taskList has no items")
			}
			testkit.AssertEqual(t, "state", taskItemState(t, items[0]), tt.want)
		})
	}
}

func TestMarkdownToADF_TaskList_HasLocalID(t *testing.T) {
	t.Parallel()

	blocks := mdBlocks(t, "- [ ] a\n- [x] b")
	taskList, ok := blocks[0].(map[string]any)
	if !ok {
		t.Fatal("first block is not a map")
	}
	listAttrs, ok := taskList["attrs"].(map[string]any)
	if !ok {
		t.Fatal("taskList has no attrs")
	}
	if id, _ := listAttrs["localId"].(string); id == "" {
		t.Error("taskList missing localId")
	}

	items, ok := taskList["content"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 taskItems, got %d", len(items))
	}
	seen := make(map[string]bool)
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			t.Fatal("taskItem is not a map")
		}
		itemAttrs, ok := m["attrs"].(map[string]any)
		if !ok {
			t.Fatal("taskItem has no attrs")
		}
		id, _ := itemAttrs["localId"].(string)
		if id == "" {
			t.Error("taskItem missing localId")
		}
		if seen[id] {
			t.Errorf("duplicate localId %q across taskItems", id)
		}
		seen[id] = true
	}
}

func TestMarkdownToADF_TaskItem_EmptyTextIsEmptyArray(t *testing.T) {
	t.Parallel()

	blocks := mdBlocks(t, "- [ ] first\n- [ ] ")
	items, ok := blocks[0].(map[string]any)["content"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("expected 2 taskItems, got %d", len(items))
	}
	empty, ok := items[1].(map[string]any)
	if !ok {
		t.Fatal("second taskItem is not a map")
	}
	content, ok := empty["content"].([]any)
	if !ok {
		t.Fatalf("empty taskItem content = %#v (%T), want an empty []any", empty["content"], empty["content"])
	}
	testkit.AssertEqual(t, "content length", len(content), 0)
}

func TestMarkdownToADF_TaskItem_NoTrailingSpace(t *testing.T) {
	t.Parallel()

	blocks := mdBlocks(t, "- [ ]")
	items, ok := blocks[0].(map[string]any)["content"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 taskItem, got %d", len(items))
	}
	testkit.AssertEqual(t, "state", taskItemState(t, items[0]), "TODO")
	content, ok := items[0].(map[string]any)["content"].([]any)
	if !ok {
		t.Fatalf("taskItem content = %#v (%T), want an empty []any", items[0].(map[string]any)["content"], items[0].(map[string]any)["content"])
	}
	testkit.AssertEqual(t, "content length", len(content), 0)
}

func TestMarkdownToADF_TaskItem_TrailingWhitespaceIsEmpty(t *testing.T) {
	t.Parallel()

	for _, md := range []string{"- [ ]\t", "- [ ]\r", "- [ ]  "} {
		t.Run(md, func(t *testing.T) {
			t.Parallel()
			blocks := mdBlocks(t, md)
			items, ok := blocks[0].(map[string]any)["content"].([]any)
			if !ok || len(items) != 1 {
				t.Fatalf("expected 1 taskItem, got %d", len(items))
			}
			content, _ := items[0].(map[string]any)["content"].([]any)
			testkit.AssertEqual(t, "content length", len(content), 0)
		})
	}
}

// Gates test TrimSpace'd lines (Unicode whitespace) while list parsers match
// raw lines; any disagreement makes markdownToADF loop forever or lose text.
func TestMarkdownToADF_UnicodeWhitespaceKeepsListText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, md, want string
	}{
		{"NBSP task item", "\u00a0- [ ] x", "- [ ] x"},
		{"vertical tab task item", "\v- [ ] x", "- [ ] x"},
		{"NBSP bullet", "\u00a0- x", "- x"},
		{"vertical tab ordered", "\v1. x", "1. x"},
		{"trailing NBSP checkbox", "- [ ]\u00a0", "- [ ]\u00a0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ADFToMarkdown(markdownToADF(tt.md))
			if !strings.Contains(got, tt.want) {
				t.Errorf("round trip of %q = %q, want it to contain %q", tt.md, got, tt.want)
			}
		})
	}
}

// TestMarkdownToADF_EmptyListItem_HasEmptyArrayContent guards against a
// pre-existing bug where a trailing bullet/ordered marker with no text after
// it (common when an editor auto-inserts the next marker) parses its
// paragraph content as JSON null instead of an empty array.
func TestMarkdownToADF_EmptyListItem_HasEmptyArrayContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		md   string
	}{
		{"bullet", "- a\n- "},
		{"ordered", "1. a\n2. "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			blocks := mdBlocks(t, tt.md)
			raw, err := json.Marshal(blocks[0])
			if err != nil {
				t.Fatalf("marshal block: %v", err)
			}
			if strings.Contains(string(raw), `"content":null`) {
				t.Errorf("block content marshaled to null, want an empty array: %s", raw)
			}
		})
	}
}

func TestADFToMarkdown_TaskItem_ParagraphWrappedContent(t *testing.T) {
	t.Parallel()

	adf := mustADF(t, `{"type":"doc","content":[{"type":"taskList","content":[
		{"type":"taskItem","attrs":{"state":"TODO"},"content":[
			{"type":"paragraph","content":[{"type":"text","text":"wrapped task"}]}
		]}
	]}]}`)
	testkit.AssertEqual(t, "markdown", ADFToMarkdown(adf), "- [ ] wrapped task")
}

func TestADFToMarkdown_TaskItem_MultipleAndEmptyParagraphs(t *testing.T) {
	t.Parallel()

	adf := mustADF(t, `{"type":"doc","content":[{"type":"taskList","content":[
		{"type":"taskItem","attrs":{"state":"TODO"},"content":[
			{"type":"paragraph","content":[{"type":"text","text":"one"}]},
			{"type":"paragraph"},
			{"type":"paragraph","content":[{"type":"text","text":"two"}]}
		]}
	]}]}`)
	testkit.AssertEqual(t, "markdown", ADFToMarkdown(adf), "- [ ] one two")
}

func TestMarkdownToADF_MalformedTaskLine_DoesNotCorruptSiblings(t *testing.T) {
	t.Parallel()

	blocks := mdBlocks(t, "- [ ] one\n- [x] two\n- []broken\n- [ ] four")
	if len(blocks) != 3 {
		t.Fatalf("expected 3 blocks (taskList, bulletList, taskList), got %d: %#v", len(blocks), blocks)
	}
	testkit.AssertEqual(t, "block 0 type", blockType(t, blocks[0]), adfTaskList)
	testkit.AssertEqual(t, "block 1 type", blockType(t, blocks[1]), adfBulletList)
	testkit.AssertEqual(t, "block 2 type", blockType(t, blocks[2]), adfTaskList)

	leading, _ := blocks[0].(map[string]any)["content"].([]any)
	if len(leading) != 2 {
		t.Errorf("expected 2 items in leading taskList, got %d", len(leading))
	}
	trailing, _ := blocks[2].(map[string]any)["content"].([]any)
	if len(trailing) != 1 {
		t.Errorf("expected 1 item in trailing taskList, got %d", len(trailing))
	}
}
