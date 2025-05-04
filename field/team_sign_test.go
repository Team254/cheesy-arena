// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"image/color"
	"testing"
)

func TestTeamSign_GenerateInMatchRearText(t *testing.T) {
	arena := setupTestArena(t)
	arena.RedRealtimeScore.CurrentScore = *game.TestScore1()
	arena.BlueRealtimeScore.CurrentScore = *game.TestScore2()

	assert.Equal(t, "01:23 R080-B162 1/4", generateInMatchTeamRearText(arena, true, "01:23"))
	assert.Equal(t, "01:23 B162-R080 1/4", generateInMatchTeamRearText(arena, false, "01:23"))
	assert.Equal(t, "1-07 2-02 3-03 4-00", generateInMatchTimerRearText(arena, true))
	assert.Equal(t, "1-15 2-03 3-05 4-03", generateInMatchTimerRearText(arena, false))
	arena.BlueRealtimeScore.CurrentScore.Reef.Branches[2] = [12]bool{true, true, true, true, true, true, true, true}
	arena.BlueRealtimeScore.CurrentScore.ProcessorAlgae = 2
	assert.Equal(t, "00:59 R080-B195 1/3", generateInMatchTeamRearText(arena, true, "00:59"))
	assert.Equal(t, "00:59 B195-R080 2/3", generateInMatchTeamRearText(arena, false, "00:59"))
	assert.Equal(t, "1-07 2-02 3-03 4-00", generateInMatchTimerRearText(arena, true))
	assert.Equal(t, "1-15 2-03 3-05 4-08", generateInMatchTimerRearText(arena, false))

	// Check that RP progress is hidden for playoff matches.
	arena.CurrentMatch.Type = model.Playoff
	assert.Equal(t, "00:45 R080-B195 ", generateInMatchTeamRearText(arena, true, "00:45"))
	assert.Equal(t, "00:45 B195-R080 ", generateInMatchTeamRearText(arena, false, "00:45"))
}

func TestTeamSign_Timer(t *testing.T) {
	arena := setupTestArena(t)
	sign := TeamSign{isTimer: true}

	// Should do nothing if no address is set.
	sign.update(arena, nil, true, "12:34", "Rear Text")
	assert.Equal(t, [128]byte{}, sign.packetData)

	// Check some basics about the data but don't unit-test the whole packet.
	sign.SetId(56)
	sign.update(arena, nil, true, "12:34", "Rear Text")
	assert.Equal(t, "CYPRX", string(sign.packetData[0:5]))
	assert.Equal(t, 56, int(sign.packetData[5]))
	assert.Equal(t, 0x04, int(sign.packetData[6]))
	assert.Equal(t, "12:34", string(sign.packetData[10:15]))
	assert.Equal(t, []byte{0, 0}, sign.packetData[15:17])
	assert.Equal(t, "Rear Text", string(sign.packetData[30:39]))
	assert.Equal(t, 40, sign.packetIndex)

	assertSign := func(expectedFrontText string, expectedFrontColor color.RGBA, expectedRearText string) {
		frontText, frontColor, rearText := generateTimerTexts(arena, "23:45", "Rear Text")
		assert.Equal(t, expectedFrontText, frontText)
		assert.Equal(t, expectedFrontColor, frontColor)
		assert.Equal(t, expectedRearText, rearText)
	}

	// Check field reset.
	arena.FieldReset = false
	assertSign("23:45", whiteColor, "Rear Text")
	arena.FieldReset = true
	assertSign("SAFE ", greenColor, "Rear Text")

	// Check timeout mode.
	arena.FieldReset = true
	arena.MatchState = TimeoutActive
	assertSign("23:45", whiteColor, "Field Break: 23:45")

	// Check blank mode.
	arena.AllianceStationDisplayMode = "blank"
	assertSign("     ", whiteColor, "")

	// Check alliance selection.
	arena.AllianceStationDisplayMode = "logo"
	arena.AudienceDisplayMode = "allianceSelection"
	arena.AllianceSelectionShowTimer = false
	assertSign("     ", whiteColor, "")
	arena.AllianceSelectionShowTimer = true
	assertSign("23:45", whiteColor, "")
	arena.AllianceStationDisplayMode = "blank"
	assertSign("     ", whiteColor, "")
}

