// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model of a game-specific rule.

package game

type Rule struct {
	Id             int
	RuleNumber     string
	IsMajor        bool
	IsRankingPoint bool
	Description    string
}

// All rules from the 2022 game that carry point penalties.
// @formatter:off
var rules = []*Rule{
	{1, "G206", false, true, "A team or ALLIANCE may not collude with another team to each purposefully violate a rule in an attempt to influence Ranking Points."},
	{2, "G210", true, false, "A strategy not consistent with standard gameplay and clearly aimed at forcing the opponent ALLIANCE to violate a rule is not in the spirit of FIRST Robotics Competition and not allowed."},
	{3, "G301", true, false, "A DRIVE TEAM member may not cause significant delays to the start of their MATCH."},
	{4, "G401", false, false, "In AUTO, each DRIVE TEAM member must remain in their staged areas. A DRIVE TEAM member staged behind a HUMAN STARTING LINE may not contact anything in front of that HUMAN STARTING LINE, unless for personal or equipment safety, to press the E-Stop or A-Stop, or granted permission by a Head REFEREE or FTA."},
	{5, "G402", false, false, "In AUTO, a DRIVE TEAM member may not directly or indirectly interact with a ROBOT or an OPERATOR CONSOLE unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop or A-Stop."},
	{6, "G403", true, false, "In AUTO, a ROBOT whose BUMPERS are completely across the BARGE ZONE (i.e. to the opposite side of the BARGE ZONE from its ROBOT STARTING LINE) may not contact an opponent ROBOT (either directly or transitively through a SCORING ELEMENT CONTROLLED by either ROBOT and regardless of who initiates contact)."},
	{7, "G404", false, false, "In AUTO, a HUMAN PLAYER may not enter ALGAE onto the field."},
	{8, "G405", true, false, "In AUTO, a ROBOT may not contact an opposing ALLIANCE’s CAGE."},
	{9, "G406", true, false, "A ROBOT may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify the challenge associated with a FIELD element."},
	{10, "G407", false, false, "A ROBOT may not intentionally eject a SCORING ELEMENT from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT) other than ALGAE through a PROCESSOR."},
	{11, "G407", true, false, "A ROBOT may not intentionally eject a SCORING ELEMENT from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT) other than ALGAE through a PROCESSOR."},
	{12, "G408", true, false, "Neither a ROBOT nor a HUMAN PLAYER may damage a SCORING ELEMENT."},
	{13, "G409", false, false, "A ROBOT may not simultaneously CONTROL more than 1 CORAL and 1 ALGAE either directly or transitively through other objects."},
	{14, "G410", true, true, "A ROBOT may not de-score a CORAL scored on the opponent’s REEF."},
	{15, "G411", true, false, "A ROBOT may not deliberately put ALGAE on their opponent’s REEF."},
	{16, "G412", true, false, "A ROBOT may not launch CORAL unless their BUMPERS are partially in their REEF ZONE."},
	{17, "G414", false, false, "BUMPERS must be in the BUMPER ZONE."},
	{18, "G415", false, false, "A ROBOT may not extend more than 1 ft. 6 in. beyond the vertical projection of its ROBOT PERIMETER."},
	{19, "G415", true, false, "A ROBOT may not extend more than 1 ft. 6 in. beyond the vertical projection of its ROBOT PERIMETER."},
	{20, "G417", true, false, "A ROBOT is prohibited from the following interactions with FIELD elements with the exception of CAGES: grabbing, grasping, attaching to, becoming entangled with, suspending from."},
	{21, "G418", true, true, "In TELEOP, a ROBOT may not contact an opponent’s CAGE."},
	{22, "G419", true, false, "A ROBOT may not contact the ANCHORS."},
	{23, "G420", true, false, "A ROBOT may not contact either NET or any ALGAE scored in an opponent NET."},
	{24, "G421", false, false, "No more than 1 ROBOT may be on the opponent’s side of the FIELD (i.e. containing the opponent REEF) with its BUMPERS fully outside and beyond the BARGE ZONES."},
	{25, "G421", true, false, "No more than 1 ROBOT may be on the opponent’s side of the FIELD (i.e. containing the opponent REEF) with its BUMPERS fully outside and beyond the BARGE ZONES."},
	{26, "G422", false, false, "A ROBOT may not use a COMPONENT outside its ROBOT PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of the opponent's ROBOT PERIMETER."},
	{27, "G423", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately. B. regardless of intent, by initiating contact, either directly or transitively via a SCORING ELEMENT CONTROLLED by the ROBOT, inside the vertical projection of an opponent's ROBOT PERIMETER."},
	{28, "G424", true, false, "A ROBOT may not deliberately attach to, tip, or entangle with an opponent ROBOT."},
	{29, "G425", false, false, "A ROBOT may not PIN an opponent’s ROBOT for more than 3 seconds."},
	{30, "G425", true, false, "A ROBOT may not PIN an opponent’s ROBOT for more than 3 seconds."},
	{31, "G426", true, false, "2 or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major element of MATCH play."},
	{32, "G427", true, false, "A ROBOT may not contact, directly or transitively through a SCORING ELEMENT, an opponent ROBOT partially or fully inside the opponent’s BARGE ZONE or REEF ZONE regardless of who initiates contact."},
	{33, "G428", true, true, "A ROBOT may not contact, directly or transitively through a SCORING ELEMENT, an opponent ROBOT in contact with an opponent CAGE during the last 20 seconds regardless of who initiates contact."},
	{34, "G429", false, false, "A DRIVE TEAM member must remain in their designated area as follows: A. DRIVERS and COACHES may not contact anything outside their ALLIANCE AREA, B. a DRIVER must use the OPERATOR CONSOLE in the DRIVER STATION to which they are assigned, as indicated on the team sign, C. a HUMAN PLAYER may not contact anything outside their ALLIANCE AREA or their PROCESSOR AREA, and D. a TECHNICIAN may not contact anything outside their designated area."},
	{35, "G430", true, false, "A ROBOT shall be operated only by the DRIVERS and/or HUMAN PLAYERS of that team. A COACH activating their E-Stop or A-Stop is the exception to this rule."},
	{36, "G431", false, false, "A DRIVE TEAM member may not extend into the CHUTE."},
	{37, "G432", true, false, "A DRIVE TEAM member may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify a challenge associated with a FIELD element."},
	{38, "G433", true, false, "SCORING ELEMENTS may only be entered onto the FIELD as follows: A. CORAL may only be introduced to the FIELD by a HUMAN PLAYER or DRIVER through the CORAL STATION and B. ALGAE may only be entered onto the FIELD by a HUMAN PLAYER in their PROCESSOR AREA."},
	{39, "G434", false, false, "COACHES may not touch SCORING ELEMENTS, unless for safety purposes."},
	{40, "G435", true, false, "HUMAN PLAYERS may not store more than 4 ALGAE in the PROCESSOR AREA."},
}

// @formatter:on
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
