// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Utilities for logging packets received from team driver stations during a match.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"log"
	"os"
	"path/filepath"
	"time"
)

const logsDir = "static/logs"

type TeamMatchLog struct {
	logger  *log.Logger
	logFile *os.File
}

// Creates a file to log to for the given match and team.
func NewTeamMatchLog(teamId int, match *model.Match) (*TeamMatchLog, error) {
	err := os.MkdirAll(filepath.Join(model.BaseDir, logsDir), 0755)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s_%s_Match_%s_%d.csv", filepath.Join(model.BaseDir, logsDir),
		time.Now().Format("20060102150405"), match.CapitalizedType(), match.DisplayName, teamId)
	logFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	log := TeamMatchLog{log.New(logFile, "", 0), logFile}
	log.logger.Println("matchTimeSec,packetType,teamId,allianceStation,dsLinked,radioLinked,robotLinked,auto,enabled," +
		"emergencyStop,batteryVoltage,missedPacketCount,dsRobotTripTimeMs")

	return &log, nil
}

// Adds a line to the log when a packet is received.
func (log *TeamMatchLog) LogDsPacket(matchTimeSec float64, packetType int, dsConn *DriverStationConnection) {
	log.logger.Printf("%f,%d,%d,%s,%v,%v,%v,%v,%v,%v,%f,%d,%d", matchTimeSec, packetType, dsConn.TeamId,
		dsConn.AllianceStation, dsConn.DsLinked, dsConn.RadioLinked, dsConn.RobotLinked, dsConn.Auto, dsConn.Enabled, dsConn.Estop,
		dsConn.BatteryVoltage, dsConn.MissedPacketCount, dsConn.DsRobotTripTimeMs)
}

func (log *TeamMatchLog) Close() {
	log.logFile.Close()
}
