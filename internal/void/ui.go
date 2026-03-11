package void

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	shipStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	marketStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	logStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	planetStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	currentPlanetStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	boxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
)

type Model struct {
	game     *Game
	selected int // Selected planet for jump
}

func NewModel() Model {
	return Model{
		game: NewGame(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "1", "2", "3", "4", "5", "6":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.game.Planets) {
				m.game.Travel(idx)
			}
		case "f": m.game.SelectedRes = Fuel
		case "o": m.game.SelectedRes = Food
		case "m": m.game.SelectedRes = Metal
		case "t": m.game.SelectedRes = Tech
		case "a": m.game.SelectedRes = Artifacts
		case "b":
			m.game.Buy(m.game.SelectedRes, 1)
		case "s":
			m.game.Sell(m.game.SelectedRes, 1)
		case "w":
			m.game.Wait()
		}
	}
	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString("\n  " + titleStyle.Render(" ATLAS VOID - STAR TRADER ") + "\n\n")

	// 1. SHIP PANEL
	shipInfo := fmt.Sprintf(
		"Credits: %d cr\nFuel:    %d/%d\nCargo:   %d/%d\n\n[Hold]\n",
		m.game.Ship.Credits, m.game.Ship.Fuel, m.game.Ship.MaxFuel,
		m.totalCargo(), m.game.Ship.MaxCargo,
	)
	for r := Fuel; r <= Artifacts; r++ {
		shipInfo += fmt.Sprintf("- %s: %d\n", r.String(), m.game.Ship.Cargo[r])
	}
	shipBox := boxStyle.Width(25).Render(shipStyle.Render("SHIP SYSTEMS") + "\n" + shipInfo)

	// 2. MARKET PANEL
	currentPlanet := m.game.Planets[m.game.Ship.CurrentPos]
	marketInfo := fmt.Sprintf("Location: %s\n%s\n\n", currentPlanet.Name, currentPlanet.Description)
	for r := Fuel; r <= Artifacts; r++ {
		prefix := "  "
		if m.game.SelectedRes == r { prefix = "> " }
		marketInfo += fmt.Sprintf("%s%-10s: %d cr (Stock: %d)\n", prefix, r.String(), currentPlanet.Resources[r], currentPlanet.Stock[r])
	}
	marketInfo += "\n[B] Buy [S] Sell (1 unit)"
	marketBox := boxStyle.Width(35).Render(marketStyle.Render("PLANETARY MARKET") + "\n" + marketInfo)

	// 3. NAVIGATION PANEL
	navInfo := ""
	for i, p := range m.game.Planets {
		prefix := fmt.Sprintf("[%d] ", i+1)
		style := planetStyle
		if i == m.game.Ship.CurrentPos {
			style = currentPlanetStyle
			prefix = "[*] "
		}
		navInfo += style.Render(prefix+p.Name) + "\n"
	}
	navInfo += "\n[W] Wait / Refuel"
	navBox := boxStyle.Width(25).Render(titleStyle.Render("NAV COMPUTER") + "\n" + navInfo)

	// Combine Middle Section
	middle := lipgloss.JoinHorizontal(lipgloss.Top, shipBox, marketBox, navBox)
	sb.WriteString("  " + middle + "\n\n")

	// 4. LOG PANEL
	logInfo := strings.Join(m.game.Log, "\n")
	logBox := boxStyle.Width(89).Render(logStyle.Render("COMMAND LOG") + "\n" + logInfo)
	sb.WriteString("  " + logBox + "\n")

	sb.WriteString("\n  [1-6] Jump [F,O,M,T,A] Select Resource [Q] Exit")
	return sb.String()
}

func (m Model) totalCargo() int {
	total := 0
	for _, a := range m.game.Ship.Cargo { total += a }
	return total
}
