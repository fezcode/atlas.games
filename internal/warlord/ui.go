package warlord

import (
	"fmt"
	"math"
	"strings"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	ViewWidth    = 40 // 40 cells = 80 terminal columns
	ViewHeight   = 25
	SidebarWidth = 34
)

var (
	waterStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	landStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	forestStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	mountainStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	
	playerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	aiStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true)
	bossStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Blink(true)
	cacheStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("46"))
	sidebarStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")).Padding(0, 1).Width(SidebarWidth)
	marketStyleUI = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Align(lipgloss.Center)
	helpStyle     = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Width(80 + SidebarWidth + 4)
	
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	dangerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	combatBoxStyle = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("196")).Padding(1, 4).Align(lipgloss.Center)
	cursorStyle    = lipgloss.NewStyle().Background(lipgloss.Color("255")).Foreground(lipgloss.Color("0"))
)

type GameStateUI int
const (
	StateMap GameStateUI = iota
	StateDuel
	StateDuelResult
	StateMarket
)

type Model struct {
	game         *GameState
	state        GameStateUI
	selectedUnit int
	cursorX, cursorY int
	lastMsg      string
	showingHelp  bool
	
	duelAttacker *Unit
	duelTarget   *Unit
	playerRoll   int
	enemyRoll    int
	duelSummary  string
}

func NewModel() Model {
	w, h := 120, 40
	return Model{
		game:         NewGame(w, h),
		state:        StateMap,
		selectedUnit: -1,
		cursorX:      w / 2,
		cursorY:      h / 2,
		lastMsg:      "Tactical ready. [H] for manual.",
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.game.GameOver || m.game.GameWon {
		if _, ok := msg.(tea.KeyMsg); ok { return m, tea.Quit }
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "q", "ctrl+c", "esc": return m, tea.Quit
		case "h": m.showingHelp = !m.showingHelp; return m, nil
		}

		if m.showingHelp { return m, nil }

		if m.state == StateMarket {
			player := m.game.Units[0]
			switch key {
			case "1":
				cost := player.Level * 60
				if player.Gold >= cost {
					player.Gold -= cost
					player.Attack += 4
					m.game.AddLog("🛒 MARKET: Weapons Upgraded (+4 ATK)")
				}
			case "2":
				cost := player.Level * 60
				if player.Gold >= cost {
					player.Gold -= cost
					player.Defense += 3
					m.game.AddLog("🛒 MARKET: Shields Upgraded (+3 DEF)")
				}
			case "enter", " ", "m": m.state = StateMap
			}
			return m, nil
		}

		if m.state == StateDuel {
			if key == " " { m.resolveDuel(); m.state = StateDuelResult }
			return m, nil
		}
		if m.state == StateDuelResult {
			if key == "enter" || key == " " { m.state = StateMap; m.selectedUnit = -1 }
			return m, nil
		}

		switch key {
		case "up", "w": if m.cursorY > 0 { m.cursorY-- }
		case "down", "s": if m.cursorY < m.game.Height-1 { m.cursorY++ }
		case "left", "a": if m.cursorX > 0 { m.cursorX-- }
		case "right", "d": if m.cursorX < m.game.Width-1 { m.cursorX++ }
		case "m": if m.game.Grid[m.cursorY][m.cursorX].Type == Market { m.state = StateMarket }
		case " ":
			if m.selectedUnit != -1 { m.selectedUnit = -1 } else {
				for i, u := range m.game.Units {
					if u.X == m.cursorX && u.Y == m.cursorY && u.Team == TeamPlayer {
						m.selectedUnit = i; break
					}
				}
			}
		case "enter":
			if m.selectedUnit != -1 {
				var target *Unit
				for _, u := range m.game.Units {
					if u.X == m.cursorX && u.Y == m.cursorY && u.Team == TeamInvader {
						target = u; break
					}
				}
				if target != nil {
					m.duelAttacker = m.game.Units[m.selectedUnit]
					m.duelTarget = target
					m.state = StateDuel
				} else {
					m.lastMsg = m.game.MoveUnit(m.selectedUnit, m.cursorX, m.cursorY)
					if m.game.Grid[m.cursorY][m.cursorX].Type == Market { m.state = StateMarket }
					m.selectedUnit = -1
				}
			}
		case "n": m.game.NextTurn(); m.selectedUnit = -1
		}
	}
	return m, nil
}

