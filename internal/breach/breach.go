package breach

import (
	"fmt"
	"math/rand"
	"time"
)

type NodeType int

const (
	Standard NodeType = iota
	Firewall
	Database
	Core
)

type Node struct {
	ID          int
	Name        string
	Type        NodeType
	Security    int // 0-100
	MaxSecurity int
	Hacked      bool
	Adjacent    []int // IDs of connected nodes
	X, Y        int   // For visual map
}

type Program int

const (
	Crack Program = iota
	Stealth
	Overclock
)

func (p Program) String() string {
	return [...]string{"Crack.exe", "Stealth.sh", "Overclock.go"}[p]
}

type Game struct {
	Nodes      []*Node
	CurrentNode int
	Trace      float64 // 0-100%
	CPU        int     // Available "power"
	MaxCPU     int
	HackedData int
	Win        bool
	GameOver   bool
	Log        []string
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		CurrentNode: 0,
		Trace:       0,
		CPU:         10,
		MaxCPU:      10,
		Log:         []string{"[CONNECTION ESTABLISHED] Local proxy active."},
	}

	// Generate nodes in a small network
	g.generateNetwork()

	return g
}

func (g *Game) generateNetwork() {
	// Simple grid-like layout for map
	for i := 0; i < 9; i++ {
		name := fmt.Sprintf("NODE-%02X", i)
		ntype := Standard
		if i == 0 { name = "ENTRY-PT" }
		if i == 4 { ntype = Firewall; name = "FWALL-01" }
		if i == 7 { ntype = Database; name = "DB-SERVER" }
		if i == 8 { ntype = Core; name = "MAIN-CORE" }

		maxSec := 20 + rand.Intn(30)
		if ntype == Firewall { maxSec = 60 }
		if ntype == Core { maxSec = 100 }

		g.Nodes = append(g.Nodes, &Node{
			ID:          i,
			Name:        name,
			Type:        ntype,
			Security:    maxSec,
			MaxSecurity: maxSec,
			X:           (i % 3) * 20 + 5,
			Y:           (i / 3) * 6 + 3,
		})
	}

	// Connect nodes (simple grid connections)
	g.Nodes[0].Adjacent = []int{1, 3}
	g.Nodes[1].Adjacent = []int{0, 2, 4}
	g.Nodes[2].Adjacent = []int{1, 5}
	g.Nodes[3].Adjacent = []int{0, 4, 6}
	g.Nodes[4].Adjacent = []int{1, 3, 5, 7}
	g.Nodes[5].Adjacent = []int{2, 4, 8}
	g.Nodes[6].Adjacent = []int{3, 7}
	g.Nodes[7].Adjacent = []int{4, 6, 8}
	g.Nodes[8].Adjacent = []int{5, 7}

	g.Nodes[0].Hacked = true
}

func (g *Game) AddLog(msg string) {
	g.Log = append(g.Log, "> "+msg)
	if len(g.Log) > 6 {
		g.Log = g.Log[1:]
	}
}

func (g *Game) Tick() {
	if g.Win || g.GameOver { return }
	
	// Natural trace increase if we are on a non-hacked node or hacking
	g.Trace += 0.05
	if g.Trace >= 100 {
		g.Trace = 100
		g.GameOver = true
		g.AddLog("[ALARM] TRACE COMPLETE. PHYSICAL LOCATION COMPROMISED.")
	}
}

func (g *Game) RunProgram(p Program) {
	if g.Win || g.GameOver { return }
	
	node := g.Nodes[g.CurrentNode]
	switch p {
	case Crack:
		damage := 5 + rand.Intn(10)
		node.Security -= damage
		g.Trace += 1.5
		g.AddLog(fmt.Sprintf("CRACK.EXE: Deployed. Sec-layer reduced by %d.", damage))
		if node.Security <= 0 {
			node.Security = 0
			node.Hacked = true
			g.AddLog(fmt.Sprintf("[SUCCESS] %s fully compromised.", node.Name))
			if node.Type == Core {
				g.Win = true
				g.AddLog("[ACCESS GRANTED] Core data decrypted.")
			}
		}
	case Stealth:
		reduction := 2 + rand.Intn(5)
		g.Trace -= float64(reduction)
		if g.Trace < 0 { g.Trace = 0 }
		g.AddLog(fmt.Sprintf("STEALTH.SH: Rerouting packets. Trace reduced by %d%%.", reduction))
	case Overclock:
		g.AddLog("OVERCLOCK.GO: Boosting CPU. Security protocols bypass initiated.")
		// Massive damage but massive trace
		node.Security -= 25
		g.Trace += 8.0
		if node.Security <= 0 { node.Security = 0; node.Hacked = true }
	}
}

func (g *Game) Move(targetID int) bool {
	if g.Win || g.GameOver { return false }
	
	// Can only move to adjacent nodes
	current := g.Nodes[g.CurrentNode]
	valid := false
	for _, adj := range current.Adjacent {
		if adj == targetID {
			valid = true
			break
		}
	}

	if valid {
		// Moving to unhacked node is risky
		if !g.Nodes[targetID].Hacked {
			g.Trace += 2.0
		}
		g.CurrentNode = targetID
		g.AddLog(fmt.Sprintf("Relocating to %s...", g.Nodes[targetID].Name))
		return true
	}
	return false
}
