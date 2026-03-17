package views

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/pkg/jira"
	"github.com/textfuel/lazyjira/pkg/tui/components"
	"github.com/textfuel/lazyjira/pkg/tui/theme"
)

type DetailTab int

const (
	TabDetails  DetailTab = iota
	TabSubtasks
	TabComments
	TabLinks
	TabInfo
	tabCount = 5
)

// MainMode controls what the right panel displays.
type MainMode int

const (
	ModeIssue   MainMode = iota
	ModeSplash
	ModeProject
)

// SplashInfo holds data for the splash/status screen.
type SplashInfo struct {
	AuthMethod string // "API Token" or "OAuth" or "env vars"
	Host       string
	Email      string
	Project    string
}

type DetailView struct {
	issue     *jira.Issue
	project   *jira.Project
	splash    SplashInfo
	mode      MainMode
	activeTab DetailTab
	scrollY   int
	width     int
	height    int
	focused   bool
	theme     *theme.Theme
}

func NewDetailView() *DetailView {
	return &DetailView{theme: theme.DefaultTheme(), mode: ModeIssue}
}

func (d *DetailView) SetIssue(issue *jira.Issue) {
	d.issue = issue
	d.mode = ModeIssue
	d.scrollY = 0
	d.activeTab = TabDetails
}

func (d *DetailView) SetProject(project *jira.Project) {
	d.project = project
	d.mode = ModeProject
	d.scrollY = 0
}

func (d *DetailView) SetSplash(info SplashInfo) {
	d.splash = info
	d.mode = ModeSplash
	d.scrollY = 0
}

func (d *DetailView) SetSize(w, h int)       { d.width = w; d.height = h }
func (d *DetailView) SetFocused(focused bool) { d.focused = focused }
func (d *DetailView) Init() tea.Cmd           { return nil }

func (d *DetailView) NextTab() {
	vt := d.visibleTabs()
	for i, t := range vt {
		if t == d.activeTab {
			d.activeTab = vt[(i+1)%len(vt)]
			d.scrollY = 0
			return
		}
	}
	if len(vt) > 0 {
		d.activeTab = vt[0]
		d.scrollY = 0
	}
}

func (d *DetailView) PrevTab() {
	vt := d.visibleTabs()
	for i, t := range vt {
		if t == d.activeTab {
			d.activeTab = vt[(i+len(vt)-1)%len(vt)]
			d.scrollY = 0
			return
		}
	}
	if len(vt) > 0 {
		d.activeTab = vt[0]
		d.scrollY = 0
	}
}

func (d *DetailView) visibleTabs() []DetailTab {
	tabs := []DetailTab{TabDetails}
	if d.issue != nil {
		if len(d.issue.Subtasks) > 0 {
			tabs = append(tabs, TabSubtasks)
		}
		if len(d.issue.Comments) > 0 {
			tabs = append(tabs, TabComments)
		}
		if len(d.issue.IssueLinks) > 0 {
			tabs = append(tabs, TabLinks)
		}
	}
	tabs = append(tabs, TabInfo)
	return tabs
}

func (d *DetailView) ScrollBy(delta int) {
	d.scrollY += delta
	if d.scrollY < 0 {
		d.scrollY = 0
	}
}

func (d *DetailView) Update(msg tea.Msg) (*DetailView, tea.Cmd) {
	if !d.focused {
		return d, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			d.scrollY++
		case "k", "up":
			if d.scrollY > 0 {
				d.scrollY--
			}
		case "tab":
			d.activeTab = (d.activeTab + 1) % tabCount
			d.scrollY = 0
		case "i":
			d.activeTab = TabInfo
			d.scrollY = 0
		case "ctrl+d":
			d.scrollY += d.visibleRows() / 2
		case "ctrl+u":
			d.scrollY -= d.visibleRows() / 2
			if d.scrollY < 0 {
				d.scrollY = 0
			}
		}
	}
	return d, nil
}

