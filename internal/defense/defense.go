package defense

import (
	"math"
	"math/rand"
	"time"
)

type CellType int

const (
	Empty CellType = iota
	Path
	TowerCell
	Base
	Spawn
)

type Enemy struct {
	HP        int
	MaxHP     int
	PathIndex int // Index in the path slice
	X, Y      int
	Killed    bool
	Reached   bool
}

type Tower struct {
	X, Y     int
	Damage   int
	Range    float64
	Cooldown int // Current cooldown counter
	MaxCD    int // Maximum cooldown
	Target   *Enemy
}

type Projectile struct {
	X, Y       float64
	TargetX, TargetY float64
	Target     *Enemy
	Speed      float64
	Damage     int
	Active     bool
}

type Game struct {
	Width, Height int
	Grid          [][]CellType
	Path          [][2]int
	Enemies       []*Enemy
	Towers        []*Tower
	Projectiles   []*Projectile
	Gold          int
	Health        int
	Wave          int
	TickCount     int
	NextWaveIn    int // Ticks until next wave
}

func NewGame(w, h int) *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		Width:      w,
		Height:     h,
		Grid:       make([][]CellType, h),
		Gold:       200, // Doubled starting gold
		Health:     50,  // Increased base health
		Wave:       0,
		NextWaveIn: 30,
	}

	for y := 0; y < h; y++ {
		g.Grid[y] = make([]CellType, w)
	}

	// Generate a simple zigzag path
	g.generatePath()

	return g
}

func (g *Game) generatePath() {
	// Simple horizontal path with some vertical jogs
	x, y := 0, g.Height/2
	g.Path = append(g.Path, [2]int{x, y})
	g.Grid[y][x] = Spawn

	for x < g.Width-2 {
		// Move right 5-8 steps
		steps := 6 + rand.Intn(4) // Slightly longer segments
		for i := 0; i < steps && x < g.Width-2; i++ {
			x++
			g.Path = append(g.Path, [2]int{x, y})
			g.Grid[y][x] = Path
		}
		// Move up or down
		dy := 2 + rand.Intn(2)
		if y > g.Height-7 { dy = -dy } 
		if y < 6 { dy = int(math.Abs(float64(dy))) } 
		if rand.Float64() < 0.5 { dy = -dy }

		direction := 1
		if dy < 0 { direction = -1; dy = -dy }

		for i := 0; i < dy; i++ {
			y += direction
			if y < 0 { y = 0; break }
			if y >= g.Height { y = g.Height - 1; break }
			g.Path = append(g.Path, [2]int{x, y})
			g.Grid[y][x] = Path
		}
	}

	// Connect to final base on the right edge
	for x < g.Width-1 {
		x++
		g.Path = append(g.Path, [2]int{x, y})
		g.Grid[y][x] = Path
	}
	g.Grid[y][x] = Base
}

func (g *Game) Tick() {
	g.TickCount++

	// 1. Spawn Enemies
	if g.NextWaveIn > 0 {
		g.NextWaveIn--
	} else {
		g.Wave++
		g.NextWaveIn = 150 // Even more time between waves
		// Spawn a wave of enemies
		for i := 0; i < 4+g.Wave; i++ {
			hp := 2 + g.Wave // Much slower HP scaling
			g.Enemies = append(g.Enemies, &Enemy{
				HP:        hp,
				MaxHP:     hp,
				PathIndex: -i * 6, // More spacing to prevent overwhelming
			})
		}
	}

	// 2. Move Enemies
	for _, e := range g.Enemies {
		if e.Killed || e.Reached { continue }
		e.PathIndex++
		if e.PathIndex >= 0 && e.PathIndex < len(g.Path) {
			pos := g.Path[e.PathIndex]
			e.X, e.Y = pos[0], pos[1]
		} else if e.PathIndex >= len(g.Path) {
			e.Reached = true
			g.Health--
			if g.Health < 0 { g.Health = 0 }
		}
	}

	// 3. Tower Logic
	for _, t := range g.Towers {
		if t.Cooldown > 0 {
			t.Cooldown--
			continue
		}

		// Find target
		var bestTarget *Enemy
		for _, e := range g.Enemies {
			if e.Killed || e.Reached || e.PathIndex < 0 { continue }
			dist := math.Sqrt(float64((t.X-e.X)*(t.X-e.X) + (t.Y-e.Y)*(t.Y-e.Y)))
			if dist <= t.Range {
				if bestTarget == nil || e.PathIndex > bestTarget.PathIndex {
					bestTarget = e
				}
			}
		}

		if bestTarget != nil {
			g.Projectiles = append(g.Projectiles, &Projectile{
				X:       float64(t.X),
				Y:       float64(t.Y),
				TargetX: float64(bestTarget.X),
				TargetY: float64(bestTarget.Y),
				Target:  bestTarget,
				Speed:   1.5,
				Damage:  3,
				Active:  true,
			})
			t.Cooldown = t.MaxCD
			t.Target = bestTarget
		} else {
			t.Target = nil
		}
	}

	// 4. Projectile Logic
	for _, p := range g.Projectiles {
		if !p.Active { continue }
		
		// Homing: Update destination to target's current position
		if p.Target != nil && !p.Target.Killed && !p.Target.Reached {
			p.TargetX = float64(p.Target.X)
			p.TargetY = float64(p.Target.Y)
		}

		dx := p.TargetX - p.X
		dy := p.TargetY - p.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 1.0 {
			// Hit! Apply damage directly to the specific target
			if p.Target != nil && !p.Target.Killed && !p.Target.Reached {
				p.Target.HP -= p.Damage
				if p.Target.HP <= 0 && !p.Target.Killed {
					p.Target.Killed = true
					g.Gold += 15
				}
			}
			p.Active = false
		} else {
			p.X += (dx / dist) * p.Speed
			p.Y += (dy / dist) * p.Speed
		}
	}

	// Clean up
	activeEnemies := []*Enemy{}
	for _, e := range g.Enemies {
		if !e.Killed && !e.Reached {
			activeEnemies = append(activeEnemies, e)
		}
	}
	g.Enemies = activeEnemies

	activeProjectiles := []*Projectile{}
	for _, p := range g.Projectiles {
		if p.Active {
			activeProjectiles = append(activeProjectiles, p)
		}
	}
	g.Projectiles = activeProjectiles
}

func (g *Game) PlaceTower(x, y int) bool {
	if g.Gold < 15 { return false } // Even cheaper towers
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height { return false }
	if g.Grid[y][x] != Empty { return false }

	for _, t := range g.Towers {
		if t.X == x && t.Y == y { return false }
	}

	g.Towers = append(g.Towers, &Tower{
		X: x, Y: y,
		Damage: 3,
		Range: 7.0, // Increased range
		MaxCD: 3,   // Faster firing
	})
	g.Gold -= 15
	g.Grid[y][x] = TowerCell
	return true
}

func (g *Game) SellTower(x, y int) bool {
	for i, t := range g.Towers {
		if t.X == x && t.Y == y {
			g.Towers = append(g.Towers[:i], g.Towers[i+1:]...)
			g.Gold += 10 // Partial refund
			g.Grid[y][x] = Empty
			return true
		}
	}
	return false
}