func (m *Model) resolveDuel() {
	m.playerRoll = rand.Intn(6) + 1
	m.enemyRoll = rand.Intn(6) + 1
	pwr := m.duelAttacker.Attack + m.playerRoll
	enemyDefRoll := m.duelTarget.Defense + m.enemyRoll
	dmg := pwr - enemyDefRoll; if dmg < 1 { dmg = 1 }
	
	m.duelTarget.Health -= dmg
	m.duelSummary = fmt.Sprintf("⚔️ COMBAT REPORT ⚔️\nYOU: %d+%d vs ENEMY: %d+%d\nDEALT: %d DAMAGE", 
		m.duelAttacker.Attack, m.playerRoll, m.duelTarget.Defense, m.enemyRoll, dmg)
	
	if m.duelTarget.Health <= 0 {
		xp, gold := 50, 50
		if m.duelTarget.Type == Dragon { xp, gold = 1000, 1000 }
		m.duelSummary += fmt.Sprintf("\n\n💀 TARGET NEUTRALIZED!\n💰 +%d GOLD | 💎 +%d XP", gold, xp)
		m.duelAttacker.Gold += gold
		m.duelAttacker.XP += xp
		for i, u := range m.game.Units {
			if u == m.duelTarget {
				m.game.Units = append(m.game.Units[:i], m.game.Units[i+1:]...)
				break
			}
		}
		m.game.CheckLevelUp(m.duelAttacker)
	} else {
		cntRoll := rand.Intn(6) + 1
		cntDmg := (m.duelTarget.Attack + cntRoll) / 2
		m.duelAttacker.Health -= cntDmg
		m.duelSummary += fmt.Sprintf("\n\n⚠️ COUNTER-ATTACK: -%d HP", cntDmg)
	}
	m.duelAttacker.Moves = 0
}

