// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"image/color"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestTeamSign_GenerateInMatchRearText(t *testing.T) {
	arena := setupTestArena(t)
	currentTimeAt := func(matchTimeSec int) time.Time {
		return arena.MatchStartTime.Add(time.Duration(matchTimeSec) * time.Second)
	}
	score := func(
		wonAuto bool, shiftCounts [game.ShiftCount]int, autoTowerStatuses [3]game.TowerStatus,
	) game.Score {
		return game.Score{
			AutoTowerStatuses: autoTowerStatuses, Hub: game.Hub{WonAuto: wonAuto, ShiftCounts: shiftCounts},
		}
	}

	testCases := []struct {
		name             string
		matchType        model.MatchType
		matchTimeSec     int
		countdown        string
		redScore         game.Score
		blueScore        game.Score
		expectedRearText [4]string
	}{
		{
			name:         "qualification auto",
			matchType:    model.Qualification,
			matchTimeSec: 5,
			countdown:    "2:37",
			redScore:     score(false, [game.ShiftCount]int{}, [3]game.TowerStatus{}),
			blueScore:    score(true, [game.ShiftCount]int{}, [3]game.TowerStatus{}),
			expectedRearText: [4]string{
				" A15   0/100  0 2:37",
				"2:37       R000-B000",
				" A15   0/100  0 2:37",
				"2:37       B000-R000",
			},
		},
		{
			name:         "playoff transition",
			matchType:    model.Playoff,
			matchTimeSec: 25,
			countdown:    "2:02",
			redScore: score(
				false,
				[game.ShiftCount]int{8, 4},
				[3]game.TowerStatus{game.TowerLevel1},
			),
			blueScore: score(
				true,
				[game.ShiftCount]int{5},
				[3]game.TowerStatus{game.TowerLevel2, game.TowerLevel3},
			),
			expectedRearText: [4]string{
				"  T07 R027-B035 2:02",
				"2:02       R027-B035",
				"  T07 B035-R027 2:02",
				"2:02       B035-R027",
			},
		},
		{
			name:         "qualification shift 1 red active",
			matchType:    model.Qualification,
			matchTimeSec: 45,
			countdown:    "1:57",
			redScore: score(
				false,
				[game.ShiftCount]int{10, 7, 13},
				[3]game.TowerStatus{game.TowerLevel2, game.TowerLevel3},
			),
			blueScore: score(
				true,
				[game.ShiftCount]int{11, 6, 0, 8},
				[3]game.TowerStatus{game.TowerLevel1},
			),
			expectedRearText: [4]string{
				" R12  30/100 30 1:57",
				"1:57       R060-B040",
				" R12  25/100 15 1:57",
				"1:57       B040-R060",
			},
		},
		{
			name:         "playoff shift 2 blue active",
			matchType:    model.Playoff,
			matchTimeSec: 73,
			countdown:    "1:32",
			redScore: score(
				false,
				[game.ShiftCount]int{20, 10, 5, 40, 0, 0, 0, 3},
				[3]game.TowerStatus{},
			),
			blueScore: score(
				true,
				[game.ShiftCount]int{22, 11, 99, 7, 0, 0, 0, 4},
				[3]game.TowerStatus{game.TowerLevel1},
			),
			expectedRearText: [4]string{
				"  B09 R035-B055 1:32",
				"1:32       R035-B055",
				"  B09 B055-R035 1:32",
				"1:32       B055-R035",
			},
		},
		{
			name:         "qualification shift 3 blue active after red won auto",
			matchType:    model.Qualification,
			matchTimeSec: 101,
			countdown:    "1:01",
			redScore: score(
				true,
				[game.ShiftCount]int{40, 15, 0, 20, 99},
				[3]game.TowerStatus{game.TowerLevel2},
			),
			blueScore: score(
				false,
				[game.ShiftCount]int{35, 5, 12, 0, 22},
				[3]game.TowerStatus{},
			),
			expectedRearText: [4]string{
				" B06  75/100 15 1:01",
				"1:01       R090-B074",
				" B06  74/100  0 1:01",
				"1:01       B074-R090",
			},
		},
		{
			name:         "qualification shift 4 red active after red won auto",
			matchType:    model.Qualification,
			matchTimeSec: 128,
			countdown:    "0:29",
			redScore: score(
				true,
				[game.ShiftCount]int{80, 20, 0, 40, 0, 230, 5, 10},
				[3]game.TowerStatus{game.TowerLevel2, game.TowerLevel3},
			),
			blueScore: score(
				false,
				[game.ShiftCount]int{45, 10, 30, 0, 30, 0, 5, 5},
				[3]game.TowerStatus{},
			),
			expectedRearText: [4]string{
				" R04 375/360 30 0:29",
				"0:29       R405-B120",
				" R04 120/360  0 0:29",
				"0:29       B120-R405",
			},
		},
		{
			name:         "playoff endgame",
			matchType:    model.Playoff,
			matchTimeSec: 151,
			countdown:    "0:11",
			redScore: score(
				false,
				[game.ShiftCount]int{7, 3, 2, 0, 4, 0, 6, 1},
				[3]game.TowerStatus{game.TowerLevel1},
			),
			blueScore: score(
				true,
				[game.ShiftCount]int{9, 5, 0, 4, 0, 7, 8, 2},
				[3]game.TowerStatus{game.TowerLevel2, game.TowerLevel3},
			),
			expectedRearText: [4]string{
				"  E11 R037-B063 0:11",
				"0:11       R037-B063",
				"  E11 B063-R037 0:11",
				"0:11       B063-R037",
			},
		},
		{
			name:         "qualification after match",
			matchType:    model.Qualification,
			matchTimeSec: 165,
			countdown:    "0:00",
			redScore:     score(false, [game.ShiftCount]int{}, [3]game.TowerStatus{}),
			blueScore:    score(true, [game.ShiftCount]int{}, [3]game.TowerStatus{}),
			expectedRearText: [4]string{
				" E00   0/100  0 0:00",
				"0:00       R000-B000",
				" E00   0/100  0 0:00",
				"0:00       B000-R000",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.name, func(t *testing.T) {
				arena.CurrentMatch.Type = testCase.matchType
				arena.RedRealtimeScore.CurrentScore = testCase.redScore
				arena.BlueRealtimeScore.CurrentScore = testCase.blueScore
				currentTime := currentTimeAt(testCase.matchTimeSec)

				actualRearText := [4]string{
					generateInMatchTeamRearText(arena, true, testCase.countdown, currentTime),
					generateInMatchTimerRearText(arena, true, testCase.countdown),
					generateInMatchTeamRearText(arena, false, testCase.countdown, currentTime),
					generateInMatchTimerRearText(arena, false, testCase.countdown),
				}
				assert.Equal(t, testCase.expectedRearText, actualRearText)
			},
		)
	}
}

