package breach

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	nodeHackedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	nodeLockedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	nodeCurrentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true).Blink(true)
	traceStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	boxStyle       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
	programStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
)

type tickMsg time.Time

type Model struct {
	game *Game
}

func NewModel() Model {
	return Model{
		game: NewGame(),
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
		case "1": m.game.RunProgram(Crack)
		case "2": m.game.RunProgram(Stealth)
		case "3": m.game.RunProgram(Overclock)
		case "4", "5", "6", "7", "8", "9":
			targetID := int(msg.String()[0] - '1')
			m.game.Move(targetID)
		case "r":
			m.game = NewGame()
		}
	case tickMsg:
		m.game.Tick()
		return m, tick()
	}
	return m, nil
}

func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString("\n  " + titleStyle.Render(" ATLAS BREACH - NETWORK INFILTRATOR ") + "\n\n")

	// Map Buffer
	mapWidth, mapHeight := 65, 20
	mapBuf := make([][]string, mapHeight)
	for y := 0; y < mapHeight; y++ {
		mapBuf[y] = make([]string, mapWidth)
		for x := 0; x < mapWidth; x++ {
			mapBuf[y][x] = " "
		}
	}

	// Draw Connections
	for _, n := range m.game.Nodes {
		for _, adjID := range n.Adjacent {
			adj := m.game.Nodes[adjID]
			m.drawLine(mapBuf, n.X, n.Y, adj.X, adj.Y)
		}
	}

	// Draw Nodes
	for i, n := range m.game.Nodes {
		style := nodeLockedStyle
		if n.Hacked { style = nodeHackedStyle }
		if i == m.game.CurrentNode { style = nodeCurrentStyle }
		
		char := "[ ]"
		if n.Type == Core { char = "[C]" }
		if n.Type == Firewall { char = "[F]" }
		
		label := fmt.Sprintf("%d:%s", i+1, n.Name)
		m.writeAt(mapBuf, n.X-1, n.Y, style.Render(char))
		m.writeAt(mapBuf, n.X-3, n.Y+1, label)
	}

	mapView := ""
	for y := 0; y < mapHeight; y++ {
		mapView += strings.Join(mapBuf[y], "") + "\n"
	}

	// Stats Panel
	currentNode := m.game.Nodes[m.game.CurrentNode]
	traceBar := fmt.Sprintf("[%s%s] %.1f%%", 
		strings.Repeat("█", int(m.game.Trace/5)),
		strings.Repeat("░", 20-int(m.game.Trace/5)),
		m.game.Trace)
	
	statusInfo := fmt.Sprintf(
		"LOCATION:  %s\nSECURITY:  %d%%\nSTATUS:    %s\n\n%s\n%s\n",
		currentNode.Name, currentNode.Security, 
		m.getStatusText(currentNode),
		traceStyle.Render("TRACE DETECTION:"), traceStyle.Render(traceBar),
	)

	programs := programStyle.Render("1. Crack.exe\n2. Stealth.sh\n3. Overclock.go\n")
	
	rightPanel := boxStyle.Width(35).Render(
		titleStyle.Render("TERMINAL STATUS") + "\n\n" + 
		statusInfo + "\n" +
		titleStyle.Render("AVAILABLE TOOLS") + "\n" + programs,
	)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, boxStyle.Render(mapView), rightPanel)
	sb.WriteString("  " + mainView + "\n")

	// Log
	logView := boxStyle.Width(102).Render(lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Render("SYSTEM LOG") + "\n" + strings.Join(m.game.Log, "\n"))
	sb.WriteString("  " + logView + "\n")

	controls := " [1-3] Run Program  [4-9] Move to Node  [R] Restart  [Q] Disconnect"
	if m.game.Win {
		controls = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Render(" [MISSION SUCCESS] CORE DATA ACQUIRED. PRESS [Q] TO EXIT.")
	} else if m.game.GameOver {
		controls = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render(" [TERMINATED] CONNECTION TRACED. PRESS [R] TO REBOOT.")
	}
	sb.WriteString("\n " + controls)

	return sb.String()
}

func (m Model) getStatusText(n *Node) string {
	if n.Hacked { return "COMPROMISED" }
	return "LOCKED"
}

func (m Model) writeAt(buf [][]string, x, y int, s string) {
	if y >= 0 && y < len(buf) && x >= 0 && x < len(buf[y]) {
		buf[y][x] = s
	}
}

func (m Model) drawLine(buf [][]string, x1, y1, x2, y2 int) {
	char := "."
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	if x1 == x2 {
		minY, maxY := y1, y2
		if minY > maxY { minY, maxY = maxY, minY }
		for y := minY; y <= maxY; y++ {
			if buf[y][x1] == " " { buf[y][x1] = style.Render("|") }
		}
	} else if y1 == y2 {
		minX, maxX := x1, x2
		if minX > maxX { minX, maxX = maxX, minX }
		for x := minX; x <= maxX; x++ {
			if buf[y1][x] == " " { buf[y1][x] = style.Render("-") }
		}
	} else {
		// Diagonal/rough line
		m.writeAt(buf, (x1+x2)/2, (y1+y2)/2, style.Render(char))
	}
}

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
