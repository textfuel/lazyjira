package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/pkg/config"
	"github.com/textfuel/lazyjira/pkg/jira"
	"github.com/textfuel/lazyjira/pkg/tui/components"
	"github.com/textfuel/lazyjira/pkg/tui/views"
)

type focusPanel int

const (
	focusStatus   focusPanel = iota
	focusIssues
	focusProjects
)

type focusSide int

const (
	sideLeft focusSide = iota
	sideRight
)

// Async messages.
type issuesLoadedMsg struct{ issues []jira.Issue }
type issueDetailLoadedMsg struct{ issue *jira.Issue }
type transitionDoneMsg struct{}
type errorMsg struct{ err error }
type projectsLoadedMsg struct{ projects []jira.Project }

type App struct {
	cfg        *config.Config
	client     *jira.Client
	splashInfo views.SplashInfo

	statusPanel *views.StatusPanel
	issuesList  *views.IssuesList
	projectList *views.ProjectList
	detailView  *views.DetailView
	logPanel    *views.LogPanel

	helpBar   components.HelpBar
	searchBar components.SearchBar

	side       focusSide
	leftFocus  focusPanel
	projectKey string
	showHelp   bool // ? popup visible

	width  int
	height int
	err    error
}

// AuthMethod describes how the user authenticated.
type AuthMethod string

const (
	AuthSaved AuthMethod = "Saved credentials (auth.json)"
	AuthEnv   AuthMethod = "Environment variables"
	AuthWizard AuthMethod = "Setup wizard"
)

func NewApp(cfg *config.Config, client *jira.Client) *App {
	return NewAppWithAuth(cfg, client, AuthEnv)
}

func NewAppWithAuth(cfg *config.Config, client *jira.Client, authMethod AuthMethod) *App {
	projectKey := ""
	if len(cfg.Projects) > 0 {
		projectKey = cfg.Projects[0].Key
	}

	statusPanel := views.NewStatusPanel(projectKey, cfg.Jira.Email, cfg.Jira.Host)
	issuesList := views.NewIssuesList()
	issuesList.SetFocused(true)
	projectList := views.NewProjectList()
	detailView := views.NewDetailView()
	logPanel := views.NewLogPanel()
	helpBar := components.NewHelpBar(nil)
	searchBar := components.NewSearchBar()

	client.SetOnRequest(func(rl jira.RequestLog) {
		logPanel.AddEntry(views.LogEntry{
			Time:    time.Now(),
			Method:  rl.Method,
			Path:    rl.Path,
			Status:  rl.Status,
			Elapsed: rl.Elapsed,
		})
	})

	splash := views.SplashInfo{
		AuthMethod: string(authMethod),
		Host:       cfg.Jira.Host,
		Email:      cfg.Jira.Email,
		Project:    projectKey,
	}

	app := &App{
		cfg:        cfg,
		client:     client,
		splashInfo: splash,
		statusPanel: statusPanel,
		issuesList:  issuesList,
		projectList: projectList,
		detailView:  detailView,
		logPanel:    logPanel,
		helpBar:     helpBar,
		searchBar:   searchBar,
		side:        sideLeft,
		leftFocus:   focusIssues,
		projectKey:  projectKey,
	}
	app.helpBar.SetItems(app.helpBarItems())
	return app
}

