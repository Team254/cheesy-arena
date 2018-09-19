// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains pixel RGB mappings for common colors.

package led

type Color int

const (
	Red Color = iota
	Orange
	Yellow
	Green
	Teal
	Blue
	Purple
	White
	Black
	PurpleRed
	PurpleBlue
	DimRed
	DimBlue
)

var Colors = map[Color][3]byte{
	Red:        {255, 0, 0},
	Orange:     {255, 50, 0},
	Yellow:     {255, 255, 0},
	Green:      {0, 255, 0},
	Teal:       {0, 100, 100},
	Blue:       {0, 0, 255},
	Purple:     {100, 0, 100},
	White:      {255, 255, 255},
	Black:      {0, 0, 0},
	PurpleRed:  {200, 0, 50},
	PurpleBlue: {50, 0, 200},
	DimRed:     {50, 0, 0},
	DimBlue:    {0, 0, 50},
}
