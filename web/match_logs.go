// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for viewing match logs

package web

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
)

type MatchLogsListItem struct {
	Id         int
	ShortName  string
	Time       string
	RedTeams   []int
	BlueTeams  []int
	ColorClass string
	IsComplete bool
}

type MatchLogRow struct {
	MatchTimeSec          float64
	PacketType            int
	TeamId                int
	AllianceStation       string
	DsLinked              bool
	RadioLinked           bool
	RioLinked             bool
	RobotLinked           bool
	Auto                  bool
	Enabled               bool
	EmergencyStop         bool
	AutonomousStop        bool
	BatteryVoltage        float64
	MissedPacketCount     int
	DsRobotTripTimeMs     int
	TxRate                float64
	RxRate                float64
	SignalNoiseRatio      int
	EthernetConnected     bool
	DsReportedStatusValid bool
	DsReportedAuto        bool
	DsReportedTeleop      bool
	DsReportedDisabled    bool
	DsReportedEnabled     bool
}

type MatchLog struct {
	Filename  string
	StartTime string
	Rows      []MatchLogRow
}

type MatchLogs struct {
	TeamId          int
	AllianceStation string
	Logs            []MatchLog
}

// Shows the match Log interface.
func (web *Web) matchLogsHandler(w http.ResponseWriter, r *http.Request) {
	practiceMatches, err := web.buildMatchLogsList(model.Practice)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := web.buildMatchLogsList(model.Qualification)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	playoffMatches, err := web.buildMatchLogsList(model.Playoff)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/match_logs.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[model.MatchType][]MatchLogsListItem{
		model.Practice:      practiceMatches,
		model.Qualification: qualificationMatches,
		model.Playoff:       playoffMatches,
	}
	currentMatchType := web.arena.CurrentMatch.Type
	if currentMatchType == model.Test {
		currentMatchType = model.Practice
	}
	data := struct {
		*model.EventSettings
		MatchesByType    map[model.MatchType][]MatchLogsListItem
		CurrentMatchType model.MatchType
	}{web.arena.EventSettings, matchesByType, currentMatchType}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Shows the page to view a log for a match.
func (web *Web) matchLogsViewGetHandler(w http.ResponseWriter, r *http.Request) {
	match, matchLogs, _, err := web.getMatchLogFromRequest(r)
	firstMatch := "0"
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template, err := web.parseFiles("templates/view_match_log.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(matchLogs.Logs) > 0 {
		firstMatch = matchLogs.Logs[0].StartTime
	}
	data := struct {
		*model.EventSettings
		Match      *model.Match
		MatchLogs  *MatchLogs
		FirstMatch string
	}{web.arena.EventSettings, match, matchLogs, firstMatch}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Load the match logs for the match referenced in the HTTP query string.
func (web *Web) getMatchLogFromRequest(r *http.Request) (*model.Match, *MatchLogs, bool, error) {
	matchId, err := strconv.Atoi(r.PathValue("matchId"))
	if err != nil {
		return nil, nil, false, err
	}
	stationId := r.PathValue("stationId")
	match, err := web.arena.Database.GetMatchById(matchId)

	logs := MatchLogs{
		TeamId:          0,
		AllianceStation: stationId,
	}
	if err != nil {
		return nil, nil, false, err
	}
	if match == nil {
		return nil, nil, false, fmt.Errorf("Error: No such match: %d", matchId)
	}
	switch stationId {
	case "R1":
		logs.TeamId = match.Red1
	case "R2":
		logs.TeamId = match.Red2
	case "R3":
		logs.TeamId = match.Red3
	case "B1":
		logs.TeamId = match.Blue1
	case "B2":
		logs.TeamId = match.Blue2
	case "B3":
		logs.TeamId = match.Blue3
	}
	// rows []MatchLogRow
	// Load a csv file.
	if logs.TeamId == 0 {
		return nil, nil, false, nil
	}
	var files []string
	files, err = filepath.Glob(
		filepath.Join(".", "static", "logs", "*_*_Match_"+match.ShortName+"_"+strconv.Itoa(logs.TeamId)+".csv"),
	)
	if err != nil {
		return nil, nil, false, err
	}
	if len(files) == 0 {
		return match, &logs, false, nil
	}

	for _, filename := range files {
		err := func() (err error) {
			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer func() {
				if closeErr := f.Close(); err == nil && closeErr != nil {
					err = closeErr
				}
			}()

			// Create a new reader.
			reader := csv.NewReader(f)

			// Read row
			header, err := reader.Read()
			if err != nil {
				return err
			}

			// Add mapping: Column/property name --> record index
			headerMap := make(map[string]int)
			for i, v := range header {
				headerMap[v] = i
			}
			records, err := reader.ReadAll()
			if err != nil {
				return err
			}

			var curlog = MatchLog{
				Filename:  filename,
				StartTime: filename[12:26],
				Rows:      make([]MatchLogRow, len(records)),
			}
			for i, record := range records {
				var curRow MatchLogRow
				curRow.MatchTimeSec = parseOptionalFloat(record, headerMap, "matchTimeSec", 0)
				curRow.PacketType = parseOptionalInt(record, headerMap, "packetType", 0)
				curRow.TeamId = parseOptionalInt(record, headerMap, "teamId", 0)
				curRow.AllianceStation = parseOptionalString(record, headerMap, "allianceStation", "")
				curRow.DsLinked = parseOptionalBool(record, headerMap, "dsLinked", false)
				curRow.RadioLinked = parseOptionalBool(record, headerMap, "radioLinked", false)
				curRow.RioLinked = parseOptionalBool(record, headerMap, "rioLinked", false)
				curRow.RobotLinked = parseOptionalBool(record, headerMap, "robotLinked", false)
				curRow.Auto = parseOptionalBool(record, headerMap, "auto", false)
				curRow.Enabled = parseOptionalBool(record, headerMap, "enabled", false)
				curRow.EmergencyStop = parseOptionalBool(record, headerMap, "emergencyStop", false)
				curRow.AutonomousStop = parseOptionalBool(record, headerMap, "autonomousStop", false)
				curRow.BatteryVoltage = parseOptionalFloat(record, headerMap, "batteryVoltage", 0)
				curRow.MissedPacketCount = parseOptionalInt(record, headerMap, "missedPacketCount", 0)
				curRow.DsRobotTripTimeMs = parseOptionalInt(record, headerMap, "dsRobotTripTimeMs", 0)
				curRow.TxRate = parseOptionalFloat(record, headerMap, "txRate", -1)
				curRow.RxRate = parseOptionalFloat(record, headerMap, "rxRate", -1)
				curRow.SignalNoiseRatio = parseOptionalInt(record, headerMap, "signalNoiseRatio", -1)
				curRow.EthernetConnected = parseOptionalBool(record, headerMap, "ethernetConnected", false)
				curRow.DsReportedStatusValid = parseOptionalBool(record, headerMap, "dsReportedStatusValid", false)
				curRow.DsReportedAuto = parseOptionalBool(record, headerMap, "dsReportedAuto", false)
				curRow.DsReportedTeleop = parseOptionalBool(record, headerMap, "dsReportedTeleop", false)
				curRow.DsReportedDisabled = parseOptionalBool(record, headerMap, "dsReportedDisabled", false)
				curRow.DsReportedEnabled = parseOptionalBool(record, headerMap, "dsReportedEnabled", false)

				// Store the parsed row in the same position as the CSV record.
				curlog.Rows[i] = curRow
			}

			logs.Logs = append(logs.Logs, curlog)
			return nil
		}()
		if err != nil {
			return nil, nil, false, err
		}
	}
	return match, &logs, false, nil
}

// parseOptionalString returns a CSV value by column name, or a default for legacy files that lack the column.
func parseOptionalString(record []string, headerMap map[string]int, columnName string, defaultValue string) string {
	index, ok := headerMap[columnName]
	if !ok || index >= len(record) {
		return defaultValue
	}
	return record[index]
}

// parseOptionalBool parses a bool CSV value by column name, preserving the default for missing or malformed values.
func parseOptionalBool(record []string, headerMap map[string]int, columnName string, defaultValue bool) bool {
	valueString := parseOptionalString(record, headerMap, columnName, "")
	if valueString == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueString)
	if err != nil {
		return defaultValue
	}
	return value
}

// parseOptionalFloat parses a float CSV value by column name, preserving the default for missing or malformed values.
func parseOptionalFloat(record []string, headerMap map[string]int, columnName string, defaultValue float64) float64 {
	valueString := parseOptionalString(record, headerMap, columnName, "")
	if valueString == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// parseOptionalInt parses an int CSV value by column name, preserving the default for missing or malformed values.
func parseOptionalInt(record []string, headerMap map[string]int, columnName string, defaultValue int) int {
	valueString := parseOptionalString(record, headerMap, columnName, "")
	if valueString == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueString)
	if err != nil {
		return defaultValue
	}
	return value
}

// Constructs the list of matches to display in the match Logs interface.
func (web *Web) buildMatchLogsList(matchType model.MatchType) ([]MatchLogsListItem, error) {
	matches, err := web.arena.Database.GetMatchesByType(matchType, false)
	if err != nil {
		return []MatchLogsListItem{}, err
	}

	matchLogsList := make([]MatchLogsListItem, len(matches))
	for i, match := range matches {
		matchLogsList[i].Id = match.Id
		matchLogsList[i].ShortName = match.ShortName
		matchLogsList[i].Time = match.Time.Local().Format("Mon 1/02 03:04 PM")
		matchLogsList[i].RedTeams = []int{match.Red1, match.Red2, match.Red3}
		matchLogsList[i].BlueTeams = []int{match.Blue1, match.Blue2, match.Blue3}
		if err != nil {
			return []MatchLogsListItem{}, err
		}
		switch match.Status {
		case game.RedWonMatch:
			matchLogsList[i].ColorClass = "red"
			matchLogsList[i].IsComplete = true
		case game.BlueWonMatch:
			matchLogsList[i].ColorClass = "blue"
			matchLogsList[i].IsComplete = true
		case game.TieMatch:
			matchLogsList[i].ColorClass = "yellow"
			matchLogsList[i].IsComplete = true
		default:
			matchLogsList[i].ColorClass = ""
			matchLogsList[i].IsComplete = false
		}
	}

	return matchLogsList, nil
}
