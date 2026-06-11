package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira"
)

func makeCommentsDetail(t *testing.T, bodies ...string) *DetailView {
	t.Helper()
	detail := makeFocusedDetail()
	comments := make([]jira.Comment, 0, len(bodies))
	for _, body := range bodies {
		comments = append(comments, jira.Comment{Body: body})
	}
	detail.SetIssue(&jira.Issue{Key: testKey, Comments: comments})
	detail.SetActiveTab(TabComments)
	return detail
}

func longCommentBody() string {
	lines := make([]string, maxBlockLines+3)
	for i := range lines {
		lines[i] = "content line"
	}
	return strings.Join(lines, "\n")
}

func TestDetailView_SelectedComment_NilWhenCursorOutOfRange(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t, "only one")
	detail.listCursor = 5
	if detail.SelectedComment() != nil {
		t.Error("out of range cursor should return nil comment")
	}
}

func TestDetailView_NextTab_FallsBackWhenActiveTabHidden(t *testing.T) {
	t.Parallel()
	detail := NewDetailView(BuiltinRenderer{})
	detail.SetActiveTab(TabComments)
	detail.NextTab()
	testkit.AssertEqual(t, "fallback to first tab", detail.ActiveTab(), TabDetails)
}

func TestDetailView_PrevTab_FallsBackWhenActiveTabHidden(t *testing.T) {
	t.Parallel()
	detail := NewDetailView(BuiltinRenderer{})
	detail.SetActiveTab(TabComments)
	detail.PrevTab()
	testkit.AssertEqual(t, "fallback to first tab", detail.ActiveTab(), TabDetails)
}

func TestDetailView_ClickTab_BeforeTabsAreaIgnored(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t, "c")
	detail.SetActiveTab(TabDetails)
	detail.ClickTab(1)
	testkit.AssertEqual(t, "tab unchanged", detail.ActiveTab(), TabDetails)
}

func TestDetailView_ScrollBy_ListTabClamps(t *testing.T) {
	t.Parallel()

	t.Run("negative delta clamps to zero", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "c1", "c2")
		detail.ScrollBy(-5)
		testkit.AssertEqual(t, "cursor", detail.listCursor, 0)
	})

	t.Run("large delta clamps to last item", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "c1", "c2")
		detail.ScrollBy(10)
		testkit.AssertEqual(t, "cursor", detail.listCursor, 1)
	})

	t.Run("empty list keeps cursor at zero", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t)
		detail.ScrollBy(3)
		testkit.AssertEqual(t, "cursor", detail.listCursor, 0)
	})
}

func TestDetailView_ClickItem_Branches(t *testing.T) {
	t.Parallel()

	t.Run("double click on truncated block expands", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, longCommentBody())
		detail.View()
		detail.ClickItem(2)
		cmd := detail.ClickItem(2)
		if cmd == nil {
			t.Fatal("double click on truncated block should return command")
		}
		msg := cmd()
		if _, ok := msg.(ExpandBlockMsg); !ok {
			t.Fatalf("expected ExpandBlockMsg, got %T", msg)
		}
	})

	t.Run("click on second block moves cursor", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "first", "second")
		detail.View()
		blockHeight := len(detail.blocks[0])
		detail.ClickItem(blockHeight + 2)
		testkit.AssertEqual(t, "cursor on second block", detail.listCursor, 1)
	})

	t.Run("click past all blocks returns nil", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "only")
		detail.View()
		if cmd := detail.ClickItem(200); cmd != nil {
			t.Error("click past blocks should return nil")
		}
	})
}

func TestDetailView_ListTabItemCount_NilIssue(t *testing.T) {
	t.Parallel()
	detail := NewDetailView(BuiltinRenderer{})
	detail.SetActiveTab(TabComments)
	testkit.AssertEqual(t, "count", detail.listTabItemCount(), 0)
}

func TestDetailView_Update_NonKeyMsgIgnored(t *testing.T) {
	t.Parallel()
	detail := makeFocusedDetail()
	detail.SetIssue(makeDetailIssue())
	_, cmd := detail.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if cmd != nil {
		t.Error("non key message should be ignored")
	}
}

func TestDetailView_Update_NavBottomScrollsTextTab(t *testing.T) {
	t.Parallel()
	detail := makeFocusedDetail()
	detail.SetIssue(&jira.Issue{Key: testKey, Description: strings.Repeat("line\n", 60)})
	detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	detail.View()
	if detail.scrollY == 0 {
		t.Error("nav bottom on text tab should scroll down")
	}
}

func TestDetailView_HandleCursorDownUp_ListWraps(t *testing.T) {
	t.Parallel()

	t.Run("down moves then wraps to top", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "c1", "c2")
		detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		testkit.AssertEqual(t, "moved down", detail.listCursor, 1)
		detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
		testkit.AssertEqual(t, "wrapped to top", detail.listCursor, 0)
	})

	t.Run("up wraps to bottom then moves up", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "c1", "c2")
		detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		testkit.AssertEqual(t, "wrapped to bottom", detail.listCursor, 1)
		detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
		testkit.AssertEqual(t, "moved up", detail.listCursor, 0)
	})
}