func (d *DetailView) visibleRows() int {
	// Total height = innerHeight + 2 (borders).
	// innerHeight = tabs line + scrollable content lines.
	// So scrollable = height - 2 (borders) - 1 (tabs) = height - 3.
	rows := d.height - 3
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (d *DetailView) View() string {
	contentWidth := d.width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}
	innerH := d.height - 2
	if innerH < 1 {
		innerH = 1
	}

	// Splash mode.
	if d.mode == ModeSplash {
		return d.renderSplash(contentWidth, innerH)
	}

	// Project mode.
	if d.mode == ModeProject && d.project != nil {
		return d.renderProjectView(contentWidth, innerH)
	}

	visible := d.visibleRows()

	// Issue mode.
	title := "[0] Detail"
	if d.issue != nil {
		title = fmt.Sprintf("[0] %s: %s", d.issue.Key, d.issue.Summary)
		title = truncateRunes(title, contentWidth-2)
	}

	if d.issue == nil {
		placeholder := lipgloss.NewStyle().Foreground(theme.ColorGray).Render("Select an issue to view details")
		return components.RenderPanel(title, placeholder, d.width, innerH, d.focused)
	}

	// Tabs line.
	tabs := d.renderTabs(contentWidth)

	// Content lines.
	var contentLines []string
	switch d.activeTab {
	case TabDetails:
		contentLines = d.renderDescription(contentWidth)
	case TabSubtasks:
		contentLines = d.renderSubtasks(contentWidth)
	case TabComments:
		contentLines = d.renderComments(contentWidth)
	case TabLinks:
		contentLines = d.renderLinks(contentWidth)
	case TabInfo:
		contentLines = d.renderInfo(contentWidth)
	}

	// Apply scroll.
	if d.scrollY > len(contentLines) {
		d.scrollY = len(contentLines)
	}
	if d.scrollY < 0 {
		d.scrollY = 0
	}
	scrolled := contentLines
	if d.scrollY < len(scrolled) {
		scrolled = scrolled[d.scrollY:]
	} else {
		scrolled = nil
	}
	if len(scrolled) > visible {
		scrolled = scrolled[:visible]
	}

	// Build content: tabs + body.
	body := tabs + "\n" + strings.Join(scrolled, "\n")

	return components.RenderPanel(title, body, d.width, innerH, d.focused)
}