func (m Model) View() string {
	if m.game.GameWon {
		return "\n\n  " + titleStyle.Render(" MISSION SUCCESS! DRAGON DEFEATED ") + 
			fmt.Sprintf("\n  The kingdom is safe. Final Score: %d\n  Press any key to exit...", m.game.Score)
	}
	if m.game.GameOver {
		return titleStyle.Render(" COMMANDER NEUTRALIZED ") + fmt.Sprintf("\nMission Failed.\nXP: %d\nPress any key...", m.game.Score)
	}

	if m.showingHelp { return m.renderHelp() }

	// Viewport logic
	startX := m.cursorX - ViewWidth/2 
	startY := m.cursorY - ViewHeight/2
	if startX < 0 { startX = 0 } else if startX > m.game.Width-ViewWidth { startX = m.game.Width - ViewWidth }
	if startY < 0 { startY = 0 } else if startY > m.game.Height-ViewHeight { startY = m.game.Height - ViewHeight }

	var mapLines []string
	for y := startY; y < startY+ViewHeight; y++ {
		var line strings.Builder
		for x := startX; x < startX+ViewWidth; x++ {
			// Range Highlight
			inRange := false
			if m.selectedUnit != -1 {
				u := m.game.Units[m.selectedUnit]
				dx := int(math.Abs(float64(u.X - x))); dy := int(math.Abs(float64(u.Y - y)))
				dist := dx; if dy > dx { dist = dy }
				if dist <= u.Moves && dist > 0 { inRange = true }
			}

			// Cursor
			if x == m.cursorX && y == m.cursorY {
				char := "X "; if m.selectedUnit != -1 { char = "M " }
				line.WriteString(cursorStyle.Render(char))
				continue
			}

			// Units
			var unitAt *Unit
			for _, u := range m.game.Units { if u.X == x && u.Y == y { unitAt = u; break } }
			if unitAt != nil {
				char := "S "; if unitAt.Type == Knight { char = "K " }
				style := aiStyle
				if unitAt.Team == TeamPlayer { char = "C "; style = playerStyle }
				if unitAt.Type == Dragon { char = "D "; style = bossStyle }
				if inRange { style = style.Copy().Background(lipgloss.Color("237")) }
				line.WriteString(style.Render(char))
				continue
			}

			// Terrain (ASCII)
			tile := m.game.Grid[y][x]
			char := "· "; style := landStyle
			switch tile.Type {
			case Water: char = "~ "; style = waterStyle
			case Forest: char = "↑ "; style = forestStyle
			case Mountain: char = "▲ "; style = mountainStyle
			case Cache: char = "$ "; style = cacheStyle
			case Market: char = "B "; style = highlightStyle
			case Stronghold: char = "H "; style = playerStyle
			case Princess: char = "P "; style = playerStyle
			}
			if inRange { style = style.Copy().Background(lipgloss.Color("237")) }
			line.WriteString(style.Render(char))
		}
		mapLines = append(mapLines, line.String())
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📜 MISSION STATUS") + "\n")
	sb.WriteString(fmt.Sprintf("📅 DAY: %d | 💰 GOLD: %d\n", m.game.Turn, m.game.Units[0].Gold))
	sb.WriteString(fmt.Sprintf("📍 LOC: [%d, %d]\n\n", m.cursorX-m.game.BaseX, m.game.BaseY-m.cursorY))
	
	player := m.game.Units[0]
	sb.WriteString(playerStyle.Render("🎖️ COMMANDER TELEMETRY") + "\n")
	sb.WriteString(fmt.Sprintf("❤️ HP:  %d/%d\n", player.Health, player.MaxHP))
	sb.WriteString(fmt.Sprintf("⚔️ ATK: %d | 🛡️ DEF: %d\n", player.Attack, player.Defense))
	sb.WriteString(fmt.Sprintf("💎 XP:  %d/%d (LVL %d)\n", player.XP, player.NextLevelXP, player.Level))
	sb.WriteString(highlightStyle.Render(fmt.Sprintf("⚡ AP:   %d\n\n", player.Moves)))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📡 SCAN INTEL") + "\n")
	var hoverUnit *Unit
	for _, u := range m.game.Units { if u.X == m.cursorX && u.Y == m.cursorY { hoverUnit = u; break } }
	if hoverUnit != nil {
		style := aiStyle; if hoverUnit.Type == Dragon { style = bossStyle } else if hoverUnit.Team == TeamPlayer { style = playerStyle }
		sb.WriteString(style.Render(fmt.Sprintf("%v (Lvl.%d)\nHP: %d/%d | ATK: %d\n\n", hoverUnit.Type, hoverUnit.Level, hoverUnit.Health, hoverUnit.MaxHP, hoverUnit.Attack)))
	} else { sb.WriteString("Sector clear.\n\n") }

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("📜 ACTION LOGS") + "\n")
	for _, log := range m.game.Logs { sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("> "+log) + "\n") }

	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("46")).Render(strings.Join(mapLines, "\n")),
		sidebarStyle.Render(sb.String()),
	)

	return titleStyle.Render(" ATLAS WARLORD - STRATEGIC COMMAND ") + "\n\n" + mainView + "\n[H] Manual | [N] End Turn | [Q] Abort"
}

func (m Model) renderHelp() string {
	var h strings.Builder
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Underline(true)
	section := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	key := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)

	h.WriteString(title.Render(" ATLAS WARLORD: TACTICAL OPS MANUAL ") + "\n\n")

	h.WriteString(section.Render("📜 MISSION DIRECTIVE") + "\n")
	h.WriteString("Locate the Princess (P) and neutralize the Dragon (D).\n\n")

	h.WriteString(section.Render("🕹️ COMMAND & CONTROL") + "\n")
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[WASD]"), "Move Tactical Cursor (X)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[SPACE]"), "Select Commander (C)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[ENTER]"), "Execute Order (Target X -> M)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[N]"), "End Day Cycle"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[H]"), "Toggle Manual"))

	h.WriteString(section.Render("🗺️ TACTICAL MAP LEGEND") + "\n")
	h.WriteString(" UNITS:  C Commander  D Dragon  S Scout  K Knight\n")
	h.WriteString(" LOCS:   B Shop  $ Loot  H Base  P Princess\n")
	h.WriteString(" TILES:  · Land  ↑ Forest  ~ Water  ▲ Mountain\n\n")

	h.WriteString(lipgloss.NewStyle().Width(80 + SidebarWidth).Align(lipgloss.Center).Render("PRESS [H] TO RESUME MISSION"))

	return "\n" + helpStyle.Render(h.String())
}
