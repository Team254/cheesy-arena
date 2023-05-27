// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a foul.

package game

type Foul struct {
	IsTechnical bool
	TeamId      int
	RuleId      int
}

// Returns the rule for which the foul was assigned.
func (foul *Foul) Rule() *Rule {
	return GetRuleById(foul.RuleId)
}

// Returns the number of points that the foul adds to the opposing alliance's score.
func (foul *Foul) PointValue() int {
	if foul.IsTechnical {
		return 12
	} else {
		return 5
	}
}
