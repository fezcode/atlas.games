package city

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	roadStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	buildingStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	parkStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	commercialStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	waterStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	titleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
)

type tickMsg time.Time

type Model struct {
	wfc         *WFC
	done        bool
	width       int
	height      int
	showingHelp bool
}

func NewModel() Model {
	w, h := 120, 40
	return Model{
		wfc:    NewWFC(w, h),
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
			m.wfc = NewWFC(m.width, m.height)
			m.done = false
			m.showingHelp = false
			return m, tick()
		case "h":
			m.showingHelp = !m.showingHelp
			return m, nil
		}
	case tickMsg:
		if !m.done && !m.showingHelp {
			for i := 0; i < 20; i++ {
				m.done = m.wfc.Step()
				if m.done { break }
			}
			return m, tick()
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.showingHelp {
		var sb strings.Builder
		sb.WriteString("\n  " + titleStyle.Render(" ATLAS CITY ENGINE - DOCUMENTATION ") + "\n\n")
		sb.WriteString("  " + roadStyle.Render("═║╔╗╝╚  Roads   ") + ": Double-line box drawing for connected roads.\n")
		sb.WriteString("  " + buildingStyle.Render("█ Building  ") + ": Residential zones (White Blocks).\n")
		sb.WriteString("  " + commercialStyle.Render("S Commercial") + ": Business districts (Yellow Shops).\n")
		sb.WriteString("  " + parkStyle.Render("♣ Park      ") + ": Green spaces for the citizens.\n")
		sb.WriteString("  " + waterStyle.Render("~ Water     ") + ": Fountains, lakes, or pools.\n\n")
		sb.WriteString("  [R] Reset City  [H] Close Documentation  [Q] Exit to Launcher\n")
		return sb.String()
	}

	var sb strings.Builder
	sb.WriteString("  " + titleStyle.Render(" ATLAS CITY ENGINE - CITY GENERATOR ") + "\n")

	for y := 0; y < m.height; y++ {
		sb.WriteString("  ")
		for x := 0; x < m.width; x++ {
			tile := m.wfc.Grid[y][x]
			if !tile.Collapsed {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Render("?"))
			} else {
				char := " "
				style := roadStyle
				switch tile.Type {
				case RoadV: char = "║"
				case RoadH: char = "═"
				case RoadTL: char = "╔"
				case RoadTR: char = "╗"
				case RoadBL: char = "╚"
				case RoadBR: char = "╝"
				case RoadTU: char = "╩"
				case RoadTD: char = "╦"
				case RoadTLT: char = "╠"
				case RoadTRT: char = "╣"
				case RoadCross: char = "╬"
				case Building: char = "█"; style = buildingStyle
				case Commercial: char = "S"; style = commercialStyle
				case Park: char = "♣"; style = parkStyle
				case Water: char = "~"; style = waterStyle
				}
				sb.WriteString(style.Render(char))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  [R] Reset City  [H] Help  [Q] Exit to Launcher")
	return sb.String()
}

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