func TestTeamSign_Timer(t *testing.T) {
	arena := setupTestArena(t)
	sign := TeamSign{isTimer: true}

	// Should do nothing if no address is set.
	sign.update(arena, "", true, "12:34", "Rear Text")
	assert.Equal(t, [128]byte{}, sign.packetData)

	// Check some basics about the data but don't unit-test the whole packet.
	sign.SetId(56)
	sign.update(arena, "", true, "12:34", "Rear Text")
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
	sign.update(arena, "R1", true, "12:34", "Rear Text")
	assert.Equal(t, [128]byte{}, sign.packetData)

	// Check some basics about the data but don't unit-test the whole packet.
	sign.SetId(53)
	sign.update(arena, "R1", true, "12:34", "Rear Text")
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
			arena, "R1", isRed, "12:34", "Rear Text",
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
	assertSign(true, "  254", greenColor, "254       Connect PC")
	assertSign(false, "  254", greenColor, "254       Connect PC")

	// Check through pre-match sequence.
	allianceStation.Ethernet = true
	assertSign(true, "  254", greenColor, "254         Start DS")
	allianceStation.DsConn = &DriverStationConnection{}
	assertSign(true, "  254", greenColor, "254         No Radio")
	allianceStation.DsConn.WrongStation = "R1"
	assertSign(true, "  254", greenColor, "254     Move Station")
	allianceStation.DsConn.WrongStation = ""
	allianceStation.DsConn.RadioLinked = true
	assertSign(true, "  254", greenColor, "254           No Rio")
	allianceStation.DsConn.RioLinked = true
	assertSign(true, "  254", greenColor, "254          No Code")
	allianceStation.DsConn.RobotLinked = true
	assertSign(true, "  254", redColor, "254            Ready")

	arena.FieldReset = true
	assertSign(true, "  254", redColor, "254            Ready")
	arena.FieldReset = false
	assertSign(true, "  254", redColor, "254            Ready")
	allianceStation.DsConn.RobotLinked = false
	assertSign(true, "  254", greenColor, "254          No Code")
	allianceStation.DsConn.RobotLinked = true
	allianceStation.Bypass = true
	assertSign(true, "  254", redColor, "254         Bypassed")

	// Check that timeout mode has no effect on the team sign.
	arena.MatchState = TimeoutActive
	assertSign(true, "  254", redColor, "254         Bypassed")
	arena.FieldReset = false

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
	allianceStation.EStop = false
	arena.FieldReset = true
	assertSign(false, "  254", greenColor, "Rear Text")
	allianceStation.EStop = true

	// Test preloading the team for the next match.
	sign.nextMatchTeamId = 1503
	assertSign(false, "  254", orangeColor, "Next Team Up: 1503")
	allianceStation.Bypass = false
	allianceStation.EStop = false
	allianceStation.Ethernet = false
	arena.MatchState = PreMatch
	arena.assignTeam(1503, "R1")
	assertSign(false, " 1503", greenColor, "1503      Connect PC")

	// Check blank mode.
	arena.AllianceStationDisplayMode = "blank"
	assertSign(true, "     ", whiteColor, "")

	// Check alliance selection.
	arena.AllianceStationDisplayMode = "logo"
	arena.AudienceDisplayMode = "allianceSelection"
	arena.AllianceSelectionShowTimer = false
	assertSign(true, " 2026", redColor, "1503      Connect PC")
	arena.AllianceSelectionShowTimer = true
	assertSign(false, " 2026", blueColor, "1503      Connect PC")
	arena.AllianceStationDisplayMode = "blank"
	assertSign(false, "     ", whiteColor, "")
}
