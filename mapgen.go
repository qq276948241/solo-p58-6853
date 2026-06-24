package main

import "math/rand"

func NewMonster(x, y int, floor int) Monster {
	if rand.Intn(2) == 0 {
		return Monster{
			Pos:      Position{X: x, Y: y},
			Rune:     TileGoblin,
			Name:     "哥布林",
			HP:       8 + floor*2,
			MaxHP:    8 + floor*2,
			Attack:   3 + floor,
			ExpValue: 5 + floor*2,
		}
	}
	return Monster{
		Pos:      Position{X: x, Y: y},
		Rune:     TileBat,
		Name:     "蝙蝠",
		HP:       5 + floor,
		MaxHP:    5 + floor,
		Attack:   2 + floor,
		ExpValue: 3 + floor,
	}
}

func (g *Game) GenerateFloor() {
	g.Map = make([][]Tile, MapHeight)
	for y := 0; y < MapHeight; y++ {
		g.Map[y] = make([]Tile, MapWidth)
		for x := 0; x < MapWidth; x++ {
			g.Map[y][x] = Tile{Rune: TileWall, Walkable: false}
		}
	}

	rooms := []struct{ X, Y, W, H int }{}
	numRooms := 5 + rand.Intn(4)

	for i := 0; i < numRooms; i++ {
		w := 4 + rand.Intn(6)
		h := 3 + rand.Intn(4)
		x := 1 + rand.Intn(MapWidth-w-2)
		y := 1 + rand.Intn(MapHeight-h-2)

		overlaps := false
		for _, r := range rooms {
			if x < r.X+r.W+1 && x+w+1 > r.X && y < r.Y+r.H+1 && y+h+1 > r.Y {
				overlaps = true
				break
			}
		}
		if overlaps {
			continue
		}

		for yy := y; yy < y+h; yy++ {
			for xx := x; xx < x+w; xx++ {
				g.Map[yy][xx] = Tile{Rune: TileFloor, Walkable: true}
			}
		}

		if len(rooms) > 0 {
			prev := rooms[len(rooms)-1]
			prevX := prev.X + prev.W/2
			prevY := prev.Y + prev.H/2
			currX := x + w/2
			currY := y + h/2

			if rand.Intn(2) == 0 {
				g.CreateHorizontalTunnel(prevX, currX, prevY)
				g.CreateVerticalTunnel(prevY, currY, currX)
			} else {
				g.CreateVerticalTunnel(prevY, currY, prevX)
				g.CreateHorizontalTunnel(prevX, currX, currY)
			}
		}

		rooms = append(rooms, struct{ X, Y, W, H int }{x, y, w, h})
	}

	if len(rooms) > 0 {
		first := rooms[0]
		g.Player.Pos.X = first.X + first.W/2
		g.Player.Pos.Y = first.Y + first.H/2

		last := rooms[len(rooms)-1]
		stairsX := last.X + last.W/2
		stairsY := last.Y + last.H/2
		g.Map[stairsY][stairsX].HasStairs = true
		g.Map[stairsY][stairsX].Rune = TileStairs
	}

	g.Monsters = nil
	monsterCount := 3 + g.Floor*2
	for i := 0; i < monsterCount; i++ {
		for tries := 0; tries < 100; tries++ {
			mx := rand.Intn(MapWidth)
			my := rand.Intn(MapHeight)
			if g.Map[my][mx].Walkable && !g.IsOccupied(mx, my) &&
				!(mx == g.Player.Pos.X && my == g.Player.Pos.Y) {
				g.Monsters = append(g.Monsters, NewMonster(mx, my, g.Floor))
				break
			}
		}
	}

	potionCount := 2 + rand.Intn(3)
	for i := 0; i < potionCount; i++ {
		for tries := 0; tries < 100; tries++ {
			px := rand.Intn(MapWidth)
			py := rand.Intn(MapHeight)
			if g.Map[py][px].Walkable && !g.Map[py][px].HasStairs &&
				!g.Map[py][px].HasPotion && !g.IsOccupied(px, py) &&
				!(px == g.Player.Pos.X && py == g.Player.Pos.Y) {
				g.Map[py][px].HasPotion = true
				break
			}
		}
	}

	g.Vision = NewVisionMap(MapWidth, MapHeight)
	g.Vision.ComputeFOV(g.Map, g.Player.Pos.X, g.Player.Pos.Y, ViewRadius)
}

func (g *Game) CreateHorizontalTunnel(x1, x2, y int) {
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		if y >= 0 && y < MapHeight && x >= 0 && x < MapWidth {
			g.Map[y][x] = Tile{Rune: TileFloor, Walkable: true}
		}
	}
}

func (g *Game) CreateVerticalTunnel(y1, y2, x int) {
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		if y >= 0 && y < MapHeight && x >= 0 && x < MapWidth {
			g.Map[y][x] = Tile{Rune: TileFloor, Walkable: true}
		}
	}
}

func (g *Game) IsOccupied(x, y int) bool {
	for _, m := range g.Monsters {
		if m.Pos.X == x && m.Pos.Y == y {
			return true
		}
	}
	return false
}

func (g *Game) GetMonsterAt(x, y int) *Monster {
	for i := range g.Monsters {
		if g.Monsters[i].Pos.X == x && g.Monsters[i].Pos.Y == y {
			return &g.Monsters[i]
		}
	}
	return nil
}
