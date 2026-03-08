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
	ViewWidth    = 80 
	ViewHeight   = 25
	SidebarWidth = 36
)

var (
	waterStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	landStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	forestStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	mountainStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	wallStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	
	playerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	aiStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true)
	bossStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Blink(true)
	cacheStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("46"))
	sidebarStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("244")).Padding(0, 1).Width(SidebarWidth)
	marketStyleUI = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Align(lipgloss.Center)
	helpStyle     = lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("214")).Padding(1, 2).Width(ViewWidth + SidebarWidth + 2)
	
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
		lastMsg:      "TACTICAL READY. [H] FOR MANUAL.",
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
		key := strings.ToLower(msg.String())
		switch key {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "h":
			m.showingHelp = !m.showingHelp
			return m, nil
		}

		if m.showingHelp {
			return m, nil
		}

		if m.state == StateMarket {
			player := m.game.Units[0]
			switch key {
			case "1":
				cost := player.Level * 60
				if player.Gold >= cost {
					player.Gold -= cost
					player.Attack += 4
					m.game.AddLog("MARKET: WEAPONS UPGRADED (+4 ATK)")
				}
			case "2":
				cost := player.Level * 60
				if player.Gold >= cost {
					player.Gold -= cost
					player.Defense += 3
					m.game.AddLog("MARKET: SHIELDS UPGRADED (+3 DEF)")
				}
			case "enter", " ", "m":
				m.state = StateMap
			}
			return m, nil
		}

		if m.state == StateDuel {
			if key == " " || key == "enter" {
				m.resolveDuel()
				m.state = StateDuelResult
			}
			return m, nil
		}
		if m.state == StateDuelResult {
			if key == "enter" || key == " " {
				m.state = StateMap
				m.selectedUnit = -1
			}
			return m, nil
		}

		switch key {
		case "up", "w":
			if m.cursorY > 0 {
				m.cursorY--
			}
		case "down", "s":
			if m.cursorY < m.game.Height-1 {
				m.cursorY++
			}
		case "left", "a":
			if m.cursorX > 0 {
				m.cursorX--
			}
		case "right", "d":
			if m.cursorX < m.game.Width-1 {
				m.cursorX++
			}
		case "m":
			if m.game.Grid[m.cursorY][m.cursorX].Type == Market {
				m.state = StateMarket
			}
		case " ":
			if m.selectedUnit != -1 {
				m.selectedUnit = -1
			} else {
				for i, u := range m.game.Units {
					if u.X == m.cursorX && u.Y == m.cursorY && u.Team == TeamPlayer {
						m.selectedUnit = i
						break
					}
				}
			}
		case "enter":
			if m.selectedUnit != -1 {
				attacker := m.game.Units[m.selectedUnit]
				dx := int(math.Abs(float64(attacker.X - m.cursorX)))
				dy := int(math.Abs(float64(attacker.Y - m.cursorY)))
				dist := dx
				if dy > dx {
					dist = dy
				}

				if dist > attacker.Moves {
					m.lastMsg = "OUT OF RANGE"
					return m, nil
				}

				var target *Unit
				for _, u := range m.game.Units {
					if u.X == m.cursorX && u.Y == m.cursorY && u.Team == TeamInvader {
						target = u
						break
					}
				}

				if target != nil {
					m.duelAttacker = attacker
					m.duelTarget = target
					m.duelAttacker.Moves = 0
					m.state = StateDuel
					m.game.AddLog(fmt.Sprintf("COMBAT: ENGAGING %v!", target.Type))
				} else {
					m.lastMsg = m.game.MoveUnit(m.selectedUnit, m.cursorX, m.cursorY)
					if m.game.Grid[m.cursorY][m.cursorX].Type == Market {
						m.state = StateMarket
					}
					m.selectedUnit = -1
				}
			}
		case "n":
			m.game.NextTurn()
			m.selectedUnit = -1
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
	m.duelSummary = fmt.Sprintf("COMBAT REPORT\nYOU: %d+%d vs ENEMY: %d+%d\nDEALT: %d DAMAGE", 
		m.duelAttacker.Attack, m.playerRoll, m.duelTarget.Defense, m.enemyRoll, dmg)
	
	if m.duelTarget.Health <= 0 {
		xp, gold := 50, 50
		if m.duelTarget.Type == Dragon { xp, gold = 1000, 1000 }
		m.duelSummary += fmt.Sprintf("\n\nTARGET ELIMINATED!\nGOLD: +%d | XP: +%d", gold, xp)
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
		if cntDmg < 1 { cntDmg = 1 }
		m.duelAttacker.Health -= cntDmg
		m.duelSummary += fmt.Sprintf("\n\nCOUNTER-ATTACK: -%d HP", cntDmg)
	}
}

