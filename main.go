package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	MapWidth     = 60
	MapHeight    = 25
	StatusWidth  = 20
	TotalFloors  = 5
	TileWall     = '#'
	TileFloor    = '.'
	TileStairs   = '>'
	TilePotion   = '+'
	TilePlayer   = '@'
	TileGoblin   = 'G'
	TileBat      = 'B'
)

type Position struct {
	X, Y int
}

type Player struct {
	Pos      Position
	HP       int
	MaxHP    int
	Attack   int
	Level    int
	Exp      int
	ExpToNext int
	Score    int
}

type Monster struct {
	Pos      Position
	Rune     rune
	Name     string
	HP       int
	MaxHP    int
	Attack   int
	ExpValue int
}

type Tile struct {
	Rune     rune
	Walkable bool
	HasPotion bool
	HasStairs bool
}

type Game struct {
	Screen    tcell.Screen
	Player    Player
	Map       [][]Tile
	Vision    *VisionMap
	Monsters  []Monster
	Floor     int
	GameOver  bool
	Victory   bool
	Message   string
}

func NewPlayer(x, y int) Player {
	return Player{
		Pos:       Position{X: x, Y: y},
		HP:        30,
		MaxHP:     30,
		Attack:    5,
		Level:     1,
		Exp:       0,
		ExpToNext: 10,
		Score:     0,
	}
}

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

func (g *Game) Render() {
	g.Screen.Clear()

	for y := 0; y < MapHeight; y++ {
		for x := 0; x < MapWidth; x++ {
			visible := g.Vision.IsVisible(x, y)
			explored := g.Vision.IsExplored(x, y)

			if !visible && !explored {
				continue
			}

			tile := g.Map[y][x]
			r := tile.Rune

			if visible && tile.HasPotion {
				r = TilePotion
			}

			style := tcell.StyleDefault
			switch r {
			case TileWall:
				if visible {
					style = style.Foreground(tcell.ColorGray)
				} else {
					style = style.Foreground(tcell.ColorDarkGray)
				}
			case TileFloor:
				if visible {
					style = style.Foreground(tcell.ColorDimGray)
				} else {
					r = '.'
					style = style.Foreground(tcell.ColorDarkGray)
				}
			case TileStairs:
				if visible {
					style = style.Foreground(tcell.ColorYellow)
				} else {
					style = style.Foreground(tcell.ColorOlive)
				}
			case TilePotion:
				if visible {
					style = style.Foreground(tcell.ColorGreen)
				}
			}

			if !visible && !tile.HasStairs && r != TileWall {
				r = '.'
			}

			g.Screen.SetContent(x, y, r, nil, style)
		}
	}

	for _, m := range g.Monsters {
		if !g.Vision.IsVisible(m.Pos.X, m.Pos.Y) {
			continue
		}
		style := tcell.StyleDefault
		switch m.Rune {
		case TileGoblin:
			style = style.Foreground(tcell.ColorRed)
		case TileBat:
			style = style.Foreground(tcell.ColorPurple)
		}
		g.Screen.SetContent(m.Pos.X, m.Pos.Y, m.Rune, nil, style)
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	g.Screen.SetContent(g.Player.Pos.X, g.Player.Pos.Y, TilePlayer, nil, style)

	g.RenderStatus()

	if g.GameOver || g.Victory {
		g.RenderGameOver()
	}

	g.Screen.Show()
}

func (g *Game) RenderStatus() {
	style := tcell.StyleDefault
	statusX := MapWidth + 2

	hpPercent := float64(g.Player.HP) / float64(g.Player.MaxHP)
	barWidth := 16
	barFilled := int(hpPercent * float64(barWidth))
	hpBar := "["
	for i := 0; i < barWidth; i++ {
		if i < barFilled {
			hpBar += "="
		} else {
			hpBar += " "
		}
	}
	hpBar += "]"

	lines := []string{
		"==== 状态 ====",
		"",
		fmt.Sprintf("楼层: %d / %d", g.Floor, TotalFloors),
		"",
		fmt.Sprintf("等级: Lv.%d", g.Player.Level),
		hpBar,
		fmt.Sprintf("HP: %d / %d", g.Player.HP, g.Player.MaxHP),
		fmt.Sprintf("攻击力: %d", g.Player.Attack),
		"",
		fmt.Sprintf("经验: %d / %d", g.Player.Exp, g.Player.ExpToNext),
		"",
		fmt.Sprintf("得分: %d", g.Player.Score),
		"",
		"=== 控制 ===",
		"W/↑ - 上",
		"S/↓ - 下",
		"A/← - 左",
		"D/→ - 右",
		"Q - 退出",
	}

	for i, line := range lines {
		for j, c := range line {
			g.Screen.SetContent(statusX+j, i, c, nil, style)
		}
	}

	for i, c := range g.Message {
		g.Screen.SetContent(statusX+i, MapHeight-2, c, nil, style.Foreground(tcell.ColorYellow))
	}
}

func (g *Game) RenderGameOver() {
	screenWidth, screenHeight := g.Screen.Size()
	var title, subtitle string
	if g.Victory {
		title = "  恭喜通关！  "
		subtitle = fmt.Sprintf(" 最终得分: %d ", g.Player.Score)
	} else {
		title = "  游戏结束  "
		subtitle = fmt.Sprintf(" 最终得分: %d ", g.Player.Score)
	}

	prompt := " 按 Q 退出 "

	titleX := (screenWidth - len(title)) / 2
	subtitleX := (screenWidth - len(subtitle)) / 2
	promptX := (screenWidth - len(prompt)) / 2
	centerY := screenHeight / 2

	boxStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	textStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow)
	scoreStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)

	boxWidth := len(title) + 4
	boxHeight := 7
	boxX := titleX - 2
	boxY := centerY - 3

	for y := 0; y < boxHeight; y++ {
		for x := 0; x < boxWidth; x++ {
			r := ' '
			if y == 0 || y == boxHeight-1 {
				r = '#'
			} else if x == 0 || x == boxWidth-1 {
				r = '#'
			}
			g.Screen.SetContent(boxX+x, boxY+y, r, nil, boxStyle)
		}
	}

	for i, c := range title {
		g.Screen.SetContent(titleX+i, centerY-2, c, nil, textStyle)
	}
	for i, c := range subtitle {
		g.Screen.SetContent(subtitleX+i, centerY, c, nil, scoreStyle)
	}
	for i, c := range prompt {
		g.Screen.SetContent(promptX+i, centerY+2, c, nil, boxStyle)
	}
}

