package views

import (
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/tui/theme"
)

func adfDoc(blocks ...any) map[string]any {
	return map[string]any{"type": "doc", "version": 1, "content": blocks}
}

func adfParagraphNode(inline ...any) map[string]any {
	return map[string]any{"type": adfParagraph, "content": inline}
}

func adfTextNode(text string) map[string]any {
	return map[string]any{"type": adfText, "text": text}
}

func renderBuiltinPlain(t *testing.T, doc map[string]any, width int) string {
	t.Helper()
	lines := BuiltinRenderer{}.Render(doc, width)
	return stripANSI(strings.Join(lines, "\n"))
}

func TestRenderADFBuiltin_InvalidInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node any
	}{
		{"non map node", "plain string"},
		{"doc without content array", map[string]any{"type": "doc", "content": "oops"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if lines := (BuiltinRenderer{}).Render(tt.node, 60); lines != nil {
				t.Errorf("Render(%v) = %v, want nil", tt.node, lines)
			}
		})
	}
}

func TestRenderADFGlamour_EdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("marshal error returns marker line", func(t *testing.T) {
		t.Parallel()
		lines := GlamourRenderer{Style: "notty"}.Render(make(chan int), 60)
		if len(lines) != 1 || !strings.Contains(lines[0], "glamour: marshal") {
			t.Errorf("marshal error output = %v", lines)
		}
	})

	t.Run("unmarshal error returns marker line", func(t *testing.T) {
		t.Parallel()
		node := map[string]any{"type": "doc", "content": 42}
		lines := GlamourRenderer{Style: "notty"}.Render(node, 60)
		if len(lines) != 1 || !strings.Contains(lines[0], "glamour: unmarshal") {
			t.Errorf("unmarshal error output = %v", lines)
		}
	})

	t.Run("narrow width clamped to minimum", func(t *testing.T) {
		t.Parallel()
		lines := GlamourRenderer{Style: "notty"}.Render(miniADF(), 3)
		if len(lines) == 0 {
			t.Error("narrow width should still render")
		}
	})

	t.Run("empty document returns nil", func(t *testing.T) {
		t.Parallel()
		if lines := (GlamourRenderer{Style: "notty"}).Render(adfDoc(), 60); lines != nil {
			t.Errorf("empty doc = %v, want nil", lines)
		}
	})
}

func TestRenderBlock_Branches(t *testing.T) {
	t.Parallel()

	t.Run("non map block skipped", func(t *testing.T) {
		t.Parallel()
		if got := renderBuiltinPlain(t, adfDoc("not a block"), 60); got != "" {
			t.Errorf("non map block = %q, want empty", got)
		}
	})

	t.Run("empty paragraph yields blank line", func(t *testing.T) {
		t.Parallel()
		lines := BuiltinRenderer{}.Render(adfDoc(adfParagraphNode()), 60)
		testkit.AssertSliceEqual(t, "blank line", lines, []string{""})
	})

	t.Run("code block without language has no language header", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type":    adfCodeBlock,
			"content": []any{adfTextNode("plain code")},
		})
		got := renderBuiltinPlain(t, doc, 60)
		if !strings.Contains(got, "plain code") {
			t.Errorf("code block = %q, want code text", got)
		}
		if strings.Contains(got, "┌") || strings.Contains(got, "└") {
			t.Errorf("code block without lang should have no header/footer, got %q", got)
		}
	})

	t.Run("code block hard wraps long lines", func(t *testing.T) {
		t.Parallel()
		longLine := strings.Repeat("x", 100)
		doc := adfDoc(map[string]any{
			"type":    adfCodeBlock,
			"content": []any{adfTextNode(longLine)},
		})
		lines := BuiltinRenderer{}.Render(doc, 40)
		if len(lines) < 2 {
			t.Errorf("long code line should wrap, got %d lines", len(lines))
		}
	})

	t.Run("blockquote prefixes content with bar", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type":    adfBlockquote,
			"content": []any{adfParagraphNode(adfTextNode("quoted words"))},
		})
		got := renderBuiltinPlain(t, doc, 60)
		if !strings.Contains(got, "│ quoted words") {
			t.Errorf("blockquote = %q, want bar prefix", got)
		}
	})

	t.Run("rule renders horizontal line", func(t *testing.T) {
		t.Parallel()
		got := renderBuiltinPlain(t, adfDoc(map[string]any{"type": adfRule}), 60)
		if !strings.Contains(got, "────") {
			t.Errorf("rule = %q, want dashes", got)
		}
	})

	t.Run("media single renders placeholder", func(t *testing.T) {
		t.Parallel()
		got := renderBuiltinPlain(t, adfDoc(map[string]any{"type": "mediaSingle"}), 60)
		if !strings.Contains(got, "[media]") {
			t.Errorf("mediaSingle = %q, want [media]", got)
		}
	})

	t.Run("unknown block type recurses into content", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type":    "panel",
			"content": []any{adfParagraphNode(adfTextNode("panel body"))},
		})
		got := renderBuiltinPlain(t, doc, 60)
		if !strings.Contains(got, "panel body") {
			t.Errorf("panel = %q, want inner paragraph", got)
		}
	})

	t.Run("long paragraph wraps to continuation lines", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(adfParagraphNode(adfTextNode(strings.Repeat("word ", 30))))
		lines := BuiltinRenderer{}.Render(doc, 30)
		if len(lines) < 2 {
			t.Errorf("long paragraph should wrap, got %d lines", len(lines))
		}
	})
}

