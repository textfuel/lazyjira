package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/config"
)

// topBorderLine extracts the first line from View() output.
func topBorderLine(m *IssuesList) string {
	v := m.View()
	if v == "" {
		return ""
	}
	return strings.SplitN(v, "\n", 2)[0]
}

// makeIssuesListWithTabs creates an IssuesList with the given tab names and panel size.
func makeIssuesListWithTabs(width, height int, tabNames ...string) *IssuesList {
	m := NewIssuesList()
	tabs := make([]config.IssueTabConfig, len(tabNames))
	for i, name := range tabNames {
		tabs[i] = config.IssueTabConfig{Name: name}
	}
	m.SetTabs(tabs)
	m.SetSize(width, height)
	return m
}

// TestIssuesList_TopBorderWidth_FewTabs verifies that with a small number of tabs
// the top border of View() is exactly width visible characters wide.
func TestIssuesList_TopBorderWidth_FewTabs(t *testing.T) {
	const width = 50
	m := makeIssuesListWithTabs(width, 8, "My Issues", "Done")

	line := topBorderLine(m)
	got := lipgloss.Width(line)
	if got != width {
		t.Errorf("top border width = %d, want %d\nline: %q", got, width, line)
	}
}

// TestIssuesList_TopBorderWidth_ManyTabsOverflow verifies that when many tabs
// are present and their combined title exceeds the panel width, the top border
// is still exactly width visible characters wide.
func TestIssuesList_TopBorderWidth_ManyTabsOverflow(t *testing.T) {
	const width = 50
	m := makeIssuesListWithTabs(width, 8,
		"My Issues", "In Progress", "Blocked", "Done", "Backlog", "JQL",
	)

	line := topBorderLine(m)
	got := lipgloss.Width(line)
	if got != width {
		t.Errorf("top border width = %d, want %d (title overflowed border)\nline: %q", got, width, line)
	}
}

// TestIssuesList_ActiveTabVisible_WhenTitleOverflows verifies that the active
// tab's name is always visible in the title even when the full tab list would
// overflow the panel width.
func TestIssuesList_ActiveTabVisible_WhenTitleOverflows(t *testing.T) {
	const width = 50
	m := makeIssuesListWithTabs(width, 8,
		"My Issues", "In Progress", "Blocked", "Done", "Backlog", "JQL",
	)

	// Activate the last tab, which is most likely to be cut off.
	for range m.tabs[1:] {
		m.NextTab()
	}
	activeTab := m.ActiveTab().Name

	line := topBorderLine(m)
	if !strings.Contains(stripANSI(line), activeTab) {
		t.Errorf("active tab %q not visible in top border\nline: %q", activeTab, line)
	}
}

// TestIssuesList_SlidingWindow_ContiguousOrder verifies that the visible tabs
// always form a contiguous slice of the full list in their original order.
// No tabs from the right of the active tab should appear to its left, and
// no tabs from the left should be skipped over.
func TestIssuesList_SlidingWindow_ContiguousOrder(t *testing.T) {
	tabNames := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta"}
	const width = 50

	for activeIdx := range tabNames {
		m := makeIssuesListWithTabs(width, 8, tabNames...)
		for i := 0; i < activeIdx; i++ {
			m.NextTab()
		}

		plain := stripANSI(topBorderLine(m))

		// Collect which tabs are visible and in what order.
		var visible []int
		for i, name := range tabNames {
			if strings.Contains(plain, name) {
				visible = append(visible, i)
			}
		}

		if len(visible) == 0 {
			t.Errorf("activeIdx=%d: no tabs visible\nline: %q", activeIdx, plain)
			continue
		}

		// Must contain the active tab.
		found := false
		for _, idx := range visible {
			if idx == activeIdx {
				found = true
			}
		}
		if !found {
			t.Errorf("activeIdx=%d: active tab %q not in visible set %v\nline: %q",
				activeIdx, tabNames[activeIdx], visible, plain)
		}

		// Visible set must be contiguous (no gaps).
		for i := 1; i < len(visible); i++ {
			if visible[i] != visible[i-1]+1 {
				t.Errorf("activeIdx=%d: visible tabs %v are not contiguous\nline: %q",
					activeIdx, visible, plain)
				break
			}
		}

		// Border width must still be exact.
		got := lipgloss.Width(topBorderLine(m))
		if got != width {
			t.Errorf("activeIdx=%d: top border width = %d, want %d", activeIdx, got, width)
		}
	}
}

// TestIssuesList_SlidingWindow_EarlyTabsHiddenWhenActiveIsLate verifies that
// when the active tab is near the end of a long list, early tabs are scrolled
// out of view (i.e. the window does not always start at tab 0).
func TestIssuesList_SlidingWindow_EarlyTabsHiddenWhenActiveIsLate(t *testing.T) {
	tabNames := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta"}
	const width = 50
	m := makeIssuesListWithTabs(width, 8, tabNames...)

	// Move to the last tab.
	for range tabNames[1:] {
		m.NextTab()
	}

	plain := stripANSI(topBorderLine(m))

	// If the window is working correctly, "Alpha" should NOT appear when the
	// active tab is "Zeta" and there is not enough room for both.
	// Only assert this when the full list provably overflows.
	fullM := makeIssuesListWithTabs(width, 8, tabNames...)
	fullBorderW := lipgloss.Width(topBorderLine(fullM))
	fullPlain := stripANSI(topBorderLine(fullM))
	if fullBorderW == width && strings.Contains(fullPlain, "Zeta") {
		// All tabs fit - skip this assertion.
		return
	}

	if strings.Contains(plain, "Alpha") {
		t.Errorf("expected early tab %q to be scrolled out of view when active tab is %q\nline: %q",
			"Alpha", "Zeta", plain)
	}
}

// TestIssuesList_ActiveTabTruncated_WhenLabelExceedsBudget verifies that when a
// single tab label is wider than the entire available title budget the border
// width is still exactly width (the label is truncated, not overflowed).
func TestIssuesList_ActiveTabTruncated_WhenLabelExceedsBudget(t *testing.T) {
	const width = 20
	// "A Very Long Tab Name" is 20 chars; prefix "[2] " is 4, so the label
	// budget is width-3-4 = 13. The label must be truncated to fit.
	m := makeIssuesListWithTabs(width, 8, "A Very Long Tab Name")

	line := topBorderLine(m)
	got := lipgloss.Width(line)
	if got != width {
		t.Errorf("top border width = %d, want %d (label overflowed border)\nline: %q", got, width, line)
	}
}


// stripANSI removes ANSI escape sequences for plain-text substring checks.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case !inEsc:
			b.WriteRune(r)
		}
	}
	return b.String()
}