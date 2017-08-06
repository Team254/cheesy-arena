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

// All rules from the 2017 game that carry point penalties.
var Rules = []Rule{{"S08", false}, {"C08", false}, {"C11", false}, {"G04", false}, {"G05", false},
	{"G08", false}, {"G09", false}, {"G11", false}, {"G11", true}, {"G12", false}, {"G13", true},
	{"G15", false}, {"G17", false}, {"G20", false}, {"G22", false}, {"G23", false}, {"G26", true},
	{"G27", false}, {"G27", true}, {"A01", false}, {"A02", false}, {"A04", false}, {"A04", true},
	{"A05", true}, {"H06", false}, {"H07", false}, {"H08", false}, {"H11", false}, {"H11", true},
	{"H12", true}, {"H13", false}}

func (foul *Foul) PointValue() int {
	if foul.IsTechnical {
		return 25
	}
	return 5
}
