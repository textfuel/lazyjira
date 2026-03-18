package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/pkg/jira"
	"github.com/textfuel/lazyjira/pkg/tui/components"
	"github.com/textfuel/lazyjira/pkg/tui/theme"
)

type IssuesLoadedMsg struct{ Issues []jira.Issue }
type IssueSelectedMsg struct{ Issue *jira.Issue }

// IssueTab defines which subset of issues to show.
type IssueTab int

const (
	IssueTabAll      IssueTab = iota
	IssueTabAssigned
)

const statusOpen = "○"

type IssuesList struct {
	issues      []jira.Issue
	allIssues   []jira.Issue
	filter      string
	tab         IssueTab
	userEmail   string
	activeKey   string // the issue currently being viewed
	keyColWidth int
	cursor      int
	offset      int
	width       int
	height      int
	focused     bool
	theme       *theme.Theme
}

func NewIssuesList() *IssuesList {
	return &IssuesList{theme: theme.DefaultTheme()}
}

func (m *IssuesList) SetUserEmail(email string) { m.userEmail = email }
func (m *IssuesList) SetActiveKey(key string)   { m.activeKey = key }
func (m *IssuesList) ClearActiveKey()            { m.activeKey = "" }

func (m *IssuesList) NextTab() {
	if m.tab == IssueTabAll {
		m.tab = IssueTabAssigned
	} else {
		m.tab = IssueTabAll
	}
	m.applyFilter()
}

func (m *IssuesList) PrevTab() { m.NextTab() }

func (m *IssuesList) SetIssues(issues []jira.Issue) {
	// Remember current selection to preserve position.
	var selectedKey string
	if sel := m.SelectedIssue(); sel != nil {
		selectedKey = sel.Key
	}

	m.allIssues = issues
	m.keyColWidth = 0
	for _, issue := range issues {
		if w := lipgloss.Width(issue.Key); w > m.keyColWidth {
			m.keyColWidth = w
		}
	}
	m.applyFilter()

	// Restore cursor position.
	if selectedKey != "" {
		m.SelectByKey(selectedKey)
	}
}

func (m *IssuesList) SetFilter(query string) {
	m.filter = query
	m.applyFilter()
}

// ClearFilter removes the search filter but keeps tab filter. Cursor preserved via SelectByKey.
func (m *IssuesList) ClearFilter() {
	m.filter = ""
	m.applyFilterKeepCursor()
}

func (m *IssuesList) applyFilterKeepCursor() {
	prevKey := ""
	if sel := m.SelectedIssue(); sel != nil {
		prevKey = sel.Key
	}
	m.applyFilter()
	if prevKey != "" {
		m.SelectByKey(prevKey)
	}
}

// SelectByKey moves cursor to the issue with the given key. Returns true if found.
func (m *IssuesList) SelectByKey(key string) bool {
	for i, issue := range m.issues {
		if issue.Key == key {
			m.cursor = i
			m.adjustOffset()
			return true
		}
	}
	return false
}

func (m *IssuesList) GetTab() IssueTab { return m.tab }

func (m *IssuesList) SetTab(tab IssueTab) {
	m.tab = tab
	m.applyFilter()
}

func (m *IssuesList) applyFilter() {
	// Start from all issues, apply tab filter first.
	source := m.allIssues
	if m.tab == IssueTabAssigned && m.userEmail != "" {
		var assigned []jira.Issue
		for _, issue := range source {
			if issue.Assignee != nil && strings.EqualFold(issue.Assignee.Email, m.userEmail) {
				assigned = append(assigned, issue)
			}
		}
		source = assigned
	}

	// Then apply text search filter.
	if m.filter == "" {
		m.issues = source
	} else {
		q := strings.ToLower(m.filter)
		var filtered []jira.Issue
		for _, issue := range source {
			haystack := strings.ToLower(issue.Key + " " + issue.Summary)
			if issue.Assignee != nil {
				haystack += " " + strings.ToLower(issue.Assignee.DisplayName)
			}
			if strings.Contains(haystack, q) {
				filtered = append(filtered, issue)
			}
		}
		m.issues = filtered
	}
	m.cursor = 0
	m.offset = 0
}

func (m *IssuesList) SetSize(w, h int) { m.width = w; m.height = h }

// ContentHeight returns natural height: items + 2 borders. Min 7 before data loads.
func (m *IssuesList) ContentHeight() int {
	return max(len(m.issues)+2, 7)
}
func (m *IssuesList) SetFocused(focused bool) { m.focused = focused }

func (m *IssuesList) ScrollBy(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.issues) {
		m.cursor = len(m.issues) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.adjustOffset()
}

func (m *IssuesList) ClickAt(relY int) {
	// relY is relative to the top of this panel.
	idx := m.offset + relY - 1 // -1 for top border
	if idx >= 0 && idx < len(m.issues) {
		m.cursor = idx
		m.adjustOffset()
	}
}

func (m *IssuesList) SelectedIssue() *jira.Issue {
	if len(m.issues) == 0 || m.cursor < 0 || m.cursor >= len(m.issues) {
		return nil
	}
	return &m.issues[m.cursor]
}

func (m *IssuesList) Init() tea.Cmd { return nil }

