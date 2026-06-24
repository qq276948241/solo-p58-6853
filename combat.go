package main

import (
	"fmt"
	"math/rand"
)

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
