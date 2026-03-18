package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SearchBar is a bottom search input that filters the current panel.
type SearchBar struct {
	query  string
	active bool
	width  int
}

func NewSearchBar() SearchBar {
	return SearchBar{}
}

func (s *SearchBar) SetWidth(w int)   { s.width = w }
func (s *SearchBar) IsActive() bool   { return s.active }
func (s *SearchBar) Query() string    { return s.query }
func (s *SearchBar) Activate()        { s.active = true; s.query = "" }
func (s *SearchBar) Deactivate()      { s.active = false; s.query = "" }

// SearchChangedMsg is sent when the search query changes.
type SearchChangedMsg struct{ Query string }

// SearchConfirmedMsg is sent when the user presses Enter.
type SearchConfirmedMsg struct{ Query string }

// SearchCancelledMsg is sent when the user presses Esc.
type SearchCancelledMsg struct{}

func (s *SearchBar) Update(msg tea.Msg) (SearchBar, tea.Cmd) {
	if !s.active {
		return *s, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyEnter:
			s.active = false
			q := s.query
			return *s, func() tea.Msg { return SearchConfirmedMsg{Query: q} }
		case tea.KeyEsc:
			s.active = false
			s.query = ""
			return *s, func() tea.Msg { return SearchCancelledMsg{} }
		case tea.KeyBackspace:
			if len(s.query) > 0 {
				s.query = s.query[:len(s.query)-1]
				q := s.query
				return *s, func() tea.Msg { return SearchChangedMsg{Query: q} }
			}
		case tea.KeyRunes:
			s.query += msg.String()
			q := s.query
			return *s, func() tea.Msg { return SearchChangedMsg{Query: q} }
		default:
			// Ignore other key types.
		}
	}
	return *s, nil
}

func (s *SearchBar) View() string {
	if !s.active {
		return ""
	}

	prefixStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Bold(true)

	queryStyle := lipgloss.NewStyle()

	cursor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Render("█")

	return prefixStyle.Render(" /") + queryStyle.Render(s.query) + cursor
}
