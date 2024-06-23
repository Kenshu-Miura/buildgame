package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

type Robot struct {
	Name      string
	HP        int
	Attack    int
	Defense   int
	Speed     int
	Weapon    string
	Armor     string
	Accessory string
}

func (r *Robot) EquipWeapon(weapon string) {
	r.Weapon = weapon
	switch weapon {
	case "Sword":
		r.Attack += 10
	case "Gun":
		r.Attack += 15
		r.Speed -= 2
	}
}

func (r *Robot) EquipArmor(armor string) {
	r.Armor = armor
	switch armor {
	case "Shield":
		r.Defense += 10
	case "Armor":
		r.Defense += 15
		r.Speed -= 3
	}
}

func (r *Robot) EquipAccessory(accessory string) {
	r.Accessory = accessory
	switch accessory {
	case "Boots":
		r.Speed += 5
	case "Helmet":
		r.Defense += 5
		r.Speed -= 1
	}
}

type Game struct {
	player         Robot
	enemy          Robot
	messages       []string
	equipment      [][]string
	selected       [3]int
	selectionPhase int
	battleStarted  bool
	battleEnded    bool
	lastAttackTime time.Time
}

func (g *Game) Update() error {
	if !g.battleStarted {
		if g.selectionPhase < 3 {
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
				g.selected[g.selectionPhase] = (g.selected[g.selectionPhase] - 1 + len(g.equipment[g.selectionPhase])) % len(g.equipment[g.selectionPhase])
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
				g.selected[g.selectionPhase] = (g.selected[g.selectionPhase] + 1) % len(g.equipment[g.selectionPhase])
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
				switch g.selectionPhase {
				case 0:
					g.player.EquipWeapon(g.equipment[0][g.selected[0]])
				case 1:
					g.player.EquipArmor(g.equipment[1][g.selected[1]])
				case 2:
					g.player.EquipAccessory(g.equipment[2][g.selected[2]])
				}
				g.selectionPhase++
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			g.battleStarted = true
			g.lastAttackTime = time.Now()
			g.messages = append(g.messages, "Battle Start!")
		}
	} else if !g.battleEnded {
		if time.Since(g.lastAttackTime) >= time.Second {
			g.lastAttackTime = time.Now()
			if g.player.HP > 0 && g.enemy.HP > 0 {
				if g.player.Speed >= g.enemy.Speed {
					g.messages = append(g.messages, g.player.AttackEnemy(&g.enemy))
					if g.enemy.HP > 0 {
						g.messages = append(g.messages, g.enemy.AttackEnemy(&g.player))
					}
				} else {
					g.messages = append(g.messages, g.enemy.AttackEnemy(&g.player))
					if g.player.HP > 0 {
						g.messages = append(g.messages, g.player.AttackEnemy(&g.enemy))
					}
				}
			} else {
				if g.player.HP <= 0 {
					g.messages = append(g.messages, "Player's robot is defeated!")
				} else if g.enemy.HP <= 0 {
					g.messages = append(g.messages, "Enemy's robot is defeated!")
				}
				g.battleEnded = true
			}
		}
	}
	return nil
}

func (r *Robot) AttackEnemy(enemy *Robot) string {
	damage := r.Attack - enemy.Defense
	if damage < 0 {
		damage = 0
	}
	enemy.HP -= damage
	return fmt.Sprintf("%s attacks %s for %d damage", r.Name, enemy.Name, damage)
}

func (g *Game) Draw(screen *ebiten.Image) {
	if !g.battleStarted {
		if g.selectionPhase < 3 {
			msg := fmt.Sprintf("Select %s:\n", []string{"Weapon", "Armor", "Accessory"}[g.selectionPhase])
			for i, item := range g.equipment[g.selectionPhase] {
				cursor := " "
				if i == g.selected[g.selectionPhase] {
					cursor = ">"
				}
				msg += fmt.Sprintf("%s %s\n", cursor, item)
			}
			ebitenutil.DebugPrint(screen, msg)
		} else {
			status := fmt.Sprintf(
				"Current Equipment:\nWeapon: %s\nArmor: %s\nAccessory: %s\n\nCurrent Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d\n\nPress Z to start the battle!",
				g.player.Weapon, g.player.Armor, g.player.Accessory,
				g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed,
			)
			ebitenutil.DebugPrint(screen, status)
		}
	} else {
		ebitenutil.DebugPrint(screen, strings.Join(g.messages, "\n"))
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1280, 720
}

func main() {
	rand.Seed(time.Now().UnixNano())
	player := Robot{Name: "PlayerBot", HP: 100, Attack: 20, Defense: 10, Speed: 5}
	enemy := Robot{Name: "EnemyBot", HP: 100, Attack: 18, Defense: 8, Speed: 4}

	enemy.EquipWeapon("Gun")
	enemy.EquipArmor("Armor")
	enemy.EquipAccessory("Helmet")

	game := &Game{
		player: player,
		enemy:  enemy,
		equipment: [][]string{
			{"Sword", "Gun"},
			{"Shield", "Armor"},
			{"Boots", "Helmet"},
		},
		selected:       [3]int{0, 0, 0},
		selectionPhase: 0,
		battleStarted:  false,
		battleEnded:    false,
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Robot Battle")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
