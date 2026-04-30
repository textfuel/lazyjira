package navstack

import (
	"testing"

	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func TestNavStack_NewStack_IsEmpty(t *testing.T) {
	s := NewNavStack()
	if got := s.Depth(); got != 0 {
		t.Fatalf("Depth() = %d, want 0", got)
	}
}

func TestNavStack_Pop_OnEmpty_ReturnsZeroFrame(t *testing.T) {
	s := NewNavStack()
	got := s.Pop()
	if !isZeroFrame(got) {
		t.Fatalf("Pop() on empty = %+v, want zero frame", got)
	}
	if s.Depth() != 0 {
		t.Fatalf("Depth() after Pop on empty = %d, want 0", s.Depth())
	}
}

func TestNavStack_Peek_OnEmpty_ReturnsZeroFrame(t *testing.T) {
	s := NewNavStack()
	got := s.Peek()
	if !isZeroFrame(got) {
		t.Fatalf("Peek() on empty = %+v, want zero frame", got)
	}
	if s.Depth() != 0 {
		t.Fatalf("Depth() after Peek on empty = %d, want 0", s.Depth())
	}
}

func TestNavStack_Push_IncreasesDepth(t *testing.T) {
	s := NewNavStack()
	s.Push(makeFrame("FOO-1", 0, SourceFromList))
	if s.Depth() != 1 {
		t.Fatalf("Depth after 1 push = %d, want 1", s.Depth())
	}
	s.Push(makeFrame("FOO-2", 0, SourceFromList))
	if s.Depth() != 2 {
		t.Fatalf("Depth after 2 pushes = %d, want 2", s.Depth())
	}
}

func TestNavStack_Peek_ReturnsTop(t *testing.T) {
	s := NewNavStack()
	s.Push(makeFrame("A", 1, SourceFromList))
	s.Push(makeFrame("B", 2, SourceFromList))
	top := s.Peek()
	if top.ParentKey != "B" || top.SelectedIdx != 2 {
		t.Fatalf("Peek = %+v, want ParentKey=B SelectedIdx=2", top)
	}
	if s.Depth() != 2 {
		t.Fatalf("Peek mutated Depth to %d", s.Depth())
	}
}

func TestNavStack_Pop_ReturnsLastPushedFrame(t *testing.T) {
	s := NewNavStack()
	s.Push(makeFrame("A", 0, SourceFromList))
	s.Push(makeFrame("B", 5, SourceFromList))
	got := s.Pop()
	if got.ParentKey != "B" || got.SelectedIdx != 5 {
		t.Fatalf("Pop = %+v, want ParentKey=B SelectedIdx=5", got)
	}
	if s.Depth() != 1 {
		t.Fatalf("Depth after Pop = %d, want 1", s.Depth())
	}
	if s.Peek().ParentKey != "A" {
		t.Fatalf("Top after Pop = %q, want A", s.Peek().ParentKey)
	}
}

func TestNavStack_Pop_MultipleTimes_UntilEmpty(t *testing.T) {
	s := NewNavStack()
	s.Push(makeFrame("A", 0, SourceFromList))
	s.Push(makeFrame("B", 0, SourceFromList))
	_ = s.Pop()
	_ = s.Pop()
	got := s.Pop()
	if !isZeroFrame(got) {
		t.Fatalf("Pop past empty = %+v, want zero frame", got)
	}
	if s.Depth() != 0 {
		t.Fatalf("Depth = %d, want 0", s.Depth())
	}
}

func TestNavStack_Push_DifferentParentKeys_Appends(t *testing.T) {
	s := NewNavStack()
	s.Push(makeFrame("FOO-1", 0, SourceFromList))
	s.Push(makeFrame("FOO-2", 0, SourceFromList))
	if s.Depth() != 2 {
		t.Fatalf("Depth after two different pushes = %d, want 2", s.Depth())
	}
}

func TestNavStack_Clear_EmptiesStack(t *testing.T) {
	s := NewNavStack()
	s.Push(makeFrame("A", 0, SourceFromList))
	s.Push(makeFrame("B", 0, SourceFromList))
	s.Clear()
	if s.Depth() != 0 {
		t.Fatalf("Depth after Clear = %d, want 0", s.Depth())
	}
	if !isZeroFrame(s.Peek()) {
		t.Fatalf("Peek after Clear = %+v, want zero frame", s.Peek())
	}
}

// testFocusIssues is a dummy non-zero FocusPanel value used in test fixtures.
// The semantic mapping lives in the tui package; for NavStack tests any
// non-zero int suffices.
const testFocusIssues FocusPanel = 1

func makeFrame(parentKey string, selectedIdx int, source Source) NavFrame {
	return NavFrame{
		Issues:      []jira.Issue{{Key: parentKey + "-dummy"}},
		SelectedIdx: selectedIdx,
		FocusPanel:  testFocusIssues,
		InfoTab:     0,
		InfoCursor:  0,
		Source:      source,
		ParentKey:   parentKey,
	}
}

func isZeroFrame(f NavFrame) bool {
	return f.Issues == nil &&
		f.SelectedIdx == 0 &&
		f.FocusPanel == FocusPanel(0) &&
		f.InfoTab == 0 &&
		f.InfoCursor == 0 &&
		f.Source == Source(0) &&
		f.ParentKey == ""
}