func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, fetchProjects(a.client))
	if a.projectKey != "" {
		cmds = append(cmds, fetchIssues(a.client, a.projectKey))
	}
	return tea.Batch(cmds...)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Search bar intercepts all keys when active.
	if a.searchBar.IsActive() {
		if _, ok := msg.(tea.KeyMsg); ok {
			updated, cmd := a.searchBar.Update(msg)
			a.searchBar = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return a, tea.Batch(cmds...)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.layoutPanels()
		a.helpBar.SetWidth(msg.Width)
		a.searchBar.SetWidth(msg.Width)
		return a, nil

	case tea.MouseMsg:
		return a.handleMouse(msg)

	case components.SearchChangedMsg:
		a.issuesList.SetFilter(msg.Query)
		a.projectList.SetFilter(msg.Query)
		return a, nil

	case components.SearchConfirmedMsg:
		// Select best match (current top item), then reset filter.
		var cmd tea.Cmd
		if a.side == sideLeft && a.leftFocus == focusProjects {
			// Trigger enter on filtered project list.
			updated, c := a.projectList.Update(tea.KeyMsg{Type: tea.KeyEnter})
			a.projectList = updated
			cmd = c
		}
		a.issuesList.SetFilter("")
		a.projectList.SetFilter("")
		if cmd != nil {
			return a, cmd
		}
		return a, nil

	case components.SearchCancelledMsg:
		a.issuesList.SetFilter("")
		a.projectList.SetFilter("")
		return a, nil

	case tea.KeyMsg:
		// Help popup: any key closes it (except / for search).
		if a.showHelp {
			if msg.String() == "/" {
				a.showHelp = false
				a.searchBar.Activate()
				return a, nil
			}
			a.showHelp = false
			return a, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit

		case "?":
			a.showHelp = true
			return a, nil

		case "/":
			a.searchBar.Activate()
			return a, nil

		case "tab":
			if a.side == sideLeft {
				a.side = sideRight
			} else {
				a.side = sideLeft
			}
			a.updateFocusState()
			return a, nil

		case "l", "right":
			if a.side == sideLeft {
				a.side = sideRight
				a.updateFocusState()
				return a, nil
			}

		case "h", "left":
			if a.side == sideRight {
				a.side = sideLeft
				a.updateFocusState()
				return a, nil
			}

		case "enter":
			if a.side == sideLeft && a.leftFocus == focusIssues {
				a.side = sideRight
				a.updateFocusState()
				if sel := a.issuesList.SelectedIssue(); sel != nil {
					return a, fetchIssueDetail(a.client, sel.Key)
				}
				return a, nil
			}

		case "[":
			if a.side == sideRight {
				a.detailView.PrevTab()
				return a, nil
			}
		case "]":
			if a.side == sideRight {
				a.detailView.NextTab()
				return a, nil
			}

		case "0":
			a.side = sideRight
			a.updateFocusState()
			return a, nil

		case "1":
			a.side = sideLeft
			a.leftFocus = focusStatus
			// Update splash with current project.
			a.splashInfo.Project = a.projectKey
			a.detailView.SetSplash(a.splashInfo)
			a.updateFocusState()
			return a, nil
		case "2":
			a.side = sideLeft
			a.leftFocus = focusIssues
			// Restore issue detail.
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				a.updateFocusState()
				return a, fetchIssueDetail(a.client, sel.Key)
			}
			a.updateFocusState()
			return a, nil
		case "3":
			a.side = sideLeft
			a.leftFocus = focusProjects
			a.updateFocusState()
			return a, nil

		case "r":
			if a.projectKey != "" {
				return a, fetchIssues(a.client, a.projectKey)
			}
			return a, nil
		}

	case issuesLoadedMsg:
		a.err = nil
		a.statusPanel.SetOnline(true)
		a.issuesList.SetIssues(msg.issues)
		if sel := a.issuesList.SelectedIssue(); sel != nil {
			cmds = append(cmds, fetchIssueDetail(a.client, sel.Key))
		}
		return a, tea.Batch(cmds...)

	case issueDetailLoadedMsg:
		a.detailView.SetIssue(msg.issue)
		return a, nil

	case transitionDoneMsg:
		if a.projectKey != "" {
			return a, fetchIssues(a.client, a.projectKey)
		}
		return a, nil

	case errorMsg:
		a.err = msg.err
		a.statusPanel.SetOnline(false)
		return a, nil

	case views.IssueSelectedMsg:
		if msg.Issue != nil {
			return a, fetchIssueDetail(a.client, msg.Issue.Key)
		}
		return a, nil

	case projectsLoadedMsg:
		a.projectList.SetProjects(msg.projects)
		if a.projectKey == "" && len(msg.projects) > 0 {
			a.projectKey = msg.projects[0].Key
			a.statusPanel.SetProject(a.projectKey)
			return a, fetchIssues(a.client, a.projectKey)
		}
		return a, nil

	case views.ProjectHoveredMsg:
		if msg.Project != nil {
			a.detailView.SetProject(msg.Project)
		}
		return a, nil

	case views.ProjectSelectedMsg:
		a.projectKey = msg.ProjectKey
		a.statusPanel.SetProject(msg.ProjectKey)
		a.leftFocus = focusIssues
		a.updateFocusState()
		return a, fetchIssues(a.client, a.projectKey)
	}

	// Route input to focused panel.
	if a.side == sideLeft {
		switch a.leftFocus {
		case focusIssues:
			updated, cmd := a.issuesList.Update(msg)
			a.issuesList = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case focusProjects:
			updated, cmd := a.projectList.Update(msg)
			a.projectList = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case focusStatus:
			updated, cmd := a.statusPanel.Update(msg)
			a.statusPanel = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	} else {
		updated, cmd := a.detailView.Update(msg)
		a.detailView = updated
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Determine which panel was clicked based on x/y coordinates.
	sideW := a.cfg.GUI.SidePanelWidth
	if sideW <= 0 {
		sideW = 40
	}
	if sideW > a.width/2 {
		sideW = a.width / 2
	}

	x, y := msg.X, msg.Y

	switch msg.Type {
	case tea.MouseWheelUp:
		if x < sideW {
			a.focusLeftPanelAt(y)
			return a.scrollFocused(-3)
		}
		a.side = sideRight
		a.updateFocusState()
		return a.scrollFocused(-3)

	case tea.MouseWheelDown:
		if x < sideW {
			a.focusLeftPanelAt(y)
			return a.scrollFocused(3)
		}
		a.side = sideRight
		a.updateFocusState()
		return a.scrollFocused(3)

	case tea.MouseLeft:
		if x < sideW {
			a.side = sideLeft
			a.focusLeftPanelAt(y)
			a.updateFocusState()
			// Click on an item in the list.
			if a.leftFocus == focusIssues {
				a.issuesList.ClickAt(y - a.statusPanelHeight())
				if sel := a.issuesList.SelectedIssue(); sel != nil {
					return a, fetchIssueDetail(a.client, sel.Key)
				}
			} else if a.leftFocus == focusProjects {
				a.projectList.ClickAt(y - a.statusPanelHeight() - a.issuesPanelHeight())
			}
		} else {
			a.side = sideRight
			a.updateFocusState()
		}
		return a, nil
	}

	return a, nil
}

func (a *App) scrollFocused(delta int) (tea.Model, tea.Cmd) {
	if a.side == sideLeft && a.leftFocus == focusIssues {
		a.issuesList.ScrollBy(delta)
		if sel := a.issuesList.SelectedIssue(); sel != nil {
			return a, fetchIssueDetail(a.client, sel.Key)
		}
	} else if a.side == sideLeft && a.leftFocus == focusProjects {
		a.projectList.ScrollBy(delta)
	} else if a.side == sideRight {
		a.detailView.ScrollBy(delta)
	}
	return a, nil
}

func (a *App) statusPanelHeight() int { return 3 }

func (a *App) issuesPanelHeight() int {
	totalH := a.height - 1
	remaining := totalH - 3 // minus status
	switch a.leftFocus {
	case focusIssues:
		return remaining - 5
	case focusProjects:
		return 5
	default:
		return remaining / 2
	}
}

func (a *App) focusLeftPanelAt(y int) {
	statusH := a.statusPanelHeight()
	issuesH := a.issuesPanelHeight()

	if y < statusH {
		a.leftFocus = focusStatus
	} else if y < statusH+issuesH {
		a.leftFocus = focusIssues
	} else {
		a.leftFocus = focusProjects
	}
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	leftCol := lipgloss.JoinVertical(lipgloss.Left,
		a.statusPanel.View(),
		a.issuesList.View(),
		a.projectList.View(),
	)

	rightCol := lipgloss.JoinVertical(lipgloss.Left,
		a.detailView.View(),
		a.logPanel.View(),
	)

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)

	if a.err != nil {
		errLine := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true).
			Width(a.width).
			Render(fmt.Sprintf(" Error: %v", a.err))
		content = lipgloss.JoinVertical(lipgloss.Left, content, errLine)
	}

	var bottomBar string
	if a.searchBar.IsActive() {
		bottomBar = a.searchBar.View()
	} else {
		bottomBar = a.helpBar.View()
	}

	full := lipgloss.JoinVertical(lipgloss.Left, content, bottomBar)

	// Help popup overlay.
	if a.showHelp {
		full = a.renderHelpOverlay(full)
	}

	return full
}

