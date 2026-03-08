package warlord

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type TileType int

const (
	Water TileType = iota
	Land
	Forest
	Mountain
	Road
	Stronghold
	Cache
	Market
	Princess
)

type UnitType int

const (
	Commander UnitType = iota
	Scout
	Archer
	Knight
	Dragon
)

type Team int

const (
	TeamNeutral Team = iota
	TeamPlayer
	TeamInvader
)

type Unit struct {
	ID          int
	Type        UnitType
	Team        Team
	X, Y        int
	Health      int
	MaxHP       int
	Level       int
	XP          int
	NextLevelXP int
	Attack      int
	Defense     int
	Moves       int
	Gold        int
}

type Tile struct {
	Type TileType
	Team Team
}

type GameState struct {
	Width, Height int
	Grid          [][]Tile
	Units         []*Unit
	Turn          int
	Score         int
	GameOver      bool
	GameWon       bool
	BaseX, BaseY  int
	Logs          []string
}

func NewGame(w, h int) *GameState {
	rand.Seed(time.Now().UnixNano())
	g := &GameState{
		Width:  w,
		Height: h,
		Grid:   make([][]Tile, h),
		Turn:   1,
		Logs:   []string{"🛰️ System Boot: Save the Princess mission active."},
	}

	for y := 0; y < h; y++ {
		g.Grid[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			r := rand.Float64()
			if r < 0.07 { g.Grid[y][x].Type = Water
			} else if r < 0.15 { g.Grid[y][x].Type = Forest
			} else if r < 0.20 { g.Grid[y][x].Type = Mountain
			} else if r < 0.22 { g.Grid[y][x].Type = Cache
			} else if r < 0.24 { g.Grid[y][x].Type = Market
			} else { g.Grid[y][x].Type = Land }
		}
	}

	g.BaseX, g.BaseY = w/2, h/2
	g.Grid[g.BaseY][g.BaseX].Type = Stronghold
	g.Grid[g.BaseY][g.BaseX].Team = TeamPlayer

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if g.BaseY+dy >= 0 && g.BaseY+dy < h && g.BaseX+dx >= 0 && g.BaseX+dx < w {
				g.Grid[g.BaseY+dy][g.BaseX+dx].Type = Land
			}
		}
	}

	// ALWAYS ADD COMMANDER FIRST (Index 0)
	g.Units = append(g.Units, &Unit{
		ID: 0, Type: Commander, Team: TeamPlayer, X: g.BaseX, Y: g.BaseY,
		Health: 60, MaxHP: 60, Level: 1, XP: 0, NextLevelXP: 100, Attack: 15, Defense: 8, Moves: 3, Gold: 100,
	})

	// Place Princess
	px, py := 10 + rand.Intn(10), 10 + rand.Intn(10)
	if rand.Float64() < 0.5 { px = w - 10 - rand.Intn(10) }
	if rand.Float64() < 0.5 { py = h - 10 - rand.Intn(10) }
	g.Grid[py][px].Type = Princess

	// Dragon Guard
	g.Units = append(g.Units, &Unit{
		ID: 999, Type: Dragon, Team: TeamInvader, X: px+1, Y: py,
		Health: 350, MaxHP: 350, Level: 10, Attack: 45, Defense: 25, Moves: 2,
	})

	// Minions
	for i := 0; i < 45; i++ {
		ex, ey := rand.Intn(w), rand.Intn(h)
		if math.Abs(float64(ex-g.BaseX)) > 10 && g.Grid[ey][ex].Type == Land {
			uType := Scout
			hp, atk, def := 15, 8, 4
			if rand.Float64() < 0.4 { uType = Knight; hp, atk, def = 30, 12, 6 }
			g.Units = append(g.Units, &Unit{
				ID: len(g.Units), Type: uType, Team: TeamInvader, X: ex, Y: ey,
				Health: hp, MaxHP: hp, Level: 1, Attack: atk, Defense: def, Moves: 2,
			})
		}
	}

	return g
}