func TestRenderListItem_Branches(t *testing.T) {
	t.Parallel()

	t.Run("non map item skipped", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{"type": adfBulletList, "content": []any{"junk"}})
		if got := renderBuiltinPlain(t, doc, 60); got != "" {
			t.Errorf("non map item = %q, want empty", got)
		}
	})

	t.Run("non map child skipped", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type": adfBulletList,
			"content": []any{
				map[string]any{"type": adfListItem, "content": []any{"junk", adfParagraphNode(adfTextNode("kept"))}},
			},
		})
		got := renderBuiltinPlain(t, doc, 60)
		if !strings.Contains(got, "kept") {
			t.Errorf("list item = %q, want kept paragraph", got)
		}
	})

	t.Run("second paragraph indented without marker", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type": adfBulletList,
			"content": []any{
				map[string]any{
					"type": adfListItem,
					"content": []any{
						adfParagraphNode(adfTextNode("first para")),
						adfParagraphNode(adfTextNode("second para")),
					},
				},
			},
		})
		lines := BuiltinRenderer{}.Render(doc, 60)
		joined := stripANSI(strings.Join(lines, "\n"))
		if !strings.Contains(joined, "• first para") {
			t.Errorf("list = %q, want marker on first para", joined)
		}
		if !strings.Contains(joined, "  second para") {
			t.Errorf("list = %q, want indented second para", joined)
		}
	})

	t.Run("non paragraph child rendered as block", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type": adfBulletList,
			"content": []any{
				map[string]any{
					"type": adfListItem,
					"content": []any{
						adfParagraphNode(adfTextNode("intro")),
						map[string]any{"type": adfCodeBlock, "content": []any{adfTextNode("code inside")}},
					},
				},
			},
		})
		got := renderBuiltinPlain(t, doc, 60)
		if !strings.Contains(got, "code inside") {
			t.Errorf("list = %q, want nested code block", got)
		}
	})
}

func TestRenderInline_NodeTypes(t *testing.T) {
	t.Parallel()
	renderer := &adfRenderer{width: 60}

	tests := []struct {
		name string
		node any
		want string
	}{
		{"non map", 42, ""},
		{"mention with text", map[string]any{"type": adfMention, "attrs": map[string]any{"text": "@Bob"}}, "\x00MENTION:@Bob\x00"},
		{"mention without attrs", map[string]any{"type": adfMention}, ""},
		{"emoji", map[string]any{"type": adfEmoji, "attrs": map[string]any{"shortName": ":smile:"}}, ":smile:"},
		{"emoji without attrs", map[string]any{"type": adfEmoji}, ""},
		{"hard break", map[string]any{"type": adfHardBreak}, "\n"},
		{"unknown type", map[string]any{"type": "mystery"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "inline", renderer.renderInline(tt.node), tt.want)
		})
	}

	t.Run("inline card renders styled url", func(t *testing.T) {
		t.Parallel()
		node := map[string]any{"type": adfInlineCard, "attrs": map[string]any{"url": "https://card.example.com"}}
		got := stripANSI(renderer.renderInline(node))
		testkit.AssertEqual(t, "inline card", got, "https://card.example.com")
	})
}

func TestRenderInlinePlain_NodeTypes(t *testing.T) {
	t.Parallel()
	renderer := &adfRenderer{width: 60}

	tests := []struct {
		name string
		node any
		want string
	}{
		{"non map", 42, ""},
		{"text with carriage return", map[string]any{"type": adfText, "text": "a\rb"}, "ab"},
		{"mention", map[string]any{"type": adfMention, "attrs": map[string]any{"text": "@Ann"}}, "@Ann"},
		{"mention without attrs", map[string]any{"type": adfMention}, ""},
		{"emoji", map[string]any{"type": adfEmoji, "attrs": map[string]any{"shortName": ":tada:"}}, ":tada:"},
		{"hard break", map[string]any{"type": adfHardBreak}, "\n"},
		{"inline card", map[string]any{"type": adfInlineCard, "attrs": map[string]any{"url": "https://x.example"}}, "https://x.example"},
		{"unknown type", map[string]any{"type": "mystery"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "inline plain", renderer.renderInlinePlain(tt.node), tt.want)
		})
	}
}