func (a *App) renderHelpOverlay(base string) string {
	bindings := a.ContextBindings()

	maxKey := 0
	for _, b := range bindings {
		if len(b.Key) > maxKey {
			maxKey = len(b.Key)
		}
	}

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	descStyle := lipgloss.NewStyle()

	var lines []string
	lines = append(lines, "")
	for _, b := range bindings {
		padded := b.Key
		for len(padded) < maxKey {
			padded += " "
		}
		lines = append(lines, "  "+keyStyle.Render(padded)+"  "+descStyle.Render(b.Description))
	}
	lines = append(lines, "")
	lines = append(lines, "  "+lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("press any key to close"))

	popupContent := strings.Join(lines, "\n")

	popupW := maxKey + 40
	if popupW > a.width-4 {
		popupW = a.width - 4
	}
	popupH := len(lines)
	if popupH > a.height-4 {
		popupH = a.height - 4
	}

	popup := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("2")).
		Width(popupW).
		Height(popupH).
		Render(popupContent)

	// Center the popup.
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, popup,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
	)
}

func (a *App) layoutPanels() {
	sideW := a.cfg.GUI.SidePanelWidth
	if sideW <= 0 {
		sideW = 40
	}
	if sideW > a.width/2 {
		sideW = a.width / 2
	}
	mainW := a.width - sideW
	totalH := a.height - 1

	statusH := 3
	remaining := totalH - statusH
	var issuesH, projectsH int

	switch a.leftFocus {
	case focusStatus:
		issuesH = remaining / 2
		projectsH = remaining - issuesH
	case focusIssues:
		projectsH = 5
		issuesH = remaining - projectsH
	case focusProjects:
		issuesH = 5
		projectsH = remaining - issuesH
	}

	if issuesH < 3 {
		issuesH = 3
	}
	if projectsH < 3 {
		projectsH = 3
	}

	a.statusPanel.SetSize(sideW, statusH)
	a.issuesList.SetSize(sideW, issuesH)
	a.projectList.SetSize(sideW, projectsH)

	logH := 8
	detailH := totalH - logH
	if detailH < 5 {
		detailH = 5
	}

	a.detailView.SetSize(mainW, detailH)
	a.logPanel.SetSize(mainW, logH)
}

