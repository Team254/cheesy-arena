// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game Manual 
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

// All rules from the 2026 REBUILT Game Manual (Section 7.1 - 7.5)
// @formatter:off
var rules = []*Rule{
	// --- 7.1 Personal Safety ---
	{1, "G101", true, false, "A team member may not reach into the FIELD with any part of their body during a MATCH."},
	{2, "G102", true, false, "A team member may only enter or exit the FIELD through open gates and only enter if FIELD lighting (FIELD facing side of the team signs and timers) is green, unless explicitly instructed by a REFEREE or an FTA."},
	{3, "G103", true, false, "A team member is prohibited from the following actions with regards to interaction with ARENA elements: A. climbing on or inside, B. hanging from, C. manipulating such that it doesn't return to its original shape without human intervention, and D. damaging."},
	{4, "G104", true, false, "Teams may not enable their ROBOTS on the FIELD. Teams may not tether to the ROBOT while on the FIELD except in special circumstances (e.g. after Opening Ceremonies, before an immediate MATCH replay, etc.) and with the express permission from the FTA or a REFEREE."},

	// --- 7.2 Conduct ---
	{5, "G201", true, false, "All teams must be civil toward everyone and respectful of team and event equipment while at a FIRST Robotics Competition event."},
	{6, "G202", true, false, "A team member may never strike or hit the DRIVER STATION plastic windows."},
	{7, "G203", true, false, "A team may not encourage an ALLIANCE of which it is not a member to play beneath its ability."},
	{8, "G204", true, false, "A team, as the result of encouragement by a team not on their ALLIANCE, may not play beneath its ability."},
	{9, "G205", true, false, "A team may not intentionally lose a MATCH or sacrifice Ranking Points in an effort to lower their own ranking or manipulate the rankings of other teams."},
	{10, "G206", false, true, "A team or ALLIANCE may not collude with another team to each purposefully violate a rule in an attempt to influence Ranking Points."},
	{11, "G207", true, false, "A team member (except DRIVERS, HUMAN PLAYERS, and DRIVE COACHES) granted access to restricted areas in and around the ARENA (e.g. via TECHNICIAN button, event issued Media badges, etc.) may not assist or use signaling devices during the MATCH."},
	{12, "G208", true, false, "If a ROBOT has passed initial, complete inspection, at least 1 member of its DRIVE TEAM must report to the ARENA and participate in each of their assigned Qualification MATCHES."},
	{13, "G209", true, false, "A ROBOT may not intentionally detach or leave a part on the FIELD."},
	{14, "G210", true, false, "A strategy not consistent with standard gameplay and clearly aimed at forcing the opponent ALLIANCE to violate a rule is not in the spirit of FIRST Robotics Competition and not allowed."},
	{15, "G211", true, false, "Egregious behavior beyond what is listed in the rules or subsequent violations of any rule or procedure during the event is prohibited."},
	{16, "G212", true, false, "A team may not encourage another team to exclude their ROBOT or be BYPASSED from a qualification MATCH for any reason."},

	// --- 7.3 Pre-MATCH ---
	{17, "G301", true, false, "A DRIVE TEAM member may not cause significant delays to the start of their MATCH."},
	{18, "G302", true, false, "Items used during a match must fit on your team's DRIVER STATION shelf, be worn or held by members from your DRIVE TEAM, or be an item used as an accommodation (e.g. stools, crutches, etc.)."},
	{19, "G303", true, false, "A ROBOT must meet all following MATCH-start requirements: A. it does not pose a hazard to humans, FIELD elements, or other ROBOTS, B. has passed initial, complete inspection, C. if modified after initial Inspection, it's compliant with I104, D. its BUMPERS overlap their ROBOT STARTING LINE, E. it's not contacting the BUMP, F. it's the only team-provided item left on the FIELD, G. it's not attached to, entangled with, or suspended from any FIELD element, H. it's confined to its STARTING CONFIGURATION, and I. it fully and solely supports not more than 8 FUEL."},

	// --- 7.4 In-MATCH ---
	// 7.4.1 AUTO
	{20, "G401", false, false, "In AUTO, each DRIVE TEAM member must remain in their staged areas. A DRIVE TEAM member staged behind a HUMAN STARTING LINE may not contact anything in front of that HUMAN STARTING LINE, unless for personal or equipment safety, to press the E-Stop or A-Stop, or granted permission by a Head REFEREE or FTA."},
	{21, "G402", false, false, "In AUTO, a DRIVE TEAM member may not directly or indirectly interact with a ROBOT or an OPERATOR CONSOLE unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop or A-Stop."},
	{22, "G402", true, false, "In AUTO, a DRIVE TEAM member may not directly or indirectly interact with a ROBOT or an OPERATOR CONSOLE unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop or A-Stop. (Repeated/Strategic)"},
	{23, "G403", true, false, "In AUTO, a ROBOT whose BUMPERS are completely across the CENTER LINE (i.e. to the opposite side of the CENTER LINE from its ROBOT STARTING LINE) may not contact an opponent ROBOT."},

	// 7.4.2 SCORING ELEMENTS
	{24, "G404", true, false, "A ROBOT may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify a challenge associated with a FIELD element."},
	{25, "G405", false, false, "A ROBOT may not intentionally eject SCORING ELEMENTS from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT) with an exception of through the opening at the base of the OUTPOST."},
	{26, "G405", true, false, "A ROBOT may not intentionally eject SCORING ELEMENTS from the FIELD (either directly or by bouncing off a FIELD element or other ROBOT) with an exception of through the opening at the base of the OUTPOST. (Repeated)"},
	{27, "G406", true, false, "Neither a ROBOT nor a HUMAN PLAYER may damage a SCORING ELEMENT."},
	{28, "G407", true, false, "A ROBOT may not launch a SCORING ELEMENT into their HUB unless their BUMPERS are partially or fully within their ALLIANCE ZONE."},
	{29, "G408", false, false, "A ROBOT may not do either of the following with FUEL released by the HUB unless and until that FUEL contacts anything else besides that ROBOT or FUEL CONTROLLED by that ROBOT: A. gain greater than MOMENTARY CONTROL of FUEL, or B. push or redirect FUEL to a desired location or in a preferred direction."},
	{30, "G408", true, false, "A ROBOT may not do either of the following with FUEL released by the HUB unless and until that FUEL contacts anything else besides that ROBOT or FUEL CONTROLLED by that ROBOT: A. gain greater than MOMENTARY CONTROL of FUEL, or B. push or redirect FUEL to a desired location or in a preferred direction. (Strategic/Repeated)"},

	// 7.4.3 ROBOT
	{31, "G409", true, false, "A ROBOT may not pose an undue hazard to a human, an ARENA element, or another ROBOT in the following ways: A. anything it CONTROLS contacts anything outside the FIELD, B. BUMPERS completely detach, C. ROBOT PERIMETER exposed, D. team number/color indeterminate, E. BUMPERS leave BUMPER ZONE REPEATEDLY, or F. unsafe operation."},
	{32, "G410", false, false, "ROBOT extensions may not interact with the carpet, BUMPS, or TOWER BASE such that the BUMPERS are lifted out of the BUMPER ZONE (see R405)."},
	{33, "G411", true, false, "A ROBOT may not damage FIELD elements."},
	{34, "G412", true, false, "A ROBOT is prohibited from the following interactions with FIELD elements (with the exception of the RUNGS and UPRIGHTS): A. grabbing, B. grasping, C. attaching to, D. becoming entangled with, and E. suspending from."},
	{35, "G413", false, false, "A ROBOT may not extend beyond any of the horizontal or vertical expansion limits described in R105, R106, and R107."},
	{36, "G413", true, false, "A ROBOT may not extend beyond any of the horizontal or vertical expansion limits described in R105, R106, and R107. (Strategic/Impedes/Enables Scoring)"},
	{37, "G414", true, false, "ROBOTS may not fully support the weight of other ROBOTS on their ALLIANCE to climb the TOWER. (Supported ROBOTS ineligible for points)."},

	// 7.4.4 Opponent Interaction
	{38, "G415", false, false, "A ROBOT may not use a COMPONENT outside its ROBOT PERIMETER (except its BUMPERS) to initiate contact with an opponent ROBOT inside the vertical projection of the opponent's ROBOT PERIMETER."},
	{39, "G416", true, false, "A ROBOT may not damage or functionally impair an opponent ROBOT in either of the following ways: A. deliberately. B. regardless of intent, by initiating contact, either directly or transitively via a SCORING ELEMENT CONTROLLED by the ROBOT, inside the vertical projection of an opponent's ROBOT PERIMETER."},
	{40, "G417", true, false, "A ROBOT may not deliberately, attach to, tip, or entangle with an opponent ROBOT."},
	{41, "G418", false, false, "A ROBOT may not PIN an opponent's ROBOT for more than 3 seconds."},
	{42, "G418", true, false, "A ROBOT may not PIN an opponent's ROBOT for more than 3 seconds. (Extended/Repeated)"},
	{43, "G419", true, false, "2 or more ROBOTS that appear to a REFEREE to be working together may not isolate or close off any major element of MATCH play."},
	{44, "G420", true, true, "A ROBOT may not contact, directly or transitively through a SCORING ELEMENT, an opponent ROBOT in contact with an opponent TOWER during the last 30 seconds of the MATCH regardless of who initiates contact. (Awards Level 3 Climb)."},

	// 7.4.5 Human
	{45, "G421", false, false, "A DRIVE TEAM member must remain in their designated area as follows: A. DRIVERS and DRIVE COACHES may not contact anything outside their ALLIANCE AREA, B. a DRIVER must use the OPERATOR CONSOLE in the DRIVER STATION to which they are assigned, C. a HUMAN PLAYER may not contact anything outside their ALLIANCE AREA, and D. a TECHNICIAN may not contact anything outside their designated area."},
	{46, "G422", true, false, "A ROBOT shall be operated only by the DRIVERS and/or HUMAN PLAYERS of that team. A DRIVE COACH activating their E-Stop or A-Stop is the exception to this rule."},
	{47, "G423", false, false, "A DRIVE TEAM member may not extend: A. into the CHUTE beyond the ALLIANCE-colored tape line while the CHUTE DOOR is open, or B. into the CORRAL beyond the ALLIANCE-colored tape line."},
	{48, "G424", true, false, "A DRIVE TEAM member may not deliberately use a SCORING ELEMENT in an attempt to ease or amplify a challenge associated with a FIELD element."},
	{49, "G425", true, false, "FUEL may only be introduced to the FIELD by a HUMAN PLAYER or DRIVER in the following ways: A. through the CHUTE, B. through the bottom opening in the OUTPOST, or C. thrown from the OUTPOST AREA."},
	{50, "G426", false, false, "DRIVE COACHES may not touch SCORING ELEMENTS, unless for safety purposes."},
	{51, "G427", false, false, "Off-FIELD FUEL may only be stored in the CHUTE and the CORRAL. Excess FUEL, defined as the CHUTE & CORRAL being full, must immediately be entered onto the FIELD."},
	{52, "G427", true, false, "Off-FIELD FUEL may only be stored in the CHUTE and the CORRAL. Excess FUEL, defined as the CHUTE & CORRAL being full, must immediately be entered onto the FIELD. (Continuous)"},

	// --- 7.5 Post-MATCH ---
	{53, "G501", true, false, "A DRIVE TEAM member may not cause significant or multiple delays to the start of a subsequent MATCH, scheduled break content, or other FIELD activities."},
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