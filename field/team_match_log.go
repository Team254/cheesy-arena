// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Utilities for logging team station snapshots during a match.

package field

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
)

const logsDir = "static/logs"

type TeamMatchLog struct {
	logger     *log.Logger
	logFile    *os.File
	wifiStatus *network.TeamWifiStatus
}

const teamMatchLogPacketType = 22

// Creates a file to log to for the given match and team.
func NewTeamMatchLog(teamId int, match *model.Match, wifiStatus *network.TeamWifiStatus) (*TeamMatchLog, error) {
	err := os.MkdirAll(filepath.Join(model.BaseDir, logsDir), 0755)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf(
		"%s/%s_%s_Match_%s_%d.csv",
		filepath.Join(model.BaseDir, logsDir),
		time.Now().Format("20060102150405"),
		match.Type.String(),
		match.ShortName,
		teamId,
	)
	logFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	log := TeamMatchLog{log.New(logFile, "", 0), logFile, wifiStatus}
	log.logger.Println(
		"matchTimeSec,packetType,teamId,allianceStation,dsLinked,radioLinked,rioLinked,robotLinked,auto,enabled," +
			"emergencyStop,autonomousStop,batteryVoltage,missedPacketCount,dsRobotTripTimeMs,rxRate,txRate," +
			"signalNoiseRatio,ethernetConnected,dsReportedStatusValid,dsReportedAuto,dsReportedTeleop," +
			"dsReportedDisabled,dsReportedEnabled",
	)

	return &log, nil
}

// Adds a line to the log for the current station status.
func (log *TeamMatchLog) LogStationSnapshot(matchTimeSec float64, station *AllianceStation, allianceStationId string) {
	dsConn := station.DsConn
	teamId := 0
	dsLinked := false
	radioLinked := false
	rioLinked := false
	robotLinked := false
	auto := false
	enabled := false
	emergencyStop := station.EStop
	autonomousStop := station.AStop
	batteryVoltage := 0.0
	missedPacketCount := 0
	dsRobotTripTimeMs := 0
	dsReportedStatusValid := false
	dsReportedAuto := false
	dsReportedTeleop := false
	dsReportedDisabled := false
	dsReportedEnabled := false
	rxRate := -1.0
	txRate := -1.0
	signalNoiseRatio := -1

	if station.Team != nil {
		teamId = station.Team.Id
	}
	if dsConn != nil {
		teamId = dsConn.TeamId
		dsLinked = dsConn.DsLinked
		radioLinked = dsConn.RadioLinked
		rioLinked = dsConn.RioLinked
		robotLinked = dsConn.RobotLinked
		auto = dsConn.Auto
		enabled = dsConn.Enabled
		emergencyStop = dsConn.EStop
		autonomousStop = dsConn.AStop
		batteryVoltage = dsConn.BatteryVoltage
		missedPacketCount = dsConn.MissedPacketCount
		dsRobotTripTimeMs = dsConn.DsRobotTripTimeMs
		dsReportedStatusValid = dsConn.DsReportedStatusValid
		dsReportedAuto = dsConn.DsReportedAuto
		dsReportedTeleop = dsConn.DsReportedTeleop
		dsReportedDisabled = dsConn.DsReportedDisabled
		dsReportedEnabled = dsConn.DsReportedEnabled
	}
	if log.wifiStatus != nil {
		rxRate = log.wifiStatus.RxRate
		txRate = log.wifiStatus.TxRate
		signalNoiseRatio = log.wifiStatus.SignalNoiseRatio
	}

	log.logger.Printf(
		"%f,%d,%d,%s,%v,%v,%v,%v,%v,%v,%v,%v,%f,%d,%d,%f,%f,%d,%v,%v,%v,%v,%v,%v",
		matchTimeSec,
		teamMatchLogPacketType,
		teamId,
		allianceStationId,
		dsLinked,
		radioLinked,
		rioLinked,
		robotLinked,
		auto,
		enabled,
		emergencyStop,
		autonomousStop,
		batteryVoltage,
		missedPacketCount,
		dsRobotTripTimeMs,
		rxRate,
		txRate,
		signalNoiseRatio,
		station.Ethernet,
		dsReportedStatusValid,
		dsReportedAuto,
		dsReportedTeleop,
		dsReportedDisabled,
		dsReportedEnabled,
	)
}

func (log *TeamMatchLog) Close() error {
	return log.logFile.Close()
}
