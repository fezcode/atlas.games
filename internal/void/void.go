package void

import (
	"fmt"
	"math/rand"
	"time"
)

type Resource int

const (
	Fuel Resource = iota
	Food
	Metal
	Tech
	Artifacts
)

func (r Resource) String() string {
	return [...]string{"Fuel", "Food", "Metal", "Tech", "Artifacts"}[r]
}

type Planet struct {
	Name        string
	Resources   map[Resource]int // Prices
	Stock       map[Resource]int // Amount available
	Description string
}

type Ship struct {
	Fuel       int
	MaxFuel    int
	Credits    int
	Cargo      map[Resource]int
	MaxCargo   int
	CurrentPos int // Index of the planet in the system
}

type Game struct {
	Planets    []*Planet
	Ship       *Ship
	Day        int
	Log        []string
	SelectedRes Resource
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		Ship: &Ship{
			Fuel:    100,
			MaxFuel: 100,
			Credits: 1000,
			Cargo:   make(map[Resource]int),
			MaxCargo: 10,
		},
		Day: 1,
	}

	names := []string{"Xylos", "Vandor", "Chronos", "Aether", "Nox", "Zephyr", "Icarus", "Solaris", "Luna", "Titan"}
	descs := []string{"A desert world rich in ancient artifacts.", "A bustling industrial hub.", "A frozen wasteland on the edge of the void.", "A lush jungle moon.", "A dark, mysterious planet with high-tech outposts."}

	for i := 0; i < 6; i++ {
		p := &Planet{
			Name:        names[rand.Intn(len(names))],
			Resources:   make(map[Resource]int),
			Stock:       make(map[Resource]int),
			Description: descs[rand.Intn(len(descs))],
		}
		// Randomize prices
		p.Resources[Fuel] = 5 + rand.Intn(10)
		p.Resources[Food] = 10 + rand.Intn(20)
		p.Resources[Metal] = 20 + rand.Intn(40)
		p.Resources[Tech] = 50 + rand.Intn(100)
		p.Resources[Artifacts] = 100 + rand.Intn(200)

		// Randomize stock
		for r := Fuel; r <= Artifacts; r++ {
			p.Stock[r] = 5 + rand.Intn(20)
		}
		g.Planets = append(g.Planets, p)
	}

	g.AddLog("Welcome to the Void, Captain. Profit awaits.")
	return g
}

func (g *Game) AddLog(msg string) {
	g.Log = append(g.Log, fmt.Sprintf("Day %d: %s", g.Day, msg))
	if len(g.Log) > 5 {
		g.Log = g.Log[1:]
	}
}

func (g *Game) Travel(planetIdx int) bool {
	if planetIdx == g.Ship.CurrentPos { return false }
	dist := 10 + rand.Intn(15)
	if g.Ship.Fuel < dist {
		g.AddLog("Insufficient fuel for jump.")
		return false
	}

	g.Ship.Fuel -= dist
	g.Ship.CurrentPos = planetIdx
	g.Day++
	g.AddLog(fmt.Sprintf("Jumped to %s. Consumed %d fuel.", g.Planets[planetIdx].Name, dist))
	
	// Refresh market prices slightly on travel
	for _, p := range g.Planets {
		for r := Fuel; r <= Artifacts; r++ {
			p.Resources[r] += rand.Intn(5) - 2
			if p.Resources[r] < 1 { p.Resources[r] = 1 }
		}
	}
	return true
}

func (g *Game) Buy(res Resource, amount int) bool {
	p := g.Planets[g.Ship.CurrentPos]
	price := p.Resources[res] * amount
	
	if p.Stock[res] < amount {
		g.AddLog("Market: Insufficient stock.")
		return false
	}
	if g.Ship.Credits < price {
		g.AddLog("Market: Insufficient credits.")
		return false
	}

	currentCargo := 0
	for _, a := range g.Ship.Cargo { currentCargo += a }
	if currentCargo+amount > g.Ship.MaxCargo {
		g.AddLog("Ship: Cargo hold full.")
		return false
	}

	if res == Fuel {
		g.Ship.Fuel += amount * 10 // Fuel resource adds to tank
		if g.Ship.Fuel > g.Ship.MaxFuel { g.Ship.Fuel = g.Ship.MaxFuel }
	} else {
		g.Ship.Cargo[res] += amount
	}

	p.Stock[res] -= amount
	g.Ship.Credits -= price
	g.AddLog(fmt.Sprintf("Bought %d units of %s.", amount, res.String()))
	return true
}

func (g *Game) Sell(res Resource, amount int) bool {
	if g.Ship.Cargo[res] < amount {
		g.AddLog("Ship: Insufficient cargo.")
		return false
	}

	p := g.Planets[g.Ship.CurrentPos]
	price := (p.Resources[res] - 2) * amount // Sell at a slight loss or market rate
	if price < 1 { price = 1 }

	g.Ship.Cargo[res] -= amount
	g.Ship.Credits += price
	p.Stock[res] += amount
	g.AddLog(fmt.Sprintf("Sold %d units of %s for %d cr.", amount, res.String(), price))
	return true
}

func (g *Game) Wait() {
	g.Day++
	g.Ship.Fuel += 5 // Passive fuel recovery
	if g.Ship.Fuel > g.Ship.MaxFuel { g.Ship.Fuel = g.Ship.MaxFuel }
	g.AddLog("Waited for one day. Systems recharged.")
}
