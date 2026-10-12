package interaksi

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// helperImage creates a dummy ebiten.Image with the specified dimensions for tests.
func helperImage(w, h int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	return ebiten.NewImageFromImage(img)
}

func TestAABBContainsAndIntersects(t *testing.T) {
	box := NewAABB(10, 20, 50, 60)

	// Interior point
	if !box.Contains(30, 40) {
		t.Fatalf("expected (30, 40) to be inside AABB")
	}

	// Boundary points (inclusive)
	if !box.Contains(10, 20) || !box.Contains(50, 60) {
		t.Fatalf("expected boundary corners to be inside AABB")
	}

	// Exterior points
	if box.Contains(9.9, 40) || box.Contains(50.1, 40) || box.Contains(30, 19.9) || box.Contains(30, 60.1) {
		t.Fatalf("expected exterior points to be outside AABB")
	}

	// Width, height, center
	if box.Width() != 40 || box.Height() != 40 {
		t.Fatalf("expected 40x40 dimensions, got %fx%f", box.Width(), box.Height())
	}
	center := box.Center()
	if center.X != 30 || center.Y != 40 {
		t.Fatalf("expected center (30, 40), got (%f, %f)", center.X, center.Y)
	}

	// Intersections
	overlap := NewAABB(40, 50, 70, 80)
	if !box.Intersects(overlap) {
		t.Fatalf("expected boxes to intersect")
	}

	disjoint := NewAABB(100, 100, 120, 120)
	if box.Intersects(disjoint) {
		t.Fatalf("expected boxes not to intersect")
	}

	// Insets and padding
	padded := box.Pad(NewInsets(5, 10, 15, 20))
	if padded.Min.X != -10 || padded.Min.Y != 15 || padded.Max.X != 60 || padded.Max.Y != 75 {
		t.Fatalf("unexpected padded box: %+v", padded)
	}

	insetted := box.Inset(NewInsets(5, 5, 5, 5))
	if insetted.Min.X != 15 || insetted.Min.Y != 25 || insetted.Max.X != 45 || insetted.Max.Y != 55 {
		t.Fatalf("unexpected inset box: %+v", insetted)
	}
}

func TestInteractableHitboxComputation(t *testing.T) {
	img := helperImage(100, 50)

	// Standard transform at (200, 100)
	it := New(img,
		WithPosition(200, 100),
		WithScale(1.0, 1.0),
		WithOrigin(0, 0),
	)

	hb := it.Hitbox()
	if hb.Min.X != 200 || hb.Min.Y != 100 || hb.Max.X != 300 || hb.Max.Y != 150 {
		t.Fatalf("unexpected standard hitbox: %+v", hb)
	}

	// Scaled 2x
	it.SetScale(2.0, 2.0)
	hb = it.Hitbox()
	if hb.Min.X != 200 || hb.Min.Y != 100 || hb.Max.X != 400 || hb.Max.Y != 200 {
		t.Fatalf("unexpected scaled 2x hitbox: %+v", hb)
	}

	// Center pivot (Origin = 50, 25) with Scale 1x
	it.SetScale(1.0, 1.0)
	it.SetPivot(0.5, 0.5)
	hb = it.Hitbox()
	if hb.Min.X != 150 || hb.Min.Y != 75 || hb.Max.X != 250 || hb.Max.Y != 125 {
		t.Fatalf("unexpected centered pivot hitbox: %+v", hb)
	}

	// Hitbox padding override (expands clickable area)
	it.SetOrigin(0, 0)
	it.SetHitboxPadding(10, 20, 30, 40)
	hb = it.Hitbox()
	// Base is (200, 100) -> (300, 150)
	// Min.X - Left (40) = 160
	// Min.Y - Top (10) = 90
	// Max.X + Right (20) = 320
	// Max.Y + Bottom (30) = 180
	if hb.Min.X != 160 || hb.Min.Y != 90 || hb.Max.X != 320 || hb.Max.Y != 180 {
		t.Fatalf("unexpected padded hitbox: %+v", hb)
	}

	// Custom Hitbox override
	custom := NewAABB(0, 0, 50, 50)
	it.SetCustomHitbox(&custom)
	hb = it.Hitbox()
	if hb.Min.X != 0 || hb.Max.X != 50 {
		t.Fatalf("expected custom hitbox override")
	}
}

