package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/katolikov/gpeace/internal/git"
	"github.com/katolikov/gpeace/internal/tui"
)

func main() {
	repoRoot, err := git.RepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: not inside a Git repository.\n")
		os.Exit(1)
	}

	model := tui.NewModel(repoRoot)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
