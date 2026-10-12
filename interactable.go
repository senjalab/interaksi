package interaksi

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Interactable is a 2D spatial interactive component with state tracking,
// transform coordinates, auto-computed AABB hitbox, and event callbacks.
type Interactable struct {
	// Identity & debug
	name     string
	userData any

	// State variants images
	image         *ebiten.Image // Idle / default image
	hoveredImage  *ebiten.Image
	pressedImage  *ebiten.Image
	disabledImage *ebiten.Image

	// Custom size for image-less interactables
	customWidth  float64
	customHeight float64

	// Transform
	X  float64 // World X position
	Y  float64 // World Y position
	Sx float64 // Horizontal scale
	Sy float64 // Vertical scale
	Ox float64 // Origin/Pivot X (in local pixels)
	Oy float64 // Origin/Pivot Y (in local pixels)

	// Hitbox tuning
	padding       Insets // Expands hitbox outwards
	inset         Insets // Contracts hitbox inwards
	customHitbox  *AABB  // Explicit override if non-nil

	// Layering & dispatcher
	zIndex       int
	order        uint64   // Stable tie-breaker for identical Z-indices
	manager      *Manager // Owning manager if registered
	blocksInput  bool     // Whether this item consumes hit-tests (default true)

	// Lifecycle state & flags
	state   State
	enabled bool
	visible bool

	// Callbacks
	onClick       ClickCallback
	onHoverEnter  HoverCallback
	onHoverExit   HoverCallback
	onPress       PressCallback
	onRelease     ReleaseCallback
	onDrag        DragCallback
	onStateChange StateChangeCallback
	customDraw    DrawCallback

	// Visual tinting per state
	defaultColorScale ebiten.ColorScale
	hasDefaultColor   bool
	stateColorScales  map[State]ebiten.ColorScale
}

// New creates a new Interactable with the provided base image and functional options.
func New(img *ebiten.Image, opts ...Option) *Interactable {
	item := &Interactable{
		image:            img,
		Sx:               1.0,
		Sy:               1.0,
		enabled:          true,
		visible:          true,
		blocksInput:      true,
		state:            StateIdle,
		stateColorScales: make(map[State]ebiten.ColorScale),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(item)
		}
	}

	if !item.enabled {
		item.state = StateDisabled
	}

	return item
}

// Name returns the descriptive name of the interactable.
func (i *Interactable) Name() string {
	return i.name
}

// SetName sets the descriptive name of the interactable.
func (i *Interactable) SetName(name string) *Interactable {
	i.name = name
	return i
}

// UserData returns arbitrary user data attached to this item.
func (i *Interactable) UserData() any {
	return i.userData
}

// SetUserData attaches arbitrary user data to this item.
func (i *Interactable) SetUserData(data any) *Interactable {
	i.userData = data
	return i
}

// State returns the current interaction lifecycle state.
func (i *Interactable) State() State {
	return i.state
}

// SetState updates the state and invokes onStateChange if the state changed.
func (i *Interactable) SetState(s State) *Interactable {
	if i.state == s {
		return i
	}
	old := i.state
	i.state = s
	if i.onStateChange != nil {
		i.onStateChange(old, s)
	}
	return i
}

// setState is an internal alias for SetState.
func (i *Interactable) setState(s State) {
	i.SetState(s)
}

// IsHovered reports whether the item is currently in StateHovered.
func (i *Interactable) IsHovered() bool {
	return i.state == StateHovered
}

// IsPressed reports whether the item is currently in StatePressed.
func (i *Interactable) IsPressed() bool {
	return i.state == StatePressed
}

// IsEnabled reports whether the item is currently enabled.
func (i *Interactable) IsEnabled() bool {
	return i.enabled
}

// SetEnabled enables or disables interaction.
func (i *Interactable) SetEnabled(enabled bool) *Interactable {
	if i.enabled == enabled {
		return i
	}
	i.enabled = enabled
	if !enabled {
		if i.state == StateHovered && i.onHoverExit != nil {
			i.onHoverExit()
		} else if i.state == StatePressed && i.onRelease != nil {
			i.onRelease()
		}
		i.setState(StateDisabled)
	} else {
		i.setState(StateIdle)
	}
	return i
}

// IsVisible reports whether the item is visible.
func (i *Interactable) IsVisible() bool {
	return i.visible
}

// SetVisible sets visibility. Hidden items are neither drawn nor hit-tested.
func (i *Interactable) SetVisible(visible bool) *Interactable {
	if i.visible == visible {
		return i
	}
	i.visible = visible
	if !visible && (i.state == StateHovered || i.state == StatePressed) {
		if i.state == StateHovered && i.onHoverExit != nil {
			i.onHoverExit()
		} else if i.state == StatePressed && i.onRelease != nil {
			i.onRelease()
		}
		i.setState(StateIdle)
	}
	return i
}

