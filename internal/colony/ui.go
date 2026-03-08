package colony

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dirtStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("130"))
	tunnelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	antStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	foodStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	queenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	spiderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
)

type tickMsg time.Time

type Model struct {
	colony      *Colony
	width       int
	height      int
	showingHelp bool
}

func NewModel() Model {
	w, h := 120, 40
	return Model{
		colony: NewColony(w, h),
		width:  w,
		height: h,
	}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			m.colony = NewColony(m.width, m.height)
			m.showingHelp = false
			return m, tick()
		case "h":
			m.showingHelp = !m.showingHelp
			return m, nil
		}
	case tickMsg:
		if !m.showingHelp {
			m.colony.Tick()
		}
		return m, tick()
	}
	return m, nil
}

func (m Model) View() string {
	if m.showingHelp {
		var sb strings.Builder
		sb.WriteString("\n  " + titleStyle.Render(" ATLAS BIOLOGICAL ENGINE - COLONY DOCUMENTATION ") + "\n\n")
		sb.WriteString("  " + antStyle.Render("x Worker   ") + ": Forages for food and returns to the queen.\n")
		sb.WriteString("  " + antStyle.Render("o Carrier  ") + ": A worker carrying a food unit (S).\n")
		sb.WriteString("  " + spiderStyle.Render("* Spider   ") + ": Lethal territorial predator. Avoid at all costs.\n")
		sb.WriteString("  " + queenStyle.Render("Q Queen    ") + ": The heart of the colony. Located in the center.\n")
		sb.WriteString("  " + foodStyle.Render("S Sugar    ") + ": Food source found on surface and underground.\n")
		sb.WriteString("  " + dirtStyle.Render("░ Dirt     ") + ": Solid earth. Spiders may be hidden inside.\n")
		sb.WriteString("  " + tunnelStyle.Render("  Tunnel   ") + ": Safe paths for movement.\n\n")
		
		sb.WriteString("  " + lipgloss.NewStyle().Bold(true).Render("THE SURVIVAL:") + "\n")
		sb.WriteString("  - Spiders (*) stay within 8 squares of their origin point.\n")
		sb.WriteString("  - Ants will try to avoid spiders, but may stumble into them while digging.\n")
		sb.WriteString("  - If an ant is eaten, the population decreases. Reset (R) to restart colony.\n\n")

		sb.WriteString("  [H] Close Documentation  [R] Reset Colony  [Q] Exit\n")
		return sb.String()
	}

	var sb strings.Builder
	sb.WriteString("\n  " + titleStyle.Render(" ATLAS BIOLOGICAL ENGINE - TACTICAL COLONY ") + "\n\n")

	// Render Grid
	buffer := make([][]string, m.height)
	for y := 0; y < m.height; y++ {
		buffer[y] = make([]string, m.width)
		for x := 0; x < m.width; x++ {
			cell := m.colony.Grid[y][x]
			switch cell {
			case Dirt:
				buffer[y][x] = dirtStyle.Render("░")
			case Tunnel:
				buffer[y][x] = tunnelStyle.Render(" ")
			case Food:
				buffer[y][x] = foodStyle.Render("S")
			case Queen:
				buffer[y][x] = queenStyle.Render("Q")
			default:
				buffer[y][x] = " "
			}
		}
	}

	// Overlay Spiders
	for _, s := range m.colony.Spiders {
		buffer[s.Y][s.X] = spiderStyle.Render("*")
	}

	// Overlay Ants
	for _, a := range m.colony.Ants {
		char := "x"
		if a.HasFood { char = "o" }
		buffer[a.Y][a.X] = antStyle.Render(char)
	}

	for y := 0; y < m.height; y++ {
		sb.WriteString("  " + strings.Join(buffer[y], "") + "\n")
	}

	sb.WriteString(fmt.Sprintf("\n  Workers: %d | [H] Help | [R] Reset | [Q] Exit", len(m.colony.Ants)))
	return sb.String()
}

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