func (d *DetailView) renderTabs(width int) string {
	activeStyle := lipgloss.NewStyle().
		Foreground(theme.ColorGreen).
		Bold(true).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(theme.ColorGray).
		Padding(0, 1)

	// Build visible tabs with counters. Hide tabs with 0 items.
	type tabDef struct {
		tab   DetailTab
		label string
	}
	var tabs []tabDef
	tabs = append(tabs, tabDef{TabDetails, "Details"})
	if d.issue != nil {
		if n := len(d.issue.Subtasks); n > 0 {
			tabs = append(tabs, tabDef{TabSubtasks, fmt.Sprintf("Subtasks(%d)", n)})
		}
		if n := len(d.issue.Comments); n > 0 {
			tabs = append(tabs, tabDef{TabComments, fmt.Sprintf("Comments(%d)", n)})
		}
		if n := len(d.issue.IssueLinks); n > 0 {
			tabs = append(tabs, tabDef{TabLinks, fmt.Sprintf("Links(%d)", n)})
		}
	}
	tabs = append(tabs, tabDef{TabInfo, "Info"})

	var parts []string
	for _, t := range tabs {
		if t.tab == d.activeTab {
			parts = append(parts, activeStyle.Render(t.label))
		} else {
			parts = append(parts, inactiveStyle.Render(t.label))
		}
	}
	return " " + lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func (d *DetailView) renderDescription(width int) []string {
	valStyle := d.theme.ValueStyle
	var lines []string

	desc := d.issue.Description
	if desc == "" {
		desc = "(no description)"
	}
	for _, line := range wrapText(desc, width-2) {
		lines = append(lines, " "+valStyle.Render(line))
	}
	return lines
}

func (d *DetailView) renderSubtasks(width int) []string {
	if d.issue == nil || len(d.issue.Subtasks) == 0 {
		return []string{" No subtasks."}
	}
	var lines []string
	for _, sub := range d.issue.Subtasks {
		emoji := statusEmoji(sub.Status)
		lines = append(lines, fmt.Sprintf(" %s %s: %s", emoji, sub.Key, sub.Summary))
	}
	return lines
}

func (d *DetailView) renderInfo(width int) []string {
	issue := d.issue
	valStyle := d.theme.ValueStyle

	var lines []string

	statusName := "Unknown"
	if issue.Status != nil {
		statusName = theme.StatusColor(issue.Status.CategoryKey).Render(issue.Status.Name)
	}
	priorityName := "None"
	if issue.Priority != nil {
		priorityName = d.priorityStyled(issue.Priority.Name)
	}
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Status:", statusName))
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Priority:", priorityName))

	assignee := "Unassigned"
	if issue.Assignee != nil {
		assignee = issue.Assignee.DisplayName
	}
	reporter := "Unknown"
	if issue.Reporter != nil {
		reporter = issue.Reporter.DisplayName
	}
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Assignee:", valStyle.Render(assignee)))
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Reporter:", valStyle.Render(reporter)))

	typeName := "Unknown"
	if issue.IssueType != nil {
		typeName = issue.IssueType.Name
	}
	sprintName := "None"
	if issue.Sprint != nil {
		sprintName = issue.Sprint.Name
	}
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Type:", valStyle.Render(typeName)))
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Sprint:", valStyle.Render(sprintName)))

	if len(issue.Labels) > 0 {
		lines = append(lines, fmt.Sprintf(" %-11s %s", "Labels:", valStyle.Render(strings.Join(issue.Labels, ", "))))
	}

	if len(issue.Components) > 0 {
		var names []string
		for _, c := range issue.Components {
			names = append(names, c.Name)
		}
		lines = append(lines, fmt.Sprintf(" %-11s %s", "Components:", valStyle.Render(strings.Join(names, ", "))))
	}

	return lines
}

func (d *DetailView) renderComments(width int) []string {
	if d.issue == nil {
		return []string{" No issue selected."}
	}
	if len(d.issue.Comments) == 0 {
		return []string{" No comments."}
	}

	keyStyle := d.theme.KeyStyle
	valStyle := d.theme.ValueStyle
	var lines []string
	for i, c := range d.issue.Comments {
		author := "Unknown"
		if c.Author != nil {
			author = c.Author.DisplayName
		}
		ago := timeAgo(c.Created)
		lines = append(lines, " "+keyStyle.Render(fmt.Sprintf("%s (%s):", author, ago)))
		for _, wl := range wrapText(c.Body, width-2) {
			lines = append(lines, " "+valStyle.Render(wl))
		}
		if i < len(d.issue.Comments)-1 {
			lines = append(lines, " "+strings.Repeat("─", width/2))
		}
	}
	return lines
}

func (d *DetailView) renderLinks(width int) []string {
	if d.issue == nil {
		return []string{" No issue selected."}
	}
	if len(d.issue.IssueLinks) == 0 {
		return []string{" No links."}
	}

	keyStyle := d.theme.KeyStyle
	valStyle := d.theme.ValueStyle
	var lines []string
	for _, link := range d.issue.IssueLinks {
		if link.Type == nil {
			continue
		}
		if link.OutwardIssue != nil {
			lines = append(lines, " "+
				keyStyle.Render(link.Type.Outward)+" "+
				valStyle.Render(fmt.Sprintf("%s: %s", link.OutwardIssue.Key, link.OutwardIssue.Summary)))
		}
		if link.InwardIssue != nil {
			lines = append(lines, " "+
				keyStyle.Render(link.Type.Inward)+" "+
				valStyle.Render(fmt.Sprintf("%s: %s", link.InwardIssue.Key, link.InwardIssue.Summary)))
		}
	}
	if len(lines) == 0 {
		lines = append(lines, " No links.")
	}
	return lines
}

func (d *DetailView) renderSplash(contentWidth, innerH int) string {
	green := lipgloss.NewStyle().Foreground(theme.ColorGreen).Bold(true)
	gray := lipgloss.NewStyle().Foreground(theme.ColorGray)
	label := lipgloss.NewStyle().Foreground(theme.ColorGreen)
	val := lipgloss.NewStyle()

	ascii := `   _                  _ _
  | |                (_|_)
  | | __ _ _____   _  _ _ _ __ __ _
  | |/ _` + "`" + ` |_  / | | || | | '__/ _` + "`" + ` |
  | | (_| |/ /| |_| || | | | | (_| |
  |_|\__,_/___|\__, || |_|_|  \__,_|
                __/ |/ |
               |___/__/`

	var lines []string
	for _, l := range strings.Split(ascii, "\n") {
		lines = append(lines, green.Render(l))
	}
	lines = append(lines, "")
	lines = append(lines, gray.Render("  lazyjira — Jira TUI"))
	lines = append(lines, gray.Render("  (c) 2026 Andrey Kondratev"))

	// Connection info.
	s := d.splash
	lines = append(lines, "")
	lines = append(lines, "  "+strings.Repeat("─", 30))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s  %s", label.Render("Auth:"), val.Render(s.AuthMethod)))
	lines = append(lines, fmt.Sprintf("  %s  %s", label.Render("Host:"), val.Render(s.Host)))
	lines = append(lines, fmt.Sprintf("  %s %s", label.Render("Email:"), val.Render(s.Email)))
	if s.Project != "" {
		lines = append(lines, fmt.Sprintf("  %s  %s", label.Render("Project:"), val.Render(s.Project)))
	}

	content := strings.Join(lines, "\n")
	return components.RenderPanel("[0] lazyjira", content, d.width, innerH, d.focused)
}

func (d *DetailView) renderProjectView(contentWidth, innerH int) string {
	p := d.project
	valStyle := d.theme.ValueStyle

	var lines []string
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Key:", valStyle.Render(p.Key)))
	lines = append(lines, fmt.Sprintf(" %-11s %s", "Name:", valStyle.Render(p.Name)))
	if p.Lead != nil {
		lines = append(lines, fmt.Sprintf(" %-11s %s", "Lead:", valStyle.Render(p.Lead.DisplayName)))
	}

	content := strings.Join(lines, "\n")
	title := fmt.Sprintf("[0] Project: %s", p.Name)
	title = truncateRunes(title, contentWidth-2)
	return components.RenderPanel(title, content, d.width, innerH, d.focused)
}

func (d *DetailView) priorityStyled(name string) string {
	switch strings.ToLower(name) {
	case "highest", "high", "critical", "blocker":
		return d.theme.PriorityHigh.Render(name)
	case "medium":
		return d.theme.PriorityMedium.Render(name)
	default:
		return d.theme.PriorityLow.Render(name)
	}
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		width = 80
	}
	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		if len(paragraph) <= width {
			lines = append(lines, paragraph)
			continue
		}
		for len(paragraph) > width {
			cut := width
			for cut > 0 && paragraph[cut] != ' ' {
				cut--
			}
			if cut == 0 {
				cut = width
			}
			lines = append(lines, paragraph[:cut])
			paragraph = strings.TrimLeft(paragraph[cut:], " ")
		}
		if paragraph != "" {
			lines = append(lines, paragraph)
		}
	}
	return lines
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dmo ago", int(d.Hours()/(24*30)))
	}
}