func (a *App) updateFocusState() {
	a.statusPanel.SetFocused(false)
	a.issuesList.SetFocused(false)
	a.projectList.SetFocused(false)
	a.detailView.SetFocused(false)

	if a.side == sideLeft {
		switch a.leftFocus {
		case focusStatus:
			a.statusPanel.SetFocused(true)
		case focusIssues:
			a.issuesList.SetFocused(true)
		case focusProjects:
			a.projectList.SetFocused(true)
		}
	} else {
		a.detailView.SetFocused(true)
	}

	a.helpBar.SetItems(a.helpBarItems())
	a.layoutPanels()
}

// Commands.

func fetchIssues(client *jira.Client, projectKey string) tea.Cmd {
	return func() tea.Msg {
		jql := fmt.Sprintf("project = %s ORDER BY updated DESC", projectKey)
		result, err := client.SearchIssues(context.Background(), jql, 0, 50)
		if err != nil {
			return errorMsg{err: err}
		}
		return issuesLoadedMsg{issues: result.Issues}
	}
}

func fetchIssueDetail(client *jira.Client, key string) tea.Cmd {
	return func() tea.Msg {
		issue, err := client.GetIssue(context.Background(), key)
		if err != nil {
			return errorMsg{err: err}
		}
		comments, err := client.GetComments(context.Background(), key)
		if err == nil {
			issue.Comments = comments
		}
		return issueDetailLoadedMsg{issue: issue}
	}
}

func fetchProjects(client *jira.Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.GetProjects(context.Background())
		if err != nil {
			return errorMsg{err: err}
		}
		return projectsLoadedMsg{projects: projects}
	}
}

func doTransition(client *jira.Client, key, transitionID string) tea.Cmd {
	return func() tea.Msg {
		err := client.DoTransition(context.Background(), key, transitionID)
		if err != nil {
			return errorMsg{err: err}
		}
		return transitionDoneMsg{}
	}
}
