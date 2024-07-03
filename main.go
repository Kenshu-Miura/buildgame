package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math/rand"
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

const (
	StateTitle = iota
	StateSelection
	StateBattle
	StateEquip
	StateBattleEnd
)

//go:embed KiwiMaru-Regular.ttf
var fontData []byte

var (
	fontFace *text.GoTextFace
)

func init() {
	reader := bytes.NewReader(fontData)
	src, err := text.NewGoTextFaceSource(reader)
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
	case "NanoSuit":
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
		case "NanoSuit":
			return "NanoSuit: Defense +20, Evasion Rate +5%"
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

func getEquipmentImageFilename(equipmentType, item string) string {
	return fmt.Sprintf("images/%s_%s.jpg", equipmentType, item)
}

type Game struct {
	state          int
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
	battleCommands []string
	commandIndex   int
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
		state:  StateTitle,
		player: player,
		enemy:  enemy,
		equipment: [][]string{
			{"Sword", "Gun", "Laser"},
			{"Shield", "Armor", "NanoSuit"},
			{"Boots", "Helmet", "Gloves"},
		},
		selected:       [3]int{0, 0, 0},
		selectionPhase: 0,
		battleStarted:  false,
		battleEnded:    false,
		rng:            rng,
		turn:           1,
		battleCommands: []string{"攻撃", "防御", "回復"},
		commandIndex:   0,
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
	if g.battleEnded {
		return
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.commandIndex = (g.commandIndex - 1 + len(g.battleCommands)) % len(g.battleCommands)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		g.commandIndex = (g.commandIndex + 1) % len(g.battleCommands)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.lastAttackTime = time.Now()
		switch g.battleCommands[g.commandIndex] {
		case "攻撃":
			g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.AttackEnemy(&g.enemy, g.rng)))
		case "防御":
			g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.Defend()))
		case "回復":
			g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.player.Heal()))
		}

		if g.enemy.HP > 0 {
			g.messages = append(g.messages, fmt.Sprintf("Turn %d: %s", g.turn, g.enemy.AttackEnemy(&g.player, g.rng)))
		}

		if g.player.HP <= 0 {
			g.reslt = false
			g.battleEnded = true
		} else if g.enemy.HP <= 0 {
			g.reslt = true
			g.battleEnded = true
		}
		g.turn++
		if len(g.messages) > maxMessages {
			g.messages = g.messages[1:]
		}
	}
}

func (g *Game) Update() error {
	switch g.state {
	case StateTitle:
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			g.state = StateSelection
		}
	case StateSelection:
		if g.selectionPhase < 3 {
			g.handleSelectionPhase()
		} else {
			g.handleBattleStart()
			if g.battleStarted {
				g.state = StateBattle
			}
		}
	case StateBattle:
		if !g.battleEnded {
			g.handleBattlePhase()
			if g.battleEnded {
				if g.player.HP > 0 {
					g.state = StateEquip
					g.selectionPhase = 0 // 新しい装備選択のためにフェーズをリセット
					// 選択をリセット
					g.selected = [3]int{0, 0, 0}
				} else {
					g.state = StateBattleEnd
				}
			}
		}
	case StateEquip:
		if g.selectionPhase < 1 {
			g.handleSelectionPhase()
		} else if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			g.state = StateBattle
			g.battleEnded = false
			g.battleStarted = false
			g.lastAttackTime = time.Now() // タイマーをリセット
			g.turn = 1                    // ターン数をリセット
			g.messages = nil              // メッセージをリセット
			g.enemy = Robot{
				Name: "NewEnemyBot", HP: 100, Attack: 18, Defense: 8, Speed: 4, CriticalRate: 0.05, EvasionRate: 0.05, HitRate: 0.75,
			}
			g.enemy.EquipWeapon("Gun")
			g.enemy.EquipArmor("Armor")
			g.enemy.EquipAccessory("Helmet")
		}
	case StateBattleEnd:
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			*g = *NewGame()
		}
	}
	return nil
}

func (r *Robot) Defend() string {
	r.Defense += 5
	return fmt.Sprintf("%s defended, increasing Defense to %d", r.Name, r.Defense)
}

