package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"atlas.games/internal/game"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("atlas.games v%s\n", Version)
		return
	}
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		fmt.Println("Atlas Games - Terminal-based game collection.")
		fmt.Println("\nUsage:")
		fmt.Println("  atlas.games        Start the game launcher")
		fmt.Println("  atlas.games -v     Show version")
		fmt.Println("  atlas.games -h     Show this help")
		return
	}

	p := tea.NewProgram(game.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
		os.Exit(1)
	}
}
