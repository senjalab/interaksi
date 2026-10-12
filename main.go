//go:build ignore

package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/senjalab/interaksi"
)

const (
	screenWidth  = 860
	screenHeight = 640
)

// Game implements ebiten.Game and hosts the interaksi.Manager.
type Game struct {
	manager *interaksi.Manager

	// References to interactables for demo interactivity
	cardRuby     *interaksi.Interactable
	cardSapphire *interaksi.Interactable
	relicProp    *interaksi.Interactable
	toggleBtn    *interaksi.Interactable
	disabledBtn  *interaksi.Interactable

	relicOrigin interaksi.Vec2
	eventLogs   []string

	debugHitbox bool
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("interaksi | AABB Interactable Component Demo")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g := newGame()
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}

func newGame() *Game {
	g := &Game{
		manager:     interaksi.NewManager(),
		eventLogs:   make([]string, 0, 10),
		debugHitbox: true,
	}
	g.manager.SetDebugDraw(g.debugHitbox)

	g.setupHUDButtons()
	g.setupOverlappingCards()
	g.setupDraggableProp()
	g.setupPaddedTrigger()

	g.logEvent("System initialized. Move mouse or click elements.")
	return g
}

func (g *Game) logEvent(msg string) {
	if len(g.eventLogs) >= 7 {
		g.eventLogs = g.eventLogs[1:]
	}
	g.eventLogs = append(g.eventLogs, msg)
}

// 1. HUD Buttons with Hover Tinting, Press Feedback, and Disabled state
func (g *Game) setupHUDButtons() {
	hoverTint := ebiten.ColorScale{}
	hoverTint.Scale(1.35, 1.35, 1.1, 1.0)

	pressTint := ebiten.ColorScale{}
	pressTint.Scale(0.7, 0.7, 0.7, 1.0)

	disabledTint := ebiten.ColorScale{}
	disabledTint.Scale(0.35, 0.35, 0.35, 0.7)

	// "CAST SPELL" Button
	btnImg := createButtonSprite(140, 42, "CAST SPELL", color.RGBA{R: 28, G: 65, B: 110, A: 255}, color.RGBA{R: 60, G: 160, B: 240, A: 255})
	spellBtn := interaksi.New(btnImg,
		interaksi.WithName("CastSpellBtn"),
		interaksi.WithPosition(40, 90),
		interaksi.WithZIndex(10),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnHoverEnter(func() { g.logEvent("[HOVER ENTER] 'CAST SPELL' button") }),
		interaksi.WithOnHoverExit(func() { g.logEvent("[HOVER EXIT]  'CAST SPELL' button") }),
		interaksi.WithOnPress(func() { g.logEvent("[PRESS]       'CAST SPELL' button") }),
		interaksi.WithOnRelease(func() { g.logEvent("[RELEASE]     'CAST SPELL' button") }),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.logEvent(fmt.Sprintf("[CLICK]       'CAST SPELL' (mouse button %v)", btn))
		}),
	)
	g.manager.Register(spellBtn)

	// "TOGGLE STATE" Button that toggles a target button's enabled state
	toggleImg := createButtonSprite(140, 42, "TOGGLE TARGET", color.RGBA{R: 85, G: 35, B: 95, A: 255}, color.RGBA{R: 200, G: 80, B: 220, A: 255})
	g.toggleBtn = interaksi.New(toggleImg,
		interaksi.WithName("ToggleBtn"),
		interaksi.WithPosition(40, 150),
		interaksi.WithZIndex(10),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			if g.disabledBtn != nil {
				newState := !g.disabledBtn.IsEnabled()
				g.disabledBtn.SetEnabled(newState)
				g.logEvent(fmt.Sprintf("[ACTION] Target button enabled = %v", newState))
			}
		}),
	)
	g.manager.Register(g.toggleBtn)

	// Target button (starts disabled)
	targetImg := createButtonSprite(140, 42, "LOCKED TARGET", color.RGBA{R: 50, G: 50, B: 55, A: 255}, color.RGBA{R: 110, G: 110, B: 120, A: 255})
	g.disabledBtn = interaksi.New(targetImg,
		interaksi.WithName("LockedTargetBtn"),
		interaksi.WithPosition(40, 210),
		interaksi.WithZIndex(10),
		interaksi.WithEnabled(false),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithStateTint(interaksi.StateDisabled, disabledTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.logEvent("[CLICK] Locked Target clicked!")
		}),
	)
	g.manager.Register(g.disabledBtn)
}

