// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Models and logic for controlling a Cypress team number / timer sign.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"image/color"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// Represents a collection of team number and timer signs.
type TeamSigns struct {
	Red1      TeamSign
	Red2      TeamSign
	Red3      TeamSign
	RedTimer  TeamSign
	Blue1     TeamSign
	Blue2     TeamSign
	Blue3     TeamSign
	BlueTimer TeamSign
}

// Represents a team number or timer sign.
type TeamSign struct {
	isTimer         bool
	address         byte
	nextMatchTeamId int
	frontText       string
	frontColor      color.RGBA
	rearText        string
	lastFrontText   string
	lastFrontColor  color.RGBA
	lastRearText    string
	udpConn         net.Conn
	packetData      [128]byte
	packetIndex     int
	lastPacketTime  time.Time
}

const (
	teamSignAddressPrefix            = "10.0.100."
	teamSignPort                     = 10011
	teamSignPacketMagicString        = "CYPRX"
	teamSignPacketHeaderLength       = 7
	teamSignCommandSetDisplay        = 0x04
	teamSignAddressSingle            = 0x01
	teamSignPacketTypeFrontText      = 0x01
	teamSignPacketTypeRearText       = 0x02
	teamSignPacketTypeFrontIntensity = 0x03
	teamSignPacketTypeColor          = 0x04
	teamSignPacketPeriodMs           = 5000
	teamSignBlinkPeriodMs            = 750
)

// Predefined colors for the team sign front text. The "A" channel is used as the intensity.
var redColor = color.RGBA{255, 0, 0, 255}
var blueColor = color.RGBA{0, 50, 255, 255}
var greenColor = color.RGBA{0, 255, 0, 255}
var orangeColor = color.RGBA{255, 50, 0, 255}
var purpleColor = color.RGBA{255, 0, 240, 255}
var whiteColor = color.RGBA{255, 200, 180, 255}

// Creates a new collection of team signs.
func NewTeamSigns() *TeamSigns {
	signs := new(TeamSigns)
	signs.RedTimer.isTimer = true
	signs.BlueTimer.isTimer = true
	return signs
}

// Updates the state of all signs with the latest data and sends packets to the signs if anything has changed.
func (signs *TeamSigns) Update(arena *Arena) {
	// Generate the countdown string which is used in multiple places.
	matchTimeSec := int(arena.MatchTimeSec())
	var countdownSec int
	switch arena.MatchState {
	case PreMatch:
		if arena.AudienceDisplayMode == "allianceSelection" {
			countdownSec = arena.AllianceSelectionTimeRemainingSec
		} else {
			countdownSec = game.MatchTiming.AutoDurationSec
		}
	case StartMatch:
		fallthrough
	case WarmupPeriod:
		countdownSec = game.MatchTiming.AutoDurationSec
	case AutoPeriod:
		countdownSec = game.MatchTiming.WarmupDurationSec + game.MatchTiming.AutoDurationSec - matchTimeSec
	case TeleopPeriod:
		countdownSec = game.MatchTiming.WarmupDurationSec + game.MatchTiming.AutoDurationSec +
			game.MatchTiming.TeleopDurationSec + game.MatchTiming.PauseDurationSec - matchTimeSec
	case TimeoutActive:
		countdownSec = game.MatchTiming.TimeoutDurationSec - matchTimeSec
	default:
		countdownSec = 0
	}
	countdown := fmt.Sprintf("%02d:%02d", countdownSec/60, countdownSec%60)

	// Generate the in-match rear text which is common to a whole alliance.
	redInMatchTeamRearText := generateInMatchTeamRearText(arena, true, countdown)
	redInMatchTimerRearText := generateInMatchTimerRearText(arena, true)
	blueInMatchTeamRearText := generateInMatchTeamRearText(arena, false, countdown)
	blueInMatchTimerRearText := generateInMatchTimerRearText(arena, false)

	signs.Red1.update(arena, arena.AllianceStations["R1"], true, countdown, redInMatchTeamRearText)
	signs.Red2.update(arena, arena.AllianceStations["R2"], true, countdown, redInMatchTeamRearText)
	signs.Red3.update(arena, arena.AllianceStations["R3"], true, countdown, redInMatchTeamRearText)
	signs.RedTimer.update(arena, nil, true, countdown, redInMatchTimerRearText)
	signs.Blue1.update(arena, arena.AllianceStations["B1"], false, countdown, blueInMatchTeamRearText)
	signs.Blue2.update(arena, arena.AllianceStations["B2"], false, countdown, blueInMatchTeamRearText)
	signs.Blue3.update(arena, arena.AllianceStations["B3"], false, countdown, blueInMatchTeamRearText)
	signs.BlueTimer.update(arena, nil, false, countdown, blueInMatchTimerRearText)
}