func (g *Game) MovePlayer(dx, dy int) {
	if g.GameOver || g.Victory {
		return
	}

	newX := g.Player.Pos.X + dx
	newY := g.Player.Pos.Y + dy

	if newX < 0 || newX >= MapWidth || newY < 0 || newY >= MapHeight {
		return
	}

	if monster := g.GetMonsterAt(newX, newY); monster != nil {
		g.Combat(monster)
		g.MonsterTurn()
		g.Vision.ComputeFOV(g.Map, g.Player.Pos.X, g.Player.Pos.Y, ViewRadius)
		return
	}

	if !g.Map[newY][newX].Walkable {
		return
	}

	g.Player.Pos.X = newX
	g.Player.Pos.Y = newY
	g.Vision.ComputeFOV(g.Map, g.Player.Pos.X, g.Player.Pos.Y, ViewRadius)

	if g.Map[newY][newX].HasPotion {
		healAmount := 10 + g.Floor*2
		g.Player.HP = min(g.Player.HP+healAmount, g.Player.MaxHP)
		g.Map[newY][newX].HasPotion = false
		g.Message = fmt.Sprintf("捡到药水！恢复 %d HP", healAmount)
		g.Player.Score += 5
	}

	if g.Map[newY][newX].HasStairs {
		if g.Floor >= TotalFloors {
			g.Victory = true
			g.Player.Score += 500
			return
		}
		g.Floor++
		g.Player.Score += 100
		g.Message = fmt.Sprintf("进入第 %d 层！", g.Floor)
		g.GenerateFloor()
		return
	}

	g.MonsterTurn()
	g.Message = ""
}

func (g *Game) Combat(monster *Monster) {
	playerDmg := g.Player.Attack + rand.Intn(3) - 1
	monster.HP -= playerDmg
	g.Message = fmt.Sprintf("你攻击%s造成 %d 伤害", monster.Name, playerDmg)

	if monster.HP <= 0 {
		g.Message += fmt.Sprintf(" | 击败%s！获得 %d 经验", monster.Name, monster.ExpValue)
		g.Player.Exp += monster.ExpValue
		g.Player.Score += monster.ExpValue * 2
		g.CheckLevelUp()
		for i := range g.Monsters {
			if &g.Monsters[i] == monster {
				g.Monsters = append(g.Monsters[:i], g.Monsters[i+1:]...)
				break
			}
		}
	}
}

