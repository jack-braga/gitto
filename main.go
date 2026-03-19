package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jack-braga/gitto/internal/app"
	"github.com/jack-braga/gitto/internal/config"
)

var version = "dev"

func main() {
	args := os.Args[1:]

	// Handle subcommands
	if len(args) > 0 {
		switch args[0] {
		case "version", "--version":
			fmt.Printf("gitto %s\n", version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		case "config":
			handleConfig(args[1:])
			os.Exit(0)
		}
	}

	// Parse flags
	targetDir := "."
	editorOverride := ""
	themeOverride := ""
	noPoll := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-e", "--editor":
			if i+1 < len(args) {
				editorOverride = args[i+1]
				i++
			}
		case "-t", "--theme":
			if i+1 < len(args) {
				themeOverride = args[i+1]
				i++
			}
		case "--no-poll":
			noPoll = true
		default:
			if !strings.HasPrefix(args[i], "-") {
				targetDir = args[i]
			}
		}
	}

	// Resolve target directory
	absDir, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a valid directory\n", absDir)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Apply overrides
	if editorOverride != "" {
		cfg.Editor = editorOverride
	}
	if themeOverride != "" {
		cfg.Theme = themeOverride
	}

	// Create and run the TUI
	model := app.New(absDir, cfg, noPoll)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`gitto — multi-repo Git TUI workspace manager

Usage:
  gitto [path] [flags]

  Run gitto in a parent directory to manage all child Git repos.

Flags:
  -e, --editor string    Override editor for this session
  -t, --theme string     Override theme (auto|dark|light)
      --no-poll          Disable background polling
      --version          Print version and exit
  -h, --help             Show help

Subcommands:
  config set <key> <value>   Set a config value
  config get <key>           Get a config value
  config list                List all config values
  version                    Print version

Config keys: editor, theme, poll_interval, show_untracked,
             show_ignored, confirm_discard, confirm_force, default_view

Examples:
  gitto                           Run in current directory
  gitto ~/projects/repos          Run in a specific directory
  gitto config set editor "nvim"  Set default editor
  gitto config set poll_interval 10
`)
}

func handleConfig(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: gitto config <get|set|list> [key] [value]")
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		values, err := config.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		// Sort keys for consistent output
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s = %s\n", k, values[k])
		}

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: gitto config get <key>")
			os.Exit(1)
		}
		val, err := config.Get(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(val)

	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: gitto config set <key> <value>")
			os.Exit(1)
		}
		if err := config.Set(args[1], args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s = %s\n", args[1], args[2])

	default:
		fmt.Printf("Unknown config command: %s\n", args[0])
		os.Exit(1)
	}
}
