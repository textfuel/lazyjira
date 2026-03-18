package tui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/textfuel/lazyjira/pkg/config"
	"github.com/textfuel/lazyjira/pkg/jira"
	"github.com/textfuel/lazyjira/pkg/tui/components"
	"github.com/textfuel/lazyjira/pkg/tui/views"
)

// Version is set from main at startup.
var Version = "dev"

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
type issuePrefetchedMsg struct {
	issue *jira.Issue
}
type autoFetchTickMsg struct{}
type transitionsLoadedMsg struct {
	issueKey    string
	transitions []jira.Transition
}

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
	modal     components.Modal

	side        focusSide
	leftFocus   focusPanel
	projectKey  string
	showHelp    bool
	logFlag     *bool
	issueCache  map[string]*jira.Issue

	// Cached panel sizes for mouse hit-testing.
	panelSideW    int
	panelStatusH  int
	panelIssuesH  int
	panelProjectsH int
	panelDetailH  int
	panelLogH     int

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
	issuesList.SetUserEmail(cfg.Jira.Email)
	projectList := views.NewProjectList()
	detailView := views.NewDetailView()
	logPanel := views.NewLogPanel()
	helpBar := components.NewHelpBar(nil)
	searchBar := components.NewSearchBar()
	modal := components.NewModal()

	logFlag := new(bool) // shared with closure, set via app.logRequests
	client.SetOnRequest(func(rl jira.RequestLog) {
		if *logFlag {
			logPanel.AddEntry(views.LogEntry{
				Time:    time.Now(),
				Method:  rl.Method,
				Path:    rl.Path,
				Status:  rl.Status,
				Elapsed: rl.Elapsed,
			})
		}
	})

	splash := views.SplashInfo{
		Version:    Version,
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
		modal:       modal,
		side:        sideLeft,
		leftFocus:   focusIssues,
		projectKey:  projectKey,
		logFlag:     logFlag,
		issueCache:  make(map[string]*jira.Issue),
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
	// Start autofetch timer.
	cmds = append(cmds, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return autoFetchTickMsg{}
	}))
	return tea.Batch(cmds...)
}

