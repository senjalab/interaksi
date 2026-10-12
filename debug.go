package interaksi

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawDebugHitboxes draws outline rectangles representing each interactable's AABB.
func (m *Manager) drawDebugHitboxes(screen *ebiten.Image) {
	for _, item := range m.items {
		if !item.visible {
			continue
		}

		aabb := item.Hitbox()
		var col color.RGBA
		switch item.state {
		case StateIdle:
			col = color.RGBA{R: 50, G: 200, B: 50, A: 200} // Green
		case StateHovered:
			col = color.RGBA{R: 255, G: 220, B: 40, A: 255} // Yellow
		case StatePressed:
			col = color.RGBA{R: 240, G: 60, B: 60, A: 255} // Red
		case StateDisabled:
			col = color.RGBA{R: 120, G: 120, B: 120, A: 160} // Gray
		}

		x := float32(aabb.Min.X)
		y := float32(aabb.Min.Y)
		w := float32(aabb.Width())
		h := float32(aabb.Height())

		if w > 0 && h > 0 {
			vector.StrokeRect(screen, x, y, w, h, 1.5, col, false)
		}
	}
}

