package interaksi

import "github.com/hajimehoshi/ebiten/v2"

// Option is a functional configuration option for constructing an Interactable.
type Option func(i *Interactable)

// WithPosition sets the world coordinate (x, y).
func WithPosition(x, y float64) Option {
	return func(i *Interactable) {
		i.X = x
		i.Y = y
	}
}

// WithScale sets horizontal and vertical scale.
func WithScale(sx, sy float64) Option {
	return func(i *Interactable) {
		i.Sx = sx
		i.Sy = sy
	}
}

// WithUniformScale sets uniform horizontal and vertical scale.
func WithUniformScale(s float64) Option {
	return func(i *Interactable) {
		i.Sx = s
		i.Sy = s
	}
}

// WithOrigin sets the origin/pivot point in pixel offsets.
func WithOrigin(ox, oy float64) Option {
	return func(i *Interactable) {
		i.Ox = ox
		i.Oy = oy
	}
}

// WithPivot sets origin normalized to image dimensions (0.0=top/left, 0.5=center, 1.0=bottom/right).
func WithPivot(nx, ny float64) Option {
	return func(i *Interactable) {
		i.SetPivot(nx, ny)
	}
}

// WithZIndex sets the interaction and rendering layer priority.
func WithZIndex(z int) Option {
	return func(i *Interactable) {
		i.zIndex = z
	}
}

// WithHitboxPadding expands the hitbox beyond the image bounds by top, right, bottom, left.
func WithHitboxPadding(top, right, bottom, left float64) Option {
	return func(i *Interactable) {
		i.padding = NewInsets(top, right, bottom, left)
	}
}

// WithHitboxPaddingUniform expands the hitbox uniformly on all 4 sides.
func WithHitboxPaddingUniform(pad float64) Option {
	return func(i *Interactable) {
		i.padding = UniformInsets(pad)
	}
}

// WithHitboxInset contracts the hitbox inwards on all 4 sides.
func WithHitboxInset(top, right, bottom, left float64) Option {
	return func(i *Interactable) {
		i.inset = NewInsets(top, right, bottom, left)
	}
}

// WithHitboxInsetUniform contracts the hitbox uniformly on all 4 sides.
func WithHitboxInsetUniform(inset float64) Option {
	return func(i *Interactable) {
		i.inset = UniformInsets(inset)
	}
}

// WithCustomHitbox manually overrides the auto-computed AABB with a custom one.
func WithCustomHitbox(box AABB) Option {
	return func(i *Interactable) {
		b := box
		i.customHitbox = &b
	}
}

// WithCustomSize sets dimensions for image-less interactables.
func WithCustomSize(w, h float64) Option {
	return func(i *Interactable) {
		i.customWidth = w
		i.customHeight = h
	}
}

// WithHoveredImage sets the image variant shown when hovered.
func WithHoveredImage(img *ebiten.Image) Option {
	return func(i *Interactable) {
		i.hoveredImage = img
	}
}

// WithPressedImage sets the image variant shown when pressed.
func WithPressedImage(img *ebiten.Image) Option {
	return func(i *Interactable) {
		i.pressedImage = img
	}
}

// WithDisabledImage sets the image variant shown when disabled.
func WithDisabledImage(img *ebiten.Image) Option {
	return func(i *Interactable) {
		i.disabledImage = img
	}
}

// WithStateImage sets an image variant for the specified state.
func WithStateImage(state State, img *ebiten.Image) Option {
	return func(i *Interactable) {
		i.SetStateImage(state, img)
	}
}

// WithStateTint sets color scale tinting for a specific interaction state.
func WithStateTint(state State, cs ebiten.ColorScale) Option {
	return func(i *Interactable) {
		i.stateColorScales[state] = cs
	}
}

// WithDefaultTint sets base color scale tinting.
func WithDefaultTint(cs ebiten.ColorScale) Option {
	return func(i *Interactable) {
		i.defaultColorScale = cs
		i.hasDefaultColor = true
	}
}

// WithOnClick sets the callback for click events.
func WithOnClick(fn ClickCallback) Option {
	return func(i *Interactable) {
		i.onClick = fn
	}
}

// WithOnHoverEnter sets the callback for hover enter events.
func WithOnHoverEnter(fn HoverCallback) Option {
	return func(i *Interactable) {
		i.onHoverEnter = fn
	}
}

// WithOnHoverExit sets the callback for hover exit events.
func WithOnHoverExit(fn HoverCallback) Option {
	return func(i *Interactable) {
		i.onHoverExit = fn
	}
}

// WithOnPress sets the callback for press events.
func WithOnPress(fn PressCallback) Option {
	return func(i *Interactable) {
		i.onPress = fn
	}
}

// WithOnRelease sets the callback for release events.
func WithOnRelease(fn ReleaseCallback) Option {
	return func(i *Interactable) {
		i.onRelease = fn
	}
}

// WithOnDrag sets the callback for drag movement.
func WithOnDrag(fn DragCallback) Option {
	return func(i *Interactable) {
		i.onDrag = fn
	}
}

// WithOnStateChange sets the callback for state transition events.
func WithOnStateChange(fn StateChangeCallback) Option {
	return func(i *Interactable) {
		i.onStateChange = fn
	}
}

// WithCustomDraw provides a custom draw routine for rendering.
func WithCustomDraw(fn DrawCallback) Option {
	return func(i *Interactable) {
		i.customDraw = fn
	}
}

// WithEnabled sets whether the item starts enabled (default true).
func WithEnabled(enabled bool) Option {
	return func(i *Interactable) {
		i.enabled = enabled
	}
}

// WithVisible sets whether the item starts visible (default true).
func WithVisible(visible bool) Option {
	return func(i *Interactable) {
		i.visible = visible
	}
}

// WithBlocksInput sets whether this item consumes hit tests (default true).
func WithBlocksInput(blocks bool) Option {
	return func(i *Interactable) {
		i.blocksInput = blocks
	}
}

// WithName sets a debug / human-readable name.
func WithName(name string) Option {
	return func(i *Interactable) {
		i.name = name
	}
}

// WithUserData attaches user-defined data.
func WithUserData(data any) Option {
	return func(i *Interactable) {
		i.userData = data
	}
}

