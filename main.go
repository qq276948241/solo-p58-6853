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
	Pos       Position
	HP        int
	MaxHP     int
	Attack    int
	Level     int
	Exp       int
	ExpToNext int
	Score     int
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
	Rune      rune
	Walkable  bool
	HasPotion bool
	HasStairs bool
}

type Game struct {
	Screen   tcell.Screen
	Player   Player
	Map      [][]Tile
	Vision   *VisionMap
	Monsters []Monster
	Floor    int
	GameOver bool
	Victory  bool
	Message  string
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
