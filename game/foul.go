// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a foul.

package game

type Foul struct {
	RuleId         int
	TeamId         int
	TimeInMatchSec float64
}

// Returns the rule for which the foul was assigned.
func (foul *Foul) Rule() *Rule {
	return GetRuleById(foul.RuleId)
}

// Returns the number of points that the foul adds to the opposing alliance's score.
func (foul *Foul) PointValue() int {
	if foul.Rule() == nil || foul.Rule().IsRankingPoint {
		return 0
	}
	if foul.Rule().IsTechnical {
		return 8
	} else {
		return 4
	}
}
