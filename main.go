package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"atlas.games/internal/menu"
	"atlas.games/internal/wilson"
	"atlas.games/internal/wfc"
	"atlas.games/internal/city"
	"atlas.games/internal/colony"
	"atlas.games/internal/warlord"
	"atlas.games/internal/defense"
	"atlas.games/internal/breach"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Printf("atlas.games v%s\n", Version)
		return
	}

	for {
		// 1. Run Menu
		m := menu.NewModel()
		p := tea.NewProgram(m, tea.WithAltScreen())
		res, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running menu: %v\n", err)
			os.Exit(1)
		}

		choice := res.(menu.Model).Choice()

		// 2. Launch Game
		var gameModel tea.Model
		switch choice {
		case "Wilson's Revenge":
			gameModel = wilson.NewModel()
		case "WFC Land Creator":
			gameModel = wfc.NewModel()
		case "WFC City Generator":
			gameModel = city.NewModel()
		case "Tactical Colony":
			gameModel = colony.NewModel()
		case "Atlas Warlord":
			gameModel = warlord.NewModel()
		case "Atlas Defense":
			gameModel = defense.NewModel()
		case "Atlas Breach":
			gameModel = breach.NewModel()
		case "Exit", "":
			fmt.Println("Goodbye, operator.")
			return
		}

		if gameModel != nil {
			gp := tea.NewProgram(gameModel, tea.WithAltScreen())
			if _, err := gp.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error running game: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
