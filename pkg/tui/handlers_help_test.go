package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/textfuel/lazyjira/v2/pkg/internal/testkit"
	"github.com/textfuel/lazyjira/v2/pkg/jira/jiratest"
)

func helpApp(t *testing.T) *App {
	t.Helper()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.showHelp = true
	app.helpCursor = 0
	app.helpFilter = ""
	app.helpSearching = false
	return app
}

func TestHandleDetailScroll_Down(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()

	_, _, ok := app.handleDetailScroll(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})

	if !ok {
		t.Error("detail scroll down should be handled")
	}
}

func TestHandleDetailScroll_Up(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()
	app.width = 120
	app.height = 40
	app.layoutPanels()

	_, _, ok := app.handleDetailScroll(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})

	if !ok {
		t.Error("detail scroll up should be handled")
	}
}

func TestHandleDetailScroll_UnknownKey(t *testing.T) {
	t.Parallel()
	app := newAppWithFake(t, &jiratest.FakeClient{T: t})
	app.keymap = DefaultKeymap()

	_, _, ok := app.handleDetailScroll(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}})

	if ok {
		t.Error("unknown key should not be handled by handleDetailScroll")
	}
}

func TestHandleHelpKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		key          tea.KeyMsg
		initialState func(app *App)
		assert       func(t *testing.T, app *App)
	}{
		{
			name: "q closes help overlay",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			assert: func(t *testing.T, app *App) {
				t.Helper()
				if app.showHelp {
					t.Error("q should close help overlay")
				}
			},
		},
		{
			name: "esc closes help overlay",
			key:  tea.KeyMsg{Type: tea.KeyEsc},
			assert: func(t *testing.T, app *App) {
				t.Helper()
				if app.showHelp {
					t.Error("esc should close help overlay")
				}
			},
		},
		{
			name: "slash enters search mode",
			key:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
			assert: func(t *testing.T, app *App) {
				t.Helper()
				if !app.helpSearching {
					t.Error("/ should enter help search mode")
				}
			},
		},
		{
			name:         "j navigates down",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			initialState: func(app *App) { app.helpCursor = 0 },
			assert: func(t *testing.T, app *App) {
				t.Helper()
				if app.helpCursor != 1 {
					t.Errorf("helpCursor = %d, want 1 after navigate down", app.helpCursor)
				}
			},
		},
		{
			name:         "k navigates up",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			initialState: func(app *App) { app.helpCursor = 3 },
			assert: func(t *testing.T, app *App) {
				t.Helper()
				if app.helpCursor != 2 {
					t.Errorf("helpCursor = %d, want 2 after navigate up", app.helpCursor)
				}
			},
		},
		{
			name:         "g jumps to top",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}},
			initialState: func(app *App) { app.helpCursor = 10 },
			assert: func(t *testing.T, app *App) {
				t.Helper()
				testkit.AssertEqual(t, "helpCursor after top", app.helpCursor, 0)
			},
		},
		{
			name:         "G jumps to bottom",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}},
			initialState: func(app *App) { app.helpCursor = 0 },
			assert: func(t *testing.T, app *App) {
				t.Helper()
				bindings := app.ContextBindings()
				if app.helpCursor != len(bindings)-1 {
					t.Errorf("helpCursor = %d, want %d after navigate to bottom", app.helpCursor, len(bindings)-1)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			app := helpApp(t)
			if testCase.initialState != nil {
				testCase.initialState(app)
			}
			app.handleHelpKeys(testCase.key)
			testCase.assert(t, app)
		})
	}
}

func TestFilteredHelpBindings_EmptyFilterReturnsAll(t *testing.T) {
	t.Parallel()
	app := helpApp(t)
	app.helpFilter = ""

	all := app.ContextBindings()
	filtered := app.filteredHelpBindings()

	if len(filtered) != len(all) {
		t.Errorf("len(filtered) = %d, want %d (all)", len(filtered), len(all))
	}
}

func TestFilteredHelpBindings_FilterReduces(t *testing.T) {
	t.Parallel()
	app := helpApp(t)
	app.helpFilter = string(ActQuit)

	filtered := app.filteredHelpBindings()
	all := app.ContextBindings()

	if len(filtered) >= len(all) {
		t.Errorf("filter should reduce bindings: filtered=%d, all=%d", len(filtered), len(all))
	}
	for _, binding := range filtered {
		if binding.Description != string(ActQuit) && binding.Key != "q" && binding.Key != "ctrl+c" {
			t.Errorf("unexpected binding %q/%q after filter=quit", binding.Key, binding.Description)
		}
	}
}

func TestHandleHelpSearchKey_EscClearsSearch(t *testing.T) {
	t.Parallel()
	app := helpApp(t)
	app.helpSearching = true
	app.helpFilter = "something"

	app.handleHelpSearchKey(tea.KeyMsg{Type: tea.KeyEsc})

	if app.helpSearching {
		t.Error("esc should exit help search mode")
	}
	if app.helpFilter != "" {
		t.Errorf("helpFilter = %q, want empty after esc", app.helpFilter)
	}
}

func TestHandleHelpSearchKey_EnterConfirms(t *testing.T) {
	t.Parallel()
	app := helpApp(t)
	app.helpSearching = true
	app.helpCursor = 0

	app.handleHelpSearchKey(tea.KeyMsg{Type: tea.KeyEnter})

	if app.helpSearching {
		t.Error("enter should confirm search and exit search mode")
	}
}

func TestHandleHelpSearchKey_DownMovesSelection(t *testing.T) {
	t.Parallel()
	app := helpApp(t)
	app.helpSearching = true
	app.helpFilter = ""
	app.helpCursor = 0

	app.handleHelpSearchKey(tea.KeyMsg{Type: tea.KeyDown})

	if app.helpCursor != 1 {
		t.Errorf("helpCursor = %d, want 1 after search key down", app.helpCursor)
	}
}

func TestHelpConfirmSearch_RestoresCursorToMatchedItem(t *testing.T) {
	t.Parallel()
	app := helpApp(t)
	app.helpSearching = true
	app.helpFilter = string(ActQuit)

	filtered := app.filteredHelpBindings()
	if len(filtered) == 0 {
		t.Fatal("expected at least one binding matching quit filter")
	}
	matchedBinding := filtered[0]
	app.helpCursor = 0

	app.helpConfirmSearch()

	if app.helpSearching {
		t.Error("should exit search mode after confirm")
	}
	if app.helpFilter != "" {
		t.Error("filter should be cleared after confirm")
	}

	allBindings := app.ContextBindings()
	wantCursor := -1
	for i, binding := range allBindings {
		if binding.Key == matchedBinding.Key && binding.Description == matchedBinding.Description {
			wantCursor = i
			break
		}
	}
	if wantCursor < 0 {
		t.Fatal("matched binding not found in full binding list")
	}
	if app.helpCursor != wantCursor {
		t.Errorf("helpCursor = %d, want %d (position of quit binding)", app.helpCursor, wantCursor)
	}
}
