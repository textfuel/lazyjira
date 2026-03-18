package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionStatus represents the current connection state.
type ConnectionStatus int

const (
	StatusOnline ConnectionStatus = iota
	StatusOffline
	StatusLoading
)

// StatusBar is a bottom bar showing project, user, and connection info.
type StatusBar struct {
	project    string
	user       string
	connStatus ConnectionStatus
	width      int
}

// NewStatusBar creates a status bar with the given project key and user email.
func NewStatusBar(project, user string) StatusBar {
	return StatusBar{
		project:    project,
		user:       user,
		connStatus: StatusOnline,
	}
}

// SetProject updates the displayed project key.
func (s *StatusBar) SetProject(project string) {
	s.project = project
}

// SetUser updates the displayed user.
func (s *StatusBar) SetUser(user string) {
	s.user = user
}

// SetConnectionStatus updates the connection indicator.
func (s *StatusBar) SetConnectionStatus(status ConnectionStatus) {
	s.connStatus = status
}

// SetWidth updates the status bar width.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

func (s StatusBar) Init() tea.Cmd {
	return nil
}

func (s StatusBar) Update(msg tea.Msg) (StatusBar, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		s.width = msg.Width
	}
	return s, nil
}

func (s StatusBar) View() string {
	var statusIcon string
	var statusStyle lipgloss.Style

	switch s.connStatus {
	case StatusOnline:
		statusIcon = "● online"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	case StatusOffline:
		statusIcon = "● offline"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
	case StatusLoading:
		statusIcon = "◌ loading"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCC00"))
	}

	barStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#353533"))

	projectPart := barStyle.Bold(true).Render(fmt.Sprintf(" %s ", s.project))
	userPart := barStyle.Render(fmt.Sprintf(" %s ", s.user))
	statusPart := barStyle.Render(" " + statusStyle.Render(statusIcon) + " ")

	leftContent := projectPart + barStyle.Render(" │ ") + userPart
	rightContent := statusPart

	leftWidth := lipgloss.Width(leftContent)
	rightWidth := lipgloss.Width(rightContent)

	gap := max(s.width-leftWidth-rightWidth, 0)
	filler := barStyle.Render(fmt.Sprintf("%*s", gap, ""))

	return leftContent + filler + rightContent
}
