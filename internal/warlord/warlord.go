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
	WallV // ║
	WallH // ═
	WallTL // ╔
	WallTR // ╗
	WallBL // ╚
	WallBR // ╝
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
	MaxMoves    int
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
		Logs:   []string{"🛰️ System Boot: Mission Red Dragon active."},
	}

	for y := 0; y < h; y++ {
		g.Grid[y] = make([]Tile, w)
		for x := 0; x < w; x++ {
			r := rand.Float64()
			if r < 0.07 { g.Grid[y][x].Type = Water
			} else if r < 0.15 { g.Grid[y][x].Type = Forest
			} else if r < 0.20 { g.Grid[y][x].Type = Mountain
			} else if r < 0.22 { g.Grid[y][x].Type = Cache
			} else { g.Grid[y][x].Type = Land }
		}
	}

	g.BaseX, g.BaseY = w/2, h/2
	g.Grid[g.BaseY][g.BaseX].Type = Stronghold
	g.Grid[g.BaseY][g.BaseX].Team = TeamPlayer

	// Ensure start area is walkable
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if g.BaseY+dy >= 0 && g.BaseY+dy < h && g.BaseX+dx >= 0 && g.BaseX+dx < w {
				g.Grid[g.BaseY+dy][g.BaseX+dx].Type = Land
			}
		}
	}

	// Place Cities with Markets
	numCities := 4
	for i := 0; i < numCities; i++ {
		cx, cy := rand.Intn(w-10)+5, rand.Intn(h-10)+5
		if math.Abs(float64(cx-g.BaseX)) < 15 && math.Abs(float64(cy-g.BaseY)) < 10 {
			i--
			continue
		}
		g.generateCity(cx, cy)
	}

	// Place Princess
	px, py := 10 + rand.Intn(10), 10 + rand.Intn(10)
	if rand.Float64() < 0.5 { px = w - 15 - rand.Intn(10) }
	if rand.Float64() < 0.5 { py = h - 15 - rand.Intn(10) }
	g.Grid[py][px].Type = Princess

	// Dragon Guard
	g.Units = append(g.Units, &Unit{
		ID: 999, Type: Dragon, Team: TeamInvader, X: px+1, Y: py,
		Health: 350, MaxHP: 350, Level: 10, Attack: 45, Defense: 25, Moves: 2, MaxMoves: 2,
	})

	// Player Start
	g.Units = append(g.Units, &Unit{
		ID: 0, Type: Commander, Team: TeamPlayer, X: g.BaseX, Y: g.BaseY,
		Health: 60, MaxHP: 60, Level: 1, XP: 0, NextLevelXP: 100, Attack: 15, Defense: 8, Moves: 3, MaxMoves: 3, Gold: 100,
	})

	// Spawn Minions
	for i := 0; i < 45; i++ {
		ex, ey := rand.Intn(w), rand.Intn(h)
		if math.Abs(float64(ex-g.BaseX)) > 10 && g.Grid[ey][ex].Type == Land {
			uType := Scout
			hp, atk, def := 15, 8, 4
			maxM := 4
			if rand.Float64() < 0.4 { uType = Knight; hp, atk, def = 30, 12, 6; maxM = 2 }
			g.Units = append(g.Units, &Unit{
				ID: len(g.Units), Type: uType, Team: TeamInvader, X: ex, Y: ey,
				Health: hp, MaxHP: hp, Level: 1, Attack: atk, Defense: def, Moves: maxM, MaxMoves: maxM,
			})
		}
	}

	return g
}

func (g *GameState) generateCity(cx, cy int) {
	size := 5
	for dy := 0; dy < size; dy++ {
		for dx := 0; dx < size; dx++ {
			x, y := cx+dx, cy+dy
			if x >= g.Width || y >= g.Height { continue }
			if (dy == 0 && dx == size/2) || (dy == size-1 && dx == size/2) {
				g.Grid[y][x].Type = Land
				continue
			}
			if dy == 0 { g.Grid[y][x].Type = WallH
			} else if dy == size-1 { g.Grid[y][x].Type = WallH
			} else if dx == 0 { g.Grid[y][x].Type = WallV
			} else if dx == size-1 { g.Grid[y][x].Type = WallV
			} else { g.Grid[y][x].Type = Land }
		}
	}
	g.Grid[cy][cx].Type = WallTL
	g.Grid[cy][cx+size-1].Type = WallTR
	g.Grid[cy+size-1][cx].Type = WallBL
	g.Grid[cy+size-1][cx+size-1].Type = WallBR
	g.Grid[cy+size/2][cx+size/2].Type = Market
}

func (g *GameState) AddLog(msg string) {
	g.Logs = append(g.Logs, msg)
	if len(g.Logs) > 6 { g.Logs = g.Logs[1:] }
}

func (g *GameState) NextTurn() {
	g.Turn++
	for _, u := range g.Units {
		if u.Team == TeamPlayer { u.Moves = u.MaxMoves }
	}
	for i := 0; i < len(g.Units); i++ {
		u := g.Units[i]
		if u.Team == TeamInvader && u.Health > 0 { g.aiAction(u) }
	}
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
			g.AddLog(fmt.Sprintf("🔥 %v ATK: -%d HP!", u.Type, dmg))
			return
		}
		moveX, moveY := u.X, u.Y
		if u.X < commander.X { moveX++ } else if u.X > commander.X { moveX-- }
		if u.Y < commander.Y { moveY++ } else if u.Y > commander.Y { moveY-- }
		if g.Grid[moveY][moveX].Type == Land || g.Grid[moveY][moveX].Type == Market {
			u.X, u.Y = moveX, moveY
		}
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
		u.Health = u.MaxHP // FULL HEAL ON LEVEL UP
		g.AddLog(fmt.Sprintf("⭐ RANK UP: Level %d achieved (HEALED).", u.Level))
	}
}

func (g *GameState) MoveUnit(unitID int, targetX, targetY int) string {
	u := g.Units[unitID]
	dx := int(math.Abs(float64(u.X - targetX)))
	dy := int(math.Abs(float64(u.Y - targetY)))
	dist := dx; if dy > dx { dist = dy }

	if dist > u.Moves { return "OUT OF RANGE" }

	if g.Grid[targetY][targetX].Type == Cache {
		xpGain := 75
		goldGain := 50
		u.XP += xpGain
		u.Gold += goldGain
		g.Grid[targetY][targetX].Type = Land
		u.X, u.Y = targetX, targetY
		u.Moves -= dist
		g.CheckLevelUp(u)
		return fmt.Sprintf("📦 CACHE: +%d XP / +%d GOLD", xpGain, goldGain)
	}

	tType := g.Grid[targetY][targetX].Type
	if tType == Water || tType == Mountain || tType == WallV || tType == WallH || tType == WallTL || tType == WallTR || tType == WallBL || tType == WallBR {
		return "PATH BLOCKED"
	}

	u.X, u.Y = targetX, targetY
	u.Moves -= dist
	return fmt.Sprintf("MOVED TO [%d, %d]", targetX-g.BaseX, g.BaseY-targetY)
}
