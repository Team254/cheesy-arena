// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Utilities for logging packets received from team driver stations during a match.

package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

const logsDir = "static/logs"

type TeamMatchLog struct {
	logger  *log.Logger
	logFile *os.File
}

func NewTeamMatchLog(teamId int, match *Match) (*TeamMatchLog, error) {
	err := os.MkdirAll(logsDir, 0755)
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s_%s_Match_%s_%d.csv", logsDir, time.Now().Format("20060102150405"),
		match.CapitalizedType(), match.DisplayName, teamId)
	logFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	log := TeamMatchLog{log.New(logFile, "", 0), logFile}
	log.logger.Println("matchTimeSec,teamId,allianceStation,robotLinked,auto,enabled,emergencyStop," +
		"batteryVoltage,dsVersion,packetCount,missedPacketCount,dsRobotTripTimeMs")

	return &log, nil
}

func (log *TeamMatchLog) LogDsStatus(matchTimeSec float64, dsStatus *DriverStationStatus) {
	log.logger.Printf("%f,%d,%s,%v,%v,%v,%v,%f,%s,%d,%d,%d", matchTimeSec, dsStatus.TeamId,
		dsStatus.AllianceStation, dsStatus.RobotLinked, dsStatus.Auto, dsStatus.Enabled,
		dsStatus.EmergencyStop, dsStatus.BatteryVoltage, dsStatus.DsVersion, dsStatus.PacketCount,
		dsStatus.MissedPacketCount, dsStatus.DsRobotTripTimeMs)
}

func (log *TeamMatchLog) Close() {
	log.logFile.Close()
}