// Sets the team numbers for the next match on all signs.
func (signs *TeamSigns) SetNextMatchTeams(teams [6]int) {
	signs.Red1.nextMatchTeamId = teams[0]
	signs.Red2.nextMatchTeamId = teams[1]
	signs.Red3.nextMatchTeamId = teams[2]
	signs.Blue1.nextMatchTeamId = teams[3]
	signs.Blue2.nextMatchTeamId = teams[4]
	signs.Blue3.nextMatchTeamId = teams[5]
}

// Sets the IP address of the sign.
func (sign *TeamSign) SetId(id int) {
	if sign.udpConn != nil {
		_ = sign.udpConn.Close()
	}
	sign.address = byte(id)
	if id == 0 {
		// The sign is not configured.
		return
	}
	ipAddress := fmt.Sprintf("%s%d", teamSignAddressPrefix, id)

	var err error
	sign.udpConn, err = net.Dial("udp4", fmt.Sprintf("%s:%d", ipAddress, teamSignPort))
	if err != nil {
		log.Printf("Failed to connect to team sign at %s: %v", ipAddress, err)
		return
	}
	addressParts := strings.Split(ipAddress, ".")
	if len(addressParts) != 4 {
		log.Printf("Failed to configure team sign: invalid IP address: %s", ipAddress)
		return
	}
	address, _ := strconv.Atoi(addressParts[3])
	sign.address = byte(address)

	// Reset the sign's state to ensure that the next packet sent will update the sign.
	sign.packetIndex = 0
	sign.lastPacketTime = time.Time{}
}

// Updates the sign's internal state with the latest data and sends packets to the sign if anything has changed.
func (sign *TeamSign) update(
	arena *Arena, allianceStation *AllianceStation, isRed bool, countdown, inMatchRearText string,
) {
	if sign.address == 0 {
		// Don't do anything if there is no sign configured in this position.
		return
	}

	if sign.isTimer {
		sign.frontText, sign.frontColor, sign.rearText = generateTimerTexts(arena, countdown, inMatchRearText)
	} else {
		sign.frontText, sign.frontColor, sign.rearText = sign.generateTeamNumberTexts(
			arena, allianceStation, isRed, countdown, inMatchRearText,
		)
	}

	if err := sign.sendPacket(); err != nil {
		log.Printf("Failed to send team sign packet: %v", err)
	}
}

// Returns the in-match rear text for the team number display that is common to the whole given alliance.
func generateInMatchTeamRearText(arena *Arena, isRed bool, countdown string) string {
	var realtimeScore, opponentRealtimeScore *RealtimeScore
	var formatString string
	if isRed {
		realtimeScore = arena.RedRealtimeScore
		opponentRealtimeScore = arena.BlueRealtimeScore
		formatString = "R%03d-B%03d"
	} else {
		realtimeScore = arena.BlueRealtimeScore
		opponentRealtimeScore = arena.RedRealtimeScore
		formatString = "B%03d-R%03d"
	}
	scoreSummary := realtimeScore.CurrentScore.Summarize(&opponentRealtimeScore.CurrentScore)
	scoreTotal := scoreSummary.Score - scoreSummary.BargePoints
	opponentScoreSummary := opponentRealtimeScore.CurrentScore.Summarize(&realtimeScore.CurrentScore)
	opponentScoreTotal := opponentScoreSummary.Score - opponentScoreSummary.BargePoints
	allianceScores := fmt.Sprintf(formatString, scoreTotal, opponentScoreTotal)

	var coralRankingPointProgress string
	if arena.CurrentMatch.Type != model.Playoff {
		coralRankingPointProgress = fmt.Sprintf("%d/%d", scoreSummary.NumCoralLevels, scoreSummary.NumCoralLevelsGoal)
	}

	return fmt.Sprintf("%s %s %s", countdown, allianceScores, coralRankingPointProgress)
}

