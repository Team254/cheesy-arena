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
	RuleNumber     string
	IsTechnical    bool
	IsRankingPoint bool
	Description    string
}

// All rules from the 2018 game that carry point penalties.
var Rules = []Rule{
	{"S6", false, false, "DRIVE TEAMS may not extend any body part into the CARGO Chute. Momentary encroachment into the Chute is an exception to this rule."},
	{"C8", false, false, "Strategies clearly aimed at forcing the opposing ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{"G3", true, false, "During the SANDSTORM PERIOD, a ROBOT may not cross the FIELD such that its BUMPERS break the plane defined by their opponent’s CARGO SHIP LINE."},
	{"G4", false, false, "ROBOTS may not have greater-than-momentary or repeated control, i.e. exercise greater-than-momentary or repeated influence, of more than one (1) GAME PIECE at a time, either directly or transitively through other objects."},
	{"G5", false, true, "A ROBOT may not remove a GAME PIECE from an opponents’ ROCKET/CARGO SHIP."},
	{"G7", false, false, "ROBOTS may not intentionally eject GAME PIECES from the FIELD."},
	{"G8", false, false, "ROBOTS may not deliberately use GAME PIECES in an attempt to ease or amplify the challenge associated with FIELD elements."},
	{"G9", false, false, "No more than one ROBOT may be positioned such that its BUMPERS are completely beyond the opponent’s CARGO SHIP LINE."},
	{"G9", true, false, "No more than one ROBOT may be positioned such that its BUMPERS are completely beyond the opponent’s CARGO SHIP LINE."},
	{"G10", false, false, "No part of a ROBOT, except its BUMPERS, may be outside its FRAME PERIMETER if its BUMPERS are completely beyond its opponent’s CARGO SHIP LINE."},
	{"G10", true, false, "No part of a ROBOT, except its BUMPERS, may be outside its FRAME PERIMETER if its BUMPERS are completely beyond its opponent’s CARGO SHIP LINE."},
	{"G12", false, false, "A ROBOT may not break the vertical plane above the ALLIANCE STATION WALL or damage the SANDSTORM."},
	{"G13", false, false, "A ROBOT may not contact an opponent ROBOT if that opponent ROBOT’S BUMPERS are fully in their HAB ZONE."},
	{"G15", false, false, "DRIVE TEAMS, ROBOTS, and OPERATOR CONSOLES are prohibited from the following actions with regards to interaction with ARENA elements: grabbing, grasping, attaching to, hanging, deforming, becoming entangled, damaging, and repositioning GAME PIECE holders."},
	{"G16", false, true, "During Qualification MATCHES, ROBOTS may not contact opponents’ ROCKETS starting at T-minus 20s."},
	{"G17", false, false, "Fallen (i.e. tipped over) ROBOTS attempting to right themselves (either by themselves or with assistance from a partner ROBOT) have one ten (10) second grace period in which they may not be contacted by an opponent ROBOT."},
	{"G18", false, false, "ROBOTS may not pin an opponent’s ROBOT for more than five (5) seconds."},
	{"G18", true, false, "ROBOTS may not pin an opponent’s ROBOT for more than five (5) seconds."},
	{"G19", true, false, "Strategies aimed at the destruction or inhibition of ROBOTS via attachment, damage, tipping, or entanglements are not allowed."},
	{"G20", true, false, "Initiating deliberate or damaging contact with an opponent ROBOT on or inside the vertical extension of its FRAME PERIMETER, including transitively through a GAME PIECE, is not allowed."},
	{"G23", false, false, "BUMPERS must be in the BUMPER ZONE during the MATCH unless a ROBOT is completely in its HAB ZONE or supported by a ROBOT completely in its HAB ZONE."},
	{"G24", false, false, "ROBOTS may not extend more than 30 in (~76 cm). beyond their FRAME PERIMETER."},
	{"H6", false, false, "During the MATCH, DRIVERS, COACHES, and HUMAN PLAYERS may not contact anything outside the ALLIANCE STATION and TECHNICIANS may not contact anything outside their designated area."},
	{"H7", false, false, "During the MATCH, team members may only enter GAME PIECES on to the FIELD through their LOADING STATIONS."},
	{"H8", false, false, "During a MATCH, COACHES may not touch GAME PIECES unless for safety purposes."},
	{"H9", true, false, "During the SANDSTORM PERIOD, COACHES, DRIVERS, HUMAN PLAYERS, and any part of the OPERATOR CONSOLE may not break the vertical planes defined by the STARTING LINES, unless for safety purposes."},
	{"H10", true, false, "During the SANDSTORM PERIOD, COACHES, DRIVERS, and HUMAN PLAYERS may not look over the top of the ALLIANCE WALL down to the FIELD to overcome the effect of the SANDSTORM."},
}

func (foul *Foul) PointValue() int {
	if foul.IsTechnical {
		return 10
	} else {
		return 3
	}
}