func TestManagerInteractionLifecycle(t *testing.T) {
	mgr := NewManager()
	img := helperImage(100, 100)

	var (
		hoverEnterCount int
		hoverExitCount  int
		pressCount      int
		releaseCount    int
		clickCount      int
		lastClickBtn    ebiten.MouseButton
	)

	it := New(img,
		WithPosition(50, 50),
		WithOnHoverEnter(func() { hoverEnterCount++ }),
		WithOnHoverExit(func() { hoverExitCount++ }),
		WithOnPress(func() { pressCount++ }),
		WithOnRelease(func() { releaseCount++ }),
		WithOnClick(func(btn ebiten.MouseButton) {
			clickCount++
			lastClickBtn = btn
		}),
	)
	mgr.Register(it)

	// Frame 1: Cursor outside (0, 0), no mouse
	mgr.Update(0, 0, false)
	if it.State() != StateIdle || hoverEnterCount != 0 {
		t.Fatalf("expected Idle, got %v", it.State())
	}

	// Frame 2: Cursor moves over item (60, 60), no mouse -> HoverEnter
	mgr.Update(60, 60, false)
	if it.State() != StateHovered || hoverEnterCount != 1 {
		t.Fatalf("expected Hovered with 1 enter, got %v, enters=%d", it.State(), hoverEnterCount)
	}

	// Frame 3: Mouse pressed while hovering (60, 60) -> Press
	mgr.Update(60, 60, true)
	if it.State() != StatePressed || pressCount != 1 {
		t.Fatalf("expected Pressed with 1 press, got %v, presses=%d", it.State(), pressCount)
	}

	// Frame 4: Mouse released while still inside (60, 60) -> Release + Click
	mgr.Update(60, 60, false)
	if it.State() != StateHovered || releaseCount != 1 || clickCount != 1 {
		t.Fatalf("expected Hovered with 1 release & 1 click, got %v, releases=%d, clicks=%d",
			it.State(), releaseCount, clickCount)
	}
	if lastClickBtn != ebiten.MouseButtonLeft {
		t.Fatalf("expected left button click, got %v", lastClickBtn)
	}

	// Frame 5: Cursor moves away (300, 300) -> HoverExit
	mgr.Update(300, 300, false)
	if it.State() != StateIdle || hoverExitCount != 1 {
		t.Fatalf("expected Idle with 1 exit, got %v, exits=%d", it.State(), hoverExitCount)
	}
}

func TestManagerClickCancelledWhenReleasedOutside(t *testing.T) {
	mgr := NewManager()
	img := helperImage(100, 100)

	var (
		pressCount   int
		releaseCount int
		clickCount   int
	)

	it := New(img,
		WithPosition(50, 50),
		WithOnPress(func() { pressCount++ }),
		WithOnRelease(func() { releaseCount++ }),
		WithOnClick(func(btn ebiten.MouseButton) { clickCount++ }),
	)
	mgr.Register(it)

	// Press inside
	mgr.Update(60, 60, true)
	if pressCount != 1 {
		t.Fatalf("expected press")
	}

	// Drag outside while holding
	mgr.Update(300, 300, true)

	// Release outside -> OnRelease fires, but OnClick MUST NOT fire!
	mgr.Update(300, 300, false)
	if releaseCount != 1 {
		t.Fatalf("expected release when mouse released")
	}
	if clickCount != 0 {
		t.Fatalf("click must NOT trigger when released outside hitbox, got clicks=%d", clickCount)
	}
	if it.State() != StateIdle {
		t.Fatalf("expected state Idle after outside release, got %v", it.State())
	}
}

