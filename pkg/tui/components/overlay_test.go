package components

import (
	"strings"
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestOverlay_CentersForegroundOnBackground(t *testing.T) {
	t.Parallel()
	bg := strings.Repeat(strings.Repeat(".", 20)+"\n", 10)
	bg = strings.TrimRight(bg, "\n")
	fg := "HELLO"
	out := Overlay(bg, fg, 20, 10)
	if !strings.Contains(out, "HELLO") {
		t.Errorf("expected 'HELLO' in overlay output, got %q", out)
	}
}

func TestOverlay_BackgroundPreservedAroundForeground(t *testing.T) {
	t.Parallel()
	bg := strings.Repeat(".", 20)
	fg := "X"
	out := Overlay(bg, fg, 20, 1)
	testkit.AssertEqual(t, "contains foreground", strings.Contains(out, "X"), true)
}

func TestOverlayAt_PlacesAtExactPosition(t *testing.T) {
	t.Parallel()
	bg := strings.Repeat(strings.Repeat(".", 20)+"\n", 5)
	bg = strings.TrimRight(bg, "\n")
	out := OverlayAt(bg, "AB", 5, 2, 20, 5)
	lines := strings.Split(out, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}
	plain := stripANSI(lines[2])
	if !strings.Contains(plain, "AB") {
		t.Errorf("expected 'AB' at row 2, got %q", plain)
	}
}

func TestOverlayLine_ForegroundReplacesBackground(t *testing.T) {
	t.Parallel()
	bg := "AAAAAA"
	fg := "BB"
	result := overlayLine(bg, fg, 2, 6)
	if !strings.Contains(result, "BB") {
		t.Errorf("expected 'BB' in overlaid line, got %q", result)
	}
	if !strings.HasPrefix(result, "AA") {
		t.Errorf("expected leading 'AA' preserved, got %q", result)
	}
}

func TestOverlayLines_DoesNotPanicOnEmptyBg(t *testing.T) {
	t.Parallel()
	out := Overlay("", "content", 20, 5)
	testkit.AssertEqual(t, "non-empty result", out != "", true)
}

func TestCenterOverlay_PlacesPopupCentered(t *testing.T) {
	t.Parallel()
	bg := strings.Repeat(strings.Repeat(" ", 40)+"\n", 20)
	bg = strings.TrimRight(bg, "\n")
	popup := "pop"
	out := centerOverlay(bg, popup, 40, 20)
	if !strings.Contains(out, "pop") {
		t.Errorf("expected popup content in result, got %q", out)
	}
}

func TestCenterOverlayWithHint_EmptyHintSkipsHintLine(t *testing.T) {
	t.Parallel()
	bg := strings.Repeat(strings.Repeat(" ", 40)+"\n", 20)
	bg = strings.TrimRight(bg, "\n")
	popup := "pop"
	out := centerOverlayWithHint(bg, popup, "", 40, 20)
	testkit.AssertEqual(t, "contains popup", strings.Contains(out, "pop"), true)
}

func TestCenterOverlayWithHint_WithHint(t *testing.T) {
	t.Parallel()
	bg := strings.Repeat(strings.Repeat(" ", 40)+"\n", 20)
	bg = strings.TrimRight(bg, "\n")
	out := centerOverlayWithHint(bg, "pop", "hint", 40, 20)
	testkit.AssertEqual(t, "contains popup", strings.Contains(out, "pop"), true)
	testkit.AssertEqual(t, "contains hint", strings.Contains(out, "hint"), true)
}
