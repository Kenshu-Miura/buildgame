package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	maxMessages  = 5 // 表示するメッセージの最大数
)

var (
	fontFace *text.GoTextFace
)

func init() {
	// ファイルオープンなどで io.Reader を得る
	f, err := os.Open("KiwiMaru-Regular.ttf")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// フォントを読み込む
	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		panic(err)
	}

	// フォントフェイスを作る
	fontFace = &text.GoTextFace{Source: src, Size: 24}
}

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
	rng            *rand.Rand
	turn           int
}

func NewGame() *Game {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	player := Robot{Name: "PlayerBot", HP: 100, Attack: 20, Defense: 10, Speed: 5}
	enemy := Robot{Name: "EnemyBot", HP: 100, Attack: 18, Defense: 8, Speed: 4}

	enemy.EquipWeapon("Gun")
	enemy.EquipArmor("Armor")
	enemy.EquipAccessory("Helmet")

	return &Game{
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
		rng:            rng,
		turn:           1,
	}
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
					g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.AttackEnemy(&g.enemy)))
					if len(g.messages) > maxMessages {
						g.messages = g.messages[1:] // 古いメッセージを削除
					}
					if g.enemy.HP > 0 {
						g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.enemy.AttackEnemy(&g.player)))
						if len(g.messages) > maxMessages {
							g.messages = g.messages[1:] // 古いメッセージを削除
						}
					}
				} else {
					g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.enemy.AttackEnemy(&g.player)))
					if len(g.messages) > maxMessages {
						g.messages = g.messages[1:] // 古いメッセージを削除
					}
					if g.player.HP > 0 {
						g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.AttackEnemy(&g.enemy)))
						if len(g.messages) > maxMessages {
							g.messages = g.messages[1:] // 古いメッセージを削除
						}
					}
				}
				g.turn++
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
	op := &text.DrawOptions{}
	op.GeoM.Translate(20, 40)
	op.LineSpacing = 24 * 1.5

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
			text.Draw(screen, msg, fontFace, op)
		} else {
			status := fmt.Sprintf(
				"Current Equipment:\nWeapon: %s\nArmor: %s\nAccessory: %s\n\nCurrent Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d\n\nPress Z to start the battle!",
				g.player.Weapon, g.player.Armor, g.player.Accessory,
				g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed,
			)
			text.Draw(screen, status, fontFace, op)
		}
	} else {
		text.Draw(screen, strings.Join(g.messages, "\n"), fontFace, op)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1280, 720
}

func main() {
	game := NewGame()
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Robot Battle")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