//nolint:gocognit // will be refactored in Phase 5
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

	// Modal intercepts keys and mouse when visible.
	if a.modal.IsVisible() {
		switch msg.(type) {
		case tea.KeyMsg, tea.MouseMsg:
			updated, cmd := a.modal.Update(msg)
			a.modal = updated
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
		a.modal.SetSize(msg.Width, msg.Height)
		a.layoutPanels()
		a.helpBar.SetWidth(msg.Width)
		a.searchBar.SetWidth(msg.Width)
		return a, nil

	case tea.MouseMsg:
		return a.handleMouse(msg)

	case components.SearchChangedMsg:
		// Filter only the active panel.
		if a.side == sideLeft && a.leftFocus == focusIssues {
			a.issuesList.SetFilter(msg.Query)
		} else if a.side == sideLeft && a.leftFocus == focusProjects {
			a.projectList.SetFilter(msg.Query)
		}
		return a, nil

	case components.SearchConfirmedMsg:
		var cmd tea.Cmd
		if a.side == sideLeft && a.leftFocus == focusIssues {
			selectedIssue := a.issuesList.SelectedIssue()
			a.issuesList.ClearFilter()
			if selectedIssue != nil {
				a.issuesList.SelectByKey(selectedIssue.Key)
				cmds = append(cmds, fetchIssueDetail(a.client, selectedIssue.Key))
			}
		} else if a.side == sideLeft && a.leftFocus == focusProjects {
			updated, c := a.projectList.Update(tea.KeyMsg{Type: tea.KeyEnter})
			a.projectList = updated
			cmd = c
			a.projectList.SetFilter("")
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

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

		case "h", "left", "esc":
			if a.side == sideRight {
				a.side = sideLeft
				a.updateFocusState()
				return a, nil
			}

		case "enter", " ":
			if a.side == sideLeft && a.leftFocus == focusIssues {
				if sel := a.issuesList.SelectedIssue(); sel != nil {
					a.issuesList.SetActiveKey(sel.Key)
					a.side = sideRight
					a.updateFocusState()
					return a, fetchIssueDetail(a.client, sel.Key)
				}
				return a, nil
			}

		case "[":
			if a.side == sideRight {
				a.detailView.PrevTab()
			} else if a.side == sideLeft && a.leftFocus == focusIssues {
				a.issuesList.PrevTab()
			}
			return a, nil
		case "]":
			if a.side == sideRight {
				a.detailView.NextTab()
			} else if a.side == sideLeft && a.leftFocus == focusIssues {
				a.issuesList.NextTab()
			}
			return a, nil

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
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				a.showCachedIssue(sel.Key)
			}
			a.updateFocusState()
			return a, nil
		case "3":
			a.side = sideLeft
			a.leftFocus = focusProjects
			a.updateFocusState()
			return a, nil

		case "y":
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				copyToClipboard(a.cfg.Jira.Host + "/browse/" + sel.Key)
			}
			return a, nil

		case "o":
			if sel := a.issuesList.SelectedIssue(); sel != nil && (a.leftFocus == focusIssues || a.side == sideRight) {
				openBrowser(a.cfg.Jira.Host + "/browse/" + sel.Key)
			}
			return a, nil

		case "u":
			// URL picker modal.
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				cached := sel
				if c, ok := a.issueCache[sel.Key]; ok {
					cached = c
				}
				groups := views.ExtractURLs(cached, a.cfg.Jira.Host)
				if len(groups) > 0 {
					var items []components.ModalItem
					for i, g := range groups {
						if i > 0 || len(groups) > 1 {
							items = append(items, components.ModalItem{Label: g.Section, Separator: true})
						}
						for _, u := range g.URLs {
							if key := a.extractIssueKey(u); key != "" {
								items = append(items, components.ModalItem{ID: u, Label: key, Internal: true})
							} else {
								items = append(items, components.ModalItem{ID: u, Label: components.TruncateMiddle(u, 50)})
							}
						}
					}
					a.modal.SetSize(a.width, a.height)
					a.modal.Show("URLs", items)
				}
			}
			return a, nil

		case "t":
			if a.side == sideLeft && a.leftFocus == focusIssues {
				if sel := a.issuesList.SelectedIssue(); sel != nil {
					*a.logFlag = true
					return a, fetchTransitions(a.client, sel.Key)
				}
			}
			return a, nil

		case "r":
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				*a.logFlag = true
				return a, fetchIssueDetail(a.client, sel.Key)
			}
			return a, nil

		case "R":
			if a.projectKey != "" {
				*a.logFlag = true
				return a, fetchIssues(a.client, a.projectKey)
			}
			return a, nil
		}

	case autoFetchTickMsg:
		var fetchCmds []tea.Cmd
		if a.projectKey != "" {
			fetchCmds = append(fetchCmds, fetchIssues(a.client, a.projectKey))
		}
		// Schedule next tick.
		fetchCmds = append(fetchCmds, tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
			return autoFetchTickMsg{}
		}))
		return a, tea.Batch(fetchCmds...)

	case issuesLoadedMsg:
		a.err = nil
		*a.logFlag = false
		a.statusPanel.SetOnline(true)
		a.issueCache = make(map[string]*jira.Issue)
		a.issuesList.SetIssues(msg.issues)
		// Prefetch all issue details in background.
		for _, issue := range msg.issues {
			cmds = append(cmds, prefetchIssue(a.client, issue.Key))
		}
		return a, tea.Batch(cmds...)

	case issueDetailLoadedMsg:
		*a.logFlag = false
		if msg.issue != nil {
			a.issueCache[msg.issue.Key] = msg.issue
		}
		if a.leftFocus == focusIssues || a.side == sideRight {
			a.detailView.SetIssue(msg.issue)
		} else {
			a.detailView.UpdateIssueData(msg.issue)
		}
		return a, nil

	case issuePrefetchedMsg:
		if msg.issue != nil {
			a.issueCache[msg.issue.Key] = msg.issue
			// If this is the currently selected issue, update detail.
			if sel := a.issuesList.SelectedIssue(); sel != nil && sel.Key == msg.issue.Key {
				if a.leftFocus == focusIssues || a.side == sideRight {
					a.detailView.SetIssue(msg.issue)
				}
			}
		}
		return a, nil

	case transitionDoneMsg:
		if a.projectKey != "" {
			var fetchCmds []tea.Cmd
			fetchCmds = append(fetchCmds, fetchIssues(a.client, a.projectKey))
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				fetchCmds = append(fetchCmds, fetchIssueDetail(a.client, sel.Key))
			}
			return a, tea.Batch(fetchCmds...)
		}
		return a, nil

	case transitionsLoadedMsg:
		if len(msg.transitions) == 0 {
			return a, nil
		}
		var items []components.ModalItem
		for _, t := range msg.transitions {
			label := t.Name
			if t.To != nil {
				label += " → " + t.To.Name
			}
			items = append(items, components.ModalItem{ID: t.ID, Label: label})
		}
		a.modal.SetSize(a.width, a.height)
		a.modal.Show("Transition: "+msg.issueKey, items)
		return a, nil

	case components.ModalSelectedMsg:
		id := msg.Item.ID
		if strings.HasPrefix(id, "http") {
			// Check if it's a Jira issue URL on our host.
			if issueKey := a.extractIssueKey(id); issueKey != "" {
				// Navigate to the issue in [2].
				a.navigateToIssue(issueKey)
				return a, nil
			}
			// External URL — open in browser.
			openBrowser(id)
		} else {
			// Transition picker — execute transition.
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				return a, doTransition(a.client, sel.Key, id)
			}
		}
		return a, nil

	case components.ModalCancelledMsg:
		return a, nil

	case views.ExpandBlockMsg:
		var items []components.ModalItem
		for _, line := range msg.Lines {
			items = append(items, components.ModalItem{ID: "", Label: line})
		}
		// Constrain modal to detail panel size (including borders).
		modalW := a.width - a.panelSideW
		modalH := a.panelDetailH
		if a.isVerticalLayout() {
			modalW = a.width
		}
		a.modal.SetSize(modalW, modalH)
		a.modal.ShowReadOnly(msg.Title, items)
		return a, nil

	case errorMsg:
		a.err = msg.err
		a.statusPanel.SetOnline(false)
		return a, nil

	case views.IssueSelectedMsg:
		if msg.Issue != nil {
			if cached, ok := a.issueCache[msg.Issue.Key]; ok {
				a.detailView.SetIssue(cached)
			} else {
				a.detailView.SetIssue(msg.Issue)
			}
		}
		return a, nil

	case projectsLoadedMsg:
		projects := msg.projects
		// Move last-used project to top.
		if creds, err := config.LoadCredentials(); err == nil && creds != nil && creds.LastProject != "" {
			for i, p := range projects {
				if p.Key == creds.LastProject {
					// Swap to front.
					projects[0], projects[i] = projects[i], projects[0]
					break
				}
			}
		}
		a.projectList.SetProjects(projects)
		if a.projectKey == "" && len(projects) > 0 {
			a.projectKey = projects[0].Key
			a.statusPanel.SetProject(a.projectKey)
			a.projectList.SetActiveKey(a.projectKey)
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
		a.projectList.SetActiveKey(msg.ProjectKey)
		a.issuesList.ClearActiveKey()
		a.leftFocus = focusIssues
		a.updateFocusState()
		*a.logFlag = true
		go saveLastProject(msg.ProjectKey)
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
	x, y := msg.X, msg.Y
	panel, relY := a.hitTest(x, y)

	switch {
	case msg.Button == tea.MouseButtonWheelUp:
		return a.mouseScroll(panel, -3)
	case msg.Button == tea.MouseButtonWheelDown:
		return a.mouseScroll(panel, 3)
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		return a.mouseClick(panel, relY, x)
	}
	return a, nil
}