func TestApplyMarks_AllMarkTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mark any
	}{
		{"strong", map[string]any{"type": "strong"}},
		{"em", map[string]any{"type": "em"}},
		{"code", map[string]any{"type": "code"}},
		{"underline", map[string]any{"type": "underline"}},
		{"strike", map[string]any{"type": "strike"}},
		{"link", map[string]any{"type": "link"}},
		{"text color", map[string]any{"type": "textColor", "attrs": map[string]any{"color": "#ff0000"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := applyMarks("marked", []any{tt.mark})
			testkit.AssertEqual(t, "plain text preserved", stripANSI(got), "marked")
		})
	}

	t.Run("non map mark ignored", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "unchanged", applyMarks("plain", []any{42}), "plain")
	})

	t.Run("unknown mark type ignored", func(t *testing.T) {
		t.Parallel()
		got := applyMarks("plain", []any{map[string]any{"type": "subsup"}})
		testkit.AssertEqual(t, "unchanged", got, "plain")
	})
}

func TestHeadingStyle_Levels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level int
		want  lipgloss.TerminalColor
	}{
		{1, theme.ColorGreen},
		{2, theme.ColorGreen},
		{3, theme.ColorWhite},
		{4, theme.ColorWhite},
		{5, theme.ColorGray},
		{6, theme.ColorGray},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.level), func(t *testing.T) {
			t.Parallel()
			style := headingStyle(tt.level)
			testkit.AssertEqual(t, "bold", style.GetBold(), true)
			testkit.AssertEqual(t, "foreground", style.GetForeground(), tt.want)
		})
	}
}

func TestRenderTable_Branches(t *testing.T) {
	t.Parallel()

	t.Run("non map rows and cells skipped", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{
			"type": adfTable,
			"content": []any{
				"junk row",
				map[string]any{
					"type": "tableRow",
					"content": []any{
						"junk cell",
						map[string]any{
							"type":    "tableCell",
							"content": []any{"junk block", adfParagraphNode(adfTextNode("good cell"))},
						},
					},
				},
			},
		})
		got := renderBuiltinPlain(t, doc, 60)
		if !strings.Contains(got, "good cell") {
			t.Errorf("table = %q, want surviving cell", got)
		}
	})

	t.Run("all rows invalid produces nothing", func(t *testing.T) {
		t.Parallel()
		doc := adfDoc(map[string]any{"type": adfTable, "content": []any{"a", "b"}})
		if got := renderBuiltinPlain(t, doc, 60); got != "" {
			t.Errorf("invalid table = %q, want empty", got)
		}
	})

	t.Run("wide cells truncated to column cap", func(t *testing.T) {
		t.Parallel()
		adf := makeTableADF([][]string{
			{"Header A", "Header B"},
			{strings.Repeat("verylongcontent", 5), strings.Repeat("anotherlongone", 5)},
		})
		lines := BuiltinRenderer{}.Render(adf, 30)
		for _, line := range lines {
			if w := lipgloss.Width(stripANSI(line)); w > 40 {
				t.Errorf("table line width %d too wide: %q", w, stripANSI(line))
			}
		}
	})
}

func TestHardWrapLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		line  string
		width int
		want  []string
	}{
		{"non positive width returns as is", "abcdef", 0, []string{"abcdef"}},
		{"short line untouched", "abc", 10, []string{"abc"}},
		{"exact width untouched", "abcde", 5, []string{"abcde"}},
		{"long line split into chunks", "abcdefgh", 3, []string{"abc", "def", "gh"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertSliceEqual(t, "wrapped", hardWrapLine(tt.line, tt.width), tt.want)
		})
	}
}

func TestHighlightCode(t *testing.T) {
	t.Parallel()

	t.Run("empty language returns code unchanged", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "unchanged", highlightCode("x := 1", ""), "x := 1")
	})

	t.Run("unknown language returns code unchanged", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "unchanged", highlightCode("x := 1", "nosuchlanguage"), "x := 1")
	})

	t.Run("known language keeps text content", func(t *testing.T) {
		t.Parallel()
		got := highlightCode("x := 1", "go")
		testkit.AssertEqual(t, "content preserved", stripANSI(got), "x := 1")
	})
}

func TestExtractADFURLs_InputShapes(t *testing.T) {
	t.Parallel()

	t.Run("slice input walks children", func(t *testing.T) {
		t.Parallel()
		nodes := []any{
			map[string]any{"type": adfInlineCard, "attrs": map[string]any{"url": "https://one.example"}},
			map[string]any{"type": adfInlineCard, "attrs": map[string]any{"url": "https://two.example"}},
		}
		urls := extractADFURLs(nodes)
		testkit.AssertSliceEqual(t, "urls", urls, []string{"https://one.example", "https://two.example"})
	})

	t.Run("scalar input returns nil", func(t *testing.T) {
		t.Parallel()
		if urls := extractADFURLs("just text"); urls != nil {
			t.Errorf("scalar input = %v, want nil", urls)
		}
	})
}
