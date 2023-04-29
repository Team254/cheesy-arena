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

// All rules from the 2022 game that carry point penalties.
var rules = []*Rule{
	{1, "G103", false, false, "BUMPERS must be in Bumper Zone (see R402) during the match."},
	{2, "G106", false, false, "ROBOT height, as measured when it’s resting normally on a flat floor, may not exceed 6 ft. 6 in. above the carpet during the MATCH."},
	{3, "G107", false, false, "ROBOTS may not extend beyond their FRAME PERIMETER in more than 48 in."},
	{4, "G107", true, false, "ROBOTS may not extend beyond their FRAME PERIMETER in more than 48 in. TECH FOUL if the over-extension scores a GAME PIECE."},
	{5, "G108", false, false, "A ROBOT whose BUMPERS are intersecting the opponent’s LOADING ZONE or COMMUNITY may not extend beyond its FRAME PERIMETER."},
	{6, "G108", true, false, "A ROBOT whose BUMPERS are intersecting the opponent’s LOADING ZONE or COMMUNITY may not extend beyond its FRAME PERIMETER. TECH FOUL if REPEATED."},
	{7, "G109", false, false, "ROBOTS may not extend beyond their FRAME PERIMETER in more than one direction (i.e. over 1 side of the ROBOT) at a time."},
	{8, "G109", true, false, "ROBOTS may not extend beyond their FRAME PERIMETER in more than one direction (i.e. over 1 side of the ROBOT) at a time. TECH FOUL if extending in multiple directions scores a GAME PIECE."},
	{9, "G201", false, false, "Strategies clearly aimed at forcing the opponent ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{10, "G201", true, false, "Strategies clearly aimed at forcing the opponent ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed. If REPEATED, TECH FOUL."},
	{11, "G202", false, false, "ROBOTS may not PIN an opponent’s ROBOT for more than 5 seconds."},
	{12, "G202", true, false, "ROBOTS may not PIN an opponent’s ROBOT for more than 5 seconds. An additional TECH FOUL for every 5 seconds in which the situation is not corrected."},
	{13, "G203", true, false, "2 or more ROBOTS that appear to a REFEREE to be working together may neither isolate nor close off any major component of MATCH play."},
	{14, "G204", false, false, "A ROBOT may not use a COMPONENT outside its FRAME PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of that opponent ROBOT’S FRAME PERIMETER."},
	{15, "G205", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately, as perceived by a REFEREE. B. regardless of intent, by initiating contact, either directly or transitively via a GAME PIECE CONTROLLED by the ROBOT, inside the vertical projection of an opponent ROBOT’S FRAME PERIMETER."},
	{16, "G206", true, false, "A ROBOT may not deliberately, as perceived by a REFEREE, attach to, tip, or entangle with an opponent ROBOT."},
	{17, "G207", false, false, "A ROBOT with any part of itself in their opponent’s LOADING ZONE or COMMUNITY may not contact an opponent ROBOT, regardless of who initiates contact."},
	{18, "G208", true, false, "A ROBOT may not be fully supported by a partner ROBOT unless the partner’s BUMPERS intersect its COMMUNITY."},
	{19, "G301", true, false, "ROBOTS and OPERATOR CONSOLES are prohibited from the following actions with regards to interaction with ARENA elements: grabbing, grasping, attaching to, deforming, becoming entangled with, suspending from, and damaging."},
	{20, "G302", false, false, "Before TELEOP, a ROBOT may not intersect the infinite vertical volume created by the opponent’s ALLIANCE WALL, the ROBOT’S DOUBLE SUBSTATION, guardrails, and CENTER LINE of the FIELD."},
	{21, "G302", true, false, "Before TELEOP, a ROBOT may not intersect the infinite vertical volume created by the opponent’s ALLIANCE WALL, the ROBOT’S DOUBLE SUBSTATION, guardrails, and CENTER LINE of the FIELD. If contact with an opponent ROBOT, TECH FOUL."},
	{22, "G303", true, false, "Before TELEOP, a ROBOT action may not cause GAME PIECES staged on the opposing side of the FIELD to move from their starting locations."},
	{23, "G304", false, false, "ROBOTS, either directly or transitively through a GAME PIECE, may not cause or prevent the movement of the opponent CHARGE STATION."},
	{24, "G401", false, false, "ROBOTS may not intentionally eject GAME PIECES from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT)."},
	{25, "G402", true, false, "ROBOTS may not deliberately use GAME PIECES in an attempt to ease or amplify the challenge associated with FIELD elements."},
	{26, "G403", false, false, "ROBOTS completely outside their LOADING ZONE or COMMUNITY may not have CONTROL of more than 1 GAME PIECE, either directly or transitively through other objects."},
	{27, "G404", true, false, "A ROBOT may not launch GAME PIECES unless any part of the ROBOT is in its own COMMUNITY."},
	{28, "G405", false, true, "A ROBOT may not move a scored GAME PIECE from an opponent's NODE."},
	{29, "H301", true, false, "DRIVE TEAMS may not cause significant delays to the start of their MATCH."},
	{30, "H401", false, false, "During AUTO, DRIVE TEAM members in ALLIANCE AREAS and HUMAN PLAYERS in their SUBSTATION AREAS may not contact anything in front of the STARTING LINES, unless for personal or equipment safety or granted permission by a Head REFEREE or FTA."},
	{31, "H403", false, false, "During AUTO, DRIVE TEAMS may not directly or indirectly interact with ROBOTS or OPERATOR CONSOLES unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop."},
	{32, "H502", false, false, "DRIVE TEAMS may not contact anything outside the area in which they started the MATCH (i.e. the ALLIANCE AREA, the SUBSTATION AREA, or the designated TECHNICIAN space). Exceptions are granted for a HUMAN PLAYER whose feet are partially outside the SUBSTATION AREA (but not in the opponent ALLIANCE AREA), in cases concerning safety, and for actions that are inadvertent, MOMENTARY, and inconsequential."},
	{33, "H503", false, false, "COACHES may not touch GAME PIECES, unless for safety purposes."},
	{34, "H504", false, false, "GAME PIECES may only be introduced to the FIELD A. by a HUMAN PLAYER, B. through a PORTAL, and C. during TELEOP."},
	{35, "H505", false, false, "DRIVE TEAMS may not extend any body part into the SINGLE SUBSTATION PORTAL for a greater-than-MOMENTARY period of time."},
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
