package interaksi

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

// Manager coordinates an ordered collection of Interactable components,
// handling top-to-bottom event propagation, input consumption, and Z-index sorted rendering.
type Manager struct {
	items []*Interactable

	dirtySort bool
	nextOrder uint64

	hoveredItem *Interactable
	pressedItem *Interactable

	prevMouseDown bool
	prevCursorX   int
	prevCursorY   int

	debugDrawHitbox bool
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{
		items: make([]*Interactable, 0),
	}
}

// Register adds an interactable to the manager.
func (m *Manager) Register(item *Interactable) {
	if item == nil {
		return
	}
	if item.manager == m {
		return
	}
	if item.manager != nil {
		item.manager.Unregister(item)
	}

	item.manager = m
	item.order = m.nextOrder
	m.nextOrder++

	m.items = append(m.items, item)
	m.dirtySort = true
}

// Unregister removes an interactable from the manager.
func (m *Manager) Unregister(item *Interactable) {
	if item == nil || item.manager != m {
		return
	}

	if m.hoveredItem == item {
		if item.state == StateHovered && item.onHoverExit != nil {
			item.onHoverExit()
		}
		item.setState(StateIdle)
		m.hoveredItem = nil
	}

	if m.pressedItem == item {
		if item.state == StatePressed && item.onRelease != nil {
			item.onRelease()
		}
		item.setState(StateIdle)
		m.pressedItem = nil
	}

	for idx, it := range m.items {
		if it == item {
			m.items = append(m.items[:idx], m.items[idx+1:]...)
			break
		}
	}

	item.manager = nil
}

// Clear removes all registered interactables and resets interaction states.
func (m *Manager) Clear() {
	for _, it := range m.items {
		if it == m.hoveredItem && it.onHoverExit != nil {
			it.onHoverExit()
		}
		if it == m.pressedItem && it.onRelease != nil {
			it.onRelease()
		}
		it.setState(StateIdle)
		it.manager = nil
	}
	m.items = m.items[:0]
	m.hoveredItem = nil
	m.pressedItem = nil
	m.prevMouseDown = false
	m.dirtySort = false
}

// Count returns the number of registered interactables.
func (m *Manager) Count() int {
	return len(m.items)
}

// Items returns a slice of registered interactables sorted by Z-index ascending.
func (m *Manager) Items() []*Interactable {
	m.ensureSorted()
	res := make([]*Interactable, len(m.items))
	copy(res, m.items)
	return res
}

// HoveredItem returns the interactable currently under cursor.
func (m *Manager) HoveredItem() *Interactable {
	return m.hoveredItem
}

// PressedItem returns the interactable currently holding pointer capture.
func (m *Manager) PressedItem() *Interactable {
	return m.pressedItem
}

// markDirty informs the manager that an item's Z-index or order changed.
func (m *Manager) markDirty() {
	m.dirtySort = true
}

// ensureSorted sorts items by ZIndex ascending (lowest first, highest last).
// In case of equal ZIndex, items registered later have higher order.
func (m *Manager) ensureSorted() {
	if !m.dirtySort {
		return
	}
	sort.SliceStable(m.items, func(a, b int) bool {
		if m.items[a].zIndex != m.items[b].zIndex {
			return m.items[a].zIndex < m.items[b].zIndex
		}
		return m.items[a].order < m.items[b].order
	})
	m.dirtySort = false
}

// BringToFront places the item at the highest Z-index and order.
func (m *Manager) BringToFront(item *Interactable) {
	if item == nil || item.manager != m || len(m.items) == 0 {
		return
	}
	maxZ := m.items[0].zIndex
	for _, it := range m.items {
		if it.zIndex > maxZ {
			maxZ = it.zIndex
		}
	}
	item.zIndex = maxZ + 1
	item.order = m.nextOrder
	m.nextOrder++
	m.dirtySort = true
}

// SendToBack places the item at the lowest Z-index and order.
func (m *Manager) SendToBack(item *Interactable) {
	if item == nil || item.manager != m || len(m.items) == 0 {
		return
	}
	minZ := m.items[0].zIndex
	for _, it := range m.items {
		if it.zIndex < minZ {
			minZ = it.zIndex
		}
	}
	item.zIndex = minZ - 1
	item.order = 0
	m.dirtySort = true
}

// HitTest finds the top-most visible and enabled item containing (px, py).
func (m *Manager) HitTest(px, py float64) *Interactable {
	m.ensureSorted()
	for i := len(m.items) - 1; i >= 0; i-- {
		item := m.items[i]
		if !item.visible || !item.enabled {
			continue
		}
		if item.Contains(px, py) {
			return item
		}
	}
	return nil
}

