package navstack

import "github.com/textfuel/lazyjira/v2/pkg/jira"

// Opaque panel identifier stored in NavFrame. The semantic mapping
// (which int = which panel) lives in the tui package; this package
// treats it as opaque snapshot data.
type FocusPanel int

type Source int

const (
	SourceFromList Source = iota
	SourceFromInfoSub
	SourceFromInfoLink
	SourceParent
)

type NavFrame struct {
	Issues       []jira.Issue
	SelectedIdx  int
	FocusPanel   FocusPanel
	InfoTab      int
	InfoCursor   int
	Source       Source
	ParentKey    string
	OriginTabIdx int
}

type NavStack struct {
	frames []NavFrame
}

func NewNavStack() *NavStack {
	return &NavStack{}
}

func (s *NavStack) Push(frame NavFrame) {
	s.frames = append(s.frames, frame)
}

func (s *NavStack) Clear() {
	s.frames = nil
}

func (s *NavStack) Depth() int {
	return len(s.frames)
}

func (s *NavStack) Peek() NavFrame {
	if len(s.frames) == 0 {
		return NavFrame{}
	}
	return s.frames[len(s.frames)-1]
}

func (s *NavStack) Pop() NavFrame {
	if len(s.frames) == 0 {
		return NavFrame{}
	}
	top := s.frames[len(s.frames)-1]
	s.frames = s.frames[:len(s.frames)-1]
	return top
}
