package game

import (
	"math/rand"
)

const (
	Lanes      = 3
	GameWidth  = 80
	LaneHeight = 4
	ChickenX   = 8
	MaxSpeed   = 3.5 // Challenging but fair for terminal refresh
)

type ObjectType int

const (
	TypeObstacle ObjectType = iota
	TypeGun
	TypeAmmo
	TypeTree
	TypeCar
	TypeEnemy
)

type GameObject struct {
	Type  ObjectType
	Lane  int
	X     float64
	Speed float64 // Individual speed modifier
}

type GameState struct {
	Started     bool
	ChickenLane int
	Jumping     bool
	JumpTimer   int
	Score       int
	GameOver    bool
	Objects     []GameObject
	Background  []GameObject
	Speed       float64
	FrameCount  int
	HasGun      bool
	Ammo        int
	IsShooting  bool
	ShootTimer  int
}

func NewGameState() *GameState {
	return &GameState{
		Started:     false,
		ChickenLane: 1,
		Score:       0,
		GameOver:    false,
		Objects:     []GameObject{},
		Background:  []GameObject{},
		Speed:       0.6,
		Ammo:        0,
	}
}

func (s *GameState) Update() {
	if !s.Started || s.GameOver {
		return
	}

	s.FrameCount++

	// Move objects
	for i := range s.Objects {
		// Individual speed + global speed
		s.Objects[i].X -= (s.Speed + s.Objects[i].Speed)
	}

	// Parallax: Background moves slower
	for i := range s.Background {
		s.Background[i].X -= s.Speed * 0.3
	}

	// Cleanup
	newObjects := []GameObject{}
	for _, obj := range s.Objects {
		if obj.X > -15 {
			newObjects = append(newObjects, obj)
		} else if obj.Type == TypeObstacle || obj.Type == TypeCar || obj.Type == TypeEnemy {
			s.Score += 10
		}
	}
	s.Objects = newObjects

	newBG := []GameObject{}
	for _, bg := range s.Background {
		if bg.X > -15 {
			newBG = append(newBG, bg)
		}
	}
	s.Background = newBG

	// Spawning Logic
	if s.FrameCount%15 == 0 {
		lane := rand.Intn(Lanes)
		
		// Check if the lane is clear enough to spawn a new object
		laneClear := true
		for _, obj := range s.Objects {
			if obj.Lane == lane && obj.X > float64(GameWidth)-40 {
				laneClear = false
				break
			}
		}

		if laneClear {
			roll := rand.Float64()
			if roll < 0.4 {
				// Regular Obstacle
				s.Objects = append(s.Objects, GameObject{Type: TypeObstacle, Lane: lane, X: float64(GameWidth), Speed: 0})
			} else if roll < 0.6 {
				// Car (Faster than global speed)
				s.Objects = append(s.Objects, GameObject{Type: TypeCar, Lane: lane, X: float64(GameWidth), Speed: 0.5})
			} else if roll < 0.75 {
				// Enemy (Slightly variable speed)
				s.Objects = append(s.Objects, GameObject{Type: TypeEnemy, Lane: lane, X: float64(GameWidth), Speed: rand.Float64() * 0.3})
			} else if roll < 0.82 {
				// Gun
				if !s.HasGun {
					s.Objects = append(s.Objects, GameObject{Type: TypeGun, Lane: lane, X: float64(GameWidth), Speed: 0})
				}
			} else if roll < 0.9 {
				// Ammo
				s.Objects = append(s.Objects, GameObject{Type: TypeAmmo, Lane: lane, X: float64(GameWidth), Speed: 0})
			}
		}
	}

	// Spawn Trees
	if s.FrameCount%40 == 0 && rand.Float64() < 0.3 {
		s.Background = append(s.Background, GameObject{Type: TypeTree, X: float64(GameWidth)})
	}

	// Collision
	for i := 0; i < len(s.Objects); i++ {
		obj := s.Objects[i]
		if obj.Lane == s.ChickenLane {
			// Adjust hitbox based on type (Cars are longer)
			hitboxWidth := 3.0
			if obj.Type == TypeCar {
				hitboxWidth = 6.0
			}
			
			if obj.X >= ChickenX-1 && obj.X <= ChickenX+hitboxWidth {
				switch obj.Type {
				case TypeObstacle, TypeCar, TypeEnemy:
					if !s.Jumping {
						s.GameOver = true
					}
				case TypeGun:
					s.HasGun = true
					s.Ammo += 5
					s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
					i--
				case TypeAmmo:
					s.Ammo += 3
					s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
					i--
				}
			}
		}
	}

	if s.IsShooting {
		s.ShootTimer--
		if s.ShootTimer <= 0 {
			s.IsShooting = false
		}
	}

	if s.Jumping {
		s.JumpTimer--
		if s.JumpTimer <= 0 {
			s.Jumping = false
		}
	}

	// Incremental speed increase - Now faster and higher cap
	if s.FrameCount%40 == 0 && s.Speed < 4.5 {
		s.Speed += 0.025
	}
}

func (s *GameState) MoveUp() {
	if s.ChickenLane > 0 {
		s.ChickenLane--
	}
}

func (s *GameState) MoveDown() {
	if s.ChickenLane < Lanes-1 {
		s.ChickenLane++
	}
}

func (s *GameState) Jump() {
	if !s.Jumping {
		s.Jumping = true
		s.JumpTimer = 15
	}
}

func (s *GameState) Shoot() {
	if s.HasGun && s.Ammo > 0 && !s.IsShooting {
		s.IsShooting = true
		s.ShootTimer = 5
		s.Ammo--

		nearestIdx := -1
		minX := float64(GameWidth)
		for i, obj := range s.Objects {
			if obj.Type == TypeObstacle || obj.Type == TypeCar || obj.Type == TypeEnemy {
				if obj.Lane == s.ChickenLane && obj.X > ChickenX && obj.X < minX {
					minX = obj.X
					nearestIdx = i
				}
			}
		}

		if nearestIdx != -1 {
			s.Objects = append(s.Objects[:nearestIdx], s.Objects[nearestIdx+1:]...)
			s.Score += 25
		}
	}
}