func TestTeamSign_TeamNumber(t *testing.T) {
	arena := setupTestArena(t)
	allianceStation := arena.AllianceStations["R1"]
	arena.Database.CreateTeam(&model.Team{Id: 254})
	sign := &TeamSign{isTimer: false}

	// Should do nothing if no address is set.
	sign.update(arena, allianceStation, true, "12:34", "Rear Text")
	assert.Equal(t, [128]byte{}, sign.packetData)

	// Check some basics about the data but don't unit-test the whole packet.
	sign.SetId(53)
	sign.update(arena, allianceStation, true, "12:34", "Rear Text")
	assert.Equal(t, "CYPRX", string(sign.packetData[0:5]))
	assert.Equal(t, 53, int(sign.packetData[5]))
	assert.Equal(t, 0x04, int(sign.packetData[6]))
	assert.Equal(t, []byte{0x01, 53, 0x01}, sign.packetData[7:10])
	assert.Equal(t, "     ", string(sign.packetData[10:15]))
	assert.Equal(t, []byte{0, 0}, sign.packetData[15:17])
	assert.Equal(t, "No Team Assigned", string(sign.packetData[34:50]))
	assert.Equal(t, 51, sign.packetIndex)

	assertSign := func(isRed bool, expectedFrontText string, expectedFrontColor color.RGBA, expectedRearText string) {
		frontText, frontColor, rearText := sign.generateTeamNumberTexts(
			arena, allianceStation, isRed, "12:34", "Rear Text",
		)
		assert.Equal(t, expectedFrontText, frontText)
		assert.Equal(t, expectedRearText, rearText)

		// Modify front color to account for time-based blinking.
		frontColor.A = 255
		assert.Equal(t, expectedFrontColor, frontColor)
	}

	assertSign(true, "     ", whiteColor, "    No Team Assigned")
	arena.FieldReset = true
	arena.assignTeam(254, "R1")
	assertSign(true, "  254", greenColor, "254       Connect PC")
	assertSign(false, "  254", greenColor, "254       Connect PC")
	arena.FieldReset = false
	assertSign(true, "  254", redColor, "254       Connect PC")
	assertSign(false, "  254", blueColor, "254       Connect PC")

	// Check through pre-match sequence.
	allianceStation.Ethernet = true
	assertSign(true, "  254", redColor, "254         Start DS")
	allianceStation.DsConn = &DriverStationConnection{}
	assertSign(true, "  254", redColor, "254         No Radio")
	allianceStation.DsConn.WrongStation = "R1"
	assertSign(true, "  254", redColor, "254     Move Station")
	allianceStation.DsConn.WrongStation = ""
	allianceStation.DsConn.RadioLinked = true
	assertSign(true, "  254", redColor, "254           No Rio")
	allianceStation.DsConn.RioLinked = true
	assertSign(true, "  254", redColor, "254          No Code")
	allianceStation.DsConn.RobotLinked = true
	assertSign(true, "  254", redColor, "254            Ready")
	allianceStation.Bypass = true
	assertSign(true, "  254", redColor, "254         Bypassed")

	// Check that timeout mode has no effect on the team sign.
	arena.MatchState = TimeoutActive
	assertSign(true, "  254", redColor, "254         Bypassed")

	// Check E-stop and A-stop.
	arena.MatchState = AutoPeriod
	assertSign(true, "  254", redColor, "Rear Text")
	allianceStation.AStop = true
	assertSign(true, "  254", orangeColor, "254           A-STOP")
	allianceStation.EStop = true
	assertSign(false, "  254", orangeColor, "254           E-STOP")
	allianceStation.EStop = false
	arena.MatchState = TeleopPeriod
	assertSign(false, "  254", blueColor, "Rear Text")
	allianceStation.EStop = true
	assertSign(false, "  254", orangeColor, "254           E-STOP")
	arena.MatchState = PostMatch
	assertSign(false, "  254", orangeColor, "254           E-STOP")

	// Test preloading the team for the next match.
	sign.nextMatchTeamId = 1503
	assertSign(false, "  254", orangeColor, "Next Team Up: 1503")
	allianceStation.Bypass = false
	allianceStation.EStop = false
	allianceStation.Ethernet = false
	arena.MatchState = PreMatch
	arena.assignTeam(1503, "R1")
	assertSign(false, " 1503", blueColor, "1503      Connect PC")

	// Check blank mode.
	arena.AllianceStationDisplayMode = "blank"
	assertSign(true, "     ", whiteColor, "")

	// Check alliance selection.
	arena.AllianceStationDisplayMode = "logo"
	arena.AudienceDisplayMode = "allianceSelection"
	arena.AllianceSelectionShowTimer = false
	assertSign(true, " 2025", redColor, "1503      Connect PC")
	arena.AllianceSelectionShowTimer = true
	assertSign(false, " 2025", blueColor, "1503      Connect PC")
	arena.AllianceStationDisplayMode = "blank"
	assertSign(false, "     ", whiteColor, "")
}
