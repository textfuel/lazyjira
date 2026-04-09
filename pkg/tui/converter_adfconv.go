package tui

import (
	"encoding/json"
	"fmt"

	"github.com/seflue/adf-converter/adf/adftypes"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/placeholder"
)

// AdfConvConverter wraps the adf-converter library for ADF<->Markdown conversion
// with placeholder-based preservation of complex elements.
// It is stateless — a fresh converter is created per call, session state is
// passed through the opaque state parameter.
type AdfConvConverter struct{}

func (AdfConvConverter) ToMarkdown(adf any) (string, any, error) {
	doc, err := toADFDocument(adf)
	if err != nil {
		return "", nil, fmt.Errorf("ADF parse: %w", err)
	}
	c := defaults.NewDefaultConverter()
	md, session, err := c.ToMarkdown(doc)
	if err != nil {
		return "", nil, fmt.Errorf("ADF->MD: %w", err)
	}
	return md, session, nil
}

func (AdfConvConverter) FromMarkdown(md string, state any) (any, error) {
	session, ok := state.(*placeholder.EditSession)
	if !ok || session == nil {
		// Greenfield paths (e.g. create-form preview) have no prior ADF and
		// thus no session. adf-converter's FromMarkdown requires a non-nil
		// session, so bootstrap an empty one — no placeholders to preserve.
		session = placeholder.NewManager().GetSession()
	}
	c := defaults.NewDefaultConverter()
	doc, _, err := c.FromMarkdown(md, session)
	if err != nil {
		return nil, fmt.Errorf("MD->ADF: %w", err)
	}
	return toMapAny(doc)
}

// toADFDocument converts a runtime any (typically map[string]any from JSON unmarshal)
// to a typed Document via JSON roundtrip.
func toADFDocument(v any) (adftypes.Document, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return adftypes.Document{}, err
	}
	var doc adftypes.Document
	err = json.Unmarshal(data, &doc)
	return doc, err
}

// toMapAny converts a typed Document back to map[string]any via JSON roundtrip,
// matching the format expected by the Jira API client.
func toMapAny(doc adftypes.Document) (any, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	err = json.Unmarshal(data, &result)
	return result, err
}
