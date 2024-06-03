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
}

const (
	teamSignPort                     = 10011
	teamSignPacketMagicString        = "CYPRX"
	teamSignPacketHeaderLength       = 7
	teamSignCommandSetDisplay        = 0x04
	teamSignAddressSingle            = 0x01
	teamSignPacketTypeFrontText      = 0x01
	teamSignPacketTypeRearText       = 0x02
	teamSignPacketTypeFrontIntensity = 0x03
	teamSignPacketTypeColor          = 0x04
)

// Predefined colors for the team sign front text. The "A" channel is used as the intensity.
var redColor = color.RGBA{255, 0, 0, 255}
var blueColor = color.RGBA{0, 0, 255, 255}
var greenColor = color.RGBA{0, 255, 0, 255}
var orangeColor = color.RGBA{255, 165, 0, 255}
var whiteColor = color.RGBA{255, 255, 255, 255}

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
		fallthrough
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
	redInMatchRearText := generateInMatchRearText(true, countdown, arena.RedRealtimeScore, arena.BlueRealtimeScore)
	blueInMatchRearText := generateInMatchRearText(false, countdown, arena.BlueRealtimeScore, arena.RedRealtimeScore)

	signs.Red1.update(arena, arena.AllianceStations["R1"], true, countdown, redInMatchRearText)
	signs.Red2.update(arena, arena.AllianceStations["R2"], true, countdown, redInMatchRearText)
	signs.Red3.update(arena, arena.AllianceStations["R3"], true, countdown, redInMatchRearText)
	signs.RedTimer.update(arena, nil, true, countdown, redInMatchRearText)
	signs.Blue1.update(arena, arena.AllianceStations["B1"], false, countdown, blueInMatchRearText)
	signs.Blue2.update(arena, arena.AllianceStations["B2"], false, countdown, blueInMatchRearText)
	signs.Blue3.update(arena, arena.AllianceStations["B3"], false, countdown, blueInMatchRearText)
	signs.BlueTimer.update(arena, nil, false, countdown, blueInMatchRearText)
}

// Sets the team numbers for the next match on all signs.
func (signs *TeamSigns) SetNextMatchTeams(match *model.Match) {
	signs.Red1.nextMatchTeamId = match.Red1
	signs.Red2.nextMatchTeamId = match.Red2
	signs.Red3.nextMatchTeamId = match.Red3
	signs.Blue1.nextMatchTeamId = match.Blue1
	signs.Blue2.nextMatchTeamId = match.Blue2
	signs.Blue3.nextMatchTeamId = match.Blue3
}

// Sets the IP address of the sign.
func (sign *TeamSign) SetAddress(ipAddress string) {
	if sign.udpConn != nil {
		_ = sign.udpConn.Close()
	}
	if ipAddress == "" {
		// The sign is not configured.
		sign.address = 0
		return
	}

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

	sign.lastFrontText = "dummy value to ensure it gets cleared"
	sign.lastFrontColor = color.RGBA{}
	sign.lastRearText = "dummy value to ensure it gets cleared"
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
		sign.frontText, sign.frontColor = generateTimerText(arena.FieldReset, countdown)
		sign.rearText = inMatchRearText
	} else {
		sign.frontText, sign.frontColor, sign.rearText = sign.generateTeamNumberTexts(
			arena, allianceStation, isRed, inMatchRearText,
		)
	}

	if err := sign.sendPacket(); err != nil {
		log.Printf("Failed to send team sign packet: %v", err)
	}
}

// Returns the in-match rear text that is common to a whole alliance.
func generateInMatchRearText(isRed bool, countdown string, realtimeScore, opponentRealtimeScore *RealtimeScore) string {
	scoreSummary := realtimeScore.CurrentScore.Summarize(&opponentRealtimeScore.CurrentScore)
	scoreTotal := scoreSummary.Score
	opponentScoreTotal := opponentRealtimeScore.CurrentScore.Summarize(&realtimeScore.CurrentScore).Score
	var allianceScores string
	if isRed {
		allianceScores = fmt.Sprintf("R%03d-B%03d", scoreTotal, opponentScoreTotal)
	} else {
		allianceScores = fmt.Sprintf("B%03d-R%03d", scoreTotal, opponentScoreTotal)
	}
	if realtimeScore.AmplifiedTimeRemainingSec > 0 {
		// Replace the total score with the amplified countdown while it's active.
		allianceScores = fmt.Sprintf("Amp:%2d", realtimeScore.AmplifiedTimeRemainingSec)
	}
	return fmt.Sprintf(
		"%s %02d/%02d %9s", countdown[1:], scoreSummary.NumNotes, scoreSummary.NumNotesGoal, allianceScores,
	)
}

// Returns the front text and color to display on the timer display.
func generateTimerText(fieldReset bool, countdown string) (string, color.RGBA) {
	var frontText string
	var frontColor color.RGBA
	if fieldReset {
		frontText = "SAFE"
		frontColor = greenColor
	} else {
		frontText = countdown
		frontColor = whiteColor
	}
	return frontText, frontColor
}

// Returns the front text, front color, and rear text to display on the sign for the given alliance station.
func (sign *TeamSign) generateTeamNumberTexts(
	arena *Arena, allianceStation *AllianceStation, isRed bool, inMatchRearText string,
) (string, color.RGBA, string) {
	if allianceStation.Team == nil {
		return "", whiteColor, fmt.Sprintf("%20s", "No Team Assigned")
	}

	frontText := fmt.Sprintf("%5d", allianceStation.Team.Id)

	var frontColor color.RGBA
	if allianceStation.EStop || allianceStation.AStop && arena.MatchState == AutoPeriod {
		frontColor = orangeColor
	} else if arena.FieldReset {
		frontColor = greenColor
	} else if isRed {
		frontColor = redColor
	} else {
		frontColor = blueColor
	}

	var message string
	if allianceStation.EStop {
		message = "E-STOP"
	} else if allianceStation.AStop && arena.MatchState == AutoPeriod {
		message = "A-STOP"
	} else if arena.MatchState == PreMatch {
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
		rearText = fmt.Sprintf("%-5d %14s", allianceStation.Team.Id, message)
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

	if sign.frontText != sign.lastFrontText {
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeFrontText})
		sign.writePacketData([]byte(sign.frontText))
		sign.writePacketData([]byte{0, 0}) // Second byte is "show decimal point".
		sign.lastFrontText = sign.frontText
	}

	if sign.frontColor != sign.lastFrontColor {
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeColor})
		sign.writePacketData([]byte{sign.frontColor.R, sign.frontColor.G, sign.frontColor.B})
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeFrontIntensity})
		sign.writePacketData([]byte{sign.frontColor.A})
		sign.lastFrontColor = sign.frontColor
	}

	if sign.rearText != sign.lastRearText {
		sign.writePacketData([]byte{teamSignAddressSingle, sign.address, teamSignPacketTypeRearText})
		sign.writePacketData([]byte(sign.rearText))
		sign.writePacketData([]byte{0})
		sign.lastRearText = sign.rearText
	}

	if sign.packetIndex > teamSignPacketHeaderLength {
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
