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
	"github.com/gorilla/mux"
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
	MatchTimeSec      float64
	PacketType        int
	TeamId            int
	AllianceStation   string
	DsLinked          bool
	RadioLinked       bool
	RioLinked         bool
	RobotLinked       bool
	Auto              bool
	Enabled           bool
	EmergencyStop     bool
	BatteryVoltage    float64
	MissedPacketCount int
	DsRobotTripTimeMs int
	TxRate            float64
	RxRate            float64
	SignalNoiseRatio  int
}

type MatchLog struct {
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

	template, err := web.parseFiles("templates/view_match_log.html")
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
	err = template.ExecuteTemplate(w, "view_match_log.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Load the match result for the match referenced in the HTTP query string.
func (web *Web) getMatchLogFromRequest(r *http.Request) (*model.Match, *MatchLogs, bool, error) {
	vars := mux.Vars(r)
	matchId, _ := strconv.Atoi(vars["matchId"])
	stationId := vars["stationId"]
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
	headerMap := make(map[string]int)
	//rows []MatchLogRow
	// Load a csv file.
	if logs.TeamId == 0 {
		return nil, nil, false, nil
	}
	var files []string
	files, _ = filepath.Glob(filepath.Join(".", "static", "logs", "*_*_Match_"+match.ShortName+"_"+strconv.Itoa(logs.TeamId)+".csv"))
	if len(files) == 0 {
		return match, &logs, false, nil
	}

	for _, v := range files {
		f, _ := os.Open(v)
		defer f.Close()
		// Create a new reader.
		reader := csv.NewReader(f)

		// Read row
		header, _ := reader.Read()

		// Add mapping: Column/property name --> record index
		for i, v := range header {
			headerMap[v] = i
		}
		records, _ := reader.ReadAll()

		var curlog = MatchLog{
			StartTime: v[12:26],
			Rows:      make([]MatchLogRow, len(records)),
		}
		for i, record := range records {
			var curRow MatchLogRow
			curRow.MatchTimeSec, _ = strconv.ParseFloat(record[headerMap["matchTimeSec"]], 64)
			curRow.PacketType, _ = strconv.Atoi(record[headerMap["packetType"]])
			curRow.TeamId, _ = strconv.Atoi(record[headerMap["teamId"]])
			curRow.AllianceStation = record[headerMap["allianceStation"]]
			curRow.DsLinked, _ = strconv.ParseBool(record[headerMap["dsLinked"]])
			curRow.RadioLinked, _ = strconv.ParseBool(record[headerMap["radioLinked"]])
			curRow.RioLinked, _ = strconv.ParseBool(record[headerMap["rioLinked"]])
			curRow.RobotLinked, _ = strconv.ParseBool(record[headerMap["robotLinked"]])
			curRow.Auto, _ = strconv.ParseBool(record[headerMap["auto"]])
			curRow.Enabled, _ = strconv.ParseBool(record[headerMap["enabled"]])
			curRow.EmergencyStop, _ = strconv.ParseBool(record[headerMap["emergencyStop"]])
			curRow.BatteryVoltage, _ = strconv.ParseFloat(record[headerMap["batteryVoltage"]], 64)
			curRow.MissedPacketCount, _ = strconv.Atoi(record[headerMap["missedPacketCount"]])
			curRow.DsRobotTripTimeMs, _ = strconv.Atoi(record[headerMap["dsRobotTripTimeMs"]])
			if len(headerMap) > 13 {

				curRow.TxRate, _ = strconv.ParseFloat(record[headerMap["txRate"]], 64)
				curRow.RxRate, _ = strconv.ParseFloat(record[headerMap["rxRate"]], 64)
				curRow.SignalNoiseRatio, _ = strconv.Atoi(record[headerMap["signalNoiseRatio"]])
			} else {
				curRow.TxRate = -1
				curRow.RxRate = -1
				curRow.SignalNoiseRatio = -1
			}

			// Create new person and add to persons array
			curlog.Rows[i] = curRow
		}

		logs.Logs = append(logs.Logs, curlog)

	}
	return match, &logs, false, nil
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
			matchLogsList[i].ColorClass = "danger"
			matchLogsList[i].IsComplete = true
		case game.BlueWonMatch:
			matchLogsList[i].ColorClass = "info"
			matchLogsList[i].IsComplete = true
		case game.TieMatch:
			matchLogsList[i].ColorClass = "warning"
			matchLogsList[i].IsComplete = true
		default:
			matchLogsList[i].ColorClass = ""
			matchLogsList[i].IsComplete = false
		}
	}

	return matchLogsList, nil
}
