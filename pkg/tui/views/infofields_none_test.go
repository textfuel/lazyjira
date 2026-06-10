package views

import (
	"testing"
)

func TestNoneStyle_NonEmpty(t *testing.T) {
	t.Parallel()
	style := noneStyle()
	rendered := style.Render(noneLabel)
	if rendered == "" {
		t.Error("noneStyle() should render non-empty styled text")
	}
}
