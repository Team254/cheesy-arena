// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestTeamMatchLogWritesStationSnapshotWithDsReportedStatus(t *testing.T) {
	oldBaseDir := model.BaseDir
	model.BaseDir = t.TempDir()
	t.Cleanup(func() { model.BaseDir = oldBaseDir })

	match := &model.Match{Type: model.Qualification, ShortName: "Q1"}
	teamMatchLog, err := NewTeamMatchLog(254, match, nil)
	assert.Nil(t, err)
	station := &AllianceStation{
		DsConn: &DriverStationConnection{
			TeamId:                254,
			AllianceStation:       "R1",
			DsLinked:              true,
			RadioLinked:           true,
			RioLinked:             true,
			RobotLinked:           true,
			Auto:                  true,
			Enabled:               true,
			BatteryVoltage:        12.5,
			MissedPacketCount:     7,
			DsRobotTripTimeMs:     11,
			DsReportedStatusValid: true,
			DsReportedAuto:        true,
			DsReportedTeleop:      false,
			DsReportedDisabled:    false,
			DsReportedEnabled:     true,
		},
		Ethernet: true,
		Team:     &model.Team{Id: 254},
	}
	teamMatchLog.LogStationSnapshot(1.25, station, "R1")
	teamMatchLog.Close()

	records := readTeamMatchLogRecords(t, teamMatchLog.logFile.Name())
	header := records[0]
	row := records[1]
	headerMap := headerIndexMap(header)

	assert.Equal(t, "ethernetConnected", header[len(header)-6])
	assert.Equal(t, "dsReportedStatusValid", header[len(header)-5])
	assert.Equal(t, "dsReportedAuto", header[len(header)-4])
	assert.Equal(t, "dsReportedTeleop", header[len(header)-3])
	assert.Equal(t, "dsReportedDisabled", header[len(header)-2])
	assert.Equal(t, "dsReportedEnabled", header[len(header)-1])
	assert.Equal(t, "true", row[headerMap["ethernetConnected"]])
	assert.Equal(t, "true", row[headerMap["dsReportedStatusValid"]])
	assert.Equal(t, "true", row[headerMap["dsReportedAuto"]])
	assert.Equal(t, "false", row[headerMap["dsReportedTeleop"]])
	assert.Equal(t, "false", row[headerMap["dsReportedDisabled"]])
	assert.Equal(t, "true", row[headerMap["dsReportedEnabled"]])
}

func TestTeamMatchLogWritesSnapshotWithoutDsConnection(t *testing.T) {
	oldBaseDir := model.BaseDir
	model.BaseDir = t.TempDir()
	t.Cleanup(func() { model.BaseDir = oldBaseDir })

	match := &model.Match{Type: model.Qualification, ShortName: "Q1"}
	teamMatchLog, err := NewTeamMatchLog(254, match, nil)
	assert.Nil(t, err)
	station := &AllianceStation{
		Ethernet: false,
		EStop:    true,
		AStop:    true,
		Team:     &model.Team{Id: 254},
	}
	teamMatchLog.LogStationSnapshot(2.5, station, "B3")
	teamMatchLog.Close()

	records := readTeamMatchLogRecords(t, teamMatchLog.logFile.Name())
	row := records[1]
	headerMap := headerIndexMap(records[0])
	assert.Equal(t, "254", row[headerMap["teamId"]])
	assert.Equal(t, "B3", row[headerMap["allianceStation"]])
	assert.Equal(t, "false", row[headerMap["dsLinked"]])
	assert.Equal(t, "true", row[headerMap["emergencyStop"]])
	assert.Equal(t, "true", row[headerMap["autonomousStop"]])
	assert.Equal(t, "false", row[headerMap["dsReportedStatusValid"]])
}

func TestArenaContinuouslyLogsAfterDsDisconnect(t *testing.T) {
	arena := setupTestArena(t)
	oldBaseDir := model.BaseDir
	model.BaseDir = t.TempDir()
	t.Cleanup(func() { model.BaseDir = oldBaseDir })

	plc := &FakePlc{isEnabled: true, ftaReady: true}
	plc.blueEthernetConnected[2] = true
	arena.Plc = plc
	arena.Database.CreateTeam(&model.Team{Id: 254})
	assert.Nil(t, arena.assignTeam(254, "B3"))
	for _, allianceStation := range arena.AllianceStations {
		allianceStation.Bypass = true
		allianceStation.aStopReset = true
	}
	arena.AllianceStations["B3"].Bypass = false
	arena.AllianceStations["B3"].DsConn = &DriverStationConnection{
		TeamId:          254,
		AllianceStation: "B3",
		DsLinked:        true,
		RobotLinked:     true,
		lastPacketTime:  time.Now(),
	}

	assert.Nil(t, arena.StartMatch())
	arena.Update()
	arena.MatchStartTime = time.Now().Add(-time.Second)
	arena.lastTeamLogTime = time.Time{}
	arena.Update()
	logFilename := arena.AllianceStations["B3"].TeamMatchLog.logFile.Name()

	arena.AllianceStations["B3"].DsConn = nil
	plc.blueEthernetConnected[2] = false
	arena.lastTeamLogTime = time.Unix(0, 0)
	arena.Update()

	arena.closeTeamMatchLogs()
	records := readTeamMatchLogRecords(t, logFilename)
	assert.GreaterOrEqual(t, len(records), 3)
	headerMap := headerIndexMap(records[0])
	connectedRow := records[len(records)-2]
	disconnectedRow := records[len(records)-1]
	assert.Equal(t, "true", connectedRow[headerMap["dsLinked"]])
	assert.Equal(t, "true", connectedRow[headerMap["ethernetConnected"]])
	assert.Equal(t, "false", disconnectedRow[headerMap["dsLinked"]])
	assert.Equal(t, "false", disconnectedRow[headerMap["ethernetConnected"]])

	files, err := filepath.Glob(filepath.Join(model.BaseDir, logsDir, "*_Match_*_254.csv"))
	assert.Nil(t, err)
	assert.Len(t, files, 1)
}

func readTeamMatchLogRecords(t *testing.T, filename string) [][]string {
	t.Helper()
	file, err := os.Open(filename)
	assert.Nil(t, err)
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	assert.Nil(t, err)
	return records
}

func headerIndexMap(header []string) map[string]int {
	headerMap := make(map[string]int)
	for i, columnName := range header {
		headerMap[columnName] = i
	}
	return headerMap
}
