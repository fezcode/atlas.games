package colony

import (
	"math"
	"math/rand"
	"time"
)

type CellType int

const (
	Empty CellType = iota
	Dirt
	Tunnel
	Food
	CellAnt
	Queen
	CellSpider
)

type Ant struct {
	X, Y     int
	HasFood  bool
	Activity string // "digging", "foraging", "returning"
	TargetX  int    
}

type Spider struct {
	X, Y     int
	OriginX  int
	OriginY  int
}

type Colony struct {
	Width, Height int
	Grid          [][]CellType
	Ants          []*Ant
	Spiders       []*Spider
}

func NewColony(w, h int) *Colony {
	rand.Seed(time.Now().UnixNano())
	c := &Colony{
		Width:  w,
		Height: h,
		Grid:   make([][]CellType, h),
	}

	for y := 0; y < h; y++ {
		c.Grid[y] = make([]CellType, w)
		for x := 0; x < w; x++ {
			if y > 5 {
				c.Grid[y][x] = Dirt
			} else {
				c.Grid[y][x] = Empty
			}
		}
	}

	cx, cy := w/2, h/2
	for dy := -2; dy <= 2; dy++ {
		for dx := -3; dx <= 3; dx++ {
			c.Grid[cy+dy][cx+dx] = Tunnel
		}
	}
	c.Grid[cy][cx] = Queen

	for i := 0; i < 10; i++ {
		activity := "foraging"
		if i < 4 { activity = "digging" } 
		c.Ants = append(c.Ants, &Ant{
			X: cx, Y: cy, 
			Activity: activity,
			TargetX: rand.Intn(w),
		})
	}

	for i := 0; i < 20; i++ {
		c.Grid[rand.Intn(4)][rand.Intn(w)] = Food
	}

	for i := 0; i < 50; i++ {
		fx, fy := rand.Intn(w), 6+rand.Intn(h-7)
		if c.Grid[fy][fx] == Dirt {
			c.Grid[fy][fx] = Food
		}
	}

	for i := 0; i < 15; i++ {
		sx, sy := rand.Intn(w), 4+rand.Intn(h-5)
		distToQueen := math.Sqrt(float64((sx-cx)*(sx-cx) + (sy-cy)*(sy-cy)))
		if distToQueen < 20.0 {
			i-- 
			continue
		}
		c.Spiders = append(c.Spiders, &Spider{X: sx, Y: sy, OriginX: sx, OriginY: sy})
	}

	return c
}

func (c *Colony) Tick() {
	for _, s := range c.Spiders {
		if rand.Float64() < 0.08 {
			dx, dy := rand.Intn(3)-1, rand.Intn(3)-1
			nx, ny := s.OriginX+dx, s.OriginY+dy
			if nx >= 0 && nx < c.Width && ny >= 0 && ny < c.Height {
				s.X, s.Y = nx, ny
			}
		}
	}

	newAnts := []*Ant{}
	for _, a := range c.Ants {
		c.updateAnt(a)
		
		killed := false
		for _, s := range c.Spiders {
			// LETHAL 3x3 AREA
			if math.Abs(float64(a.X-s.X)) <= 1 && math.Abs(float64(a.Y-s.Y)) <= 1 {
				killed = true
				break
			}
		}
		if !killed {
			newAnts = append(newAnts, a)
		}
	}
	c.Ants = newAnts
}

func (c *Colony) updateAnt(a *Ant) {
	cx, cy := c.Width/2, c.Height/2
	
	// 1. CARRIER LOGIC: LASER FOCUS ON HOME
	if a.HasFood {
		// Use BFS to find path through tunnels, or greedy if blocked
		nextX, nextY := c.findNextStepHome(a.X, a.Y, cx, cy)
		
		if c.Grid[nextY][nextX] == Dirt {
			c.Grid[nextY][nextX] = Tunnel // Force dig if that's the way home
		}
		
		a.X, a.Y = nextX, nextY

		// Release proximity
		if math.Abs(float64(a.X-cx)) <= 3 && math.Abs(float64(a.Y-cy)) <= 2 {
			a.HasFood = false
			a.Activity = "foraging"
			if rand.Float64() < 0.5 { a.Activity = "digging" }
			a.TargetX = rand.Intn(c.Width)
		}
		return
	}

	// 2. FORAGER/DIGGER LOGIC
	tx, ty := a.TargetX, a.Y
	
	// Scan for food
	foundFood := false
	for dy := -8; dy <= 8; dy++ {
		for dx := -15; dx <= 15; dx++ {
			nx, ny := a.X+dx, a.Y+dy
			if nx >= 0 && nx < c.Width && ny >= 0 && ny < c.Height {
				if c.Grid[ny][nx] == Food {
					tx, ty = nx, ny
					foundFood = true
					break
				}
			}
		}
		if foundFood { break }
	}

	if !foundFood {
		if a.Activity == "digging" {
			ty = c.Height - 1
		} else {
			ty = 0
		}
	}

	// Choose move
	moves := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	bestDX, bestDY := 0, 0
	maxScore := -100000.0

	for _, m := range moves {
		nx, ny := a.X+m[0], a.Y+m[1]
		if nx < 0 || nx >= c.Width || ny < 0 || ny >= c.Height { continue }

		cell := c.Grid[ny][nx]
		dist := math.Sqrt(float64((nx-tx)*(nx-tx) + (ny-ty)*(ny-ty)))
		score := 2000.0 - dist

		// Avoid Spiders
		for _, s := range c.Spiders {
			if math.Abs(float64(nx-s.X)) <= 2 && math.Abs(float64(ny-s.Y)) <= 2 {
				score -= 10000.0
			}
		}

		if cell == Dirt {
			if a.Activity == "digging" {
				score += 500.0 
			} else {
				score -= 1000.0 
			}
		} else {
			score += 100.0 
		}

		score += rand.Float64() * 20.0

		if score > maxScore {
			maxScore = score
			bestDX, bestDY = m[0], m[1]
		}
	}

	if bestDX != 0 || bestDY != 0 {
		a.X += bestDX
		a.Y += bestDY
		if c.Grid[a.Y][a.X] == Dirt {
			c.Grid[a.Y][a.X] = Tunnel
		}
		if c.Grid[a.Y][a.X] == Food {
			a.HasFood = true
			c.Grid[a.Y][a.X] = Empty
			a.Activity = "returning"
		}
	} else {
		a.TargetX = rand.Intn(c.Width)
	}
}

// findNextStepHome uses simple greedy logic with tunnel bias for home-bound carriers
func (c *Colony) findNextStepHome(startX, startY, targetX, targetY int) (int, int) {
	moves := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
	bestX, bestY := startX, startY
	maxScore := -100000.0

	for _, m := range moves {
		nx, ny := startX+m[0], startY+m[1]
		if nx < 0 || nx >= c.Width || ny < 0 || ny >= c.Height { continue }

		dist := math.Sqrt(float64((nx-targetX)*(nx-targetX) + (ny-targetY)*(ny-targetY)))
		score := 5000.0 - dist

		// Carriers hate dirt but will dig if it gets them home
		if c.Grid[ny][nx] == Dirt {
			score -= 50.0 
		} else {
			score += 100.0 
		}

		// Avoid spiders even when carrying food
		for _, s := range c.Spiders {
			if math.Abs(float64(nx-s.X)) <= 1 && math.Abs(float64(ny-s.Y)) <= 1 {
				score -= 20000.0
			}
		}

		if score > maxScore {
			maxScore = score
			bestX, bestY = nx, ny
		}
	}
	return bestX, bestY
}
