// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a foul and game-specific rules.

package game

type Foul struct {
	Rule
	TeamId         int
	TimeInMatchSec float64
}

type Rule struct {
	RuleNumber  string
	IsTechnical bool
}

// All rules from the 2018 game that carry point penalties.
var Rules = []Rule{{"S06", false}, {"C07", false}, {"C07", true}, {"G05", false}, {"G07", false},
	{"G09", true}, {"G10", false}, {"G11", false}, {"G13", false}, {"G14", false}, {"G15", false},
	{"G16", true}, {"G17", true}, {"G19", false}, {"G20", true}, {"G21", false}, {"G22", false},
	{"G23", false}, {"G24", true}, {"G25", false}, {"G25", true}, {"A01", false}, {"A02", false},
	{"A03", false}, {"A04", false}, {"A04", true}, {"A05", false}, {"H06", false}, {"H11", true},
	{"H12", true}, {"H13", false}, {"H14", false}}

func (foul *Foul) PointValue() int {
	if foul.IsTechnical {
		return 25
	}
	return 5
}