// 2. Overlapping Cards showing Z-Index Priority and Click Consumption
func (g *Game) setupOverlappingCards() {
	hoverTint := ebiten.ColorScale{}
	hoverTint.Scale(1.3, 1.3, 1.1, 1.0)

	pressTint := ebiten.ColorScale{}
	pressTint.Scale(0.75, 0.75, 0.75, 1.0)

	// Card Ruby (Z=10, drawn below Sapphire initially)
	rubyImg := createCardSprite(160, 200, "RUBY DRAGON", "ATK: 85 | DEF: 60",
		color.RGBA{R: 90, G: 20, B: 30, A: 255}, color.RGBA{R: 240, G: 60, B: 80, A: 255})

	g.cardRuby = interaksi.New(rubyImg,
		interaksi.WithName("Ruby Dragon Card"),
		interaksi.WithPosition(240, 90),
		interaksi.WithZIndex(10),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.logEvent(fmt.Sprintf("[CLICK] Ruby Dragon Card (Z=%d)", g.cardRuby.ZIndex()))
		}),
	)
	g.manager.Register(g.cardRuby)

	// Card Sapphire (Z=20, overlaps Ruby by 60px horizontally and vertically)
	sapphireImg := createCardSprite(160, 200, "SAPPHIRE GOLEM", "ATK: 40 | DEF: 95",
		color.RGBA{R: 20, G: 50, B: 100, A: 255}, color.RGBA{R: 60, G: 140, B: 240, A: 255})

	g.cardSapphire = interaksi.New(sapphireImg,
		interaksi.WithName("Sapphire Golem Card"),
		interaksi.WithPosition(310, 140),
		interaksi.WithZIndex(20),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.logEvent(fmt.Sprintf("[CLICK] Sapphire Golem Card (Z=%d)", g.cardSapphire.ZIndex()))
		}),
	)
	g.manager.Register(g.cardSapphire)

	// Button to swap layers
	swapBtnImg := createButtonSprite(150, 34, "SWAP LAYERS [Z]", color.RGBA{R: 40, G: 70, B: 50, A: 255}, color.RGBA{R: 90, G: 200, B: 120, A: 255})
	swapBtn := interaksi.New(swapBtnImg,
		interaksi.WithName("SwapLayersBtn"),
		interaksi.WithPosition(270, 360),
		interaksi.WithZIndex(15),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.swapLayers()
		}),
	)
	g.manager.Register(swapBtn)
}

func (g *Game) swapLayers() {
	if g.cardRuby.ZIndex() < g.cardSapphire.ZIndex() {
		g.manager.BringToFront(g.cardRuby)
		g.logEvent(fmt.Sprintf("[LAYERS] Ruby Dragon brought to front (Z=%d)", g.cardRuby.ZIndex()))
	} else {
		g.manager.BringToFront(g.cardSapphire)
		g.logEvent(fmt.Sprintf("[LAYERS] Sapphire Golem brought to front (Z=%d)", g.cardSapphire.ZIndex()))
	}
}

// 3. Draggable Prop with OnDrag and Click Detection
func (g *Game) setupDraggableProp() {
	hoverTint := ebiten.ColorScale{}
	hoverTint.Scale(1.4, 1.3, 1.0, 1.0)

	pressTint := ebiten.ColorScale{}
	pressTint.Scale(0.8, 0.8, 0.7, 1.0)

	g.relicOrigin = interaksi.V2(560, 130)
	relicImg := createRelicSprite(74)

	g.relicProp = interaksi.New(relicImg,
		interaksi.WithName("Golden Relic"),
		interaksi.WithPosition(g.relicOrigin.X, g.relicOrigin.Y),
		interaksi.WithZIndex(30),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnDrag(func(dx, dy float64) {
			g.relicProp.Move(dx, dy)
			g.logEvent(fmt.Sprintf("[DRAG] Relic moved by (dx=%.0f, dy=%.0f)", dx, dy))
		}),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.logEvent("[CLICK] Relic clicked without losing position!")
		}),
	)
	g.manager.Register(g.relicProp)

	// Reset position button
	resetBtnImg := createButtonSprite(140, 34, "RESET RELIC [R]", color.RGBA{R: 70, G: 60, B: 30, A: 255}, color.RGBA{R: 210, G: 180, B: 60, A: 255})
	resetBtn := interaksi.New(resetBtnImg,
		interaksi.WithName("ResetRelicBtn"),
		interaksi.WithPosition(530, 240),
		interaksi.WithZIndex(15),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithStateTint(interaksi.StatePressed, pressTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.relicProp.SetPosition(g.relicOrigin.X, g.relicOrigin.Y)
			g.logEvent("[RESET] Relic reset to original coordinates")
		}),
	)
	g.manager.Register(resetBtn)
}

