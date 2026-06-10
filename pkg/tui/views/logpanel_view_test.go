package views

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLogPanel_NewLogPanel_EmptyState(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	if panel == nil {
		t.Fatal("NewLogPanel returned nil")
	}
}

func TestLogPanel_Init_ReturnsNil(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	if cmd := panel.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestLogPanel_Update_ReturnsSelf(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	result, cmd := panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if result != panel {
		t.Error("Update() should return same panel pointer")
	}
	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
}

func TestLogPanel_View_EmptyShowsNoRequests(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	panel.SetSize(80, 10)
	output := stripANSI(panel.View())
	if !strings.Contains(output, "No requests yet") {
		t.Errorf("empty log view = %q, want 'No requests yet'", output)
	}
}

func TestLogPanel_AddEntry_ShowsInView(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	panel.SetSize(80, 10)
	panel.AddEntry(LogEntry{
		Time:    time.Now(),
		Method:  "GET",
		Path:    "/rest/api/3/issue/PLAT-1",
		Status:  200,
		Elapsed: 50 * time.Millisecond,
	})
	output := stripANSI(panel.View())
	if !strings.Contains(output, "GET") {
		t.Errorf("log view = %q, want to contain method GET", output)
	}
	if !strings.Contains(output, "200") {
		t.Errorf("log view = %q, want to contain status 200", output)
	}
}

func TestLogPanel_AddEntry_ErrorStatusColored(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	panel.SetSize(80, 10)
	panel.AddEntry(LogEntry{
		Time:    time.Now(),
		Method:  "POST",
		Path:    "/rest/api/3/issue",
		Status:  400,
		Elapsed: 10 * time.Millisecond,
	})
	output := stripANSI(panel.View())
	if !strings.Contains(output, "400") {
		t.Errorf("log view = %q, want to contain status 400", output)
	}
}

func TestLogPanel_AddEntry_CapAt100(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	panel.SetSize(80, 200)
	for i := range 110 {
		panel.AddEntry(LogEntry{
			Time:    time.Now(),
			Method:  "GET",
			Path:    "/rest/api/3/issue",
			Status:  200 + i,
			Elapsed: time.Millisecond,
		})
	}
	panel.mu.Lock()
	count := len(panel.entries)
	panel.mu.Unlock()
	if count != 100 {
		t.Errorf("entries count = %d, want 100", count)
	}
}

func TestLogPanel_SetSize_ZeroSizeNocrash(t *testing.T) {
	t.Parallel()
	panel := NewLogPanel()
	panel.SetSize(0, 0)
	_ = panel.View()
}