func TestDetailView_HandleHalfPage_ListTab(t *testing.T) {
	t.Parallel()

	t.Run("half page down clamps to last item", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "c1", "c2", "c3")
		detail.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
		testkit.AssertEqual(t, "cursor", detail.listCursor, 2)
	})

	t.Run("half page up clamps to zero", func(t *testing.T) {
		t.Parallel()
		detail := makeCommentsDetail(t, "c1", "c2", "c3")
		detail.listCursor = 2
		detail.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
		testkit.AssertEqual(t, "cursor", detail.listCursor, 0)
	})
}

func TestDetailView_HandleActivation_NonListTabNoop(t *testing.T) {
	t.Parallel()
	detail := makeFocusedDetail()
	detail.SetIssue(makeDetailIssue())
	_, cmd := detail.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on text tab should return nil command")
	}
}

func TestDetailView_HandleActivation_ShortBlockNoop(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t, "short")
	detail.View()
	_, cmd := detail.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("enter on short block should return nil command")
	}
}

func TestDetailView_View_EmptyCommentsTabShowsNoContent(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t)
	output := stripANSI(detail.View())
	if !strings.Contains(output, "No content.") {
		t.Errorf("empty comments view = %q, want 'No content.'", output)
	}
}

func TestDetailView_RenderBlockList_ClampsCursor(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t, "c1", "c2")
	detail.listCursor = 99
	detail.View()
	testkit.AssertEqual(t, "cursor clamped", detail.listCursor, 1)
}

func TestDetailView_RenderBlockList_TruncatedBlockUnfocused(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t, longCommentBody(), "short")
	detail.View()
	output := stripANSI(detail.View())
	if !strings.Contains(output, "...") {
		t.Errorf("truncated block view = %q, want ellipsis", output)
	}
}

func TestDetailView_AutoScroll_FollowsCursorInSmallView(t *testing.T) {
	t.Parallel()
	detail := makeCommentsDetail(t, "c1", "c2", "c3", "c4", "c5", "c6")
	detail.SetSize(80, 5)
	detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	detail.View()
	if detail.scrollY == 0 {
		t.Error("view should scroll to keep last block visible")
	}
	detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	detail.View()
	testkit.AssertEqual(t, "scrolled back to top", detail.scrollY, 0)
}

func TestDetailView_ClampAndSliceScroll(t *testing.T) {
	t.Parallel()

	makeLines := func(n int) []string {
		lines := make([]string, n)
		for i := range lines {
			lines[i] = "row"
		}
		return lines
	}

	t.Run("scroll beyond end clamped to max", func(t *testing.T) {
		t.Parallel()
		detail := NewDetailView(BuiltinRenderer{})
		detail.scrollY = 100
		got := detail.clampAndSliceScroll(makeLines(10), 4)
		testkit.AssertEqual(t, "scroll", detail.scrollY, 6)
		testkit.AssertEqual(t, "visible lines", len(got), 4)
	})

	t.Run("negative scroll clamped to zero", func(t *testing.T) {
		t.Parallel()
		detail := NewDetailView(BuiltinRenderer{})
		detail.scrollY = -3
		got := detail.clampAndSliceScroll(makeLines(2), 4)
		testkit.AssertEqual(t, "scroll", detail.scrollY, 0)
		testkit.AssertEqual(t, "all lines", len(got), 2)
	})

	t.Run("empty content returns nil", func(t *testing.T) {
		t.Parallel()
		detail := NewDetailView(BuiltinRenderer{})
		got := detail.clampAndSliceScroll(nil, 4)
		testkit.AssertEqual(t, "no lines", len(got), 0)
	})
}

func TestColorURLs_NoURLUnchanged(t *testing.T) {
	t.Parallel()
	testkit.AssertEqual(t, "unchanged", colorURLs("no links here"), "no links here")
}

func TestColorURLsWrapped_PlainLines(t *testing.T) {
	t.Parallel()
	lines := colorURLsWrapped([]string{"first plain", "second plain"})
	testkit.AssertSliceEqual(t, "unchanged", lines, []string{"first plain", "second plain"})
}

func TestLineEndsInURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
		want bool
	}{
		{"no url", "plain text", false},
		{"url at end", "see https://example.com/long", true},
		{"url followed by space", "see https://example.com done", false},
		{"http later than https", "https://a.example then http://tail.example", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "ends in url", lineEndsInURL(tt.line), tt.want)
		})
	}
}