func TestManagerZIndexEventPropagationAndConsumption(t *testing.T) {
	mgr := NewManager()
	img := helperImage(100, 100)

	var (
		cardAClicks int
		cardBClicks int
	)

	// Card A (Lower Z: 1) at (50, 50)
	cardA := New(img,
		WithName("CardA"),
		WithPosition(50, 50),
		WithZIndex(1),
		WithOnClick(func(btn ebiten.MouseButton) { cardAClicks++ }),
	)

	// Card B (Higher Z: 2) overlapping at (80, 80)
	// Overlapping region is (80, 80) to (150, 150)
	cardB := New(img,
		WithName("CardB"),
		WithPosition(80, 80),
		WithZIndex(2),
		WithOnClick(func(btn ebiten.MouseButton) { cardBClicks++ }),
	)

	mgr.Register(cardA)
	mgr.Register(cardB)

	// Hover in overlapping region (90, 90)
	mgr.Update(90, 90, false)
	if mgr.HoveredItem() != cardB {
		t.Fatalf("expected top cardB to be hovered in overlapping area, got %v", mgr.HoveredItem().Name())
	}
	if cardA.IsHovered() {
		t.Fatalf("lower cardA must not be hovered when higher cardB consumes hover")
	}

	// Click in overlapping region (90, 90)
	mgr.Update(90, 90, true)  // Press
	mgr.Update(90, 90, false) // Release

	if cardBClicks != 1 {
		t.Fatalf("expected cardB click count = 1, got %d", cardBClicks)
	}
	if cardAClicks != 0 {
		t.Fatalf("cardA must not receive click in overlapping region, got %d", cardAClicks)
	}

	// Now bring cardA to front!
	mgr.BringToFront(cardA)
	if cardA.ZIndex() <= cardB.ZIndex() {
		t.Fatalf("cardA ZIndex (%d) must now be higher than cardB (%d)", cardA.ZIndex(), cardB.ZIndex())
	}

	// Click again in overlapping region (90, 90) -> Now Card A must receive click!
	mgr.Update(90, 90, true)  // Press
	mgr.Update(90, 90, false) // Release

	if cardAClicks != 1 {
		t.Fatalf("expected cardA click count = 1 after BringToFront, got %d", cardAClicks)
	}
	if cardBClicks != 1 {
		t.Fatalf("cardB should remain at 1 click, got %d", cardBClicks)
	}
}

func TestManagerDragDelta(t *testing.T) {
	mgr := NewManager()
	img := helperImage(50, 50)

	var totalDx, totalDy float64

	prop := New(img,
		WithPosition(100, 100),
		WithOnDrag(func(dx, dy float64) {
			totalDx += dx
			totalDy += dy
		}),
	)
	mgr.Register(prop)

	// Press inside prop at (110, 110)
	mgr.Update(110, 110, true)

	// Move cursor by +5, +10
	mgr.Update(115, 120, true)
	// Move cursor by +15, -5
	mgr.Update(130, 115, true)

	if totalDx != 20 || totalDy != 5 {
		t.Fatalf("expected drag delta (20, 5), got (%f, %f)", totalDx, totalDy)
	}

	// Release
	mgr.Update(130, 115, false)
	if prop.State() != StateHovered {
		t.Fatalf("expected Hovered on release inside, got %v", prop.State())
	}
}

func TestDisabledAndHiddenState(t *testing.T) {
	mgr := NewManager()
	img := helperImage(50, 50)

	var clicks int
	btn := New(img,
		WithPosition(50, 50),
		WithEnabled(false),
		WithOnClick(func(b ebiten.MouseButton) { clicks++ }),
	)
	mgr.Register(btn)

	// Hover and click while disabled
	mgr.Update(60, 60, false)
	if btn.State() != StateDisabled || mgr.HoveredItem() != nil {
		t.Fatalf("disabled item must not be hovered, state=%v", btn.State())
	}

	mgr.Update(60, 60, true)
	mgr.Update(60, 60, false)
	if clicks != 0 {
		t.Fatalf("disabled item must not receive clicks, got %d", clicks)
	}

	// Re-enable
	btn.SetEnabled(true)
	mgr.Update(60, 60, false)
	if btn.State() != StateHovered {
		t.Fatalf("expected Hovered after re-enabling, got %v", btn.State())
	}

	// Hide
	btn.SetVisible(false)
	mgr.Update(60, 60, false)
	if btn.State() != StateIdle || mgr.HoveredItem() != nil {
		t.Fatalf("hidden item must not be hovered")
	}
}

