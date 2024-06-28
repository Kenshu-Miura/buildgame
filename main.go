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
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	maxMessages  = 12 // 表示するメッセージの最大数
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
	HitRate      float64 // 命中率 (0.0 - 1.0)
	Weapon       string
	Armor        string
	Accessory    string
}

func (r *Robot) EquipWeapon(weapon string) {
	r.Weapon = weapon
	switch weapon {
	case "Sword":
		r.Attack += 10
		r.HitRate += 0.1
	case "Gun":
		r.Attack += 15
		r.Speed -= 2
		r.HitRate += 0.05
	case "Laser":
		r.Attack += 20
		r.CriticalRate += 0.05
		r.HitRate += 0.15
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
	case "Nano Suit":
		r.Defense += 20
		r.EvasionRate += 0.05
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
	case "Gloves":
		r.CriticalRate += 0.05
		r.Attack += 5
	}
}

func getEquipmentDetails(equipmentType, item string) string {
	switch equipmentType {
	case "Weapon":
		switch item {
		case "Sword":
			return "Sword: Attack +10, Hit Rate +10%"
		case "Gun":
			return "Gun: Attack +15, Speed -2, Hit Rate +5%"
		case "Laser":
			return "Laser: Attack +20, Critical Rate +5%, Hit Rate +15%"
		}
	case "Armor":
		switch item {
		case "Shield":
			return "Shield: Defense +10"
		case "Armor":
			return "Armor: Defense +15, Speed -3"
		case "Nano Suit":
			return "Nano Suit: Defense +20, Evasion Rate +5%"
		}
	case "Accessory":
		switch item {
		case "Boots":
			return "Boots: Speed +5"
		case "Helmet":
			return "Helmet: Defense +5, Speed -1"
		case "Gloves":
			return "Gloves: Critical Rate +5%, Attack +5"
		}
	}
	return "No details available."
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
	player := Robot{Name: "PlayerBot", HP: 100, Attack: 20, Defense: 10, Speed: 5, CriticalRate: 0.1, EvasionRate: 0.1, HitRate: 0.8}
	enemy := Robot{Name: "EnemyBot", HP: 100, Attack: 18, Defense: 8, Speed: 4, CriticalRate: 0.05, EvasionRate: 0.05, HitRate: 0.75}

	enemy.EquipWeapon("Gun")
	enemy.EquipArmor("Armor")
	enemy.EquipAccessory("Helmet")

	return &Game{
		player: player,
		enemy:  enemy,
		equipment: [][]string{
			{"Sword", "Gun", "Laser"},
			{"Shield", "Armor", "Nano Suit"},
			{"Boots", "Helmet", "Gloves"},
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
	// 命中判定
	if rng.Float64() > r.HitRate {
		return fmt.Sprintf("%s attacks %s but misses!", r.Name, enemy.Name)
	}

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
		criticalMsg = "\nIt's a critical hit!"
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

func drawBattleStatus(g *Game, msgWindow *ebiten.Image, screen *ebiten.Image, windowWidth, windowHeight int, startMsg string) {
	leftColumn := fmt.Sprintf(
		"Current Equipment:\nWeapon: %s\nArmor: %s\nAccessory: %s",
		g.player.Weapon, g.player.Armor, g.player.Accessory,
	)

	// 左カラムの表示
	drawText(msgWindow, leftColumn, 10, 10)

	// 中央メッセージの表示
	drawText(screen, startMsg, 10, 10)
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

	// 追加: 上部に2つのウィンドウを追加
	leftWindow := ebiten.NewImage(screenWidth/3, screenHeight/2+40)
	rightWindow := ebiten.NewImage(screenWidth/2+170, screenHeight/2+40)
	leftWindow.Fill(color.RGBA{0, 0, 0, 255})
	rightWindow.Fill(color.RGBA{0, 0, 0, 255})

	// 左ウィンドウの枠を描画
	vector.DrawFilledRect(leftWindow, 0, 0, float32(screenWidth/3), 2, color.White, false)
	vector.DrawFilledRect(leftWindow, 0, float32(screenHeight/2+40-2), float32(screenWidth/3), 2, color.White, false)
	vector.DrawFilledRect(leftWindow, 0, 0, 2, float32(screenHeight/2+40), color.White, false)
	vector.DrawFilledRect(leftWindow, float32(screenWidth/3-2), 0, 2, float32(screenHeight/2+40), color.White, false)

	// 右ウィンドウの枠を描画
	vector.DrawFilledRect(rightWindow, 0, 0, float32(screenWidth/2+170), 2, color.White, false)
	vector.DrawFilledRect(rightWindow, 0, float32(screenHeight/2+40-2), float32(screenWidth/2+170), 2, color.White, false)
	vector.DrawFilledRect(rightWindow, 0, 0, 2, float32(screenHeight/2+40), color.White, false)
	vector.DrawFilledRect(rightWindow, float32(screenWidth/2+170-2), 0, 2, float32(screenHeight/2+40), color.White, false)

	status := fmt.Sprintf(
		"Current Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d\nCritical Rate: %.2f\nEvasion Rate: %.2f\nHit Rate: %.2f",
		g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed, g.player.CriticalRate, g.player.EvasionRate, g.player.HitRate,
	)

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
			drawText(msgWindow, status, float64(windowWidth/2)+10, 10)

			// 右ウィンドウに選択中の装備の詳細を表示
			equipmentType := []string{"Weapon", "Armor", "Accessory"}[g.selectionPhase]
			selectedItem := g.equipment[g.selectionPhase][g.selected[g.selectionPhase]]
			details := getEquipmentDetails(equipmentType, selectedItem)
			drawText(rightWindow, details, 10, 10)
		} else {
			drawText(msgWindow, status, float64(windowWidth/2)+10, 10)
			drawBattleStatus(g, msgWindow, rightWindow, windowWidth, windowHeight, "Press Z to start the battle!")
		}
	} else {
		if g.battleEnded {
			rightMsg := "You win!"
			if !g.reslt {
				rightMsg = "You lose!"
			}
			drawText(rightWindow, rightMsg, 10, 10)
			drawMessages(msgWindow, "Press Z to reset the game.")
			// Z キーでゲームをリセット
			if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
				*g = *NewGame()
			}
		} else {
			if g.turn == 1 {
				drawText(rightWindow, "Battle Start!", 10, 10)
			}
			msg := strings.Join(g.messages, "\n")
			drawMessages(rightWindow, msg)
			drawText(msgWindow, status, float64(windowWidth/2)+10, 10)
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(windowX), float64(windowY))
	screen.DrawImage(msgWindow, op)

	// 左ウィンドウに画像を表示
	image, _, err := ebitenutil.NewImageFromFile("image/test.jpg")
	if err == nil {
		op := &ebiten.DrawImageOptions{}
		// 画像を表示する位置とスケールを調整
		scaleX := float64(screenWidth/3-20) / float64(image.Bounds().Dx())
		scaleY := float64(screenHeight/2+20) / float64(image.Bounds().Dy())
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(10, 10) // 枠と画像の間に10ピクセルの隙間を設定
		leftWindow.DrawImage(image, op)
	}

	opLeft := &ebiten.DrawImageOptions{}
	opLeft.GeoM.Translate(10, 10)
	screen.DrawImage(leftWindow, opLeft)

	opRight := &ebiten.DrawImageOptions{}
	opRight.GeoM.Translate(float64(screenWidth)/2-180, 10)
	screen.DrawImage(rightWindow, opRight)
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