type panelID int

const (
	panelStatus panelID = iota
	panelIssues
	panelProjects
	panelDetail
	panelLog
)

// hitTest determines which panel the coordinates fall in and the relative Y.
func (a *App) hitTest(x, y int) (panelID, int) {
	if a.isVerticalLayout() {
		// Vertical: all stacked.
		top := 0
		if y < top+a.panelStatusH {
			return panelStatus, y - top
		}
		top += a.panelStatusH
		if y < top+a.panelIssuesH {
			return panelIssues, y - top
		}
		top += a.panelIssuesH
		if y < top+a.panelProjectsH {
			return panelProjects, y - top
		}
		top += a.panelProjectsH
		if y < top+a.panelDetailH {
			return panelDetail, y - top
		}
		return panelLog, y - top - a.panelDetailH
	}

	// Horizontal layout.
	if x < a.panelSideW {
		top := 0
		if y < top+a.panelStatusH {
			return panelStatus, y - top
		}
		top += a.panelStatusH
		if y < top+a.panelIssuesH {
			return panelIssues, y - top
		}
		top += a.panelIssuesH
		return panelProjects, y - top
	}

	// Right side.
	if y < a.panelDetailH {
		return panelDetail, y
	}
	return panelLog, y - a.panelDetailH
}

func (a *App) mouseScroll(panel panelID, delta int) (tea.Model, tea.Cmd) {
	switch panel { //nolint:exhaustive // only scrollable panels handled
	case panelIssues:
		if a.side != sideLeft || a.leftFocus != focusIssues {
			a.side = sideLeft
			a.leftFocus = focusIssues
			a.updateFocusState()
		}
		if delta > 0 {
			a.issuesList.ScrollBy(1)
		} else {
			a.issuesList.ScrollBy(-1)
		}
		if sel := a.issuesList.SelectedIssue(); sel != nil {
			a.showCachedIssue(sel.Key)
		}
	case panelProjects:
		if a.side != sideLeft || a.leftFocus != focusProjects {
			a.side = sideLeft
			a.leftFocus = focusProjects
			a.updateFocusState()
		}
		if delta > 0 {
			a.projectList.ScrollBy(1)
		} else {
			a.projectList.ScrollBy(-1)
		}
		if p := a.projectList.SelectedProject(); p != nil {
			a.detailView.SetProject(p)
		}
	case panelDetail:
		if a.side != sideRight {
			a.side = sideRight
			a.updateFocusState()
		}
		if delta > 0 {
			a.detailView.ScrollBy(1)
		} else {
			a.detailView.ScrollBy(-1)
		}
	}
	return a, nil
}