// Update propagates interaction events top-to-bottom using primary mouse button.
// cursorX and cursorY are decoupled coordinates (screen, world, or viewport).
// Higher Z-index items receive and consume hits first.
func (m *Manager) Update(cursorX, cursorY int, isMouseDown bool) {
	m.UpdateWithButton(cursorX, cursorY, ebiten.MouseButtonLeft, isMouseDown)
}

// UpdateWithButton propagates interaction events top-to-bottom with a specific mouse button.
func (m *Manager) UpdateWithButton(cursorX, cursorY int, btn ebiten.MouseButton, isMouseDown bool) {
	m.ensureSorted()

	cx := float64(cursorX)
	cy := float64(cursorY)

	justPressed := isMouseDown && !m.prevMouseDown
	justReleased := !isMouseDown && m.prevMouseDown
	isHeld := isMouseDown && m.prevMouseDown
	dx := cursorX - m.prevCursorX
	dy := cursorY - m.prevCursorY

	// 1. Pointer Press Transition (Top-to-Bottom hit testing)
	if justPressed {
		var hitItem *Interactable
		for i := len(m.items) - 1; i >= 0; i-- {
			item := m.items[i]
			if !item.visible || !item.enabled {
				continue
			}
			if item.Contains(cx, cy) {
				hitItem = item
				if item.blocksInput {
					break
				}
			}
		}

		if hitItem != nil {
			m.pressedItem = hitItem

			// If another item was hovered, exit it
			if m.hoveredItem != nil && m.hoveredItem != hitItem {
				if m.hoveredItem.state == StateHovered && m.hoveredItem.onHoverExit != nil {
					m.hoveredItem.onHoverExit()
				}
				m.hoveredItem.setState(StateIdle)
			}

			m.hoveredItem = hitItem
			hitItem.setState(StatePressed)
			if hitItem.onPress != nil {
				hitItem.onPress()
			}
		}
	} else if isHeld {
		// 2. Dragging & Holding Pointer Capture
		if m.pressedItem != nil {
			item := m.pressedItem
			// Ensure item is still valid
			if !item.visible || !item.enabled {
				if item.onRelease != nil {
					item.onRelease()
				}
				item.setState(StateIdle)
				m.pressedItem = nil
			} else {
				if (dx != 0 || dy != 0) && item.onDrag != nil {
					item.onDrag(float64(dx), float64(dy))
				}
			}
		}
	} else if justReleased {
		// 3. Pointer Release Transition
		if m.pressedItem != nil {
			item := m.pressedItem
			m.pressedItem = nil

			if item.visible && item.enabled {
				if item.onRelease != nil {
					item.onRelease()
				}

				if item.Contains(cx, cy) {
					// Click confirmed (pressed and released inside)
					if item.onClick != nil {
						item.onClick(btn)
					}
					item.setState(StateHovered)
					m.hoveredItem = item
				} else {
					item.setState(StateIdle)
					m.hoveredItem = nil
				}
			} else {
				item.setState(StateIdle)
				m.hoveredItem = nil
			}
		}
	}

	// 4. Hover State Evaluation (Top-to-Bottom)
	if !isMouseDown && m.pressedItem == nil {
		var topHover *Interactable
		for i := len(m.items) - 1; i >= 0; i-- {
			item := m.items[i]
			if !item.visible || !item.enabled {
				continue
			}
			if item.Contains(cx, cy) {
				topHover = item
				if item.blocksInput {
					break
				}
			}
		}

		if topHover != m.hoveredItem {
			if m.hoveredItem != nil {
				if m.hoveredItem.state == StateHovered && m.hoveredItem.onHoverExit != nil {
					m.hoveredItem.onHoverExit()
				}
				m.hoveredItem.setState(StateIdle)
			}

			m.hoveredItem = topHover
			if topHover != nil {
				topHover.setState(StateHovered)
				if topHover.onHoverEnter != nil {
					topHover.onHoverEnter()
				}
			}
		}
	}

	m.prevMouseDown = isMouseDown
	m.prevCursorX = cursorX
	m.prevCursorY = cursorY
}

// SetDebugDraw toggles rendering of debug hitbox overlays on Draw.
func (m *Manager) SetDebugDraw(enabled bool) {
	m.debugDrawHitbox = enabled
}

// IsDebugDraw reports whether debug hitbox drawing is enabled.
func (m *Manager) IsDebugDraw() bool {
	return m.debugDrawHitbox
}

// Draw renders registered items sorted by Z-index (lowest to highest).
func (m *Manager) Draw(screen *ebiten.Image) {
	if screen == nil {
		return
	}
	m.ensureSorted()

	for _, item := range m.items {
		if !item.visible {
			continue
		}
		item.Draw(screen)
	}

	if m.debugDrawHitbox {
		m.drawDebugHitboxes(screen)
	}
}