func (m Model) View() string {
	if m.game.GameWon {
		return "\n\n  " + titleStyle.Render(" MISSION SUCCESS! DRAGON DEFEATED ") + 
			fmt.Sprintf("\n  THE KINGDOM IS SECURE. FINAL XP: %d\n  PRESS ANY KEY TO EXIT...", m.game.Score)
	}
	if m.game.GameOver {
		return titleStyle.Render(" COMMANDER NEUTRALIZED ") + fmt.Sprintf("\nMISSION FAILED.\nXP: %d\nPRESS ANY KEY...", m.game.Score)
	}

	if m.showingHelp { return m.renderHelp() }

	startX := m.cursorX - ViewWidth/2 
	startY := m.cursorY - ViewHeight/2
	if startX < 0 { startX = 0 } else if startX > m.game.Width-ViewWidth { startX = m.game.Width - ViewWidth }
	if startY < 0 { startY = 0 } else if startY > m.game.Height-ViewHeight { startY = m.game.Height - ViewHeight }

	var mapLines []string
	for y := startY; y < startY+ViewHeight; y++ {
		var line strings.Builder
		for x := startX; x < startX+ViewWidth; x++ {
			inRange := false
			if m.selectedUnit != -1 {
				u := m.game.Units[m.selectedUnit]
				dx := int(math.Abs(float64(u.X - x))); dy := int(math.Abs(float64(u.Y - y)))
				dist := dx; if dy > dx { dist = dy }
				if dist <= u.Moves && dist > 0 { inRange = true }
			}

			if x == m.cursorX && y == m.cursorY {
				char := "X"; if m.selectedUnit != -1 { char = "M" }
				line.WriteString(cursorStyle.Render(char) + " ")
				continue
			}

			var unitAt *Unit
			for _, u := range m.game.Units { if u.X == x && u.Y == y { unitAt = u; break } }
			if unitAt != nil {
				char := "S "; style := aiStyle
				if unitAt.Team == TeamPlayer { char = "C "; style = playerStyle }
				if unitAt.Type == Dragon { char = "D "; style = bossStyle }
				if unitAt.Type == Knight { char = "K "; style = aiStyle }
				if inRange { style = style.Copy().Background(lipgloss.Color("237")) }
				line.WriteString(style.Render(char))
				continue
			}

			tile := m.game.Grid[y][x]
			char := ". "; style := landStyle
			switch tile.Type {
			case Water: char = "~ "; style = waterStyle
			case Forest: char = "↑ "; style = forestStyle
			case Mountain: char = "▲ "; style = mountainStyle
			case Cache: char = "$ "; style = cacheStyle
			case Market: char = "B "; style = highlightStyle
			case Stronghold: char = "H "; style = playerStyle
			case Princess: char = "P "; style = playerStyle
			case WallV: char = "| "; style = wallStyle
			case WallH: char = "- "; style = wallStyle
			case WallTL, WallTR, WallBL, WallBR: char = "+ "; style = wallStyle
			}
			if inRange { style = style.Copy().Background(lipgloss.Color("237")) }
			line.WriteString(style.Render(char))
		}
		mapLines = append(mapLines, line.String())
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("MISSION STATUS") + "\n")
	sb.WriteString(fmt.Sprintf("DAY: %d | XP: %d\n", m.game.Turn, m.game.Score))
	sb.WriteString(fmt.Sprintf("LOC: [%d, %d]\n\n", m.cursorX-m.game.BaseX, m.game.BaseY-m.cursorY))
	
	player := m.game.Units[0]
	sb.WriteString(playerStyle.Render("COMMANDER TELEMETRY") + "\n")
	sb.WriteString(fmt.Sprintf("HP:    %d/%d\n", player.Health, player.MaxHP))
	sb.WriteString(fmt.Sprintf("ATK:   %d | DEF: %d\n", player.Attack, player.Defense))
	sb.WriteString(fmt.Sprintf("GOLD:  %d | LVL: %d\n", player.Gold, player.Level))
	sb.WriteString(fmt.Sprintf("XP:    %d/%d\n", player.XP, player.NextLevelXP))
	sb.WriteString(highlightStyle.Render(fmt.Sprintf("AP:    %d/%d\n\n", player.Moves, player.MaxMoves)))

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("SCAN INTEL") + "\n")
	var hoverUnit *Unit
	for _, u := range m.game.Units { if u.X == m.cursorX && u.Y == m.cursorY { hoverUnit = u; break } }
	if hoverUnit != nil {
		style := aiStyle; if hoverUnit.Type == Dragon { style = bossStyle } else if hoverUnit.Team == TeamPlayer { style = playerStyle }
		sb.WriteString(style.Render(fmt.Sprintf("%v (LVL %d)\nHP:  %d/%d\nATK: %d\n\n", hoverUnit.Type, hoverUnit.Level, hoverUnit.Health, hoverUnit.MaxHP, hoverUnit.Attack)))
	} else { sb.WriteString("SECTOR CLEAR.\n\n") }

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("ACTION LOGS") + "\n")
	for _, log := range m.game.Logs { sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("> "+log) + "\n") }

	mainView := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().BorderStyle(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("46")).Render(strings.Join(mapLines, "\n")),
		sidebarStyle.Render(sb.String()),
	)

	finalView := titleStyle.Render(" ATLAS WARLORD - STRATEGIC COMMAND ") + "\n\n" + mainView
	
	if m.state == StateDuel {
		duelUI := fmt.Sprintf("COMBAT INTERCEPT\n\n%v VS %v\n\nPRESS [SPACE] TO ROLL", m.duelAttacker.Type, m.duelTarget.Type)
		finalView = lipgloss.Place(ViewWidth+SidebarWidth+6, ViewHeight+10, lipgloss.Center, lipgloss.Center, combatBoxStyle.Render(duelUI))
	} else if m.state == StateDuelResult {
		finalView = lipgloss.Place(ViewWidth+SidebarWidth+6, ViewHeight+10, lipgloss.Center, lipgloss.Center, combatBoxStyle.Render(m.duelSummary+"\n\n[ENTER] CONTINUE"))
	} else if m.state == StateMarket {
		cost := player.Level * 60
		marketUI := fmt.Sprintf("BLACK MARKET - LVL %d\n\nGOLD: %d\n\n[1] UPGRADE WEAPONRY (+4 ATK) - %d G\n[2] UPGRADE ARMOR    (+3 DEF) - %d G\n\n[ENTER] EXIT MARKET", player.Level, player.Gold, cost, cost)
		finalView = lipgloss.Place(ViewWidth+SidebarWidth+6, ViewHeight+10, lipgloss.Center, lipgloss.Center, marketStyleUI.Render(marketUI))
	}

	return finalView + "\n[H] Manual | [N] End Turn | [Q] Abort"
}

