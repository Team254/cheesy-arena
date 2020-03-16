// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a game-specific rule.

package game

type Rule struct {
	Id             int
	RuleNumber     string
	IsTechnical    bool
	IsRankingPoint bool
	Description    string
}

// All rules from the 2018 game that carry point penalties.
var rules = []*Rule{
	{1, "S6", false, false, "DRIVE TEAMS may not extend any body part into the CARGO Chute. Momentary encroachment into the Chute is an exception to this rule."},
	{2, "C8", false, false, "Strategies clearly aimed at forcing the opposing ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{3, "G3", true, false, "During the SANDSTORM PERIOD, a ROBOT may not cross the FIELD such that its BUMPERS break the plane defined by their opponent’s CARGO SHIP LINE."},
	{4, "G4", false, false, "ROBOTS may not have greater-than-momentary or repeated control, i.e. exercise greater-than-momentary or repeated influence, of more than one (1) GAME PIECE at a time, either directly or transitively through other objects."},
	{5, "G5", false, true, "A ROBOT may not remove a GAME PIECE from an opponents’ ROCKET/CARGO SHIP."},
	{6, "G7", false, false, "ROBOTS may not intentionally eject GAME PIECES from the FIELD."},
	{7, "G8", false, false, "ROBOTS may not deliberately use GAME PIECES in an attempt to ease or amplify the challenge associated with FIELD elements."},
	{8, "G9", false, false, "No more than one ROBOT may be positioned such that its BUMPERS are completely beyond the opponent’s CARGO SHIP LINE."},
	{9, "G9", true, false, "No more than one ROBOT may be positioned such that its BUMPERS are completely beyond the opponent’s CARGO SHIP LINE."},
	{10, "G10", false, false, "No part of a ROBOT, except its BUMPERS, may be outside its FRAME PERIMETER if its BUMPERS are completely beyond its opponent’s CARGO SHIP LINE."},
	{11, "G10", true, false, "No part of a ROBOT, except its BUMPERS, may be outside its FRAME PERIMETER if its BUMPERS are completely beyond its opponent’s CARGO SHIP LINE."},
	{12, "G12", false, false, "A ROBOT may not break the vertical plane above the ALLIANCE STATION WALL or damage the SANDSTORM."},
	{13, "G13", false, false, "A ROBOT may not contact an opponent ROBOT if that opponent ROBOT’S BUMPERS are fully in their HAB ZONE."},
	{14, "G15", false, false, "DRIVE TEAMS, ROBOTS, and OPERATOR CONSOLES are prohibited from the following actions with regards to interaction with ARENA elements: grabbing, grasping, attaching to, hanging, deforming, becoming entangled, damaging, and repositioning GAME PIECE holders."},
	{15, "G16", false, true, "During Qualification MATCHES, ROBOTS may not contact opponents’ ROCKETS starting at T-minus 20s."},
	{16, "G17", false, false, "Fallen (i.e. tipped over) ROBOTS attempting to right themselves (either by themselves or with assistance from a partner ROBOT) have one ten (10) second grace period in which they may not be contacted by an opponent ROBOT."},
	{17, "G18", false, false, "ROBOTS may not pin an opponent’s ROBOT for more than five (5) seconds."},
	{18, "G18", true, false, "ROBOTS may not pin an opponent’s ROBOT for more than five (5) seconds."},
	{19, "G19", true, false, "Strategies aimed at the destruction or inhibition of ROBOTS via attachment, damage, tipping, or entanglements are not allowed."},
	{20, "G20", true, false, "Initiating deliberate or damaging contact with an opponent ROBOT on or inside the vertical extension of its FRAME PERIMETER, including transitively through a GAME PIECE, is not allowed."},
	{21, "G23", false, false, "BUMPERS must be in the BUMPER ZONE during the MATCH unless a ROBOT is completely in its HAB ZONE or supported by a ROBOT completely in its HAB ZONE."},
	{22, "G24", false, false, "ROBOTS may not extend more than 30 in (~76 cm). beyond their FRAME PERIMETER."},
	{23, "H6", false, false, "During the MATCH, DRIVERS, COACHES, and HUMAN PLAYERS may not contact anything outside the ALLIANCE STATION and TECHNICIANS may not contact anything outside their designated area."},
	{24, "H7", false, false, "During the MATCH, team members may only enter GAME PIECES on to the FIELD through their LOADING STATIONS."},
	{25, "H8", false, false, "During a MATCH, COACHES may not touch GAME PIECES unless for safety purposes."},
	{26, "H9", true, false, "During the SANDSTORM PERIOD, COACHES, DRIVERS, HUMAN PLAYERS, and any part of the OPERATOR CONSOLE may not break the vertical planes defined by the STARTING LINES, unless for safety purposes."},
	{27, "H10", true, false, "During the SANDSTORM PERIOD, COACHES, DRIVERS, and HUMAN PLAYERS may not look over the top of the ALLIANCE WALL down to the FIELD to overcome the effect of the SANDSTORM."},
}
var ruleMap map[int]*Rule

// Returns the rule having the given ID, or nil if no such rule exists.
func GetRuleById(id int) *Rule {
	return GetAllRules()[id]
}

// Returns a slice of all defined rules that carry point penalties.
func GetAllRules() map[int]*Rule {
	if ruleMap == nil {
		ruleMap = make(map[int]*Rule, len(rules))
		for _, rule := range rules {
			ruleMap[rule.Id] = rule
		}
	}
	return ruleMap
}