func (g *Game) CheckLevelUp() {
	for g.Player.Exp >= g.Player.ExpToNext {
		g.Player.Exp -= g.Player.ExpToNext
		g.Player.Level++
		g.Player.MaxHP += 5
		g.Player.HP = g.Player.MaxHP
		g.Player.Attack += 2
		g.Player.ExpToNext = int(float64(g.Player.ExpToNext) * 1.5)
		g.Message += fmt.Sprintf(" | 升级！Lv.%d", g.Player.Level)
		g.Player.Score += 50
	}
}

func (g *Game) MonsterTurn() {
	for i := range g.Monsters {
		m := &g.Monsters[i]
		dx := 0
		dy := 0

		if m.Pos.X < g.Player.Pos.X {
			dx = 1
		} else if m.Pos.X > g.Player.Pos.X {
			dx = -1
		}
		if m.Pos.Y < g.Player.Pos.Y {
			dy = 1
		} else if m.Pos.Y > g.Player.Pos.Y {
			dy = -1
		}

		if rand.Intn(2) == 0 {
			if dx != 0 {
				newX := m.Pos.X + dx
				if g.Map[m.Pos.Y][newX].Walkable && !g.IsOccupied(newX, m.Pos.Y) {
					if newX == g.Player.Pos.X && m.Pos.Y == g.Player.Pos.Y {
						dmg := m.Attack + rand.Intn(2)
						g.Player.HP -= dmg
						g.Message += fmt.Sprintf(" | %s攻击你造成 %d 伤害", m.Name, dmg)
					} else {
						m.Pos.X = newX
					}
				}
			} else if dy != 0 {
				newY := m.Pos.Y + dy
				if g.Map[newY][m.Pos.X].Walkable && !g.IsOccupied(m.Pos.X, newY) {
					if m.Pos.X == g.Player.Pos.X && newY == g.Player.Pos.Y {
						dmg := m.Attack + rand.Intn(2)
						g.Player.HP -= dmg
						g.Message += fmt.Sprintf(" | %s攻击你造成 %d 伤害", m.Name, dmg)
					} else {
						m.Pos.Y = newY
					}
				}
			}
		} else {
			if dy != 0 {
				newY := m.Pos.Y + dy
				if g.Map[newY][m.Pos.X].Walkable && !g.IsOccupied(m.Pos.X, newY) {
					if m.Pos.X == g.Player.Pos.X && newY == g.Player.Pos.Y {
						dmg := m.Attack + rand.Intn(2)
						g.Player.HP -= dmg
						g.Message += fmt.Sprintf(" | %s攻击你造成 %d 伤害", m.Name, dmg)
					} else {
						m.Pos.Y = newY
					}
				}
			} else if dx != 0 {
				newX := m.Pos.X + dx
				if g.Map[m.Pos.Y][newX].Walkable && !g.IsOccupied(newX, m.Pos.Y) {
					if newX == g.Player.Pos.X && m.Pos.Y == g.Player.Pos.Y {
						dmg := m.Attack + rand.Intn(2)
						g.Player.HP -= dmg
						g.Message += fmt.Sprintf(" | %s攻击你造成 %d 伤害", m.Name, dmg)
					} else {
						m.Pos.X = newX
					}
				}
			}
		}
	}

	if g.Player.HP <= 0 {
		g.Player.HP = 0
		g.GameOver = true
	}
}

func (g *Game) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyCtrlC || ev.Rune() == 'q' || ev.Rune() == 'Q' {
			return false
		}
		if g.GameOver || g.Victory {
			return true
		}
		switch ev.Rune() {
		case 'w', 'W':
			g.MovePlayer(0, -1)
		case 's', 'S':
			g.MovePlayer(0, 1)
		case 'a', 'A':
			g.MovePlayer(-1, 0)
		case 'd', 'D':
			g.MovePlayer(1, 0)
		}
		switch ev.Key() {
		case tcell.KeyUp:
			g.MovePlayer(0, -1)
		case tcell.KeyDown:
			g.MovePlayer(0, 1)
		case tcell.KeyLeft:
			g.MovePlayer(-1, 0)
		case tcell.KeyRight:
			g.MovePlayer(1, 0)
		}
	}
	return true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	rand.Seed(time.Now().UnixNano())

	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer screen.Fini()

	game := &Game{
		Screen: screen,
		Player: NewPlayer(0, 0),
		Floor:  1,
	}
	game.GenerateFloor()

	for {
		game.Render()
		ev := screen.PollEvent()
		if !game.HandleEvent(ev) {
			break
		}
	}
}
