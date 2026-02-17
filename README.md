# Wilson's Revenge (atlas.games)

![banner.png](./banner.png)

Wilson the Chicken is back, and this time it's personal. A high-speed, horizontal terminal runner inspired by classic arcade games.

## Features
- **Horizontal Scrolling**: A dynamic 3-lane world that moves from right to left.
- **Variable Enemies**: 
  - **Barricades**: Standard obstacles that block your path.
  - **Cars**: High-speed vehicles that can catch you off guard.
  - **Enemies**: Humanoids on the run.
- **Arsenal**: Wilson can pick up guns and ammo to clear obstacles in his lane.
- **Parallax Background**: Beautifully rendered trees and a sun that move at a slower speed for a sense of depth.
- **Progressive Difficulty**: The game speed increases every 40 frames, capping at a challenging **4.5**.
- **Polish**: Distinct ASCII art for all entities and states, including Wilson's blue-winged jump!

## How to Play

### Controls
- **W / X** or **K / J**: Move Up/Down across lanes
- **D** or **L**: Jump (turn blue and avoid collisions)
- **Space**: Start Game / Shoot (if gun and ammo are equipped)
- **R**: Restart after "lo siento, Wilson"
- **Q / Ctrl+C**: Quit

### Objective
Dodge or destroy obstacles to score points. Every obstacle passed or destroyed increases your score. Watch out—the faster you go, the less time you have to react!

## Development

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

### Running locally
```bash
go run main.go
```

### Building
```bash
# Use the bake system
go run Recipe.go run-win
```

## License
MIT
