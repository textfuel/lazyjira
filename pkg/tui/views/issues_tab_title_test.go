package views

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/v2/pkg/config"
)

func topBorderLine(m *IssuesList) string {
	v := m.View()
	if v == "" {
		return ""
	}
	return strings.SplitN(v, "\n", 2)[0]
}

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

func TestIssuesList_TopBorderWidth_FewTabs(t *testing.T) {
	t.Parallel()
	const width = 50
	m := makeIssuesListWithTabs(width, 8, "My Issues", "Done")

	line := topBorderLine(m)
	got := lipgloss.Width(line)
	if got != width {
		t.Errorf("top border width = %d, want %d\nline: %q", got, width, line)
	}
}

func TestIssuesList_TopBorderWidth_ManyTabsOverflow(t *testing.T) {
	t.Parallel()
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

func TestIssuesList_ActiveTabVisible_WhenTitleOverflows(t *testing.T) {
	t.Parallel()
	const width = 50
	m := makeIssuesListWithTabs(width, 8,
		"My Issues", "In Progress", "Blocked", "Done", "Backlog", "JQL",
	)

	for range m.tabs[1:] {
		m.NextTab()
	}
	activeTab := m.ActiveTab().Name

	line := topBorderLine(m)
	if !strings.Contains(stripANSI(line), activeTab) {
		t.Errorf("active tab %q not visible in top border\nline: %q", activeTab, line)
	}
}

func TestIssuesList_SlidingWindow_ContiguousOrder(t *testing.T) {
	t.Parallel()
	tabNames := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta"}
	const width = 50

	for activeIdx, activeTabName := range tabNames {
		m := makeIssuesListWithTabs(width, 8, tabNames...)
		for range activeIdx {
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

		found := false
		for _, idx := range visible {
			if idx == activeIdx {
				found = true
			}
		}
		if !found {
			t.Errorf("activeIdx=%d: active tab %q not in visible set %v\nline: %q",
				activeIdx, activeTabName, visible, plain)
		}

		for i := 1; i < len(visible); i++ {
			if visible[i] != visible[i-1]+1 {
				t.Errorf("activeIdx=%d: visible tabs %v are not contiguous\nline: %q",
					activeIdx, visible, plain)
				break
			}
		}

		got := lipgloss.Width(topBorderLine(m))
		if got != width {
			t.Errorf("activeIdx=%d: top border width = %d, want %d", activeIdx, got, width)
		}
	}
}

func TestIssuesList_SlidingWindow_EarlyTabsHiddenWhenActiveIsLate(t *testing.T) {
	t.Parallel()
	tabNames := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta"}
	const width = 50
	m := makeIssuesListWithTabs(width, 8, tabNames...)

	for range tabNames[1:] {
		m.NextTab()
	}

	plain := stripANSI(topBorderLine(m))

	fullM := makeIssuesListWithTabs(width, 8, tabNames...)
	fullBorderW := lipgloss.Width(topBorderLine(fullM))
	fullPlain := stripANSI(topBorderLine(fullM))
	if fullBorderW == width && strings.Contains(fullPlain, "Zeta") {
		return
	}

	if strings.Contains(plain, "Alpha") {
		t.Errorf("expected early tab %q to be scrolled out of view when active tab is %q\nline: %q",
			"Alpha", "Zeta", plain)
	}
}

func TestIssuesList_ActiveTabTruncated_WhenLabelExceedsBudget(t *testing.T) {
	t.Parallel()
	const width = 20
	m := makeIssuesListWithTabs(width, 8, "A Very Long Tab Name")

	line := topBorderLine(m)
	got := lipgloss.Width(line)
	if got != width {
		t.Errorf("top border width = %d, want %d (label overflowed border)\nline: %q", got, width, line)
	}
}
