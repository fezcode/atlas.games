package wfc

import (
	"math/rand"
	"time"
)

type TileType int

const (
	Empty TileType = iota
	Water
	Land
	Forest
	Mountain
	Lava
)

var weights = map[TileType]int{
	Land:     50,
	Water:    35,
	Forest:   30,
	Mountain: 20,
	Lava:     12,
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

	allTypes := []TileType{Water, Land, Forest, Mountain, Lava}

	for y := 0; y < height; y++ {
		w.Grid[y] = make([]Tile, 0, width)
		for x := 0; x < width; x++ {
			w.Grid[y] = append(w.Grid[y], Tile{
				Possibilities: append([]TileType{}, allTypes...),
				Collapsed:     false,
			})
		}
	}

	// SEEDING: Plant random biomes to ensure diversity
	biomes := []TileType{Water, Forest, Mountain, Lava}
	for _, b := range biomes {
		numSeeds := 2 + rand.Intn(3)
		for i := 0; i < numSeeds; i++ {
			sx, sy := rand.Intn(width), rand.Intn(height)
			w.collapseTile(sx, sy, b)
		}
	}

	return w
}

func (w *WFC) collapseTile(x, y int, t TileType) {
	tile := &w.Grid[y][x]
	if tile.Collapsed {
		return
	}
	tile.Type = t
	tile.Collapsed = true
	tile.Possibilities = []TileType{t}
	w.Propagate(x, y)
}

func (w *WFC) GetEntropy(x, y int) int {
	if w.Grid[y][x].Collapsed {
		return 999
	}
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

	if minEntropy == 999 {
		return true 
	}

	c := candidates[rand.Intn(len(candidates))]
	x, y := c[0], c[1]

	tile := &w.Grid[y][x]
	if len(tile.Possibilities) == 0 {
		return false
	}

	typeWeights := make(map[TileType]int)
	for _, p := range tile.Possibilities {
		typeWeights[p] = weights[p]
	}

	// Neighbor Bonus
	neighbors := [][2]int{{x, y - 1}, {x, y + 1}, {x - 1, y}, {x + 1, y}}
	for _, n := range neighbors {
		nx, ny := n[0], n[1]
		if nx >= 0 && nx < w.Width && ny >= 0 && ny < w.Height {
			neighbor := w.Grid[ny][nx]
			if neighbor.Collapsed {
				if _, ok := typeWeights[neighbor.Type]; ok {
					typeWeights[neighbor.Type] *= 8 
				}
			}
		}
	}

	totalWeight := 0
	for _, weight := range typeWeights {
		totalWeight += weight
	}

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
	rules := map[TileType][]TileType{
		Water:    {Water, Land},
		Land:     {Water, Land, Forest, Mountain},
		Forest:   {Forest, Land},
		Mountain: {Mountain, Land, Lava},
		Lava:     {Lava, Mountain},
	}

	stack := [][2]int{{x, y}}
	seen := make(map[[2]int]bool)

	for len(stack) > 0 {
		curr := stack[0]
		stack = stack[1:]
		cx, cy := curr[0], curr[1]

		neighbors := [][2]int{
			{cx, cy - 1}, {cx, cy + 1}, {cx - 1, cy}, {cx + 1, cy},
		}

		for _, n := range neighbors {
			nx, ny := n[0], n[1]
			if nx < 0 || nx >= w.Width || ny < 0 || ny >= w.Height {
				continue
			}

			if w.Grid[ny][nx].Collapsed {
				continue
			}

			possibleForNeighbor := make(map[TileType]bool)
			for _, p := range w.Grid[cy][cx].Possibilities {
				for _, allowed := range rules[p] {
					possibleForNeighbor[allowed] = true
				}
			}

			newPossibilities := []TileType{}
			changed := false
			for _, p := range w.Grid[ny][nx].Possibilities {
				if possibleForNeighbor[p] {
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