func (g *GameState) AddLog(msg string) {
	g.Logs = append(g.Logs, msg)
	if len(g.Logs) > 6 { g.Logs = g.Logs[1:] }
}

func (g *GameState) NextTurn() {
	g.Turn++
	
	// Reset Moves
	for _, u := range g.Units {
		if u.Team == TeamPlayer { u.Moves = g.getUnitMaxMoves(u.Type) }
	}

	// AI Turn
	for i := 0; i < len(g.Units); i++ {
		u := g.Units[i]
		if u.Team == TeamInvader && u.Health > 0 {
			g.aiAction(u)
		}
	}
	
	// Cleanup and Check Win/Loss
	active := []*Unit{}
	commanderAlive := false
	dragonAlive := false
	for _, u := range g.Units {
		if u.Health > 0 { 
			active = append(active, u) 
			if u.Type == Commander { commanderAlive = true }
			if u.Type == Dragon { dragonAlive = true }
		}
	}
	g.Units = active
	
	if !dragonAlive { g.GameWon = true }
	if !commanderAlive { g.GameOver = true }
}

func (g *GameState) aiAction(u *Unit) {
	// Find commander safely
	var commander *Unit
	for _, unit := range g.Units {
		if unit.Type == Commander { commander = unit; break }
	}
	if commander == nil { return }

	distX := int(math.Abs(float64(u.X - commander.X)))
	distY := int(math.Abs(float64(u.Y - commander.Y)))
	dist := distX; if distY > distX { dist = distY }

	huntRange := 15
	if u.Type == Dragon { huntRange = 8 }

	if dist < huntRange {
		if dist <= 1 {
			roll := rand.Intn(6) + 1
			dmg := (u.Attack + roll) - commander.Defense
			if dmg < 1 { dmg = 1 }
			commander.Health -= dmg
			g.AddLog(fmt.Sprintf("🔥 %v ATK: -%d HP to you!", u.Type, dmg))
			return
		}
		moveX, moveY := u.X, u.Y
		if u.X < commander.X { moveX++ } else if u.X > commander.X { moveX-- }
		if u.Y < commander.Y { moveY++ } else if u.Y > commander.Y { moveY-- }
		if g.Grid[moveY][moveX].Type != Water && g.Grid[moveY][moveX].Type != Mountain {
			u.X, u.Y = moveX, moveY
		}
	}
}

func (g *GameState) getUnitMaxMoves(t UnitType) int {
	switch t {
	case Scout: return 4
	case Commander: return 3
	case Dragon: return 2
	default: return 2
	}
}

func (g *GameState) CheckLevelUp(u *Unit) {
	for u.XP >= u.NextLevelXP {
		u.Level++
		u.XP -= u.NextLevelXP
		u.NextLevelXP = u.Level * u.Level * 100
		u.Attack += 4
		u.Defense += 3
		u.MaxHP += 20
		u.Health = u.MaxHP
		g.AddLog(fmt.Sprintf("⭐ RANK UP: Level %d achieved.", u.Level))
	}
}

func (g *GameState) MoveUnit(unitID int, targetX, targetY int) string {
	u := g.Units[unitID]
	dx := int(math.Abs(float64(u.X - targetX)))
	dy := int(math.Abs(float64(u.Y - targetY)))
	dist := dx; if dy > dx { dist = dy }

	if dist > u.Moves { return "OUT OF RANGE" }

	if g.Grid[targetY][targetX].Type == Cache {
		u.XP += 75; u.Gold += 40
		g.Grid[targetY][targetX].Type = Land
		u.X, u.Y = targetX, targetY
		u.Moves -= dist
		g.CheckLevelUp(u)
		return "📦 CACHE: +75 XP / +40G"
	}

	if g.Grid[targetY][targetX].Type == Water || g.Grid[targetY][targetX].Type == Mountain {
		return "PATH BLOCKED"
	}

	u.X, u.Y = targetX, targetY
	u.Moves -= dist
	return fmt.Sprintf("MOVED TO [%d, %d]", targetX-g.BaseX, g.BaseY-targetY)
}
