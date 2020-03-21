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

// All rules from the 2020 game that carry point penalties.
var rules = []*Rule{
	{1, "S6", false, false, "DRIVE TEAMS may not extend any body part into the LOADING BAY Chute."},
	{2, "C8", false, false, "Strategies clearly aimed at forcing the opposing ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{3, "C8", true, false, "Strategies clearly aimed at forcing the opposing ALLIANCE to violate a rule are not in the spirit of FIRST Robotics Competition and not allowed."},
	{4, "G3", false, false, "During AUTO, a ROBOT’s BUMPERS may not break the plane of their ALLIANCE’s SECTOR."},
	{5, "G3", true, false, "During AUTO, a ROBOT’s BUMPERS may not break the plane of their ALLIANCE’s SECTOR."},
	{6, "G4", false, false, "During AUTO, DRIVE TEAM members in ALLIANCE STATIONS may not contact anything in front of the STARTING LINES, unless for personal or equipment safety."},
	{7, "G5", false, false, "During AUTO, DRIVE TEAMS may not directly or indirectly interact with ROBOTS or OPERATOR CONSOLES unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop."},
	{8, "G6", false, false, "ROBOTS may not have greater-than-momentary CONTROL of more than five (5) POWER CELLS at a time, either directly or transitively through other objects."},
	{9, "G7", false, false, "ROBOTS may not intentionally eject POWER CELLS from the FIELD other than through the POWER PORT."},
	{10, "G8", true, false, "ROBOTS may not deliberately use POWER CELLS in an attempt to ease or amplify the challenge associated with FIELD elements."},
	{11, "G9", true, false, "A ROBOT whose BUMPERS are fully contained by their SECTOR may not cause POWER CELLS to travel into or through their opponent’s SECTOR."},
	{12, "G10", true, false, "A ROBOT whose BUMPERS are intersecting the opponent’s TARGET ZONE, TRENCH RUN, or LOADING ZONE may not contact opponent ROBOTS, regardless of who initiates contact."},
	{13, "G11", true, false, "An opponent ROBOT may not contact a ROBOT whose BUMPERS are intersecting its TARGET ZONE or LOADING ZONE, regardless of who initiates contact."},
	{14, "G12", false, true, "A ROBOT may not contact the opponent’s CONTROL PANEL, either directly, or transitively through a POWER CELL, if A. the opponent ROBOT is contacting that CONTROL PANEL, and B. the opponent’s POWER PORT has reached CAPACITY."},
	{15, "G12", true, false, "A ROBOT may not contact the opponent’s CONTROL PANEL, either directly, or transitively through a POWER CELL, if A. the opponent ROBOT is contacting that CONTROL PANEL, and B. the opponent’s POWER PORT has reached CAPACITY."},
	{16, "G13", true, false, "A ROBOT may not be fully supported by a partner ROBOT unless the partner ROBOT’S BUMPERS are intersecting its RENDEZVOUS POINT."},
	{17, "G14", true, false, "During the ENDGAME, a ROBOT may not contact, either directly or transitively through a POWER CELL, an opponent ROBOT whose BUMPERS are completely contained in its RENDEZVOUS POINT and not in contact with its GENERATOR SWITCH."},
	{18, "G16", false, false, "BUMPERS must be in the BUMPER ZONE during the MATCH, unless during the ENDGAME and A. a ROBOT’s BUMPERS are intersecting its RENDEZVOUS POINT or B. a ROBOT is supported by a partner ROBOT whose BUMPERS are intersecting its RENDEZVOUS POINT."},
	{19, "G17", true, false, "ROBOT height, as measured when it’s resting normally on a flat floor, may not exceed 45 in. (~114 cm) above the carpet during the MATCH, with the exception of ROBOTS intersecting their ALLIANCE’S RENDEZVOUS POINT during the ENDGAME."},
	{20, "G18", false, false, "ROBOTS may not extend more than 12 inches (~30 cm) beyond their FRAME PERIMETER."},
	{21, "G21", false, false, "ROBOTS may not PIN an opponent’s ROBOT for more than five (5) seconds."},
	{22, "G21", true, false, "ROBOTS may not PIN an opponent’s ROBOT for more than five (5) seconds."},
	{23, "G22", true, false, "Two or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major component of MATCH play."},
	{24, "G23", true, false, "ROBOT actions that appear to be deliberate to a REFEREE and that cause damage or inhibition via attaching, tipping, or entangling to an opponent ROBOT are not allowed."},
	{25, "G24", false, false, "A ROBOT with a COMPONENT(S) outside its FRAME PERIMETER, other than BUMPERS, may not initiate direct contact with an opponent ROBOT inside the vertical projection of its FRAME PERIMETER using that COMPONENT."},
	{26, "G25", true, false, "Regardless of intent, a ROBOT may not initiate direct contact inside the vertical projection of an opponent ROBOT’S FRAME PERIMETER that damages or functionally impairs the opponent ROBOT."},
	{27, "G26", true, false, "ROBOTS and OPERATOR CONSOLES are prohibited from the following actions with regards to interaction with ARENA elements: grabbing, grasping, attaching, deforming, becoming entangled, damaging, suspending from."},
	{28, "H5", false, false, "During the MATCH, DRIVERS, COACHES, and HUMAN PLAYERS may not contact anything outside the ALLIANCE STATION and TECHNICIANS may not contact anything outside their designated area."},
	{29, "H6", false, false, "POWER CELLS may only be introduced to the FIELD A. during TELEOP, B. by a DRIVER or HUMAN PLAYER, and C. through the LOADING BAY."},
	{30, "H7", false, false, "During a MATCH, COACHES may not touch POWER CELLS, unless for safety purposes."},
	{31, "H9", false, false, "During TELEOP, an ALLIANCE may not have more than fifteen (15) POWER CELLS in their ALLIANCE STATION."},
	{32, "H10", false, false, "POWER CELLS must be stored on the LOADING BAY racks."},
	{33, "H10", true, false, "POWER CELLS must be stored on the LOADING BAY racks."},
	{34, "H16", true, false, "DRIVE TEAMS are prohibited from the following actions with regards to interaction with ARENA elements: climbing, hanging, deforming, damaging."},
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