// Returns the in-match rear text for the timer display for the given alliance.
func generateInMatchTimerRearText(arena *Arena, isRed bool) string {
	var reef *game.Reef
	if isRed {
		reef = &arena.RedRealtimeScore.CurrentScore.Reef
	} else {
		reef = &arena.BlueRealtimeScore.CurrentScore.Reef
	}

	return fmt.Sprintf(
		"1-%02d 2-%02d 3-%02d 4-%02d",
		reef.CountTotalCoralByLevel(game.Level1),
		reef.CountTotalCoralByLevel(game.Level2),
		reef.CountTotalCoralByLevel(game.Level3),
		reef.CountTotalCoralByLevel(game.Level4),
	)
}

// Returns the front text, front color, and rear text to display on the timer display.
func generateTimerTexts(arena *Arena, countdown, inMatchRearText string) (string, color.RGBA, string) {
	if arena.AllianceStationDisplayMode == "blank" {
		return "     ", whiteColor, ""
	}
	if arena.AudienceDisplayMode == "allianceSelection" {
		if arena.AllianceSelectionShowTimer {
			return countdown, whiteColor, ""
		} else {
			return "     ", whiteColor, ""
		}
	}

	var frontText string
	var frontColor color.RGBA
	rearText := inMatchRearText
	if arena.AllianceStationDisplayMode == "logo" {
		frontText = fmt.Sprintf("%5d", time.Now().Year())
		frontColor = whiteColor
	} else if arena.AllianceStationDisplayMode == "timeout" {
		frontText = countdown
		frontColor = whiteColor
	} else if arena.FieldReset && arena.MatchState != TimeoutActive {
		frontText = "SAFE "
		frontColor = greenColor
	} else if arena.FieldVolunteers && arena.MatchState != TimeoutActive {
		frontText = "count"
		frontColor = purpleColor
	} else {
		frontText = countdown
		frontColor = whiteColor
	}
	if arena.MatchState == TimeoutActive {
		rearText = fmt.Sprintf("Field Break: %s", countdown)
	}
	return frontText, frontColor, rearText
}