func (r *Robot) Heal() string {
	r.HP += 10
	if r.HP > 100 {
		r.HP = 100
	}
	return fmt.Sprintf("%s healed, increasing HP to %d", r.Name, r.HP)
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

func drawBattleStatus(g *Game, msgWindow *ebiten.Image, screen *ebiten.Image, startMsg string) {
	leftColumn := fmt.Sprintf(
		"Current Equipment:\nWeapon: %s\nArmor: %s\nAccessory: %s",
		g.player.Weapon, g.player.Armor, g.player.Accessory,
	)

	// 左カラムの表示
	drawText(msgWindow, leftColumn, 10, 10)

	// 中央メッセージの表示
	drawText(screen, startMsg, 10, 10)
}

func createWindow(windowWidth, windowHeight int) *ebiten.Image {
	window := ebiten.NewImage(windowWidth, windowHeight)
	window.Fill(color.Black)
	vector.DrawFilledRect(window, 0, 0, float32(windowWidth), 2, color.White, false)
	vector.DrawFilledRect(window, 0, float32(windowHeight-2), float32(windowWidth), 2, color.White, false)
	vector.DrawFilledRect(window, 0, 0, 2, float32(windowHeight), color.White, false)
	vector.DrawFilledRect(window, float32(windowWidth-2), 0, 2, float32(windowHeight), color.White, false)
	return window
}

func createSubWindow(width, height int) *ebiten.Image {
	subWindow := ebiten.NewImage(width, height)
	subWindow.Fill(color.RGBA{0, 0, 0, 255})
	vector.DrawFilledRect(subWindow, 0, 0, float32(width), 2, color.White, false)
	vector.DrawFilledRect(subWindow, 0, float32(height-2), float32(width), 2, color.White, false)
	vector.DrawFilledRect(subWindow, 0, 0, 2, float32(height), color.White, false)
	vector.DrawFilledRect(subWindow, float32(width-2), 0, 2, float32(height), color.White, false)
	return subWindow
}

func drawWindows(screen *ebiten.Image, msgWindow, leftWindow, rightWindow *ebiten.Image, windowX, windowY, screenWidth int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(windowX), float64(windowY))
	screen.DrawImage(msgWindow, op)

	opLeft := &ebiten.DrawImageOptions{}
	opLeft.GeoM.Translate(10, 10)
	screen.DrawImage(leftWindow, opLeft)

	opRight := &ebiten.DrawImageOptions{}
	opRight.GeoM.Translate(float64(screenWidth)/2-180, 10)
	screen.DrawImage(rightWindow, opRight)
}

func (g *Game) drawTitleScreen(screen *ebiten.Image) {
	msg := "Robot Battle\nPress Z to Start"
	drawText(screen, msg, screenWidth/2-100, screenHeight/2)
}

func (g *Game) drawSelectionScreen(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	windowX, windowY := 10, screenHeight/2+70
	windowWidth, windowHeight := screenWidth-20, screenHeight/2-80

	msgWindow := createWindow(windowWidth, windowHeight-10)
	leftWindow := createSubWindow(screenWidth/3, screenHeight/2+40)
	rightWindow := createSubWindow(screenWidth/2+170, screenHeight/2+40)

	status := fmt.Sprintf(
		"Current Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d\nCritical Rate: %.2f\nEvasion Rate: %.2f\nHit Rate: %.2f",
		g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed, g.player.CriticalRate, g.player.EvasionRate, g.player.HitRate,
	)

	if g.selectionPhase < 3 {
		msg := fmt.Sprintf("Select %s:\n", []string{"Weapon", "Armor", "Accessory"}[g.selectionPhase])
		for i, item := range g.equipment[g.selectionPhase] {
			cursor := " "
			if i == g.selected[g.selectionPhase] {
				cursor = ">"
			}
			msg += fmt.Sprintf("%s %s\n", cursor, item)
		}
		drawText(msgWindow, msg, 10, 10)

		equipmentType := []string{"Weapon", "Armor", "Accessory"}[g.selectionPhase]
		selectedItem := g.equipment[g.selectionPhase][g.selected[g.selectionPhase]]
		details := getEquipmentDetails(equipmentType, selectedItem)
		drawText(rightWindow, details, 10, 10)

		imageFilename := getEquipmentImageFilename(equipmentType, selectedItem)
		image, _, err := ebitenutil.NewImageFromFile(imageFilename)
		if err == nil {
			op := &ebiten.DrawImageOptions{}
			scaleX := float64(screenWidth/3-20) / float64(image.Bounds().Dx())
			scaleY := float64(screenHeight/2+20) / float64(image.Bounds().Dy())
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(10, 10)
			leftWindow.DrawImage(image, op)
		}
		drawText(msgWindow, status, float64(windowWidth/2)+10, 10)
	} else {
		drawText(msgWindow, status, float64(windowWidth/2)+10, 10)
		drawBattleStatus(g, msgWindow, rightWindow, "Press Z to start the battle!")
	}

	drawWindows(screen, msgWindow, leftWindow, rightWindow, windowX, windowY, screenWidth)
}

