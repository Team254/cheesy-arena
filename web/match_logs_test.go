// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestGetMatchLogFromRequestParsesLegacyAndNewColumns(t *testing.T) {
	web := setupTestWeb(t)
	match := &model.Match{Type: model.Qualification, ShortName: "Q9998", Red1: 9998}
	assert.Nil(t, web.arena.Database.CreateMatch(match))

	logsDir := filepath.Join(".", "static", "logs")
	assert.Nil(t, os.MkdirAll(logsDir, 0755))
	legacyFilename := filepath.Join(logsDir, "20260606000000_Qualification_Match_Q9998_9998.csv")
	newFilename := filepath.Join(logsDir, "20260606000001_Qualification_Match_Q9998_9998.csv")
	t.Cleanup(func() {
		os.Remove(legacyFilename)
		os.Remove(newFilename)
	})

	legacyColumns := []string{
		"matchTimeSec", "packetType", "teamId", "allianceStation", "dsLinked", "radioLinked",
		"rioLinked", "robotLinked", "auto", "enabled", "emergencyStop", "autonomousStop",
		"batteryVoltage", "missedPacketCount", "dsRobotTripTimeMs", "rxRate", "txRate",
		"signalNoiseRatio",
	}
	newColumns := append(legacyColumns,
		"ethernetConnected", "dsReportedStatusValid", "dsReportedAuto", "dsReportedTeleop",
		"dsReportedDisabled", "dsReportedEnabled",
	)
	legacyCsv := strings.Join(legacyColumns, ",") + "\n" +
		"1.000000,22,9998,R1,true,true,true,true,true,true,false,false,12.500000,1,2,3.000000,4.000000,5\n"
	newCsv := strings.Join(newColumns, ",") + "\n" +
		strings.Join([]string{
			"2.000000", "22", "9998", "R1", "true", "false", "false", "false", "false", "false",
			"false", "false", "0.000000", "6", "7", "8.000000", "9.000000", "10", "true", "true",
			"false", "true", "true", "false",
		}, ",") + "\n"
	assert.Nil(t, os.WriteFile(legacyFilename, []byte(legacyCsv), 0644))
	assert.Nil(t, os.WriteFile(newFilename, []byte(newCsv), 0644))

	request := httptest.NewRequest("GET", "/match_logs/"+strconv.Itoa(match.Id)+"/R1/log", nil)
	request.SetPathValue("matchId", strconv.Itoa(match.Id))
	request.SetPathValue("stationId", "R1")
	_, matchLogs, _, err := web.getMatchLogFromRequest(request)
	assert.Nil(t, err)
	if assert.NotNil(t, matchLogs) && assert.Len(t, matchLogs.Logs, 2) {
		var legacyRow *MatchLogRow
		var newRow *MatchLogRow
		for i := range matchLogs.Logs {
			row := &matchLogs.Logs[i].Rows[0]
			if row.MatchTimeSec == 1 {
				legacyRow = row
			} else if row.MatchTimeSec == 2 {
				newRow = row
			}
		}

		if assert.NotNil(t, legacyRow) {
			assert.False(t, legacyRow.EthernetConnected)
			assert.False(t, legacyRow.DsReportedStatusValid)
			assert.Equal(t, 4.0, legacyRow.TxRate)
			assert.Equal(t, 3.0, legacyRow.RxRate)
			assert.Equal(t, 5, legacyRow.SignalNoiseRatio)
		}
		if assert.NotNil(t, newRow) {
			assert.True(t, newRow.EthernetConnected)
			assert.True(t, newRow.DsReportedStatusValid)
			assert.False(t, newRow.DsReportedAuto)
			assert.True(t, newRow.DsReportedTeleop)
			assert.True(t, newRow.DsReportedDisabled)
			assert.False(t, newRow.DsReportedEnabled)
		}
	}
}
