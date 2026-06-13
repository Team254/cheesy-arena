// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains pixel RGB mappings for common LED colors.

package led

type Color struct {
	R byte
	G byte
	B byte
}

var (
	Black  = Color{0, 0, 0}
	Red    = Color{255, 0, 0}
	Green  = Color{0, 255, 0}
	Blue   = Color{0, 0, 255}
	Purple = Color{100, 0, 100}
	White  = Color{255, 255, 255}
)

// Scale Returns the color dimmed by the given factor.
func (color Color) Scale(factor float64) Color {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return Color{
		byte(float64(color.R)*factor + 0.5),
		byte(float64(color.G)*factor + 0.5),
		byte(float64(color.B)*factor + 0.5),
	}
}
