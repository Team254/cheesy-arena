// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains pixel RGB mappings for common colors.

package led

type color int

const (
	red color = iota
	orange
	yellow
	green
	teal
	blue
	purple
	white
	black
	purpleRed
	purpleBlue
	dimRed
	dimBlue
)

var colors = map[color][3]byte{
	red:        {255, 0, 0},
	orange:     {255, 50, 0},
	yellow:     {255, 255, 0},
	green:      {0, 255, 0},
	teal:       {0, 100, 100},
	blue:       {0, 0, 255},
	purple:     {100, 0, 100},
	white:      {255, 255, 255},
	black:      {0, 0, 0},
	purpleRed:  {200, 0, 50},
	purpleBlue: {50, 0, 200},
	dimRed:     {50, 0, 0},
	dimBlue:    {0, 0, 50},
}
