package views

import (
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestBlockToMarkdown_NonMapNode(t *testing.T) {
	t.Parallel()
	testkit.AssertEqual(t, "non map block", blockToMarkdown(42, 0), "")
}

func TestListItemToMarkdown_Branches(t *testing.T) {
	t.Parallel()

	t.Run("non map item", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "non map item", listItemToMarkdown("junk", 0, "- "), "")
	})

	t.Run("non map child skipped", func(t *testing.T) {
		t.Parallel()
		item := map[string]any{
			"type":    adfListItem,
			"content": []any{"junk", adfParagraphNode(adfTextNode("kept"))},
		}
		testkit.AssertEqual(t, "kept paragraph", listItemToMarkdown(item, 0, "- "), "- kept")
	})

	t.Run("second paragraph continues with indent", func(t *testing.T) {
		t.Parallel()
		item := map[string]any{
			"type": adfListItem,
			"content": []any{
				adfParagraphNode(adfTextNode("first")),
				adfParagraphNode(adfTextNode("second")),
			},
		}
		got := listItemToMarkdown(item, 0, "- ")
		testkit.AssertEqual(t, "continuation", got, "- first\n  second")
	})

	t.Run("nested list child", func(t *testing.T) {
		t.Parallel()
		item := map[string]any{
			"type": adfListItem,
			"content": []any{
				adfParagraphNode(adfTextNode("outer")),
				map[string]any{
					"type": adfBulletList,
					"content": []any{
						map[string]any{
							"type":    adfListItem,
							"content": []any{adfParagraphNode(adfTextNode("inner"))},
						},
					},
				},
			},
		}
		got := listItemToMarkdown(item, 0, "- ")
		testkit.AssertEqual(t, "nested", got, "- outer\n  - inner")
	})

	t.Run("non list non paragraph child becomes block", func(t *testing.T) {
		t.Parallel()
		item := map[string]any{
			"type": adfListItem,
			"content": []any{
				map[string]any{"type": adfCodeBlock, "content": []any{adfTextNode("code")}},
			},
		}
		got := listItemToMarkdown(item, 0, "- ")
		if !strings.Contains(got, "code") {
			t.Errorf("code child = %q, want code fence", got)
		}
	})
}

func TestInlineToMarkdown_NodeTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content []any
		want    string
	}{
		{"non map child skipped", []any{42, adfTextNode("ok")}, "ok"},
		{"emoji", []any{map[string]any{"type": adfEmoji, "attrs": map[string]any{"shortName": ":fire:"}}}, ":fire:"},
		{"hard break", []any{map[string]any{"type": adfHardBreak}}, "  \n"},
		{"inline card", []any{map[string]any{"type": adfInlineCard, "attrs": map[string]any{"url": "https://c.example"}}}, "https://c.example"},
		{
			"mention",
			[]any{map[string]any{"type": adfMention, "attrs": map[string]any{"text": "@Ann", "id": "uid-1"}}},
			"[@Ann](accountid:uid-1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "markdown", inlineToMarkdown(tt.content), tt.want)
		})
	}

	t.Run("unknown inline becomes opaque marker", func(t *testing.T) {
		t.Parallel()
		got := inlineToMarkdown([]any{map[string]any{"type": "status"}})
		if !strings.HasPrefix(got, "<!-- adf:status") {
			t.Errorf("opaque marker = %q", got)
		}
	})
}

func TestApplyMarksMD_Branches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		marks []any
		want  string
	}{
		{"non map mark ignored", []any{42}, "txt"},
		{"strike", []any{map[string]any{"type": "strike"}}, "~~txt~~"},
		{"underline", []any{map[string]any{"type": "underline"}}, "<u>txt</u>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "marked", applyMarksMD("txt", tt.marks), tt.want)
		})
	}
}

func TestTableToMarkdown_Branches(t *testing.T) {
	t.Parallel()

	t.Run("invalid rows cells and blocks skipped", func(t *testing.T) {
		t.Parallel()
		rows := []any{
			"junk row",
			map[string]any{
				"type": "tableRow",
				"content": []any{
					"junk cell",
					map[string]any{
						"type":    "tableCell",
						"content": []any{"junk block", adfParagraphNode(adfTextNode("cell text"))},
					},
				},
			},
		}
		got := tableToMarkdown(rows, 0)
		if !strings.Contains(got, "cell text") {
			t.Errorf("table = %q, want surviving cell", got)
		}
	})

	t.Run("no valid rows returns empty", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "empty table", tableToMarkdown([]any{"a"}, 0), "")
	})
}