func (m *IssuesList) Update(msg tea.Msg) (*IssuesList, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		prevCursor := m.cursor
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "g", "home":
			m.cursor = 0
		case "G", "end":
			if len(m.issues) > 0 {
				m.cursor = len(m.issues) - 1
			}
		case "ctrl+d":
			m.cursor += m.visibleRows() / 2
			if m.cursor >= len(m.issues) {
				m.cursor = len(m.issues) - 1
			}
		case "ctrl+u":
			m.cursor -= m.visibleRows() / 2
			if m.cursor < 0 {
				m.cursor = 0
			}
		}
		m.adjustOffset()
		if prevCursor != m.cursor {
			return m, func() tea.Msg {
				return IssueSelectedMsg{Issue: m.SelectedIssue()}
			}
		}
	}
	return m, nil
}

func (m *IssuesList) visibleRows() int {
	return max(m.height-2, 1) // top + bottom border
}

func (m *IssuesList) adjustOffset() {
	m.offset = components.AdjustOffset(m.cursor, m.offset, m.visibleRows(), len(m.issues))
}

func (m *IssuesList) View() string {
	contentWidth := max(m.width-2, 10)
	visible := m.visibleRows()

	var rows []string
	end := min(m.offset+visible, len(m.issues))
	for i := m.offset; i < end; i++ {
		rows = append(rows, m.renderIssueRow(m.issues[i], contentWidth, i == m.cursor))
	}

	content := strings.Join(rows, "\n")
	title := m.buildTitle()
	footer := ""
	if len(m.issues) > 0 {
		footer = fmt.Sprintf("%d of %d", m.cursor+1, len(m.issues))
	}
	scroll := &components.ScrollInfo{Total: len(m.issues), Visible: visible, Offset: m.offset}
	return components.RenderPanelFull(title, footer, content, m.width, visible, m.focused, scroll)
}

// ClickTabAt handles clicks on the title bar to switch All/Assigned tabs.
func (m *IssuesList) ClickTabAt(x int) {
	// Title: "[2] All - Assigned"
	// "[2] " = 4 chars. "All" starts at 4, " - " at 7, "Assigned" at 10.
	// Zone: x < midpoint → All, x >= midpoint → Assigned.
	prefix := 4 // "[2] "
	allW := 3   // "All"
	sepW := 3   // " - "
	mid := prefix + allW + sepW/2

	if x >= prefix {
		if x < mid {
			if m.tab != IssueTabAll {
				m.tab = IssueTabAll
				m.applyFilter()
			}
		} else {
			if m.tab != IssueTabAssigned {
				m.tab = IssueTabAssigned
				m.applyFilter()
			}
		}
	}
}

func (m *IssuesList) buildTitle() string {
	active := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true)
	inactive := lipgloss.NewStyle().Foreground(theme.ColorWhite)
	sep := lipgloss.NewStyle().Foreground(theme.ColorGray).Render(" - ")

	// Count assigned issues.
	assignedCount := 0
	for _, issue := range m.allIssues {
		if issue.Assignee != nil && strings.EqualFold(issue.Assignee.Email, m.userEmail) {
			assignedCount++
		}
	}

	allLabel := "All"
	assignedLabel := "Assigned"

	if m.tab == IssueTabAll {
		allLabel = active.Render(allLabel)
		assignedLabel = inactive.Render(assignedLabel)
	} else {
		allLabel = inactive.Render(allLabel)
		assignedLabel = active.Render(assignedLabel)
	}

	return "[2] " + allLabel + sep + assignedLabel
}

func (m *IssuesList) renderIssueRow(issue jira.Issue, width int, selected bool) string {
	key := issue.Key

	var emoji string
	if selected {
		emoji = statusEmojiPlain(issue.Status)
	} else {
		emoji = statusEmoji(issue.Status)
	}

	// Pad key to fixed column width.
	keyPad := max(m.keyColWidth-lipgloss.Width(key), 0)
	paddedKey := key + strings.Repeat(" ", keyPad)

	separators := 4 // leading space + space after key + space after emoji + trailing
	summaryWidth := max(width-m.keyColWidth-1-separators, 5)

	summary := truncateRunes(issue.Summary, summaryWidth)

	active := issue.Key == m.activeKey
	marker := " "
	if active {
		marker = "*"
	}

	if selected && m.focused {
		line := fmt.Sprintf("%s%s %s %s", marker, paddedKey, emoji, summary)
		return m.theme.SelectedItem.Width(width).Render(line)
	}
	line := fmt.Sprintf("%s%s %s %s", marker, paddedKey, emoji, summary)
	return m.theme.NormalItem.Width(width).Render(line)
}

// truncateRunes truncates a string to fit within maxWidth display columns,
// respecting multi-byte UTF-8 characters (Cyrillic, etc).
func truncateRunes(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i])
		if lipgloss.Width(candidate)+1 <= maxWidth { // +1 for "…"
			return candidate + "…"
		}
	}
	return "…"
}

// statusEmojiPlain returns uncolored status char for selected rows.
func statusEmojiPlain(status *jira.Status) string {
	if status == nil {
		return statusOpen
	}
	switch status.CategoryKey {
	case "done":
		return "✓"
	case "indeterminate":
		return "→"
	default:
		return statusOpen
	}
}

func statusEmoji(status *jira.Status) string {
	if status == nil {
		return statusOpen
	}
	switch status.CategoryKey {
	case "done":
		return theme.StatusColor("done").Render("✓")
	case "indeterminate":
		return theme.StatusColor("indeterminate").Render("→")
	default:
		return theme.StatusColor("new").Render("○")
	}
}
