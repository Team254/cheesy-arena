// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Utilities for logging packets received from team driver stations during a match.

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

// Creates a file to log to for the given match and team.
func NewTeamMatchLog(teamId int, match *model.Match, wifiStatus *network.TeamWifiStatus) (*TeamMatchLog, error) {
	err := os.MkdirAll(filepath.Join(model.BaseDir, logsDir), 0755)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s_%s_Match_%s_%d.csv", filepath.Join(model.BaseDir, logsDir),
		time.Now().Format("20060102150405"), match.Type.String(), match.ShortName, teamId)
	logFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	log := TeamMatchLog{log.New(logFile, "", 0), logFile, wifiStatus}
	log.logger.Println("matchTimeSec,packetType,teamId,allianceStation,dsLinked,radioLinked,rioLinked,robotLinked,auto,enabled," +
		"emergencyStop,batteryVoltage,missedPacketCount,dsRobotTripTimeMs,rxRate,txRate,signalNoiseRatio")

	return &log, nil
}

// Adds a line to the log when a packet is received.
func (log *TeamMatchLog) LogDsPacket(matchTimeSec float64, packetType int, dsConn *DriverStationConnection) {
	log.logger.Printf(
		"%f,%d,%d,%s,%v,%v,%v,%v,%v,%v,%v,%f,%d,%d,%f,%f,%d",
		matchTimeSec,
		packetType,
		dsConn.TeamId,
		dsConn.AllianceStation,
		dsConn.DsLinked,
		dsConn.RadioLinked,
		dsConn.RioLinked,
		dsConn.RobotLinked,
		dsConn.Auto,
		dsConn.Enabled,
		dsConn.Estop,
		dsConn.BatteryVoltage,
		dsConn.MissedPacketCount,
		dsConn.DsRobotTripTimeMs,
		log.wifiStatus.RxRate,
		log.wifiStatus.TxRate,
		log.wifiStatus.SignalNoiseRatio,
	)
}

func (log *TeamMatchLog) Close() {
	log.logFile.Close()
}
