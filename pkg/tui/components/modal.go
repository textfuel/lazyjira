package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ModalItem is one option in the modal.
type ModalItem struct {
	ID       string
	Label    string
	Internal bool // true = handled in-app (e.g. Jira issue), styled differently
}

// ModalSelectedMsg is sent when user picks an item.
type ModalSelectedMsg struct {
	Item ModalItem
}

// ModalCancelledMsg is sent when user presses Esc.
type ModalCancelledMsg struct{}

// Modal is a centered popup list for picking an option (transitions, etc).
type Modal struct {
	title   string
	items   []ModalItem
	cursor  int
	visible  bool
	readOnly bool // scroll-only, no selection highlight
	offset   int
	width    int
	height   int
}

func NewModal() Modal {
	return Modal{}
}

func (m *Modal) Show(title string, items []ModalItem) {
	m.title = title
	m.items = items
	m.cursor = 0
	m.offset = 0
	m.visible = true
	m.readOnly = false
}

func (m *Modal) ShowReadOnly(title string, items []ModalItem) {
	m.title = title
	m.items = items
	m.cursor = 0
	m.offset = 0
	m.visible = true
	m.readOnly = true
}

func (m *Modal) Hide()          { m.visible = false }
func (m *Modal) IsVisible() bool { return m.visible }

func (m *Modal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if !m.visible {
		return *m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.readOnly {
				m.offset++
			} else if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.readOnly {
				if m.offset > 0 {
					m.offset--
				}
			} else if m.cursor > 0 {
				m.cursor--
			}
		case "enter", " ":
			if m.readOnly {
				m.visible = false
				return *m, func() tea.Msg { return ModalCancelledMsg{} }
			}
			if m.cursor >= 0 && m.cursor < len(m.items) {
				selected := m.items[m.cursor]
				m.visible = false
				return *m, func() tea.Msg { return ModalSelectedMsg{Item: selected} }
			}
		case "esc", "q", "h":
			m.visible = false
			return *m, func() tea.Msg { return ModalCancelledMsg{} }
		}
	case tea.MouseMsg:
		switch {
		case msg.Button == tea.MouseButtonWheelDown:
			if m.readOnly {
				m.offset++
			} else if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case msg.Button == tea.MouseButtonWheelUp:
			if m.readOnly {
				if m.offset > 0 {
					m.offset--
				}
			} else if m.cursor > 0 {
				m.cursor--
			}
		case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
			if !m.readOnly {
				// Click on item: title=1 line + blank=1 line, items start at y offset ~2 from modal top.
				// Approximate: map click Y to item index.
				// Modal is centered, so we use relative positioning.
				clickY := msg.Y
				// Items start after title + blank (2 lines) + modal border (1) + centering offset.
				// Simple approach: just select based on relative position in items area.
				idx := clickY - 3 // rough: border + title + blank
				if m.height > 0 {
					// Adjust for centering.
					modalH := min(len(m.items)+5, m.height-2) // title + blank + items + blank + hint
					topOffset := (m.height - modalH) / 2
					idx = clickY - topOffset - 3
				}
				if idx >= 0 && idx < len(m.items) {
					m.cursor = idx
					selected := m.items[m.cursor]
					m.visible = false
					return *m, func() tea.Msg { return ModalSelectedMsg{Item: selected} }
				}
			}
		}
	}
	return *m, nil
}

func (m *Modal) View() string {
	if !m.visible || len(m.items) == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Bold(true)

	internalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	if m.readOnly {
		// Collect item lines only (title is in panel border).
		var lines []string
		for _, item := range m.items {
			lines = append(lines, " "+item.Label)
		}
		totalLines := len(lines)
		visibleH := max(m.height-2, 3)
		if m.offset > totalLines-visibleH {
			m.offset = totalLines - visibleH
		}
		if m.offset < 0 {
			m.offset = 0
		}
		scrolled := lines
		if m.offset < len(scrolled) {
			scrolled = scrolled[m.offset:]
		}
		if len(scrolled) > visibleH {
			scrolled = scrolled[:visibleH]
		}
		content := strings.Join(scrolled, "\n")
		return RenderPanelFull(m.title, "", content, m.width, visibleH, true,
			&ScrollInfo{Total: totalLines, Visible: visibleH, Offset: m.offset})
	}

	// Calculate content width from longest item.
	contentW := lipgloss.Width(m.title) + 2
	for _, item := range m.items {
		if w := lipgloss.Width(item.Label) + 2; w > contentW {
			contentW = w
		}
	}
	maxW := min(55, m.width-6)
	if contentW > maxW {
		contentW = maxW
	}

	// Normal modal (selection) — no hints, title + blank + items.
	var lines []string
	lines = append(lines, " "+titleStyle.Render(m.title))
	lines = append(lines, "")

	for i, item := range m.items {
		label := " " + item.Label
		switch {
		case i == m.cursor:
			lines = append(lines, lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("4")).Width(contentW).Render(label))
		case item.Internal:
			lines = append(lines, internalStyle.Width(contentW).Render(label))
		default:
			lines = append(lines, lipgloss.NewStyle().Width(contentW).Render(label))
		}
	}

	popupH := len(lines)
	maxH := max(m.height-2, 5)
	if popupH > maxH {
		popupH = maxH
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("2")).
		Width(contentW).
		Height(popupH).
		Render(strings.Join(lines, "\n"))
}
