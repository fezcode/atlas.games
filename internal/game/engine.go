package game

import (
	"math/rand"
)

const (
	Lanes      = 3
	GameWidth  = 80
	LaneHeight = 4
	ChickenX   = 8
	MaxSpeed   = 5.5 // Increased max speed for more challenge
)

type ObjectType int

const (
	TypeObstacle ObjectType = iota
	TypeGun
	TypeAmmo
	TypeTree
	TypeCar
	TypeEnemy
	TypePowerUp
	TypePowerDown
	TypeDisableJump
	TypeEnableJump
)

type GameObject struct {
	Type  ObjectType
	Lane  int
	X     float64
	Speed float64
}

type GameState struct {
	Started     bool
	Paused      bool
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
	
	// Power-up state
	InvincibleTimer int
	DoubleScoreTimer int
	JumpDisabled     bool
}

func NewGameState() *GameState {
	return &GameState{
		Started:     false,
		Paused:      false,
		ChickenLane: 1,
		Score:       0,
		GameOver:    false,
		Objects:     []GameObject{},
		Background:  []GameObject{},
		Speed:       0.8, // Start slightly faster
		Ammo:        0,
		JumpDisabled: false,
	}
}

func (s *GameState) Update() {
	if !s.Started || s.GameOver || s.Paused {
		return
	}

	s.FrameCount++

	// Move objects
	for i := range s.Objects {
		s.Objects[i].X -= (s.Speed + s.Objects[i].Speed)
	}

	// Parallax
	for i := range s.Background {
		s.Background[i].X -= s.Speed * 0.3
	}

	// Cleanup
	newObjects := []GameObject{}
	for _, obj := range s.Objects {
		if obj.X > -15 {
			newObjects = append(newObjects, obj)
		} else if obj.Type == TypeObstacle || obj.Type == TypeCar || obj.Type == TypeEnemy {
			scoreGain := 10
			if s.DoubleScoreTimer > 0 {
				scoreGain *= 2
			}
			s.Score += scoreGain
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

	// MORE AGGRESSIVE SPAWNING
	spawnInterval := 15 - int(s.Speed * 2.8)
	if spawnInterval < 5 { spawnInterval = 5 }

	if s.FrameCount % spawnInterval == 0 {
		// Occasionally spawn in two lanes at once to force movement
		numToSpawn := 1
		if s.Speed > 2.0 && rand.Float64() < 0.25 { numToSpawn = 2 }
		if s.Speed > 3.5 && rand.Float64() < 0.45 { numToSpawn = 2 }

		spawnedLanes := make(map[int]bool)

		for i := 0; i < numToSpawn; i++ {
			lane := rand.Intn(Lanes)
			if spawnedLanes[lane] { continue }
			
			// Tighter buffer for denser action
			buffer := 35.0 - (s.Speed * 4.0)
			if buffer < 12 { buffer = 12 }

			laneClear := true
			for _, obj := range s.Objects {
				if obj.Lane == lane && obj.X > float64(GameWidth) - buffer {
					laneClear = false
					break
				}
			}

			if laneClear {
				spawnedLanes[lane] = true
				roll := rand.Float64()
				
				// Higher probability for enemies/cars
				carThreshold := 0.35 - (s.Speed * 0.04)
				if carThreshold < 0.2 { carThreshold = 0.2 }
				
				enemyThreshold := 0.70 + (s.Speed * 0.04)
				if enemyThreshold > 0.88 { enemyThreshold = 0.88 }

				if roll < carThreshold {
					s.Objects = append(s.Objects, GameObject{Type: TypeObstacle, Lane: lane, X: float64(GameWidth), Speed: 0})
				} else if roll < 0.60 {
					s.Objects = append(s.Objects, GameObject{Type: TypeCar, Lane: lane, X: float64(GameWidth), Speed: 0.6 + (s.Speed * 0.12)})
				} else if roll < enemyThreshold {
					s.Objects = append(s.Objects, GameObject{Type: TypeEnemy, Lane: lane, X: float64(GameWidth), Speed: rand.Float64() * (0.4 + s.Speed*0.12)})
				} else {
					// Items
					itemRoll := rand.Float64()
					if itemRoll < 0.15 {
						if !s.HasGun {
							s.Objects = append(s.Objects, GameObject{Type: TypeGun, Lane: lane, X: float64(GameWidth), Speed: 0})
						}
					} else if itemRoll < 0.35 {
						s.Objects = append(s.Objects, GameObject{Type: TypeAmmo, Lane: lane, X: float64(GameWidth), Speed: 0})
					} else if itemRoll < 0.50 {
						s.Objects = append(s.Objects, GameObject{Type: TypePowerUp, Lane: lane, X: float64(GameWidth), Speed: 0})
					} else if itemRoll < 0.65 {
						s.Objects = append(s.Objects, GameObject{Type: TypePowerDown, Lane: lane, X: float64(GameWidth), Speed: 0})
					} else if itemRoll < 0.80 {
						s.Objects = append(s.Objects, GameObject{Type: TypeDisableJump, Lane: lane, X: float64(GameWidth), Speed: 0})
					} else {
						s.Objects = append(s.Objects, GameObject{Type: TypeEnableJump, Lane: lane, X: float64(GameWidth), Speed: 0})
					}
				}
			}
		}
	}

	// Trees
	if s.FrameCount%40 == 0 && rand.Float64() < 0.3 {
		s.Background = append(s.Background, GameObject{Type: TypeTree, X: float64(GameWidth)})
	}

	// Collision
	for i := 0; i < len(s.Objects); i++ {
		obj := s.Objects[i]
		if obj.Lane == s.ChickenLane {
			hitboxWidth := 3.0
			if obj.Type == TypeCar {
				hitboxWidth = 6.0
			}
			
			if obj.X >= ChickenX-1 && obj.X <= ChickenX+hitboxWidth {
				switch obj.Type {
				case TypeObstacle, TypeCar, TypeEnemy:
					if !s.Jumping && s.InvincibleTimer <= 0 {
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
				case TypePowerUp:
					s.InvincibleTimer = 125
					s.DoubleScoreTimer = 125
					s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
					i--
				case TypePowerDown:
					s.Score -= 100
					if s.Score < 0 { s.Score = 0 }
					s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
					i--
				case TypeDisableJump:
					s.JumpDisabled = true
					s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
					i--
				case TypeEnableJump:
					s.JumpDisabled = false
					s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
					i--
				}
			}
		}
	}

	// Timers
	if s.ShootTimer > 0 { s.ShootTimer-- ; if s.ShootTimer == 0 { s.IsShooting = false } }
	if s.JumpTimer > 0 { s.JumpTimer-- ; if s.JumpTimer == 0 { s.Jumping = false } }
	if s.InvincibleTimer > 0 { s.InvincibleTimer-- }
	if s.DoubleScoreTimer > 0 { s.DoubleScoreTimer-- }

	// Speed
	if s.FrameCount%40 == 0 && s.Speed < MaxSpeed {
		s.Speed += 0.035 // Faster speed increment
	}
}

func (s *GameState) MoveUp() {
	if s.ChickenLane > 0 { s.ChickenLane-- }
}

func (s *GameState) MoveDown() {
	if s.ChickenLane < Lanes-1 { s.ChickenLane++ }
}

func (s *GameState) Jump() {
	if !s.Jumping && !s.JumpDisabled {
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
			scoreGain := 25
			if s.DoubleScoreTimer > 0 {
				scoreGain *= 2
			}
			s.Score += scoreGain
		}
	}
}

func (s *GameState) TogglePause() {
	if s.Started && !s.GameOver {
		s.Paused = !s.Paused
	}
}
