package interaksi

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// State represents the current interaction lifecycle state of an Interactable.
type State int

const (
	// StateIdle indicates the item is enabled, not hovered, and not pressed.
	StateIdle State = iota

	// StateHovered indicates the cursor is currently over the item's hitbox.
	StateHovered

	// StatePressed indicates the mouse button is held down over this item.
	StatePressed

	// StateDisabled indicates the item is inactive and ignores interactions.
	StateDisabled
)

// String returns the human-readable name of the state.
func (s State) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateHovered:
		return "Hovered"
	case StatePressed:
		return "Pressed"
	case StateDisabled:
		return "Disabled"
	default:
		return fmt.Sprintf("State(%d)", int(s))
	}
}

// ClickCallback is invoked when an interactable is clicked.
type ClickCallback func(btn ebiten.MouseButton)

// HoverCallback is invoked when cursor enters or leaves the interactable's hitbox.
type HoverCallback func()

// PressCallback is invoked when the mouse button is pressed on the interactable.
type PressCallback func()

// ReleaseCallback is invoked when the mouse button is released on the interactable.
type ReleaseCallback func()

// DragCallback is invoked on mouse movement while the interactable is pressed.
// dx and dy indicate the delta movement from the previous frame.
type DragCallback func(dx, dy float64)

// StateChangeCallback is invoked whenever the interactable transitions between states.
type StateChangeCallback func(oldState, newState State)

// DrawCallback is an optional hook allowing custom rendering for an interactable.
type DrawCallback func(item *Interactable, screen *ebiten.Image, op *ebiten.DrawImageOptions)

