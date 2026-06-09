package components

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
)

func TestDblClickDetector_SameIndexTwiceIsDouble(t *testing.T) {
	t.Parallel()
	var detector DblClickDetector

	testkit.AssertEqual(t, "first click", detector.Click(3), false)
	testkit.AssertEqual(t, "second click same index", detector.Click(3), true)
}

func TestDblClickDetector_DifferentIndexIsNotDouble(t *testing.T) {
	t.Parallel()
	var detector DblClickDetector

	testkit.AssertEqual(t, "first click", detector.Click(3), false)
	testkit.AssertEqual(t, "second click other index", detector.Click(4), false)
}

func TestDblClickDetector_TripleClickResetsAfterDouble(t *testing.T) {
	t.Parallel()
	var detector DblClickDetector

	detector.Click(3)
	testkit.AssertEqual(t, "double", detector.Click(3), true)
	testkit.AssertEqual(t, "third click does not retrigger", detector.Click(3), false)
}
