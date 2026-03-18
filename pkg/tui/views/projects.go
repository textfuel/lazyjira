package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/pkg/jira"
	"github.com/textfuel/lazyjira/pkg/tui/components"
	"github.com/textfuel/lazyjira/pkg/tui/theme"
)

type ProjectSelectedMsg struct {
	ProjectKey string
}

type ProjectHoveredMsg struct {
	Project *jira.Project
}

type ProjectList struct {
	projects    []jira.Project
	allProjects []jira.Project
	filter      string
	activeKey   string // the project currently in use
	cursor      int
	offset      int
	width       int
	height      int
	focused     bool
	theme       *theme.Theme
}

func NewProjectList() *ProjectList {
	return &ProjectList{theme: theme.DefaultTheme()}
}

func (p *ProjectList) SetProjects(projects []jira.Project) {
	p.allProjects = projects
	p.applyFilter()
}

func (p *ProjectList) SetFilter(query string) {
	p.filter = query
	p.applyFilter()
}

func (p *ProjectList) applyFilter() {
	if p.filter == "" {
		p.projects = p.allProjects
	} else {
		q := strings.ToLower(p.filter)
		var filtered []jira.Project
		for _, proj := range p.allProjects {
			haystack := strings.ToLower(proj.Key + " " + proj.Name)
			if strings.Contains(haystack, q) {
				filtered = append(filtered, proj)
			}
		}
		p.projects = filtered
	}
	p.cursor = 0
}

func (p *ProjectList) SetSize(w, h int)       { p.width = w; p.height = h }
func (p *ProjectList) SetActiveKey(key string) { p.activeKey = key }

func (p *ProjectList) ContentHeight() int {
	return max(len(p.projects)+2, 5)
}
func (p *ProjectList) SetFocused(focused bool) { p.focused = focused }

func (p *ProjectList) SelectedProject() *jira.Project {
	if p.cursor >= 0 && p.cursor < len(p.projects) {
		proj := p.projects[p.cursor]
		return &proj
	}
	return nil
}
func (p *ProjectList) Init() tea.Cmd           { return nil }

func (p *ProjectList) visibleRows() int {
	return max(p.height-2, 1)
}

func (p *ProjectList) adjustOffset() {
	p.offset = components.AdjustOffset(p.cursor, p.offset, p.visibleRows(), len(p.projects))
}

func (p *ProjectList) ScrollBy(delta int) {
	p.cursor += delta
	if p.cursor < 0 {
		p.cursor = 0
	}
	if p.cursor >= len(p.projects) {
		p.cursor = len(p.projects) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
	p.adjustOffset()
}

func (p *ProjectList) ClickAt(relY int) {
	idx := p.offset + relY - 1 // -1 for top border
	if idx >= 0 && idx < len(p.projects) {
		p.cursor = idx
		p.adjustOffset()
	}
}

func (p *ProjectList) Update(msg tea.Msg) (*ProjectList, tea.Cmd) {
	if !p.focused {
		return p, nil
	}
	if msg, ok := msg.(tea.KeyMsg); ok {
		prevCursor := p.cursor
		switch msg.String() {
		case "j", "down":
			if p.cursor < len(p.projects)-1 {
				p.cursor++
			}
		case "k", "up":
			if p.cursor > 0 {
				p.cursor--
			}
		case "enter", " ":
			if p.cursor >= 0 && p.cursor < len(p.projects) {
				selected := p.projects[p.cursor]
				return p, func() tea.Msg {
					return ProjectSelectedMsg{ProjectKey: selected.Key}
				}
			}
		}
		p.adjustOffset()
		if prevCursor != p.cursor && p.cursor >= 0 && p.cursor < len(p.projects) {
			proj := p.projects[p.cursor]
			return p, func() tea.Msg {
				return ProjectHoveredMsg{Project: &proj}
			}
		}
	}
	return p, nil
}

func (p *ProjectList) View() string {
	contentWidth := max(p.width-2, 10)
	innerHeight := max(p.height-2, 1)

	var rows []string
	end := min(p.offset+innerHeight, len(p.projects))
	for i := p.offset; i < end; i++ {
		proj := p.projects[i]
		lead := ""
		if proj.Lead != nil {
			lead = " · " + proj.Lead.DisplayName
		}
		// Truncate name+lead to fit.
		maxName := contentWidth - 10 - len(lead)
		if maxName < 5 {
			lead = "" // drop lead if no space
			maxName = contentWidth - 10
		}
		namePart := proj.Name
		if len(namePart) > maxName {
			namePart = namePart[:maxName-1] + "…"
		}
		active := proj.Key == p.activeKey
		marker := " "
		if active {
			marker = "*"
		}

		line := fmt.Sprintf("%s%-8s %s%s", marker, proj.Key, namePart, lead)
		if i == p.cursor && p.focused {
			rows = append(rows, p.theme.SelectedItem.Width(contentWidth).Render(line))
		} else {
			rows = append(rows, p.theme.NormalItem.Width(contentWidth).Render(line))
		}
	}

	content := strings.Join(rows, "\n")
	footer := ""
	if len(p.projects) > 0 {
		footer = fmt.Sprintf("%d of %d", p.cursor+1, len(p.projects))
	}
	scroll := &components.ScrollInfo{Total: len(p.projects), Visible: innerHeight, Offset: p.offset}
	return components.RenderPanelFull("[3] Projects", footer, content, p.width, innerHeight, p.focused, scroll)
}
