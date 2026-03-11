package defense

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	pathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	towerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	enemyStyleFull = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))   // Green
	enemyStyleMed  = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))  // Yellow
	enemyStyleLow  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))  // Red
	bulletStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	baseStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	spawnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	cursorStyle  = lipgloss.NewStyle().Background(lipgloss.Color("240"))
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	goldStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	healthStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	rangeStyle   = lipgloss.NewStyle().Background(lipgloss.Color("17")) // Deep Blue background
)

type tickMsg time.Time

type Model struct {
	game        *Game
	cursorX     int
	cursorY     int
	width       int
	height      int
	showingHelp bool
	gameOver    bool
}

func NewModel() Model {
	w, h := 60, 25
	return Model{
		game:    NewGame(w, h),
		width:   w,
		height:  h,
		cursorX: w / 2,
		cursorY: h / 2,
	}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.game.Health <= 0 {
		m.gameOver = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "w":
			if m.cursorY > 0 { m.cursorY-- }
		case "down", "s":
			if m.cursorY < m.height-1 { m.cursorY++ }
		case "left", "a":
			if m.cursorX > 0 { m.cursorX-- }
		case "right", "d":
			if m.cursorX < m.width-1 { m.cursorX++ }
		case "enter", " ":
			if !m.gameOver {
				m.game.PlaceTower(m.cursorX, m.cursorY)
			}
		case "backspace", "delete", "x":
			if !m.gameOver {
				m.game.SellTower(m.cursorX, m.cursorY)
			}
		case "r":
			m.game = NewGame(m.width, m.height)
			m.gameOver = false
			return m, tick()
		case "h":
			m.showingHelp = !m.showingHelp
		}
	case tickMsg:
		if !m.showingHelp && !m.gameOver {
			m.game.Tick()
		}
		return m, tick()
	}
	return m, nil
}

func (m Model) View() string {
	if m.showingHelp {
		var sb strings.Builder
		sb.WriteString("\n  " + titleStyle.Render(" ATLAS TACTICAL DEFENSE - MANUAL ") + "\n\n")
		sb.WriteString("  Defend the Atlas core from incoming data corruption (enemies).\n\n")
		sb.WriteString("  " + towerStyle.Render("T Tower    ") + ": Fires at the furthest enemy in range. Cost: 15 Gold.\n")
		sb.WriteString("  " + bulletStyle.Render("* Bullet   ") + ": Projectiles traveling toward targets.\n")
		sb.WriteString("  " + enemyStyleFull.Render("e Enemy    ") + ": Moves on path. Color changes: Green (High HP) > Yellow > Red (Low HP).\n")
		sb.WriteString("  " + baseStyle.Render("B Base     ") + ": Protect this at all costs.\n")
		sb.WriteString("  " + pathStyle.Render("░ Path     ") + ": Enemies only move on this designated route.\n\n")
		sb.WriteString("  CONTROLS:\n")
		sb.WriteString("  - Arrows/WASD: Move cursor\n")
		sb.WriteString("  - Enter/Space: Build Tower (15 Gold)\n")
		sb.WriteString("  - Backspace/X: Sell Tower (10 Gold Refund)\n")
		sb.WriteString("  - [H] Close Help  [R] Reset  [Q] Exit\n")
		return sb.String()
	}

	var sb strings.Builder
	sb.WriteString("\n  " + titleStyle.Render(" ATLAS TACTICAL DEFENSE ") + "\n\n")

	// Render Grid
	buffer := make([][]string, m.height)
	for y := 0; y < m.height; y++ {
		buffer[y] = make([]string, m.width)
		for x := 0; x < m.width; x++ {
			cell := m.game.Grid[y][x]
			char := " "
			style := lipgloss.NewStyle()

			switch cell {
			case Path:
				char = "░"
				style = pathStyle
			case TowerCell:
				char = "T"
				style = towerStyle
			case Base:
				char = "B"
				style = baseStyle
			case Spawn:
				char = "S"
				style = spawnStyle
			default:
				char = "."
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
			}

			// Range visualization
			distToCursor := math.Sqrt(float64((x-m.cursorX)*(x-m.cursorX) + (y-m.cursorY)*(y-m.cursorY)))
			if distToCursor <= 7.0 {
				style = style.Copy().Background(lipgloss.Color("17"))
			}

			if x == m.cursorX && y == m.cursorY {
				buffer[y][x] = cursorStyle.Render(char)
			} else {
				buffer[y][x] = style.Render(char)
			}
		}
	}

	// Overlay Enemies
	for _, e := range m.game.Enemies {
		if e.PathIndex >= 0 && e.PathIndex < len(m.game.Path) {
			char := "e"
			hpRatio := float64(e.HP) / float64(e.MaxHP)
			style := enemyStyleFull
			if hpRatio < 0.4 {
				style = enemyStyleLow
				char = "x"
			} else if hpRatio < 0.8 {
				style = enemyStyleMed
			}
			
			if e.X == m.cursorX && e.Y == m.cursorY {
				buffer[e.Y][e.X] = cursorStyle.Render(char)
			} else {
				buffer[e.Y][e.X] = style.Render(char)
			}
		}
	}

	// Overlay Projectiles
	for _, p := range m.game.Projectiles {
		px, py := int(p.X), int(p.Y)
		if px >= 0 && px < m.width && py >= 0 && py < m.height {
			buffer[py][px] = bulletStyle.Render("*")
		}
	}

	for y := 0; y < m.height; y++ {
		sb.WriteString("  " + strings.Join(buffer[y], "") + "\n")
	}

	status := fmt.Sprintf("\n  " + goldStyle.Render("GOLD: %d") + " | " + healthStyle.Render("HEALTH: %d") + " | WAVE: %d | NEXT: %d", 
		m.game.Gold, m.game.Health, m.game.Wave, m.game.NextWaveIn)
	
	if m.gameOver {
		status += " | " + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("GAME OVER (R to Restart)")
	}

	sb.WriteString(status + "\n  [WASD] Move [Space] Build [H] Help [Q] Exit")
	return sb.String()
}

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