func (a *App) mouseClick(panel panelID, relY int, x int) (tea.Model, tea.Cmd) {
	switch panel { //nolint:exhaustive // only clickable panels handled
	case panelStatus:
		a.side = sideLeft
		a.leftFocus = focusStatus
		a.splashInfo.Project = a.projectKey
		a.detailView.SetSplash(a.splashInfo)
		a.updateFocusState()

	case panelIssues:
		a.side = sideLeft
		a.leftFocus = focusIssues
		a.updateFocusState()
		if relY == 0 {
			// Title bar — tab click (All / Assigned).
			a.issuesList.ClickTabAt(x)
		} else {
			a.issuesList.ClickAt(relY)
			if sel := a.issuesList.SelectedIssue(); sel != nil {
				a.showCachedIssue(sel.Key)
			}
		}

	case panelProjects:
		a.side = sideLeft
		a.leftFocus = focusProjects
		a.updateFocusState()
		a.projectList.ClickAt(relY)
		// Emit hover for preview.
		if p := a.projectList.SelectedProject(); p != nil {
			a.detailView.SetProject(p)
		}

	case panelDetail:
		a.side = sideRight
		a.updateFocusState()
		if relY == 0 {
			// Title bar → tab click.
			relX := x
			if !a.isVerticalLayout() {
				relX = x - a.panelSideW
			}
			a.detailView.ClickTab(relX)
		} else {
			if cmd := a.detailView.ClickItem(relY); cmd != nil {
				return a, cmd
			}
		}
	}
	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	var content string

	if a.isVerticalLayout() {
		// Vertical: all stacked.
		var detailArea string
		if a.modal.IsVisible() {
			detailH := max(a.height-1-3-5-3-5, 5) // rough: total - status - issues - projects - log
			detailArea = lipgloss.Place(a.width, detailH,
				lipgloss.Center, lipgloss.Center,
				a.modal.View(),
			)
		} else {
			detailArea = a.detailView.View()
		}
		content = lipgloss.JoinVertical(lipgloss.Left,
			a.statusPanel.View(),
			a.issuesList.View(),
			a.projectList.View(),
			detailArea,
			a.logPanel.View(),
		)
	} else {
		leftCol := lipgloss.JoinVertical(lipgloss.Left,
			a.statusPanel.View(),
			a.issuesList.View(),
			a.projectList.View(),
		)

		var detailArea string
		if a.modal.IsVisible() {
			sideW := a.sideWidth()
			mainW := a.width - sideW
			detailH := max(a.height-1-8, 5)
			popup := a.modal.View()
			detailArea = lipgloss.Place(mainW, detailH,
				lipgloss.Center, lipgloss.Center,
				popup,
			)
		} else {
			detailArea = a.detailView.View()
		}
		rightCol := lipgloss.JoinVertical(lipgloss.Left,
			detailArea,
			a.logPanel.View(),
		)

		content = lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
	}

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

	// Overlays.
	if a.showHelp {
		full = a.renderHelpOverlay(full)
	}
	// Modal is rendered inline in the right column, not as overlay.

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

	lines := make([]string, 0, len(bindings)+3)
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

	popupW := min(maxKey+40, a.width-4)
	popupH := min(len(lines), a.height-4)

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

// showCachedIssue updates the detail view with the cached version of the given issue key.
func (a *App) showCachedIssue(key string) {
	if cached, ok := a.issueCache[key]; ok {
		a.detailView.SetIssue(cached)
	}
}

// extractIssueKey checks if a URL points to our Jira and extracts the issue key.
// e.g. https://didlogic.atlassian.net/browse/DR-13819 → "DR-13819"
func (a *App) extractIssueKey(url string) string {
	host := strings.TrimRight(a.cfg.Jira.Host, "/")
	prefix := host + "/browse/"
	key, found := strings.CutPrefix(url, prefix)
	if found {
		// Strip any trailing query params or fragments.
		if idx := strings.IndexAny(key, "?#&/"); idx != -1 {
			key = key[:idx]
		}
		if key != "" {
			return key
		}
	}
	return ""
}

// navigateToIssue switches to the issue in the issues list.
// If found in current tab (All/Assigned), selects it there.
// If not, switches to All tab and tries again.
func (a *App) navigateToIssue(key string) {
	// Try current tab first.
	if a.issuesList.SelectByKey(key) {
		a.side = sideLeft
		a.leftFocus = focusIssues
		a.updateFocusState()
		a.showCachedIssue(key)
		return
	}
	// Switch to All tab and try again.
	if a.issuesList.GetTab() != views.IssueTabAll {
		a.issuesList.SetTab(views.IssueTabAll)
		if a.issuesList.SelectByKey(key) {
			a.side = sideLeft
			a.leftFocus = focusIssues
			a.updateFocusState()
			a.showCachedIssue(key)
			return
		}
	}
	// Not in our list — open in browser as fallback.
	openBrowser(a.cfg.Jira.Host + "/browse/" + key)
}

func saveLastProject(projectKey string) {
	creds, err := config.LoadCredentials()
	if err != nil || creds == nil {
		return
	}
	creds.LastProject = projectKey
	_ = config.SaveCredentials(creds)
}

// platformCommand returns the OS-specific command name and args for the given action.
func platformCommand(action string, arg string) (name string, args []string) {
	switch action {
	case "open":
		switch runtime.GOOS {
		case "darwin":
			return "open", []string{arg}
		case "windows":
			return "rundll32", []string{"url.dll,FileProtocolHandler", arg}
		default:
			return "xdg-open", []string{arg}
		}
	case "copy":
		switch runtime.GOOS {
		case "darwin":
			return "pbcopy", nil
		case "windows":
			return "clip", nil
		default:
			return "xclip", []string{"-selection", "clipboard"}
		}
	}
	return "", nil
}

func copyToClipboard(text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	name, args := platformCommand("copy", "")
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = strings.NewReader(text)
	_ = cmd.Run()
}

func openBrowser(url string) {
	name, args := platformCommand("open", url)
	cmd := exec.CommandContext(context.Background(), name, args...)
	_ = cmd.Start()
}

// isVerticalLayout returns true when terminal is too narrow for side-by-side.
func (a *App) isVerticalLayout() bool {
	return a.width < 80
}

// sideWidth calculates the left panel width, shrinking for narrow terminals.
func (a *App) sideWidth() int {
	if a.isVerticalLayout() {
		return a.width
	}
	sideW := a.cfg.GUI.SidePanelWidth
	if sideW <= 0 {
		sideW = 40
	}
	// Shrink side panel for medium terminals to fit [0] tabs.
	if a.width < 120 && sideW > a.width*35/100 {
		sideW = a.width * 35 / 100
	}
	if sideW > a.width/2 {
		sideW = a.width / 2
	}
	if sideW < 25 {
		sideW = 25
	}
	return sideW
}

func (a *App) layoutPanels() {
	totalH := a.height - 1

	if a.isVerticalLayout() {
		w := a.width
		statusH := 3
		logH := 5
		projectsH := 3 // projects always minimal in vertical
		remaining := totalH - statusH - logH - projectsH

		var issuesH, detailH int

		issuesCollapsed := 7 // 5 items + 2 borders

		switch {
		case a.side == sideRight:
			issuesH = issuesCollapsed
			detailH = remaining - issuesH
		case a.leftFocus == focusIssues:
			detailH = 6
			issuesH = remaining - detailH
		case a.leftFocus == focusProjects:
			issuesH = issuesCollapsed
			detailH = remaining - issuesH
		default:
			issuesH = issuesCollapsed
			detailH = remaining - issuesH
		}

		if issuesH < 3 {
			issuesH = 3
		}
		if detailH < 3 {
			detailH = 3
		}

		a.statusPanel.SetSize(w, statusH)
		a.issuesList.SetSize(w, issuesH)
		a.projectList.SetSize(w, projectsH)
		a.detailView.SetSize(w, detailH)
		a.logPanel.SetSize(w, logH)
		a.panelSideW = w
		a.panelStatusH = statusH
		a.panelIssuesH = issuesH
		a.panelProjectsH = projectsH
		a.panelDetailH = detailH
		a.panelLogH = logH
		return
	}

	sideW := a.sideWidth()
	mainW := a.width - sideW

	statusH := 3
	remaining := totalH - statusH

	// Natural content heights (capped at remaining space).
	issuesNat := a.issuesList.ContentHeight()
	projectsNat := a.projectList.ContentHeight()
	minH := 3 // minimum panel height

	if issuesNat < minH {
		issuesNat = minH
	}
	if projectsNat < minH {
		projectsNat = minH
	}

	var issuesH, projectsH int

	if issuesNat+projectsNat <= remaining {
		// Both fit — give extra space to the focused panel.
		extra := remaining - issuesNat - projectsNat
		if a.leftFocus == focusProjects {
			projectsH = projectsNat + extra
			issuesH = issuesNat
		} else {
			issuesH = issuesNat + extra
			projectsH = projectsNat
		}
	} else {
		// Not enough space — focused panel gets remaining, other gets natural or min.
		switch a.leftFocus {
		case focusIssues:
			projectsH = max(min(projectsNat, remaining/3), minH)
			issuesH = remaining - projectsH
		case focusProjects:
			issuesH = max(min(issuesNat, remaining/3), minH)
			projectsH = remaining - issuesH
		default:
			issuesH = remaining / 2
			projectsH = remaining - issuesH
		}
	}

	if issuesH < minH {
		issuesH = minH
	}
	if projectsH < minH {
		projectsH = minH
	}

	a.statusPanel.SetSize(sideW, statusH)
	a.issuesList.SetSize(sideW, issuesH)
	a.projectList.SetSize(sideW, projectsH)

	// Right column: log fits content or max 8, detail gets the rest.
	logH := 8
	detailH := max(totalH-logH, 5)

	a.detailView.SetSize(mainW, detailH)
	a.logPanel.SetSize(mainW, logH)

	// Cache sizes for mouse.
	a.panelSideW = sideW
	a.panelStatusH = statusH
	a.panelIssuesH = issuesH
	a.panelProjectsH = projectsH
	a.panelDetailH = detailH
	a.panelLogH = logH
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

// fetchFullIssue fetches issue + comments + changelog, returning the given message type.
func fetchFullIssue(client *jira.Client, key string, mkMsg func(*jira.Issue) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		issue, err := client.GetIssue(context.Background(), key)
		if err != nil {
			return mkMsg(nil)
		}
		comments, err := client.GetComments(context.Background(), key)
		if err == nil {
			issue.Comments = comments
		}
		changelog, err := client.GetChangelog(context.Background(), key)
		if err == nil {
			issue.Changelog = changelog
		}
		return mkMsg(issue)
	}
}

func fetchIssueDetail(client *jira.Client, key string) tea.Cmd {
	return fetchFullIssue(client, key, func(issue *jira.Issue) tea.Msg {
		if issue == nil {
			return errorMsg{err: fmt.Errorf("failed to fetch issue %s", key)}
		}
		return issueDetailLoadedMsg{issue: issue}
	})
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

func prefetchIssue(client *jira.Client, key string) tea.Cmd {
	return fetchFullIssue(client, key, func(issue *jira.Issue) tea.Msg {
		if issue == nil {
			return nil // silent fail for prefetch
		}
		return issuePrefetchedMsg{issue: issue}
	})
}

func fetchTransitions(client *jira.Client, issueKey string) tea.Cmd {
	return func() tea.Msg {
		transitions, err := client.GetTransitions(context.Background(), issueKey)
		if err != nil {
			return errorMsg{err: err}
		}
		return transitionsLoadedMsg{issueKey: issueKey, transitions: transitions}
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
