package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	maxMessages  = 5  // 表示するメッセージの最大数
	lineHeight   = 26 // 行の高さを適切に設定（フォントサイズに応じて調整）
)

var (
	fontFace *text.GoTextFace
)

func init() {
	f, err := os.Open("KiwiMaru-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	src, err := text.NewGoTextFaceSource(f)
	if err != nil {
		log.Fatal(err)
	}

	fontFace = &text.GoTextFace{Source: src, Size: 24}
}

type Robot struct {
	Name         string
	HP           int
	Attack       int
	Defense      int
	Speed        int
	CriticalRate float64 // クリティカル率 (0.0 - 1.0)
	EvasionRate  float64 // 回避率 (0.0 - 1.0)
	Weapon       string
	Armor        string
	Accessory    string
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
	reslt          bool
}

func NewGame() *Game {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	player := Robot{Name: "PlayerBot", HP: 100, Attack: 20, Defense: 10, Speed: 5, CriticalRate: 0.1, EvasionRate: 0.1}
	enemy := Robot{Name: "EnemyBot", HP: 100, Attack: 18, Defense: 8, Speed: 4, CriticalRate: 0.05, EvasionRate: 0.05}

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

func (g *Game) handleSelectionPhase() {
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
}

func (g *Game) handleBattleStart() {
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.battleStarted = true
		g.lastAttackTime = time.Now()
	}
}

func (g *Game) handleBattlePhase() {
	if time.Since(g.lastAttackTime) >= time.Second {
		g.lastAttackTime = time.Now()
		if g.player.HP > 0 && g.enemy.HP > 0 {
			if g.player.Speed >= g.enemy.Speed {
				g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.AttackEnemy(&g.enemy, g.rng)))
				if len(g.messages) > maxMessages {
					g.messages = g.messages[1:]
				}
				if g.enemy.HP > 0 {
					g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.enemy.AttackEnemy(&g.player, g.rng)))
					if len(g.messages) > maxMessages {
						g.messages = g.messages[1:]
					}
				}
			} else {
				g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.enemy.AttackEnemy(&g.player, g.rng)))
				if len(g.messages) > maxMessages {
					g.messages = g.messages[1:]
				}
				if g.player.HP > 0 {
					g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.AttackEnemy(&g.enemy, g.rng)))
					if len(g.messages) > maxMessages {
						g.messages = g.messages[1:]
					}
				}
			}
			g.turn++
		} else {
			if g.player.HP <= 0 {
				g.reslt = false
			} else if g.enemy.HP <= 0 {
				g.reslt = true
			}
			g.battleEnded = true
		}
	}
}

func (g *Game) Update() error {
	if !g.battleStarted {
		if g.selectionPhase < 3 {
			g.handleSelectionPhase()
		} else {
			g.handleBattleStart()
		}
	} else if !g.battleEnded {
		g.handleBattlePhase()
	}
	return nil
}

func (r *Robot) AttackEnemy(enemy *Robot, rng *rand.Rand) string {
	// 回避判定
	if rng.Float64() < enemy.EvasionRate {
		return fmt.Sprintf("%s attacks %s but misses!", r.Name, enemy.Name)
	}

	// クリティカルヒット判定
	critical := 1.0
	if rng.Float64() < r.CriticalRate {
		critical = 2.0
	}

	// ランダムなダメージ修正 (-3 から +3)
	randomDamage := rng.Intn(7) - 3

	// ダメージ計算
	damage := int(float64(r.Attack-enemy.Defense+randomDamage) * critical)
	if damage < 0 {
		damage = 0
	}
	enemy.HP -= damage

	// クリティカルヒットかどうかのメッセージ
	criticalMsg := ""
	if critical > 1.0 {
		criticalMsg = " It's a critical hit!"
	}

	return fmt.Sprintf("%s attacks %s for %d damage.%s", r.Name, enemy.Name, damage, criticalMsg)
}

func drawText(msgWindow *ebiten.Image, msg string, x, y float64) {
	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		textOp := &text.DrawOptions{}
		textOp.GeoM.Translate(x, y+float64(i*lineHeight)) // 各行の表示位置を調整
		text.Draw(msgWindow, line, fontFace, textOp)
	}
}

func drawCenterText(screen *ebiten.Image, msg string, fontFace *text.GoTextFace, windowWidth, windowHeight int) {
	centerMsgWidth, _ := text.Measure(msg, fontFace, 1.0)
	textOp := &text.DrawOptions{}
	textOp.GeoM.Translate(float64((windowWidth-int(centerMsgWidth))/2), (float64(windowHeight) / 1.25)) // 中央のメッセージの表示位置
	text.Draw(screen, msg, fontFace, textOp)
}

func drawBattleStatus(g *Game, msgWindow *ebiten.Image, screen *ebiten.Image, windowWidth, windowHeight int, centerMsg string) {
	leftColumn := fmt.Sprintf(
		"Current Equipment:\nWeapon: %s\nArmor: %s\nAccessory: %s",
		g.player.Weapon, g.player.Armor, g.player.Accessory,
	)
	rightColumn := fmt.Sprintf(
		"Current Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d",
		g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed,
	)

	// 左カラムの表示
	drawText(msgWindow, leftColumn, 10, 10)

	// 右カラムの表示
	drawText(msgWindow, rightColumn, float64(windowWidth/2)+10, 10)

	// 中央メッセージの表示
	drawCenterText(screen, centerMsg, fontFace, windowWidth, windowHeight)
}

func drawMessages(msgWindow *ebiten.Image, msg string) {
	lines := strings.Split(msg, "\n")
	for i, line := range lines {
		textOp := &text.DrawOptions{}
		textOp.GeoM.Translate(10, 10+float64(i*lineHeight)) // 各行の表示位置を調整
		text.Draw(msgWindow, line, fontFace, textOp)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	windowX, windowY := 10, screenHeight/2+70
	windowWidth, windowHeight := screenWidth-20, screenHeight/2-80

	msgWindow := ebiten.NewImage(windowWidth, windowHeight-10)
	msgWindow.Fill(color.Black)
	vector.DrawFilledRect(msgWindow, 0, 0, float32(windowWidth), 2, color.White, false)
	vector.DrawFilledRect(msgWindow, 0, float32(windowHeight-12), float32(windowWidth), 2, color.White, false)
	vector.DrawFilledRect(msgWindow, 0, 0, 2, float32(windowHeight-10), color.White, false)
	vector.DrawFilledRect(msgWindow, float32(windowWidth-2), 0, 2, float32(windowHeight-10), color.White, false)

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
			drawMessages(msgWindow, msg) // 左カラムとして表示
		} else {
			drawBattleStatus(g, msgWindow, screen, windowWidth, windowHeight, "Press Z to start the battle!")
		}
	} else {
		if g.battleEnded {
			centerMsg := "You win!"
			if !g.reslt {
				centerMsg = "You lose!"
			}
			drawCenterText(screen, centerMsg, fontFace, windowWidth, windowHeight)
		} else {
			if g.turn == 1 {
				drawCenterText(screen, "Battle Start!", fontFace, windowWidth, windowHeight)
			}
			msg := strings.Join(g.messages, "\n")
			drawMessages(msgWindow, msg) // 左カラムとして表示
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(windowX), float64(windowY))
	screen.DrawImage(msgWindow, op)
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