// BlocksInput reports whether this item absorbs hit-tests to block items underneath.
func (i *Interactable) BlocksInput() bool {
	return i.blocksInput
}

// SetBlocksInput sets whether this item consumes hit-tests.
func (i *Interactable) SetBlocksInput(blocks bool) *Interactable {
	i.blocksInput = blocks
	return i
}

// ZIndex returns the current layer priority.
func (i *Interactable) ZIndex() int {
	return i.zIndex
}

// SetZIndex updates the layer priority and triggers re-sorting in the manager.
func (i *Interactable) SetZIndex(z int) *Interactable {
	if i.zIndex == z {
		return i
	}
	i.zIndex = z
	if i.manager != nil {
		i.manager.markDirty()
	}
	return i
}

// Position returns the current (X, Y) coordinates.
func (i *Interactable) Position() (float64, float64) {
	return i.X, i.Y
}

// SetPosition updates the (X, Y) world position.
func (i *Interactable) SetPosition(x, y float64) *Interactable {
	i.X = x
	i.Y = y
	return i
}

// Move adds (dx, dy) to the current world position.
func (i *Interactable) Move(dx, dy float64) *Interactable {
	i.X += dx
	i.Y += dy
	return i
}

// Scale returns the horizontal and vertical scale factors.
func (i *Interactable) Scale() (float64, float64) {
	return i.Sx, i.Sy
}

// SetScale updates horizontal and vertical scale.
func (i *Interactable) SetScale(sx, sy float64) *Interactable {
	i.Sx = sx
	i.Sy = sy
	return i
}

// SetUniformScale updates both scale factors equally.
func (i *Interactable) SetUniformScale(s float64) *Interactable {
	i.Sx = s
	i.Sy = s
	return i
}

// Origin returns the origin/pivot offset in pixels.
func (i *Interactable) Origin() (float64, float64) {
	return i.Ox, i.Oy
}

// SetOrigin updates the origin/pivot point in pixel coordinates.
func (i *Interactable) SetOrigin(ox, oy float64) *Interactable {
	i.Ox = ox
	i.Oy = oy
	return i
}

// SetPivot sets origin as normalized ratios (0.0=top/left, 0.5=center, 1.0=bottom/right).
func (i *Interactable) SetPivot(nx, ny float64) *Interactable {
	w, h := i.BaseSize()
	i.Ox = w * nx
	i.Oy = h * ny
	return i
}

// Image returns the default/idle image.
func (i *Interactable) Image() *ebiten.Image {
	return i.image
}

// SetImage updates the default/idle image.
func (i *Interactable) SetImage(img *ebiten.Image) *Interactable {
	i.image = img
	return i
}

// SetStateImage assigns an image variant for a specific state.
func (i *Interactable) SetStateImage(state State, img *ebiten.Image) *Interactable {
	switch state {
	case StateIdle:
		i.image = img
	case StateHovered:
		i.hoveredImage = img
	case StatePressed:
		i.pressedImage = img
	case StateDisabled:
		i.disabledImage = img
	}
	return i
}

// ActiveImage returns the image corresponding to the current state, falling back to base image.
func (i *Interactable) ActiveImage() *ebiten.Image {
	switch i.state {
	case StateHovered:
		if i.hoveredImage != nil {
			return i.hoveredImage
		}
	case StatePressed:
		if i.pressedImage != nil {
			return i.pressedImage
		}
	case StateDisabled:
		if i.disabledImage != nil {
			return i.disabledImage
		}
	}
	return i.image
}

// BaseSize returns the width and height of the active image, or custom size if image-less.
func (i *Interactable) BaseSize() (float64, float64) {
	img := i.ActiveImage()
	if img != nil {
		b := img.Bounds()
		return float64(b.Dx()), float64(b.Dy())
	}
	if i.image != nil {
		b := i.image.Bounds()
		return float64(b.Dx()), float64(b.Dy())
	}
	return i.customWidth, i.customHeight
}

// SetCustomSize assigns manual width and height for invisible or procedurally drawn interactables.
func (i *Interactable) SetCustomSize(w, h float64) *Interactable {
	i.customWidth = w
	i.customHeight = h
	return i
}