// Returns the front text, front color, and rear text to display on the sign for the given alliance station.
func (sign *TeamSign) generateTeamNumberTexts(
	arena *Arena, allianceStation *AllianceStation, isRed bool, countdown, inMatchRearText string,
) (string, color.RGBA, string) {
	allianceColor := redColor
	if !isRed {
		allianceColor = blueColor
	}

	if arena.AllianceStationDisplayMode == "blank" {
		return "     ", whiteColor, ""
	}

	var frontText string
	var frontColor color.RGBA
	if arena.AllianceStationDisplayMode == "logo" {
		frontText = fmt.Sprintf("%5d", time.Now().Year())
		frontColor = allianceColor
	} else {
		if allianceStation.Team == nil {
			return "     ", whiteColor, fmt.Sprintf("%20s", "No Team Assigned")
		}

		frontText = fmt.Sprintf("%5d", allianceStation.Team.Id)

		if allianceStation.EStop {
			frontColor = orangeColor
		} else if allianceStation.AStop && arena.MatchState == AutoPeriod {
			frontColor = blinkColor(orangeColor)
		} else if arena.FieldReset {
			frontColor = greenColor
		} else if arena.FieldVolunteers {
			frontColor = purpleColor
		} else if allianceStation.DsConn != nil && !allianceStation.DsConn.RobotLinked &&
			(arena.MatchState == AutoPeriod || arena.MatchState == PausePeriod || arena.MatchState == TeleopPeriod) {
			// Blink the display to indicate that the robot is not linked while the match is in progress.
			frontColor = blinkColor(allianceColor)
		} else {
			frontColor = allianceColor
		}
	}

	var message string
	if allianceStation.EStop {
		message = "E-STOP"
	} else if allianceStation.AStop && arena.MatchState == AutoPeriod {
		message = "A-STOP"
	} else if arena.MatchState == PreMatch || arena.MatchState == TimeoutActive {
		if allianceStation.Bypass {
			message = "Bypassed"
		} else if !allianceStation.Ethernet {
			message = "Connect PC"
		} else if allianceStation.DsConn == nil {
			message = "Start DS"
		} else if allianceStation.DsConn.WrongStation != "" {
			message = "Move Station"
		} else if !allianceStation.DsConn.RadioLinked {
			message = "No Radio"
		} else if !allianceStation.DsConn.RioLinked {
			message = "No Rio"
		} else if !allianceStation.DsConn.RobotLinked {
			message = "No Code"
		} else {
			message = "Ready"
		}
	}

	var rearText string
	if arena.MatchState == PostMatch && sign.nextMatchTeamId > 0 && sign.nextMatchTeamId != allianceStation.Team.Id {
		// Show the next match team number on the rear display before the score is committed so that queueing teams know
		// where to go.
		rearText = fmt.Sprintf("Next Team Up: %d", sign.nextMatchTeamId)
	} else if len(message) > 0 {
		teamId := 0
		if allianceStation.Team != nil {
			teamId = allianceStation.Team.Id
		}
		rearText = fmt.Sprintf("%-5d %14s", teamId, message)
	} else {
		rearText = inMatchRearText
	}

	return frontText, frontColor, rearText
}

// Sends a UDP packet to the sign if its state has changed.
func (sign *TeamSign) sendPacket() error {
	if sign.packetIndex == 0 {
		// Write the static packet header the first time this method is invoked.
		sign.writePacketData([]byte(teamSignPacketMagicString))
		sign.writePacketData([]byte{sign.address, teamSignCommandSetDisplay})
	} else {
		// Reset the write index to just after the header.
		sign.packetIndex = teamSignPacketHeaderLength
	}

	isStale := time.Now().Sub(sign.lastPacketTime).Milliseconds() >= teamSignPacketPeriodMs

	if sign.frontText != sign.lastFrontText || isStale {
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeFrontText})
		sign.writePacketData([]byte(sign.frontText))
		sign.writePacketData([]byte{0, 0}) // Second byte is "show decimal point".
		sign.lastFrontText = sign.frontText
	}

	if sign.frontColor != sign.lastFrontColor || isStale {
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeColor})
		sign.writePacketData([]byte{sign.frontColor.R, sign.frontColor.G, sign.frontColor.B})
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeFrontIntensity})
		sign.writePacketData([]byte{sign.frontColor.A})
		sign.lastFrontColor = sign.frontColor
	}

	if sign.rearText != sign.lastRearText || isStale {
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeRearText})
		sign.writePacketData([]byte(sign.rearText))
		sign.writePacketData([]byte{0})
		sign.lastRearText = sign.rearText
	}

	if sign.packetIndex > teamSignPacketHeaderLength && sign.udpConn != nil {
		sign.lastPacketTime = time.Now()
		if _, err := sign.udpConn.Write(sign.packetData[:sign.packetIndex]); err != nil {
			return err
		}
	}

	return nil
}

// Writes the given data to the packet buffer and advances the write index.
func (sign *TeamSign) writePacketData(data []byte) {
	for _, value := range data {
		sign.packetData[sign.packetIndex] = value
		sign.packetIndex++
	}
}

// Periodically modifies the given color to zero brightness to create a blinking effect.
func blinkColor(originalColor color.RGBA) color.RGBA {
	if time.Now().UnixMilli()%teamSignBlinkPeriodMs < teamSignBlinkPeriodMs/2 {
		return originalColor
	}
	return color.RGBA{originalColor.R, originalColor.G, originalColor.B, 0}
}
