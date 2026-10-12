package interaksi

import "math"

// Vec2 represents a 2D vector or point with 64-bit floating point precision.
type Vec2 struct {
	X float64
	Y float64
}

// V2 constructs a new Vec2.
func V2(x, y float64) Vec2 {
	return Vec2{X: x, Y: y}
}

// Insets defines padding or insets for all 4 borders of a 2D rectangle.
type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// NewInsets constructs Insets with individual top, right, bottom, and left values.
func NewInsets(top, right, bottom, left float64) Insets {
	return Insets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

// UniformInsets constructs Insets with identical values on all four edges.
func UniformInsets(v float64) Insets {
	return Insets{
		Top:    v,
		Right:  v,
		Bottom: v,
		Left:   v,
	}
}

// AABB represents an Axis-Aligned Bounding Box defined by its Min and Max corners.
type AABB struct {
	Min Vec2
	Max Vec2
}

// NewAABB constructs an AABB ensuring Min is always <= Max.
func NewAABB(minX, minY, maxX, maxY float64) AABB {
	return AABB{
		Min: Vec2{
			X: math.Min(minX, maxX),
			Y: math.Min(minY, maxY),
		},
		Max: Vec2{
			X: math.Max(minX, maxX),
			Y: math.Max(minY, maxY),
		},
	}
}

// NewAABBCentered constructs an AABB around a center point with the given half-extents.
func NewAABBCentered(center Vec2, halfWidth, halfHeight float64) AABB {
	hw := math.Abs(halfWidth)
	hh := math.Abs(halfHeight)
	return AABB{
		Min: Vec2{X: center.X - hw, Y: center.Y - hh},
		Max: Vec2{X: center.X + hw, Y: center.Y + hh},
	}
}

// Contains tests whether the given (px, py) coordinate lies within the AABB (inclusive).
func (a AABB) Contains(px, py float64) bool {
	return px >= a.Min.X && px <= a.Max.X && py >= a.Min.Y && py <= a.Max.Y
}

// ContainsVec tests whether the given point vector lies within the AABB.
func (a AABB) ContainsVec(p Vec2) bool {
	return a.Contains(p.X, p.Y)
}

// Intersects checks if this AABB intersects another AABB.
func (a AABB) Intersects(other AABB) bool {
	return a.Min.X <= other.Max.X &&
		a.Max.X >= other.Min.X &&
		a.Min.Y <= other.Max.Y &&
		a.Max.Y >= other.Min.Y
}

// Width returns the horizontal extent of the AABB.
func (a AABB) Width() float64 {
	return math.Max(0, a.Max.X-a.Min.X)
}

// Height returns the vertical extent of the AABB.
func (a AABB) Height() float64 {
	return math.Max(0, a.Max.Y-a.Min.Y)
}

// Center returns the midpoint of the AABB.
func (a AABB) Center() Vec2 {
	return Vec2{
		X: (a.Min.X + a.Max.X) * 0.5,
		Y: (a.Min.Y + a.Max.Y) * 0.5,
	}
}

// Pad returns a new AABB expanded outwards by the given insets.
func (a AABB) Pad(pad Insets) AABB {
	return AABB{
		Min: Vec2{X: a.Min.X - pad.Left, Y: a.Min.Y - pad.Top},
		Max: Vec2{X: a.Max.X + pad.Right, Y: a.Max.Y + pad.Bottom},
	}
}

// Inset returns a new AABB contracted inwards by the given insets.
func (a AABB) Inset(inset Insets) AABB {
	return AABB{
		Min: Vec2{X: a.Min.X + inset.Left, Y: a.Min.Y + inset.Top},
		Max: Vec2{X: a.Max.X - inset.Right, Y: a.Max.Y - inset.Bottom},
	}
}

// IsZero reports whether both Min and Max are at origin (0, 0).
func (a AABB) IsZero() bool {
	return a.Min.X == 0 && a.Min.Y == 0 && a.Max.X == 0 && a.Max.Y == 0
}

