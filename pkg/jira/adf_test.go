package jira

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func adfTextNode(text string) map[string]any {
	return map[string]any{"type": "text", "text": text}
}

func adfParagraph(children ...any) map[string]any {
	return map[string]any{"type": "paragraph", "content": children}
}

func adfListItem(children ...any) map[string]any {
	return map[string]any{"type": "listItem", "content": children}
}

func TestExtractADFText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node any
		want string
	}{
		{
			name: "plain string passes through",
			node: "already plain",
			want: "already plain",
		},
		{
			name: "unsupported scalar yields empty",
			node: 42.0,
			want: "",
		},
		{
			name: "text node",
			node: adfTextNode("hello"),
			want: "hello",
		},
		{
			name: "paragraph appends newline",
			node: adfParagraph(adfTextNode("alpha")),
			want: "alpha\n",
		},
		{
			name: "doc joins children without separator",
			node: map[string]any{"type": "doc", "content": []any{adfParagraph(adfTextNode("alpha")), adfParagraph(adfTextNode("beta"))}},
			want: "alpha\nbeta\n",
		},
		{
			name: "top level slice concatenates",
			node: []any{adfTextNode("alpha"), adfTextNode("beta")},
			want: "alphabeta",
		},
		{
			name: "mention wraps text in markers",
			node: map[string]any{"type": "mention", "attrs": map[string]any{"text": "@ada"}},
			want: "\x00MENTION:@ada\x00",
		},
		{
			name: "mention without attrs yields empty",
			node: map[string]any{"type": "mention"},
			want: "",
		},
		{
			name: "mention with non string text yields empty",
			node: map[string]any{"type": "mention", "attrs": map[string]any{"text": 7.0}},
			want: "",
		},
		{
			name: "emoji uses short name",
			node: map[string]any{"type": "emoji", "attrs": map[string]any{"shortName": ":tada:"}},
			want: ":tada:",
		},
		{
			name: "hard break becomes newline",
			node: adfParagraph(adfTextNode("first"), map[string]any{"type": "hardBreak"}, adfTextNode("second")),
			want: "first\nsecond\n",
		},
		{
			name: "inline card uses url",
			node: map[string]any{"type": "inlineCard", "attrs": map[string]any{"url": "https://example.com/page"}},
			want: "https://example.com/page",
		},
		{
			name: "list item gets bullet",
			node: adfListItem(adfParagraph(adfTextNode("buy milk"))),
			want: "• buy milk\n",
		},
		{
			name: "bullet list joins items and appends newline",
			node: map[string]any{"type": "bulletList", "content": []any{adfListItem(adfParagraph(adfTextNode("one"))), adfListItem(adfParagraph(adfTextNode("two")))}},
			want: "• one\n• two\n\n",
		},
		{
			name: "code block appends newline",
			node: map[string]any{"type": "codeBlock", "content": []any{adfTextNode("x := 1")}},
			want: "x := 1\n",
		},
		{
			name: "blockquote appends newline",
			node: map[string]any{"type": "blockquote", "content": []any{adfParagraph(adfTextNode("quoted"))}},
			want: "quoted\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "extracted text", extractADFText(tt.node), tt.want)
		})
	}
}
