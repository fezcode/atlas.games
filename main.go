package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"atlas.games/internal/game"
)

func main() {
	p := tea.NewProgram(game.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
		os.Exit(1)
	}
}