func TestCollectPlainText_Branches(t *testing.T) {
	t.Parallel()
	content := []any{
		42,
		adfTextNode("line one"),
		map[string]any{"type": adfHardBreak},
		adfTextNode("line two"),
	}
	testkit.AssertEqual(t, "plain text", collectPlainText(content), "line one\nline two")
}

func TestOpaqueMarker_MarshalError(t *testing.T) {
	t.Parallel()
	node := map[string]any{"type": "broken", "bad": make(chan int)}
	got := opaqueMarker(node)
	testkit.AssertEqual(t, "marshal error marker", got, "<!-- adf:broken (marshal error) -->")
}

func TestMarkdownToADF_Branches(t *testing.T) {
	t.Parallel()

	contentOf := func(t *testing.T, md string) []any {
		t.Helper()
		doc, ok := markdownToADF(md).(map[string]any)
		if !ok {
			t.Fatal("markdownToADF should return a map")
		}
		blocks, _ := doc["content"].([]any)
		return blocks
	}

	blockType := func(t *testing.T, block any) string {
		t.Helper()
		m, ok := block.(map[string]any)
		if !ok {
			t.Fatalf("block is %T, want map", block)
		}
		typ, _ := m["type"].(string)
		return typ
	}

	t.Run("blank line separates paragraphs", func(t *testing.T) {
		t.Parallel()
		blocks := contentOf(t, "first\n\nsecond")
		testkit.AssertEqual(t, "block count", len(blocks), 2)
	})

	t.Run("table followed by paragraph", func(t *testing.T) {
		t.Parallel()
		blocks := contentOf(t, "| a | b |\n| --- | --- |\n| c | d |\ntrailing text")
		testkit.AssertEqual(t, "block count", len(blocks), 2)
		testkit.AssertEqual(t, "first is table", blockType(t, blocks[0]), adfTable)
		testkit.AssertEqual(t, "second is paragraph", blockType(t, blocks[1]), adfParagraph)
	})

	t.Run("paragraph stops at heading", func(t *testing.T) {
		t.Parallel()
		blocks := contentOf(t, "para text\n# Heading")
		testkit.AssertEqual(t, "block count", len(blocks), 2)
		testkit.AssertEqual(t, "first is paragraph", blockType(t, blocks[0]), adfParagraph)
		testkit.AssertEqual(t, "second is heading", blockType(t, blocks[1]), adfHeading)
	})

	t.Run("opaque marker without json falls through to paragraph", func(t *testing.T) {
		t.Parallel()
		blocks := contentOf(t, "<!-- adf:rule -->")
		testkit.AssertEqual(t, "block count", len(blocks), 1)
		testkit.AssertEqual(t, "fallback paragraph", blockType(t, blocks[0]), adfParagraph)
	})
}

func TestParseInlineWithHardBreaks_Branches(t *testing.T) {
	t.Parallel()

	t.Run("double space newline becomes hard break", func(t *testing.T) {
		t.Parallel()
		nodes := parseInlineWithHardBreaks("alpha  \nbeta")
		testkit.AssertEqual(t, "node count", len(nodes), 3)
		middle, _ := nodes[1].(map[string]any)
		testkit.AssertEqual(t, "hard break", middle["type"], adfHardBreak)
	})

	t.Run("plain newline joins with space", func(t *testing.T) {
		t.Parallel()
		nodes := parseInlineWithHardBreaks("alpha\nbeta")
		testkit.AssertEqual(t, "node count", len(nodes), 3)
		middle, _ := nodes[1].(map[string]any)
		testkit.AssertEqual(t, "joiner text", middle["text"], " ")
	})
}