// 4. Small Sprite with Expanded Hitbox Padding
func (g *Game) setupPaddedTrigger() {
	hoverTint := ebiten.ColorScale{}
	hoverTint.Scale(1.4, 1.4, 1.4, 1.0)

	ringImg := createRingSprite(28)
	ring := interaksi.New(ringImg,
		interaksi.WithName("Crystal Ring"),
		interaksi.WithPosition(730, 150),
		interaksi.WithZIndex(25),
		interaksi.WithHitboxPaddingUniform(22),
		interaksi.WithStateTint(interaksi.StateHovered, hoverTint),
		interaksi.WithOnClick(func(btn ebiten.MouseButton) {
			g.logEvent("[CLICK] Crystal Ring touched via padded hitbox (+22px)")
		}),
	)
	g.manager.Register(ring)
}

// Update handles frame logic and routes input into the interaksi.Manager.
func (g *Game) Update() error {
	cx, cy := ebiten.CursorPosition()
	isMouseDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	g.manager.Update(cx, cy, isMouseDown)

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		g.debugHitbox = !g.debugHitbox
		g.manager.SetDebugDraw(g.debugHitbox)
		g.logEvent(fmt.Sprintf("[KEY] Toggled Hitbox Outlines = %v", g.debugHitbox))
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.swapLayers()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.relicProp.SetPosition(g.relicOrigin.X, g.relicOrigin.Y)
		g.logEvent("[KEY] Reset Relic position")
	}

	return nil
}

// Draw renders backgrounds, sections, interactables, and debug HUD.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 16, G: 20, B: 28, A: 255})
	drawGrid(screen, 40, color.RGBA{R: 24, G: 32, B: 44, A: 255})

	ebitenutil.DebugPrintAt(screen, "=== INTERAKSI 2D INTERACTION ENGINE ===", 40, 20)
	ebitenutil.DebugPrintAt(screen, "[HUD Buttons]", 40, 65)
	ebitenutil.DebugPrintAt(screen, "[Overlapping Cards: Z-Priority]", 240, 65)
	ebitenutil.DebugPrintAt(screen, "[Draggable Prop]", 530, 65)
	ebitenutil.DebugPrintAt(screen, "[Padded Hitbox]", 700, 65)

	g.manager.Draw(screen)
	g.drawStatusOverlay(screen)
}

func (g *Game) drawStatusOverlay(screen *ebiten.Image) {
	vector.FillRect(screen, 0, 420, screenWidth, 220, color.RGBA{R: 10, G: 13, B: 18, A: 235}, false)
	vector.StrokeRect(screen, 0, 420, screenWidth, 2, 1, color.RGBA{R: 45, G: 60, B: 85, A: 255}, false)

	ebitenutil.DebugPrintAt(screen, "CONTROLS: [D] Toggle Hitbox Outlines | [Z] Swap Card Layers | [R] Reset Relic", 20, 432)

	cx, cy := ebiten.CursorPosition()
	isDown := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	hoverName := "None"
	hoverZ := 0
	if h := g.manager.HoveredItem(); h != nil {
		hoverName = h.Name()
		hoverZ = h.ZIndex()
	}

	pressedName := "None"
	if p := g.manager.PressedItem(); p != nil {
		pressedName = p.Name()
	}

	ptrInfo := fmt.Sprintf("Pointer: (%3d, %3d) | MouseDown: %-5v | Hovered: %-18s (Z=%2d) | Pressed: %s",
		cx, cy, isDown, hoverName, hoverZ, pressedName)
	ebitenutil.DebugPrintAt(screen, ptrInfo, 20, 452)

	ebitenutil.DebugPrintAt(screen, "HITBOX LEGEND: [Green: Idle]  [Yellow: Hovered]  [Red: Pressed]  [Gray: Disabled]", 20, 472)

	ebitenutil.DebugPrintAt(screen, "LIVE INTERACTION EVENT LOG:", 20, 498)
	startLogY := 518
	for idx, line := range g.eventLogs {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("> %s", line), 30, startLogY+idx*16)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func drawGrid(screen *ebiten.Image, step int, col color.Color) {
	for x := 0; x < screenWidth; x += step {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(screenHeight), 1, col, false)
	}
	for y := 0; y < screenHeight; y += step {
		vector.StrokeLine(screen, 0, float32(y), float32(screenWidth), float32(y), 1, col, false)
	}
}

