package city

import (
	"math/rand"
	"time"
)

type TileType int

const (
	Empty TileType = iota
	RoadV      // ║
	RoadH      // ═
	RoadTL     // ╔
	RoadTR     // ╗
	RoadBL     // ╚
	RoadBR     // ╝
	RoadTU     // ╩
	RoadTD     // ╦
	RoadTLT    // ╠
	RoadTRT    // ╣
	RoadCross  // ╬
	Building
	Park
	Commercial
	Water
)

// Sockets: [Top, Right, Bottom, Left]
// 0: No connection, 1: Road connection
var sockets = map[TileType][4]int{
	RoadV:     {1, 0, 1, 0},
	RoadH:     {0, 1, 0, 1},
	RoadTL:    {0, 1, 1, 0},
	RoadTR:    {0, 0, 1, 1},
	RoadBL:    {1, 1, 0, 0},
	RoadBR:    {1, 0, 0, 1},
	RoadTU:    {1, 1, 0, 1},
	RoadTD:    {0, 1, 1, 1},
	RoadTLT:   {1, 1, 1, 0},
	RoadTRT:   {1, 0, 1, 1},
	RoadCross: {1, 1, 1, 1},
	Building:  {0, 0, 0, 0},
	Park:      {0, 0, 0, 0},
	Commercial: {0, 0, 0, 0},
	Water:     {0, 0, 0, 0},
}

var weights = map[TileType]int{
	RoadV:     10, RoadH: 10,
	RoadTL:    5, RoadTR: 5, RoadBL: 5, RoadBR: 5,
	RoadTU:    3, RoadTD: 3, RoadTLT: 3, RoadTRT: 3,
	RoadCross: 2,
	Building:  40,
	Park:      15,
	Commercial: 20,
	Water:     10,
}

type Tile struct {
	Type          TileType
	Possibilities []TileType
	Collapsed     bool
}

type WFC struct {
	Width  int
	Height int
	Grid   [][]Tile
}

func NewWFC(width, height int) *WFC {
	rand.Seed(time.Now().UnixNano())
	w := &WFC{
		Width:  width,
		Height: height,
		Grid:   make([][]Tile, height),
	}

	allTypes := []TileType{
		RoadV, RoadH, RoadTL, RoadTR, RoadBL, RoadBR,
		RoadTU, RoadTD, RoadTLT, RoadTRT, RoadCross,
		Building, Park, Commercial, Water,
	}

	for y := 0; y < height; y++ {
		w.Grid[y] = make([]Tile, 0, width)
		for x := 0; x < width; x++ {
			w.Grid[y] = append(w.Grid[y], Tile{
				Possibilities: append([]TileType{}, allTypes...),
				Collapsed:     false,
			})
		}
	}

	// Seed some starting points
	w.collapseTile(width/2, height/2, RoadCross)
	w.collapseTile(rand.Intn(width), rand.Intn(height), Water)

	return w
}

func (w *WFC) collapseTile(x, y int, t TileType) {
	tile := &w.Grid[y][x]
	if tile.Collapsed { return }
	tile.Type = t
	tile.Collapsed = true
	tile.Possibilities = []TileType{t}
	w.Propagate(x, y)
}

func (w *WFC) GetEntropy(x, y int) int {
	if w.Grid[y][x].Collapsed { return 999 }
	return len(w.Grid[y][x].Possibilities)
}

func (w *WFC) Collapse() bool {
	minEntropy := 999
	var candidates [][2]int

	for y := 0; y < w.Height; y++ {
		for x := 0; x < w.Width; x++ {
			e := w.GetEntropy(x, y)
			if e < minEntropy {
				minEntropy = e
				candidates = [][2]int{{x, y}}
			} else if e == minEntropy {
				candidates = append(candidates, [2]int{x, y})
			}
		}
	}

	if minEntropy == 999 { return true }

	c := candidates[rand.Intn(len(candidates))]
	x, y := c[0], c[1]

	tile := &w.Grid[y][x]
	if len(tile.Possibilities) == 0 { return false }

	typeWeights := make(map[TileType]int)
	for _, p := range tile.Possibilities {
		typeWeights[p] = weights[p]
	}

	totalWeight := 0
	for _, weight := range typeWeights { totalWeight += weight }
	if totalWeight == 0 { return false }

	pick := rand.Intn(totalWeight)
	current := 0
	var selected TileType
	for _, p := range tile.Possibilities {
		current += typeWeights[p]
		if pick < current {
			selected = p
			break
		}
	}

	tile.Type = selected
	tile.Collapsed = true
	tile.Possibilities = []TileType{selected}

	w.Propagate(x, y)
	return false
}

func (w *WFC) Propagate(x, y int) {
	stack := [][2]int{{x, y}}
	seen := make(map[[2]int]bool)

	for len(stack) > 0 {
		curr := stack[0]
		stack = stack[1:]
		cx, cy := curr[0], curr[1]

		// Directions: 0:Up, 1:Right, 2:Down, 3:Left
		dirs := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

		for d, offset := range dirs {
			nx, ny := cx+offset[0], cy+offset[1]
			if nx < 0 || nx >= w.Width || ny < 0 || ny >= w.Height { continue }
			if w.Grid[ny][nx].Collapsed { continue }

			// Opposite direction index for neighbor
			// Up(0) -> Down(2), Right(1) -> Left(3), Down(2) -> Up(0), Left(3) -> Right(1)
			oppDir := (d + 2) % 4

			// Valid possibilities for neighbor at (nx, ny)
			// based on the possibilities of current tile at (cx, cy)
			allowedForNeighbor := make(map[TileType]bool)
			for _, p := range w.Grid[cy][cx].Possibilities {
				pSocket := sockets[p][d] // My socket in this direction
				
				// Neighbor's socket in the opposite direction must match
				for targetType, targetSockets := range sockets {
					if targetSockets[oppDir] == pSocket {
						allowedForNeighbor[targetType] = true
					}
				}
			}

			newPossibilities := []TileType{}
			changed := false
			for _, p := range w.Grid[ny][nx].Possibilities {
				if allowedForNeighbor[p] {
					newPossibilities = append(newPossibilities, p)
				} else {
					changed = true
				}
			}

			if changed {
				w.Grid[ny][nx].Possibilities = newPossibilities
				if !seen[[2]int{nx, ny}] {
					stack = append(stack, [2]int{nx, ny})
					seen[[2]int{nx, ny}] = true
				}
			}
		}
	}
}

func (w *WFC) Step() bool {
	return w.Collapse()
}
