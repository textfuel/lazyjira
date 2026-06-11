package tui

import (
	"testing"

	"github.com/seflue/adf-converter/placeholder"
)

func TestBuiltinConverter_Roundtrip(t *testing.T) {
	t.Parallel()
	converter := BuiltinConverter{}
	adf := map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": "Hello world"},
				},
			},
		},
	}

	md, state, err := converter.ToMarkdown(adf)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	if state != nil {
		t.Errorf("BuiltinConverter.ToMarkdown should return nil state, got %v", state)
	}
	if md == "" {
		t.Error("ToMarkdown returned empty markdown for non-empty ADF")
	}

	back, err := converter.FromMarkdown(md, nil)
	if err != nil {
		t.Fatalf("FromMarkdown returned error: %v", err)
	}
	if back == nil {
		t.Error("FromMarkdown returned nil ADF for non-empty markdown")
	}
}

func TestAdfConvConverter_FromMarkdown_GreenfieldBootstrap(t *testing.T) {
	t.Parallel()
	converter := AdfConvConverter{}

	doc, err := converter.FromMarkdown("# Heading\n\nparagraph", nil)
	if err != nil {
		t.Fatalf("nil state should bootstrap an empty session, got error: %v", err)
	}
	if doc == nil {
		t.Fatal("FromMarkdown returned nil ADF for non-empty markdown")
	}

	m, ok := doc.(map[string]any)
	if !ok {
		t.Fatalf("FromMarkdown should return map[string]any, got %T", doc)
	}
	if m["type"] != "doc" {
		t.Errorf("expected doc root, got type=%v", m["type"])
	}
}

func TestAdfConvConverter_FromMarkdown_EmptyInput(t *testing.T) {
	t.Parallel()
	doc, err := AdfConvConverter{}.FromMarkdown("", nil)
	if err != nil {
		t.Fatalf("empty markdown should not error, got: %v", err)
	}
	if doc == nil {
		t.Fatal("empty markdown should still return an ADF document (empty doc)")
	}
}

func TestAdfConvConverter_RoundtripWithSession(t *testing.T) {
	t.Parallel()
	converter := AdfConvConverter{}
	adf := map[string]any{
		"type":    "doc",
		"version": 1,
		"content": []any{
			map[string]any{
				"type": "paragraph",
				"content": []any{
					map[string]any{"type": "text", "text": "edit me"},
				},
			},
		},
	}

	md, state, err := converter.ToMarkdown(adf)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	if _, ok := state.(*placeholder.EditSession); !ok {
		t.Fatalf("ToMarkdown should return *placeholder.EditSession, got %T", state)
	}

	back, err := converter.FromMarkdown(md, state)
	if err != nil {
		t.Fatalf("FromMarkdown with session: %v", err)
	}
	if back == nil {
		t.Error("FromMarkdown returned nil ADF on roundtrip")
	}
}
