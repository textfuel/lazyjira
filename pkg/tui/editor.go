package tui

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// editorFinishedMsg is sent when $EDITOR exits.
type editorFinishedMsg struct {
	original string // original content (trimmed) for comparison
	tempPath string // path to temp file
	err      error  // non-nil if editor launch failed
}

var errNoEditor = errors.New("no editor found - set $EDITOR environment variable")

// resolveEditor returns the editor command: $EDITOR -> $VISUAL -> vi.
func resolveEditor() (string, error) {
	if e := os.Getenv("EDITOR"); e != "" {
		slog.Debug("editor: resolved", "source", "EDITOR", "editor", e)
		return e, nil
	}
	if e := os.Getenv("VISUAL"); e != "" {
		slog.Debug("editor: resolved", "source", "VISUAL", "editor", e)
		return e, nil
	}
	if path, err := exec.LookPath("vi"); err == nil {
		slog.Debug("editor: resolved", "source", "fallback", "editor", path)
		return path, nil
	}
	slog.Debug("editor: no editor found")
	return "", errNoEditor
}

// launchEditor writes content to a temp file and opens it in $EDITOR.
// Returns a tea.Cmd that suspends the TUI via tea.ExecProcess
func launchEditor(content, suffix string) tea.Cmd {
	editor, err := resolveEditor()
	if err != nil {
		slog.Debug("editor: resolve failed", "err", err)
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}

	tmpFile, err := os.CreateTemp("", "lazyjira-*"+suffix)
	if err != nil {
		slog.Debug("editor: temp file create failed", "err", err)
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}
	tmpPath := tmpFile.Name()
	n, err := tmpFile.WriteString(content)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		slog.Debug("editor: temp file write failed", "temp_path", tmpPath, "err", err)
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}
	_ = tmpFile.Close()
	slog.Debug("editor: launching", "editor", editor, "temp_path", tmpPath, "temp_size_bytes", n)

	original := strings.TrimRight(content, "\n")
	start := time.Now()
	parts := strings.Fields(editor)
	cmd := exec.CommandContext(context.Background(), parts[0], append(parts[1:], tmpPath)...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		elapsed := time.Since(start)
		slog.Debug("editor: exited", "elapsed_ms", elapsed.Milliseconds(), "err", err)
		return editorFinishedMsg{
			original: original,
			tempPath: tmpPath,
			err:      err,
		}
	})
}

// readAndCheckEditor reads the temp file and checks if content changed.
func readAndCheckEditor(msg editorFinishedMsg) (string, bool, error) {
	if msg.err != nil {
		slog.Debug("editor: read skipped due to launch error", "err", msg.err)
		return "", false, msg.err
	}
	data, err := os.ReadFile(msg.tempPath)
	if err != nil {
		slog.Debug("editor: temp file read failed", "temp_path", msg.tempPath, "err", err)
		return "", false, err
	}
	newContent := strings.TrimRight(string(data), "\n")
	changed := newContent != msg.original
	slog.Debug("editor: read back", "temp_size_bytes", len(data), "new_content_bytes", len(newContent), "changed", changed)
	return newContent, changed, nil
}

// cleanupEditor removes the temp file.
func cleanupEditor(path string) {
	if path != "" {
		_ = os.Remove(path)
	}
}