// Hitbox returns the auto-computed AABB synced with image dimensions, transform, origin, and insets.
func (i *Interactable) Hitbox() AABB {
	if i.customHitbox != nil {
		return *i.customHitbox
	}

	w, h := i.BaseSize()
	if w <= 0 && h <= 0 {
		return AABB{Min: V2(i.X, i.Y), Max: V2(i.X, i.Y)}
	}

	// Calculate transformed corner bounds taking scale and origin into account
	x1 := i.X - i.Ox*i.Sx
	y1 := i.Y - i.Oy*i.Sy
	x2 := i.X + (w-i.Ox)*i.Sx
	y2 := i.Y + (h-i.Oy)*i.Sy

	minX := math.Min(x1, x2)
	maxX := math.Max(x1, x2)
	minY := math.Min(y1, y2)
	maxY := math.Max(y1, y2)

	// Apply padding (expands clickable area)
	minX -= i.padding.Left
	maxX += i.padding.Right
	minY -= i.padding.Top
	maxY += i.padding.Bottom

	// Apply inset (contracts clickable area)
	minX += i.inset.Left
	maxX -= i.inset.Right
	minY += i.inset.Top
	maxY -= i.inset.Bottom

	return NewAABB(minX, minY, maxX, maxY)
}

// Contains tests whether the given coordinate lies within this interactable's hitbox.
func (i *Interactable) Contains(px, py float64) bool {
	return i.Hitbox().Contains(px, py)
}

// SetHitboxPadding expands the hitbox outwards from the transformed image bounds.
func (i *Interactable) SetHitboxPadding(top, right, bottom, left float64) *Interactable {
	i.padding = NewInsets(top, right, bottom, left)
	return i
}

// SetHitboxInset contracts the hitbox inwards from the transformed image bounds.
func (i *Interactable) SetHitboxInset(top, right, bottom, left float64) *Interactable {
	i.inset = NewInsets(top, right, bottom, left)
	return i
}

// SetCustomHitbox manually overrides the hitbox with an explicit AABB.
func (i *Interactable) SetCustomHitbox(box *AABB) *Interactable {
	if box == nil {
		i.customHitbox = nil
	} else {
		b := *box
		i.customHitbox = &b
	}
	return i
}

// SetStateColorScale defines color tinting for a specific interaction state.
func (i *Interactable) SetStateColorScale(state State, cs ebiten.ColorScale) *Interactable {
	i.stateColorScales[state] = cs
	return i
}

// SetDefaultColorScale defines default base color tinting.
func (i *Interactable) SetDefaultColorScale(cs ebiten.ColorScale) *Interactable {
	i.defaultColorScale = cs
	i.hasDefaultColor = true
	return i
}

// OnClick assigns the callback invoked when the item is clicked.
func (i *Interactable) OnClick(fn ClickCallback) *Interactable {
	i.onClick = fn
	return i
}

// OnHoverEnter assigns the callback invoked when cursor enters the hitbox.
func (i *Interactable) OnHoverEnter(fn HoverCallback) *Interactable {
	i.onHoverEnter = fn
	return i
}

// OnHoverExit assigns the callback invoked when cursor exits the hitbox.
func (i *Interactable) OnHoverExit(fn HoverCallback) *Interactable {
	i.onHoverExit = fn
	return i
}

// OnPress assigns the callback invoked when mouse presses down inside hitbox.
func (i *Interactable) OnPress(fn PressCallback) *Interactable {
	i.onPress = fn
	return i
}

// OnRelease assigns the callback invoked when mouse button is released.
func (i *Interactable) OnRelease(fn ReleaseCallback) *Interactable {
	i.onRelease = fn
	return i
}

// OnDrag assigns the callback invoked when moving while pressed.
func (i *Interactable) OnDrag(fn DragCallback) *Interactable {
	i.onDrag = fn
	return i
}

// OnStateChange assigns the callback invoked whenever state transitions.
func (i *Interactable) OnStateChange(fn StateChangeCallback) *Interactable {
	i.onStateChange = fn
	return i
}

// OnCustomDraw assigns a custom draw function overriding standard image draw.
func (i *Interactable) OnCustomDraw(fn DrawCallback) *Interactable {
	i.customDraw = fn
	return i
}

// Draw renders the interactable onto the screen using its active image, transform, and state tint.
func (i *Interactable) Draw(screen *ebiten.Image) {
	if !i.visible || screen == nil {
		return
	}

	img := i.ActiveImage()
	if img == nil && i.customDraw == nil {
		return
	}

	var op ebiten.DrawImageOptions
	op.GeoM.Translate(-i.Ox, -i.Oy)
	op.GeoM.Scale(i.Sx, i.Sy)
	op.GeoM.Translate(i.X, i.Y)

	// Apply state tint if registered, or default tint
	if cs, ok := i.stateColorScales[i.state]; ok {
		op.ColorScale = cs
	} else if i.hasDefaultColor {
		op.ColorScale = i.defaultColorScale
	}

	if i.customDraw != nil {
		i.customDraw(i, screen, &op)
		return
	}

	if img != nil {
		screen.DrawImage(img, &op)
	}
}
