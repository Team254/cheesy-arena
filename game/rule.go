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
	{2, "G106", false, false, "ROBOT height, as measured when it’s resting normally on a flat floor, may not exceed the maximum STARTING CONFIGURATION height (4 ft. 4 in.) unless any part of the ROBOT’S BUMPERS is in its HANGAR ZONE, in which case its height may not exceed 5 ft. 6 in."},
	{3, "G106", true, false, "ROBOT height, as measured when it’s resting normally on a flat floor, may not exceed the maximum STARTING CONFIGURATION height (4 ft. 4 in.) unless any part of the ROBOT’S BUMPERS is in its HANGAR ZONE, in which case its height may not exceed 5 ft. 6 in. TECH FOUL if the over-extension blocks an opponent’s shot, scores a CARGO, or is\nthe first thing that contacts CARGO exiting from an UPPER EXIT"},
	{4, "G107", false, false, "ROBOTS may not extend for more than 16 in. beyond their FRAME PERIMETER."},
	{5, "G107", true, false, "ROBOTS may not extend for more than 16 in. beyond their FRAME PERIMETER. TECH FOUL if the over-extension blocks an opponent’s shot, scores a CARGO, or is the first thing that contacts CARGO exiting from an UPPER EXIT."},
	{6, "G201", false, false, "Strategies clearly aimed at forcing the opponent ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{7, "G201", true, false, "Strategies clearly aimed at forcing the opponent ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed. If REPEATED, TECH FOUL."},
	{8, "G202", false, false, "ROBOTS may not PIN an opponent’s ROBOT for more than five seconds."},
	{9, "G202", true, false, "ROBOTS may not PIN an opponent’s ROBOT for more than five seconds. An additional TECH FOUL for every 5 seconds in which the situation is not corrected."},
	{10, "G203", true, false, "Two or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major component of MATCH play."},
	{11, "G204", false, false, "A ROBOT may not use a COMPONENT outside its FRAME PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of that opponent ROBOT’S FRAME PERIMETER."},
	{12, "G205", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately, as perceived by a REFEREE. B. regardless of intent, by initiating contact inside the vertical projection of an opponent ROBOT’S FRAME PERIMETER."},
	{13, "G206", true, false, "A ROBOT may not deliberately, as perceived by a REFEREE, attach to, tip, or entangle with an opponent ROBOT."},
	{14, "G207", false, false, "A ROBOT may not contact (either directly or transitively through CARGO and regardless of who initiates contact) an opponent ROBOT whose BUMPERS are contacting their LAUNCH PAD."},
	{15, "G209", true, false, "A ROBOT may not be fully supported by a partner ROBOT."},
	{16, "G210", true, false, "During AUTO, a ROBOT with any part of its BUMPERS on the opposite side of the FIELD (i.e. on the other side of the CENTER LINE from its ALLIANCE'S TARMACS) may contact neither CARGO still in its staged location on the opposite side of the FIELD nor an opponent ROBOT."},
	{17, "G301", true, false, "ROBOTS and OPERATOR CONSOLES are prohibited from the following actions with regards to interaction with ARENA elements: grabbing, grasping, attaching to, becoming entangled with, suspending from, and damaging."},
	{18, "G302", true, false, "ROBOTS may not reach into or straddle the LOWER EXIT. MOMENTARY reaching into and/or MOMENTARY straddling of the LOWER EXIT are exceptions to this rule."},
	{19, "G401", false, false, "ROBOTS may not eject opponent CARGO from the FIELD other than through the TERMINAL (either directly or by bouncing off a FIELD element or other ROBOT)."},
	{20, "G402", true, false, "ROBOTS may neither deliberately use CARGO in an attempt to ease or amplify the challenge associated with FIELD elements nor deliberately strand opponent CARGO on top of a HANGAR or HUB."},
	{21, "G403", false, false, "ROBOTS may not have greater-than-MOMENTARY CONTROL of more than 2 CARGO at a time, either directly or transitively through other objects."},
	{22, "G404", false, false, "A ROBOT may not restrict access to more than 3 opposing ALLIANCE CARGO except during the final 30 seconds of the MATCH."},
	{23, "G404", true, false, "A ROBOT may not restrict access to more than 3 opposing ALLIANCE CARGO except during the final 30 seconds of the MATCH. An additional TECH FOUL for every 5 seconds in which the situation is not corrected"},
	{24, "G405", false, false, "A ROBOT may not REPEATEDLY score or gain greater-than-MOMENTARY CONTROL of CARGO released by an UPPER EXIT until and unless that CARGO contacts anything else besides that ROBOT or CARGO controlled by that ROBOT."},
	{25, "H312", false, false, "After the end of the MATCH (i.e. when the timer displays 0 seconds following TELEOP), DRIVE TEAMS may not enter CARGO into the FIELD."},
	{26, "H401", false, false, "During AUTO, DRIVE TEAM members staged behind a STARTING LINE or TERMINAL STARTING LINE may not contact anything in front of those lines, unless for personal or equipment safety or granted permission by a Head REFEREE or FTA."},
	{27, "H403", false, false, "During AUTO, DRIVE TEAMS may not directly or indirectly interact with ROBOTS or OPERATOR CONSOLES unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop. A HUMAN PLAYER feeding CARGO to their ROBOT is an exception to this rule."},
	{28, "H404", false, false, "During AUTO, CARGO may only be introduced to the FIELD by a HUMAN PLAYER in a TERMINAL AREA."},
	{29, "H502", false, false, "DRIVERS, COACHES, and HUMAN PLAYERS may not contact anything outside the area in which they started the MATCH (i.e. the ALLIANCE AREA or the TERMINAL AREA). TECHNICIANS may not contact anything outside their designated area. Exceptions are granted in cases concerning safety and for actions that are inadvertent, MOMENTARY, and inconsequential."},
	{30, "H503", false, false, "COACHES may not touch CARGO, unless for safety purposes."},
	{31, "H504", false, false, "During TELEOP, CARGO may only be introduced to the FIELD A. by a HUMAN PLAYER and B. through the GUARD."},
	{32, "H505", false, false, "During a MATCH, HUMAN PLAYERS may not contact opponent CARGO. Inconsequential and MOMENTARY contact, and/or contact that, as perceived by a REFEREE, is intended to be helpful, are exceptions to this rule."},
	{33, "H506", false, false, "HUMAN PLAYERS may not deliver their CARGO to opponent ROBOTS."},
	{34, "H507", false, false, "HUMAN PLAYERS may not reach beyond the PURPLE PLANE."},
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