func createButtonSprite(w, h int, label string, bgCol, borderCol color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	fw, fh := float32(w), float32(h)

	vector.FillRect(img, 0, 0, fw, fh, bgCol, false)
	vector.StrokeRect(img, 1, 1, fw-2, fh-2, 2, borderCol, false)
	vector.StrokeLine(img, 3, 3, fw-4, 3, 1, color.RGBA{R: 255, G: 255, B: 255, A: 70}, false)

	textX := (w - len(label)*6) / 2
	textY := (h - 12) / 2
	ebitenutil.DebugPrintAt(img, label, textX, textY)
	return img
}

func createCardSprite(w, h int, title, stats string, bgCol, borderCol color.RGBA) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	fw, fh := float32(w), float32(h)

	vector.FillRect(img, 0, 0, fw, fh, bgCol, false)
	vector.StrokeRect(img, 2, 2, fw-4, fh-4, 2, borderCol, false)
	vector.StrokeRect(img, 6, 6, fw-12, fh-12, 1, color.RGBA{R: 255, G: 255, B: 255, A: 50}, false)

	vector.FillRect(img, 8, 8, fw-16, 26, color.RGBA{R: 15, G: 15, B: 20, A: 210}, false)
	ebitenutil.DebugPrintAt(img, title, 14, 14)

	centerX, centerY := fw*0.5, fh*0.48
	vector.FillRect(img, centerX-25, centerY-25, 50, 50, borderCol, false)
	vector.FillRect(img, centerX-18, centerY-18, 36, 36, color.RGBA{R: 240, G: 240, B: 240, A: 200}, false)

	vector.FillRect(img, 8, fh-34, fw-16, 24, color.RGBA{R: 15, G: 15, B: 20, A: 210}, false)
	ebitenutil.DebugPrintAt(img, stats, 12, int(fh-28))

	return img
}

func createRelicSprite(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	s := float32(size)
	center := s * 0.5

	goldOuter := color.RGBA{R: 240, G: 185, B: 30, A: 255}
	goldInner := color.RGBA{R: 255, G: 230, B: 90, A: 255}
	darkGold := color.RGBA{R: 160, G: 110, B: 15, A: 255}

	vector.FillRect(img, 4, 4, s-8, s-8, goldOuter, false)
	vector.StrokeRect(img, 2, 2, s-4, s-4, 2, darkGold, false)
	vector.FillRect(img, center-16, center-16, 32, 32, goldInner, false)
	vector.StrokeRect(img, center-16, center-16, 32, 32, 1.5, darkGold, false)

	ebitenutil.DebugPrintAt(img, "RELIC", int(center)-14, int(center)-6)
	return img
}

func createRingSprite(size int) *ebiten.Image {
	img := ebiten.NewImage(size, size)
	s := float32(size)
	vector.FillRect(img, 2, 2, s-4, s-4, color.RGBA{R: 20, G: 180, B: 120, A: 255}, false)
	vector.StrokeRect(img, 1, 1, s-2, s-2, 1.5, color.RGBA{R: 80, G: 240, B: 180, A: 255}, false)
	vector.FillRect(img, s/2-3, s/2-3, 6, 6, color.RGBA{R: 255, G: 255, B: 255, A: 230}, false)
	return img
}

