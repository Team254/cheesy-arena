// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a foul.

package game

type Foul struct {
	IsMajor bool
	TeamId  int
	RuleId  int
}

// Returns the rule for which the foul was assigned.
func (foul *Foul) Rule() *Rule {
	return GetRuleById(foul.RuleId)
}

// Returns the number of points that the foul adds to the opposing alliance's score.
func (foul *Foul) PointValue() int {
	if foul.IsMajor {
		return 6
	} else {
		if foul.Rule() != nil && foul.Rule().RuleNumber == "G206" {
			// Special case in 2025 for G206, which is not actually a foul but does make the alliance ineligible for
			// some bonus RPs.
			return 0
		}
		return 2
	}
}
