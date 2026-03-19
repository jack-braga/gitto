package editor

import (
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/config"
)

// EditorFinishedMsg is sent when an external editor process completes.
type EditorFinishedMsg struct {
	Err error
}

// terminalEditors is the set of editors that run inside the terminal.
var terminalEditors = map[string]bool{
	"vim":   true,
	"nvim":  true,
	"vi":    true,
	"nano":  true,
	"micro": true,
	"helix": true,
	"hx":    true,
	"emacs": true,
	"joe":   true,
	"ne":    true,
}

// Open returns a Bubble Tea command that opens the given file in the user's
// preferred editor. For terminal editors it uses tea.ExecProcess to suspend
// the TUI. For GUI editors it spawns the process in the background.
func Open(filePath string, cfg *config.Config) tea.Cmd {
	editor := resolveEditor(cfg)

	if isTerminalEditor(editor) {
		cmd := exec.Command(editor, filePath)
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return EditorFinishedMsg{Err: err}
		})
	}

	// GUI editor — spawn in background, don't block TUI
	return func() tea.Msg {
		cmd := exec.Command(editor, filePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		return EditorFinishedMsg{Err: err}
	}
}

// resolveEditor determines which editor to use.
// Resolution order: config → $EDITOR → $VISUAL → code → vim → nano
func resolveEditor(cfg *config.Config) string {
	if cfg != nil && cfg.Editor != "" {
		return cfg.Editor
	}
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}

	// Try to find a common editor on PATH
	for _, e := range []string{"code", "vim", "nano"} {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}

	return "nano"
}

// isTerminalEditor returns true if the editor runs inside the terminal
// (as opposed to opening a GUI window).
func isTerminalEditor(editor string) bool {
	// Extract the base command name (handle paths and flags)
	base := editor
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	// Handle "emacs -nw" style args
	base = strings.Fields(base)[0]

	return terminalEditors[base]
}