func (m Model) renderHelp() string {
	var h strings.Builder
	title := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Underline(true)
	section := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	key := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)

	h.WriteString(title.Render(" ATLAS WARLORD: TACTICAL OPS MANUAL ") + "\n\n")

	h.WriteString(section.Render("MISSION DIRECTIVE") + "\n")
	h.WriteString("LOCATE THE PRINCESS (P) AND NEUTRALIZE THE DRAGON (D).\n")
	h.WriteString("DRAGON STATS: 350 HP | 45 ATK | 25 DEF.\n")
	h.WriteString("NEUTRALIZE MINIONS (S/K) TO EARN GOLD AND XP.\n\n")

	h.WriteString(section.Render("COMMAND & CONTROL") + "\n")
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[WASD]"), "MOVE TACTICAL CURSOR (X)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[SPACE]"), "SELECT COMMANDER (C)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[ENTER]"), "EXECUTE ORDER (TARGET X -> M)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[N]"), "END DAY CYCLE (REFILLS AP)"))
	h.WriteString(fmt.Sprintf(" %-12s %s\n", key.Render("[H]"), "TOGGLE MANUAL"))

	h.WriteString(section.Render("COMBAT DATA") + "\n")
	h.WriteString("ATTACKING SETS AP TO 0. COMBAT IS A DICE DUEL:\n")
	h.WriteString(" - POWER: [ATK + 1D6] | SHIELD: [DEF + 1D6]\n")
	h.WriteString(" - SURVIVORS LAUNCH A COUNTER-STRIKE FOR 50% POWER.\n\n")

	h.WriteString(section.Render("LOGISTICS & LEVELING") + "\n")
	h.WriteString(" - CACHE ($): GRANTS INSTANT XP AND GOLD.\n")
	h.WriteString(" - MARKET (B): BUY PERMANENT ATK/DEF UPGRADES.\n")
	h.WriteString(" - RANK UP: EVERY 150 XP INCREASES STATS AND FULL HEALS.\n\n")

	h.WriteString(section.Render("MAP LEGEND") + "\n")
	h.WriteString(" UNITS:  C YOU  D DRAGON  P PRINCESS  S SCOUT  K KNIGHT\n")
	h.WriteString(" LOCS:   B MARKET  $ CACHE  H BASE  | - + CITY WALLS\n")
	h.WriteString(" TILES:  . LAND  ↑ FOREST  ~ WATER  ▲ MOUNTAIN\n\n")

	h.WriteString(lipgloss.NewStyle().Width(ViewWidth + SidebarWidth).Align(lipgloss.Center).Render("PRESS [H] TO RESUME MISSION"))

	return "\n" + helpStyle.Render(h.String())
}
