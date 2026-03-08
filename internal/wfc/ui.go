package wfc

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	waterStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	landStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("185"))
	forestStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	mountainStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true) // White Mountains
	lavaStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
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
	// Massive grid
	w, h := 200, 60
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
			// Higher burst for massive grid
			for i := 0; i < 50; i++ {
				m.done = m.wfc.Step()
				if m.done {
					break
				}
			}
			return m, tick()
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.showingHelp {
		var sb strings.Builder
		sb.WriteString("\n  " + titleStyle.Render(" ATLAS LANDSCAPE ENGINE - LAND CREATOR DOCUMENTATION ") + "\n\n")
		sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("THE ALGORITHMS") + "\n")
		sb.WriteString("  1. " + lipgloss.NewStyle().Bold(true).Render("ENTROPY:") + " Every cell starts in a state of 'Superposition'.\n")
		sb.WriteString("  2. " + lipgloss.NewStyle().Bold(true).Render("OBSERVATION:") + " The system collapses cells based on weighted probability.\n")
		sb.WriteString("  3. " + lipgloss.NewStyle().Bold(true).Render("PROPAGATION:") + " Decisions ripple to neighbors, enforcing biome logic.\n")
		sb.WriteString("  4. " + lipgloss.NewStyle().Bold(true).Render("SEEDING:") + " Every map starts with 'Primordial Seeds' of all biomes to ensure diversity.\n\n")

		sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("THE BIOMES") + "\n")
		sb.WriteString("  " + waterStyle.Render("~ Water   ") + ": Expansive oceans and lakes.\n")
		sb.WriteString("  " + landStyle.Render("█ Land    ") + ": The primary substrate (Solid Block).\n")
		sb.WriteString("  " + forestStyle.Render("↑ Forest  ") + ": Dense wooded clusters.\n")
		sb.WriteString("  " + mountainStyle.Render("▲ Mountain") + ": Jagged ridges and peaks (White Peaks).\n")
		sb.WriteString("  " + lavaStyle.Render("░ Lava    ") + ": Volcanic flows near mountains.\n\n")

		sb.WriteString("  " + lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("CONTROLS") + "\n")
		sb.WriteString("  [R] Reset Map  [H] Close Documentation  [Q] Exit to Launcher\n")
		return sb.String()
	}

	var sb strings.Builder
	sb.WriteString("  " + titleStyle.Render(" ATLAS LANDSCAPE ENGINE - LAND CREATOR ") + "\n")

	for y := 0; y < m.height; y++ {
		sb.WriteString("  ")
		for x := 0; x < m.width; x++ {
			tile := m.wfc.Grid[y][x]
			if !tile.Collapsed {
				sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Render("?"))
			} else {
				switch tile.Type {
				case Water:
					sb.WriteString(waterStyle.Render("~"))
				case Land:
					sb.WriteString(landStyle.Render("█"))
				case Forest:
					sb.WriteString(forestStyle.Render("↑"))
				case Mountain:
					sb.WriteString(mountainStyle.Render("▲"))
				case Lava:
					sb.WriteString(lavaStyle.Render("░"))
				default:
					sb.WriteString(" ")
				}
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  [R] Reset Map  [H] Help  [Q] Exit to Launcher")
	return sb.String()
}

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