func (g *Game) drawBattleScreen(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	windowX, windowY := 10, screenHeight/2+70
	windowWidth, windowHeight := screenWidth-20, screenHeight/2-80

	msgWindow := createWindow(windowWidth, windowHeight-10)
	leftWindow := createSubWindow(screenWidth/3, screenHeight/2+40)
	rightWindow := createSubWindow(screenWidth/2+170, screenHeight/2+40)

	status := fmt.Sprintf(
		"Current Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d\nCritical Rate: %.2f\nEvasion Rate: %.2f\nHit Rate: %.2f",
		g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed, g.player.CriticalRate, g.player.EvasionRate, g.player.HitRate,
	)
	drawText(msgWindow, status, float64(windowWidth/2)+10, 10)

	if g.turn == 1 {
		drawText(rightWindow, "Battle Start!", 10, 10)
	}
	msg := strings.Join(g.messages, "\n")
	drawText(rightWindow, msg, 10, 10)

	commandMsg := "Commands:\n"
	for i, cmd := range g.battleCommands {
		cursor := " "
		if i == g.commandIndex {
			cursor = ">"
		}
		commandMsg += fmt.Sprintf("%s %s\n", cursor, cmd)
	}
	drawText(msgWindow, commandMsg, 10, 10)

	drawWindows(screen, msgWindow, leftWindow, rightWindow, windowX, windowY, screenWidth)
}

func (g *Game) drawBattleEndScreen(screen *ebiten.Image) {
	msg := "You win!"
	if !g.reslt {
		msg = "You lose!"
	}
	drawText(screen, msg, screenWidth/2-50, screenHeight/2)
	drawText(screen, "Press Z to restart", screenWidth/2-100, screenHeight/2+30)
}

func (g *Game) drawEquipScreen(screen *ebiten.Image) {
	screenWidth, screenHeight := screen.Bounds().Dx(), screen.Bounds().Dy()
	windowX, windowY := 10, screenHeight/2+70
	windowWidth, windowHeight := screenWidth-20, screenHeight/2-80

	msgWindow := createWindow(windowWidth, windowHeight-10)
	leftWindow := createSubWindow(screenWidth/3, screenHeight/2+40)
	rightWindow := createSubWindow(screenWidth/2+170, screenHeight/2+40)

	status := fmt.Sprintf(
		"Current Status:\nHP: %d\nAttack: %d\nDefense: %d\nSpeed: %d\nCritical Rate: %.2f\nEvasion Rate: %.2f\nHit Rate: %.2f",
		g.player.HP, g.player.Attack, g.player.Defense, g.player.Speed, g.player.CriticalRate, g.player.EvasionRate, g.player.HitRate,
	)

	if g.selectionPhase < 1 {
		msg := fmt.Sprintf("Select %s:\n", []string{"Weapon", "Armor", "Accessory"}[g.selectionPhase])
		for i, item := range g.equipment[g.selectionPhase] {
			cursor := " "
			if i == g.selected[g.selectionPhase] {
				cursor = ">"
			}
			msg += fmt.Sprintf("%s %s\n", cursor, item)
		}
		drawText(msgWindow, msg, 10, 10)

		equipmentType := []string{"Weapon", "Armor", "Accessory"}[g.selectionPhase]
		selectedItem := g.equipment[g.selectionPhase][g.selected[g.selectionPhase]]
		details := getEquipmentDetails(equipmentType, selectedItem)
		drawText(rightWindow, details, 10, 10)

		imageFilename := getEquipmentImageFilename(equipmentType, selectedItem)
		image, _, err := ebitenutil.NewImageFromFile(imageFilename)
		if err == nil {
			op := &ebiten.DrawImageOptions{}
			scaleX := float64(screenWidth/3-20) / float64(image.Bounds().Dx())
			scaleY := float64(screenHeight/2+20) / float64(image.Bounds().Dy())
			op.GeoM.Scale(scaleX, scaleY)
			op.GeoM.Translate(10, 10)
			leftWindow.DrawImage(image, op)
		}
		drawText(msgWindow, status, float64(windowWidth/2)+10, 10)
	} else {
		drawText(msgWindow, status, float64(windowWidth/2)+10, 10)
		drawBattleStatus(g, msgWindow, rightWindow, "Press Z to start the battle!")
	}

	drawWindows(screen, msgWindow, leftWindow, rightWindow, windowX, windowY, screenWidth)
}

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateTitle:
		g.drawTitleScreen(screen)
	case StateSelection:
		g.drawSelectionScreen(screen)
	case StateBattle:
		g.drawBattleScreen(screen)
	case StateEquip:
		g.drawEquipScreen(screen)
	case StateBattleEnd:
		g.drawBattleEndScreen(screen)
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
