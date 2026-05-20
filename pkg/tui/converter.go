package tui

// ADFConverter abstracts ADF<->Markdown conversion.
// state is opaque roundtrip data (e.g. placeholder session) that must be
// passed from ToMarkdown back into FromMarkdown for the same edit cycle.
type ADFConverter interface {
	ToMarkdown(adf any) (markdown string, state any, err error)
	FromMarkdown(md string, state any) (adf any, err error)
}