func TestUnregisterWhileActive(t *testing.T) {
	mgr := NewManager()
	img := helperImage(50, 50)

	var exitCalled bool
	btn := New(img,
		WithPosition(50, 50),
		WithOnHoverExit(func() { exitCalled = true }),
	)
	mgr.Register(btn)

	mgr.Update(60, 60, false)
	if mgr.HoveredItem() != btn {
		t.Fatalf("expected hovered")
	}

	// Unregister while hovered
	mgr.Unregister(btn)
	if mgr.HoveredItem() != nil {
		t.Fatalf("hovered item should be cleared after unregister")
	}
	if !exitCalled {
		t.Fatalf("expected onHoverExit callback on unregistering hovered item")
	}
	if mgr.Count() != 0 {
		t.Fatalf("expected manager count 0, got %d", mgr.Count())
	}
}

func TestManagerDrawAndDebugHitbox(t *testing.T) {
	mgr := NewManager()
	img := helperImage(50, 50)
	screen := helperImage(200, 200)

	it1 := New(img, WithPosition(10, 10), WithZIndex(1))
	it2 := New(img, WithPosition(30, 30), WithZIndex(2))
	mgr.Register(it1)
	mgr.Register(it2)

	// Draw standard
	mgr.Draw(screen)

	// Draw with debug hitbox enabled
	mgr.SetDebugDraw(true)
	if !mgr.IsDebugDraw() {
		t.Fatalf("expected debug draw to be true")
	}
	mgr.Draw(screen)
}

func TestStateImageVariantsAndFallback(t *testing.T) {
	idleImg := helperImage(30, 30)
	hoverImg := helperImage(30, 30)
	pressImg := helperImage(30, 30)
	disableImg := helperImage(30, 30)

	it := New(idleImg,
		WithHoveredImage(hoverImg),
		WithPressedImage(pressImg),
		WithDisabledImage(disableImg),
	)

	if it.ActiveImage() != idleImg {
		t.Fatalf("expected idle image")
	}

	it.SetState(StateHovered)
	if it.ActiveImage() != hoverImg {
		t.Fatalf("expected hover image")
	}

	it.SetState(StatePressed)
	if it.ActiveImage() != pressImg {
		t.Fatalf("expected pressed image")
	}

	it.SetState(StateDisabled)
	if it.ActiveImage() != disableImg {
		t.Fatalf("expected disabled image")
	}

	// Without variants, fallback to base image
	it2 := New(idleImg)
	it2.SetState(StateHovered)
	if it2.ActiveImage() != idleImg {
		t.Fatalf("expected fallback to base image")
	}
}

func TestPivotNormalized(t *testing.T) {
	img := helperImage(80, 40)
	it := New(img, WithPosition(100, 100))
	it.SetPivot(0.5, 0.5)

	ox, oy := it.Origin()
	if ox != 40 || oy != 20 {
		t.Fatalf("expected origin (40, 20), got (%f, %f)", ox, oy)
	}

	hb := it.Hitbox()
	if hb.Min.X != 60 || hb.Min.Y != 80 || hb.Max.X != 140 || hb.Max.Y != 120 {
		t.Fatalf("unexpected center pivot hitbox: %+v", hb)
	}
}

func TestCustomDrawCallback(t *testing.T) {
	img := helperImage(40, 40)
	screen := helperImage(100, 100)

	var customDrawCalled bool
	it := New(img,
		WithCustomDraw(func(item *Interactable, scr *ebiten.Image, op *ebiten.DrawImageOptions) {
			customDrawCalled = true
		}),
	)

	it.Draw(screen)
	if !customDrawCalled {
		t.Fatalf("expected customDraw to be called")
	}
}

