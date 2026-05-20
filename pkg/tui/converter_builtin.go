package tui

import "github.com/textfuel/lazyjira/v2/pkg/tui/views"

// BuiltinConverter wraps lazyjira's built-in ADF<->Markdown conversion.
// It is stateless — state is always nil.
type BuiltinConverter struct{}

func (BuiltinConverter) ToMarkdown(adf any) (string, any, error) {
	return views.ADFToMarkdown(adf), nil, nil
}

func (BuiltinConverter) FromMarkdown(md string, _ any) (any, error) {
	return views.MarkdownToADF(md), nil
}
