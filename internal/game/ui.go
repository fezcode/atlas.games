package game

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	chickenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	jumpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	obsStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	carStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true)
	enemyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true)
	treeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))
	sunStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	shootStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Underline(true)
	groundStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	deathStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Italic(true)
)

type tickMsg time.Time

type Model struct {
	State *GameState
}

func NewModel() Model {
	return Model{
		State: NewGameState(),
	}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if m.State.GameOver {
				m.State = NewGameState()
				m.State.Started = true
				return m, tick()
			}
		case "s", " ":
			if !m.State.Started {
				m.State.Started = true
				return m, tick()
			}
			if !m.State.GameOver {
				m.State.Shoot()
			}
		case "up", "k", "w":
			if m.State.Started {
				m.State.MoveUp()
			}
		case "down", "j", "x":
			if m.State.Started {
				m.State.MoveDown()
			}
		case "right", "l", "d":
			if m.State.Started {
				m.State.Jump()
			}
		}

	case tickMsg:
		if m.State.Started && !m.State.GameOver {
			m.State.Update()
			return m, tick()
		}
	}

	return m, nil
}

// stamp renders a string into the buffer at (x, y) with a specific style,
// ensuring each character takes exactly one cell.
func stamp(buffer [][]string, x, y int, s string, style lipgloss.Style) {
	if y < 0 || y >= len(buffer) {
		return
	}
	for i, char := range s {
		if x+i >= 0 && x+i < len(buffer[y]) {
			buffer[y][x+i] = style.Render(string(char))
		}
	}
}

func (m Model) View() string {
	if !m.State.Started {
		return fmt.Sprintf("\n\n   %s\n\n   Wilson is back, and this time it's personal.\n\n   [W/X] Up/Down\n   [D]   Jump\n   [SPC] Shoot (Requires Gun)\n\n   Press [S] or [SPACE] to Start\n", titleStyle.Render("WILSON'S REVENGE"))
	}

	if m.State.GameOver {
		return fmt.Sprintf("\n\n   %s\n\n   WILSON MET HIS FATE...\n\n   FINAL SCORE: %d\n\n   [R] RESTART  [Q] QUIT\n", deathStyle.Render("lo siento, Wilson"), m.State.Score)
	}

	SkyHeight := 6
	BGHeight := 5
	TotalViewHeight := SkyHeight + BGHeight + (Lanes * LaneHeight)

	buffer := make([][]string, TotalViewHeight)
	for i := range buffer {
		buffer[i] = make([]string, GameWidth)
		for j := range buffer[i] {
			buffer[i][j] = " "
		}
	}

	// Sun
	stamp(buffer, GameWidth-15, 1, "\\ | /", sunStyle)
	stamp(buffer, GameWidth-15, 2, "-- O --", sunStyle)
	stamp(buffer, GameWidth-15, 3, "/ | \\", sunStyle)

	// Background (Trees)
	for _, bg := range m.State.Background {
		x := int(bg.X)
		stamp(buffer, x, SkyHeight+1, "  ^  ", treeStyle)
		stamp(buffer, x, SkyHeight+2, " / \\ ", treeStyle)
		stamp(buffer, x, SkyHeight+3, "/   \\", treeStyle)
		stamp(buffer, x, SkyHeight+4, "  |  ", treeStyle)
	}

	// Lanes
	for l := 0; l < Lanes; l++ {
		laneY := SkyHeight + BGHeight + (l * LaneHeight)
		
		// Ground
		for x := 0; x < GameWidth; x++ {
			buffer[laneY+LaneHeight-1][x] = groundStyle.Render("_")
		}

		// Objects
		for _, obj := range m.State.Objects {
			if obj.Lane == l {
				x := int(obj.X)
				switch obj.Type {
				case TypeObstacle:
					stamp(buffer, x, laneY+1, "[XX]", obsStyle)
					stamp(buffer, x, laneY+2, "[XX]", obsStyle)
					stamp(buffer, x, laneY+3, "[XX]", obsStyle)
				case TypeCar:
					stamp(buffer, x, laneY+1, " ____ ", carStyle)
					stamp(buffer, x, laneY+2, "[_||_\\", carStyle)
					stamp(buffer, x, laneY+3, "\"O--O\"", carStyle)
				case TypeEnemy:
					stamp(buffer, x, laneY+1, " (X) ", enemyStyle)
					stamp(buffer, x, laneY+2, " /|\\ ", enemyStyle)
					stamp(buffer, x, laneY+3, " / \\ ", enemyStyle)
				case TypeGun:
					stamp(buffer, x, laneY+2, "~-=", lipgloss.NewStyle().Foreground(lipgloss.Color("214")))
				case TypeAmmo:
					stamp(buffer, x, laneY+2, "[A]", lipgloss.NewStyle().Foreground(lipgloss.Color("45")))
				}
			}
		}

		// Wilson
		if m.State.ChickenLane == l {
			wx := ChickenX
			wy := laneY + 1
			style := chickenStyle
			if m.State.Jumping {
				wy -= 1
				style = jumpStyle
				stamp(buffer, wx, wy, " (o> ", style)
				stamp(buffer, wx, wy+1, "<| |>", style)
				stamp(buffer, wx, wy+2, " L L ", style)
			} else {
				stamp(buffer, wx, wy, " (o> ", style)
				if m.State.HasGun {
					stamp(buffer, wx+4, wy, "--=", style)
					if m.State.IsShooting {
						for bx := wx + 7; bx < GameWidth; bx++ {
							buffer[wy][bx] = shootStyle.Render("-")
						}
					}
				}
				stamp(buffer, wx, wy+1, " / ) ", style)
				stamp(buffer, wx, wy+2, " L L ", style)
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(" %s | Score: %d | Ammo: %d | Speed: %.2f\n\n", titleStyle.Render("WILSON'S REVENGE"), m.State.Score, m.State.Ammo, m.State.Speed))
	for _, line := range buffer {
		sb.WriteString(strings.Join(line, "") + "\n")
	}
	sb.WriteString("\n [W/X] Up/Down | [D] Jump | [Space] Shoot | [Q] Quit")
	return sb.String()
}

func tick() tea.Cmd {
	return tea.Every(time.Millisecond*40, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