func TestParseInlineSegment_Branches(t *testing.T) {
	t.Parallel()

	t.Run("empty text returns nil", func(t *testing.T) {
		t.Parallel()
		if nodes := parseInlineSegment(""); nodes != nil {
			t.Errorf("empty segment = %v, want nil", nodes)
		}
	})

	t.Run("leading plain text before match", func(t *testing.T) {
		t.Parallel()
		nodes := parseInlineSegment("intro **bold**")
		testkit.AssertEqual(t, "node count", len(nodes), 2)
		first, _ := nodes[0].(map[string]any)
		testkit.AssertEqual(t, "leading text", first["text"], "intro ")
	})

	t.Run("trailing text after match parsed recursively", func(t *testing.T) {
		t.Parallel()
		nodes := parseInlineSegment("**bold** tail")
		testkit.AssertEqual(t, "node count", len(nodes), 2)
		last, _ := nodes[1].(map[string]any)
		testkit.AssertEqual(t, "trailing text", last["text"], " tail")
	})
}

func TestParseList_Branches(t *testing.T) {
	t.Parallel()

	itemCount := func(t *testing.T, list map[string]any) int {
		t.Helper()
		items, _ := list["content"].([]any)
		return len(items)
	}

	t.Run("stops at blank line", func(t *testing.T) {
		t.Parallel()
		lines := []string{"- one", "", "- two"}
		index := 0
		list := parseList(lines, &index, "bullet")
		testkit.AssertEqual(t, "items before blank", itemCount(t, list), 1)
		testkit.AssertEqual(t, "index after parse", index, 1)
	})

	t.Run("stops at deeper non list line", func(t *testing.T) {
		t.Parallel()
		lines := []string{"- one", "    stray indented text"}
		index := 0
		list := parseList(lines, &index, "bullet")
		testkit.AssertEqual(t, "items", itemCount(t, list), 1)
	})

	t.Run("stops at non marker line", func(t *testing.T) {
		t.Parallel()
		lines := []string{"- one", "plain text"}
		index := 0
		list := parseList(lines, &index, "bullet")
		testkit.AssertEqual(t, "items", itemCount(t, list), 1)
	})

	t.Run("nested bullet list attached to item", func(t *testing.T) {
		t.Parallel()
		lines := []string{"- outer", "  - inner"}
		index := 0
		list := parseList(lines, &index, "bullet")
		items, _ := list["content"].([]any)
		testkit.AssertEqual(t, "outer items", len(items), 1)
		first, _ := items[0].(map[string]any)
		itemContent, _ := first["content"].([]any)
		testkit.AssertEqual(t, "paragraph plus nested list", len(itemContent), 2)
		nested, _ := itemContent[1].(map[string]any)
		testkit.AssertEqual(t, "nested type", nested["type"], adfBulletList)
	})

	t.Run("nested ordered list detected", func(t *testing.T) {
		t.Parallel()
		lines := []string{"- outer", "  1. inner"}
		index := 0
		list := parseList(lines, &index, "bullet")
		items, _ := list["content"].([]any)
		first, _ := items[0].(map[string]any)
		itemContent, _ := first["content"].([]any)
		nested, _ := itemContent[1].(map[string]any)
		testkit.AssertEqual(t, "nested type", nested["type"], adfOrderedList)
	})

	t.Run("dedent returns to outer list", func(t *testing.T) {
		t.Parallel()
		lines := []string{"  - inner one", "- outer"}
		index := 0
		list := parseList(lines, &index, "bullet")
		testkit.AssertEqual(t, "inner items", itemCount(t, list), 1)
		testkit.AssertEqual(t, "stopped at dedent", index, 1)
	})
}

func TestIsSeparatorRow_NonSeparator(t *testing.T) {
	t.Parallel()
	testkit.AssertEqual(t, "data row", isSeparatorRow([]string{"c", "d"}), false)
	testkit.AssertEqual(t, "separator row", isSeparatorRow([]string{"---", ":---:"}), true)
}

func TestRestoreOpaqueMarker_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
	}{
		{"missing suffix", "<!-- adf:rule"},
		{"no json payload", "<!-- adf:rule -->"},
		{"invalid json", "<!-- adf:rule {not json} -->"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if node := restoreOpaqueMarker(tt.line); node != nil {
				t.Errorf("restoreOpaqueMarker(%q) = %v, want nil", tt.line, node)
			}
		})
	}
}
