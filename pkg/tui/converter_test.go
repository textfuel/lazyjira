package tui

import (
	"testing"

	"github.com/seflue/adf-converter/placeholder"
)

// TestBuiltinConverter_Roundtrip pins the adapter contract: ToMarkdown/FromMarkdown
// produce non-empty markdown and ADF for ordinary input, and state is nil since
// the builtin converter does not use sessions.
func TestBuiltinConverter_Roundtrip(t *testing.T) {
	t.Parallel()
	c := BuiltinConverter{}
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

	md, state, err := c.ToMarkdown(adf)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	if state != nil {
		t.Errorf("BuiltinConverter.ToMarkdown should return nil state, got %v", state)
	}
	if md == "" {
		t.Error("ToMarkdown returned empty markdown for non-empty ADF")
	}

	back, err := c.FromMarkdown(md, nil)
	if err != nil {
		t.Fatalf("FromMarkdown returned error: %v", err)
	}
	if back == nil {
		t.Error("FromMarkdown returned nil ADF for non-empty markdown")
	}
}

// TestAdfConvConverter_FromMarkdown_GreenfieldBootstrap pins the create-form
// preview path: when no prior ADF exists (state is nil), the wrapper must
// bootstrap a fresh session instead of forwarding nil into the library.
// Regression test for the silent fallback that hid AdfConvConverter being
// unused on the preview path entirely.
func TestAdfConvConverter_FromMarkdown_GreenfieldBootstrap(t *testing.T) {
	t.Parallel()
	c := AdfConvConverter{}

	doc, err := c.FromMarkdown("# Heading\n\nparagraph", nil)
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

// TestAdfConvConverter_FromMarkdown_EmptyInput covers the boundary case where
// the user has not yet typed anything in the create form. Must not error.
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

// TestAdfConvConverter_RoundtripWithSession pins the edit-roundtrip contract:
// ToMarkdown produces a session, FromMarkdown consumes it. Asserts that the
// any-typed state opaque blob passes through unmodified.
func TestAdfConvConverter_RoundtripWithSession(t *testing.T) {
	t.Parallel()
	c := AdfConvConverter{}
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

	md, state, err := c.ToMarkdown(adf)
	if err != nil {
		t.Fatalf("ToMarkdown: %v", err)
	}
	if _, ok := state.(*placeholder.EditSession); !ok {
		t.Fatalf("ToMarkdown should return *placeholder.EditSession, got %T", state)
	}

	back, err := c.FromMarkdown(md, state)
	if err != nil {
		t.Fatalf("FromMarkdown with session: %v", err)
	}
	if back == nil {
		t.Error("FromMarkdown returned nil ADF on roundtrip")
	}
}