func TestStableOrderSameZIndex(t *testing.T) {
	mgr := NewManager()
	img := helperImage(40, 40)

	// Both have ZIndex = 5, but item1 is registered first, item2 second
	item1 := New(img, WithName("Item1"), WithPosition(10, 10), WithZIndex(5))
	item2 := New(img, WithName("Item2"), WithPosition(10, 10), WithZIndex(5))

	mgr.Register(item1)
	mgr.Register(item2)

	// Hit test at (15, 15) should return item2 because it was registered later
	hit := mgr.HitTest(15, 15)
	if hit != item2 {
		t.Fatalf("expected item2 to be on top due to registration order, got %v", hit.Name())
	}
}

func TestFunctionalOptionsAndMethods(t *testing.T) {
	img := helperImage(50, 50)
	var stateChanged bool

	it := New(img,
		WithPosition(15, 25),
		WithScale(1.5, 2.0),
		WithUniformScale(1.2),
		WithOrigin(5, 10),
		WithPivot(0.5, 0.5),
		WithZIndex(42),
		WithHitboxPaddingUniform(10),
		WithHitboxInset(2, 3, 4, 5),
		WithHitboxInsetUniform(1),
		WithName("CustomItem"),
		WithUserData("test_tag"),
		WithBlocksInput(true),
		WithOnStateChange(func(old, new State) { stateChanged = true }),
	)

	if it.Name() != "CustomItem" {
		t.Fatalf("expected name CustomItem, got %s", it.Name())
	}
	if it.UserData() != "test_tag" {
		t.Fatalf("expected user data test_tag, got %v", it.UserData())
	}
	if it.ZIndex() != 42 {
		t.Fatalf("expected ZIndex 42, got %d", it.ZIndex())
	}
	if !it.BlocksInput() {
		t.Fatalf("expected BlocksInput true")
	}

	it.Move(10, 5)
	x, y := it.Position()
	if x != 25 || y != 30 {
		t.Fatalf("expected position (25, 30), got (%f, %f)", x, y)
	}

	it.SetState(StateHovered)
	if !stateChanged {
		t.Fatalf("expected onStateChange callback")
	}

	// State string representation
	if StateIdle.String() != "Idle" || StateHovered.String() != "Hovered" ||
		StatePressed.String() != "Pressed" || StateDisabled.String() != "Disabled" {
		t.Fatalf("unexpected state string values")
	}

	// Centered AABB and IsZero
	centerBox := NewAABBCentered(V2(50, 50), 20, 10)
	if centerBox.Min.X != 30 || centerBox.Max.X != 70 || centerBox.Min.Y != 40 || centerBox.Max.Y != 60 {
		t.Fatalf("unexpected centered AABB: %+v", centerBox)
	}
	var zeroBox AABB
	if !zeroBox.IsZero() {
		t.Fatalf("expected zeroBox to report IsZero = true")
	}

	// Image-less item with CustomSize
	blankIt := New(nil, WithCustomSize(100, 80), WithPosition(10, 10))
	hb := blankIt.Hitbox()
	if hb.Width() != 100 || hb.Height() != 80 {
		t.Fatalf("expected image-less hitbox to be 100x80, got %fx%f", hb.Width(), hb.Height())
	}
}

func TestManagerSendToBackAndClear(t *testing.T) {
	mgr := NewManager()
	img := helperImage(20, 20)

	it1 := New(img, WithZIndex(10))
	it2 := New(img, WithZIndex(20))
	mgr.Register(it1)
	mgr.Register(it2)

	// Send it2 to back
	mgr.SendToBack(it2)
	if it2.ZIndex() >= it1.ZIndex() {
		t.Fatalf("expected it2 to have lower Z-index than it1 after SendToBack")
	}

	items := mgr.Items()
	if len(items) != 2 || items[0] != it2 || items[1] != it1 {
		t.Fatalf("expected [it2, it1] in items after sort")
	}

	mgr.Clear()
	if mgr.Count() != 0 {
		t.Fatalf("expected manager to be empty after Clear")
	}
}
