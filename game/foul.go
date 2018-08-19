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
	RuleNumber  string
	IsTechnical bool
	Description string
}

// All rules from the 2018 game that carry point penalties.
var Rules = []Rule{
	{"S06", false, "DRIVE TEAMS may not extend any body part into the RETURN chute, the PORTAL chute, or the EXCHANGE tunnel."},
	{"C07", false, "Strategies clearly aimed at forcing the opposing ALLIANCE to violate a rule are not in the spirit of FIRST® Robotics Competition and not allowed."},
	{"C07", true, "Strategies clearly aimed at forcing the opposing ALLIANCE to violate a rule are not in the spirit of FIRST® Robotics Competition and not allowed."},
	{"G05", false, "ROBOTS may not extend more than 16 in (41 cm). beyond their FRAME PERIMETER."},
	{"G07", false, "ROBOTS must be in compliance with BUMPER rules throughout the MATCH."},
	{"G10", false, "Strategies aimed at the destruction or inhibition of ROBOTS via attachment, damage, tipping, or entanglements are not allowed."},
	{"G11", false, "Initiating deliberate or damaging contact with an opponent ROBOT on or inside the vertical extension of its FRAME PERIMETER, including transitively through a POWER CUBE, is not allowed."},
	{"G13", false, "Fallen (i.e. tipped over) ROBOTS attempting to right themselves (either by themselves or with assistance from a partner ROBOT) have one ten (10) second grace period in which they may not be contacted by an opponent ROBOT."},
	{"G14", false, "ROBOTS may not pin an opponent’s ROBOT for more than five (5) seconds."},
	{"G15", false, "A ROBOT may not block their opponent’s EXCHANGE ZONE for more than five (5) seconds."},
	{"G16", true, "A ROBOT whose BUMPERS are breaking the plane of or completely contained by its NULL TERRITORY and not breaking the plane of the opponent’s PLATFORM ZONE may not be contacted by an opposing ROBOT either directly or transitively through a POWER CUBE, regardless of who initiates the contact."},
	{"G17", true, "Unless during the ENDGAME, or attempting to right a fallen (i.e. tipped over) ALLIANCE partner, ROBOTS may neither fully nor partially strategically support the weight of partner ROBOTS."},
	{"G19", false, "DRIVE TEAMS, ROBOTS, and OPERATOR CONSOLES are prohibited from the following actions with regards to interaction with ARCADE elements: grabbing, grasping, attaching to, hanging, deforming, becoming entangled, and damaging."},
	{"G20", true, "With the exception of placing a POWER CUBES on PLATES, ROBOTS may not deliberately use POWER CUBES in an attempt to ease or amplify the challenge associated with FIELD elements."},
	{"G21", false, "With the exception of feeding POWER CUBES through the lower opening of the EXCHANGE, ROBOTS may not intentionally eject POWER CUBES from the FIELD."},
	{"G22", false, "ROBOTS may not control more than one (1) POWER CUBE at a time, except when breaking the plane of their own EXCHANGE ZONE."},
	{"G23", false, "ROBOTS may not remove POWER CUBES, or cause POWER CUBES to be removed, from the opponent’s POWER CUBE ZONE."},
	{"G24", true, "Strategies aimed at removing POWER CUBES from PLATES are prohibited."},
	{"G25", false, "Except via the weight of placed POWER CUBES, ROBOTS may not directly or transitively cause or prevent the movement of PLATES to their ALLIANCE's advantage."},
	{"G25", true, "Except via the weight of placed POWER CUBES, ROBOTS may not directly or transitively cause or prevent the movement of PLATES to their ALLIANCE's advantage."},
	{"A01", false, "During AUTO, DRIVE TEAM members in ALLIANCE STATIONS and PORTALS may not contact anything in front of the STARTING LINES, unless for personal or equipment safety."},
	{"A02", false, "During AUTO, DRIVE TEAMS may not directly or indirectly interact with ROBOTS or OPERATOR CONSOLES unless for personal safety, OPERATOR CONSOLE safety, or pressing an E-Stop for ROBOT safety."},
	{"A03", false, "During AUTO, any control devices worn or held by the DRIVERS and/or HUMAN PLAYERS must be disconnected from the OPERATOR CONSOLE."},
	{"A04", false, "During AUTO, no part of a ROBOT’S BUMPERS may pass from the NULL TERRITORY to the opponent’s side of the FIELD."},
	{"A04", true, "During AUTO, no part of a ROBOT’S BUMPERS may pass from the NULL TERRITORY to the opponent’s side of the FIELD."},
	{"A05", false, "During AUTO, DRIVE TEAMS may not contact any POWER CUBES, unless for personal safety."},
	{"H06", false, "DRIVE TEAM members may not contact anything outside the zone in which they started the MATCH (e.g. the ALLIANCE STATION, PORTAL, designated area for the TECHNICIAN) during the MATCH."},
	{"H11", true, "During a MATCH, COACHES may not touch POWER CUBES unless for safety purposes."},
	{"H12", true, "During a MATCH, COACHES may not touch any component of the VAULT (including the buttons) unless for safety purposes."},
	{"H13", false, "DRIVE TEAMS may only deliberately cause POWER CUBES to leave an ALLIANCE STATION or PORTAL during TELEOP, by a HUMAN PLAYER or DRIVER, and through a PORTAL wall or the RETURN."},
	{"H14", false, "POWER CUBES may not be removed from the VAULT."},
}

func (foul *Foul) PointValue() int {
	if foul.IsTechnical {
		return 25
	}
	return 5
}