func TestColorMentions(t *testing.T) {
	t.Parallel()

	t.Run("marker replaced with colored name", func(t *testing.T) {
		t.Parallel()
		got := colorMentions("ping \x00MENTION:@Ann\x00 now")
		testkit.AssertEqual(t, "marker resolved", stripANSI(got), "ping @Ann now")
	})

	t.Run("marker without terminator left alone", func(t *testing.T) {
		t.Parallel()
		input := "broken \x00MENTION:@Ann tail"
		testkit.AssertEqual(t, "unchanged", colorMentions(input), input)
	})

	t.Run("no marker unchanged", func(t *testing.T) {
		t.Parallel()
		testkit.AssertEqual(t, "unchanged", colorMentions("plain"), "plain")
	})
}

func TestDetailView_RenderHistoryBlocks_EmptyValuesBecomeNone(t *testing.T) {
	t.Parallel()
	detail := makeFocusedDetail()
	detail.SetIssue(&jira.Issue{
		Key: testKey,
		Changelog: []jira.ChangelogEntry{
			{
				Author:  &jira.User{DisplayName: "Mover"},
				Created: time.Now().Add(-time.Hour),
				Items:   []jira.ChangeItem{{Field: "summary", FromString: "", ToString: ""}},
			},
		},
	})
	detail.SetActiveTab(TabHistory)
	output := stripANSI(detail.View())
	if !strings.Contains(output, noneLabel) {
		t.Errorf("history view = %q, want %q for empty values", output, noneLabel)
	}
}

func TestRenderDiff_SkipsNoneValues(t *testing.T) {
	t.Parallel()
	lines := renderDiff(noneLabel, noneLabel, 80)
	joined := stripANSI(strings.Join(lines, "\n"))
	if !strings.Contains(joined, "content changed") {
		t.Errorf("none diff = %q, want '(content changed)'", joined)
	}
}

func TestExtractURLs_CommentSources(t *testing.T) {
	t.Parallel()

	t.Run("plain comment urls", func(t *testing.T) {
		t.Parallel()
		issue := &jira.Issue{
			Key:      testKey,
			Comments: []jira.Comment{{Body: "see https://comment.example now"}},
		}
		groups := ExtractURLs(issue, "https://h")
		testkit.AssertEqual(t, "group count", len(groups), 1)
		testkit.AssertEqual(t, "section", groups[0].Section, "Comments")
		testkit.AssertSliceEqual(t, "urls", groups[0].URLs, []string{"https://comment.example"})
	})

	t.Run("adf comment urls", func(t *testing.T) {
		t.Parallel()
		issue := &jira.Issue{
			Key: testKey,
			Comments: []jira.Comment{{
				BodyADF: adfDoc(map[string]any{
					"type":  adfInlineCard,
					"attrs": map[string]any{"url": "https://adfcomment.example"},
				}),
			}},
		}
		groups := ExtractURLs(issue, "https://h")
		testkit.AssertEqual(t, "group count", len(groups), 1)
		testkit.AssertSliceEqual(t, "urls", groups[0].URLs, []string{"https://adfcomment.example"})
	})

	t.Run("inward link url", func(t *testing.T) {
		t.Parallel()
		issue := &jira.Issue{
			Key:        testKey,
			IssueLinks: []jira.IssueLink{{InwardIssue: &jira.Issue{Key: testKey2}}},
		}
		groups := ExtractURLs(issue, "https://h")
		testkit.AssertEqual(t, "group count", len(groups), 1)
		testkit.AssertSliceEqual(t, "urls", groups[0].URLs, []string{"https://h/browse/" + testKey2})
	})
}

func TestCleanWikiMarkup_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"accountid without closing bracket", "ping [~accountid:abc", "ping [~accountid:abc"},
		{"code without closing brace", "x {code", "x {code"},
		{"code without closing tag strips opening", "x {code:go} y", "x  y"},
		{"bracket without closing kept", "open [tail", "open [tail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testkit.AssertEqual(t, "cleaned", cleanWikiMarkup(tt.in), strings.TrimSpace(tt.want))
		})
	}
}

func TestWrapText_WideRuneNarrowWidth(t *testing.T) {
	t.Parallel()
	lines := wrapText("你你", 1)
	testkit.AssertSliceEqual(t, "one rune per line", lines, []string{"你", "你"})
}

func TestRenderDescriptionPreview_CloudUsesADF(t *testing.T) {
	t.Parallel()
	lines := RenderDescriptionPreview("# Title\n\nbody text", 40, true, BuiltinRenderer{})
	joined := stripANSI(strings.Join(lines, "\n"))
	if !strings.Contains(joined, "Title") || !strings.Contains(joined, "body text") {
		t.Errorf("cloud preview = %q, want rendered markdown", joined)
	}
}

func TestRenderDescriptionPreview_ZeroWidthNil(t *testing.T) {
	t.Parallel()
	if lines := RenderDescriptionPreview("text", 0, true, BuiltinRenderer{}); lines != nil {
		t.Errorf("zero width = %v, want nil", lines)
	}
}
