// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for controlling match play.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
)

type MatchPlayListItem struct {
	Id          int
	DisplayName string
	Time        string
	Status      string
	ColorClass  string
}

type MatchPlayList []MatchPlayListItem

type MatchTimeMessage struct {
	MatchState   int
	MatchTimeSec int
}

// Global var to hold the current active tournament so that its matches are displayed by default.
var currentMatchType string

// Shows the match play control interface.
func MatchPlayHandler(w http.ResponseWriter, r *http.Request) {
	practiceMatches, err := buildMatchPlayList("practice")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	qualificationMatches, err := buildMatchPlayList("qualification")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	eliminationMatches, err := buildMatchPlayList("elimination")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	template := template.New("").Funcs(templateHelpers)
	_, err = template.ParseFiles("templates/match_play.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	matchesByType := map[string]MatchPlayList{"practice": practiceMatches,
		"qualification": qualificationMatches, "elimination": eliminationMatches}
	if currentMatchType == "" {
		currentMatchType = "practice"
	}
	allowSubstitution := mainArena.currentMatch.Type == "test" || mainArena.currentMatch.Type == "practice"
	matchResult, err := db.GetMatchResultForMatch(mainArena.currentMatch.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	isReplay := matchResult != nil
	data := struct {
		*EventSettings
		MatchesByType     map[string]MatchPlayList
		CurrentMatchType  string
		Match             *Match
		AllowSubstitution bool
		IsReplay          bool
	}{eventSettings, matchesByType, currentMatchType, mainArena.currentMatch, allowSubstitution, isReplay}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func MatchPlayLoadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchId, _ := strconv.Atoi(vars["matchId"])
	var match *Match
	var err error
	if matchId == 0 {
		err = mainArena.LoadTestMatch()
	} else {
		match, err = db.GetMatchById(matchId)
		if err != nil {
			handleWebErr(w, err)
			return
		}
		if match == nil {
			handleWebErr(w, fmt.Errorf("Invalid match ID %d.", matchId))
			return
		}
		err = mainArena.LoadMatch(match)
	}
	if err != nil {
		handleWebErr(w, err)
		return
	}
	currentMatchType = mainArena.currentMatch.Type

	http.Redirect(w, r, "/match_play", 302)
}

// The websocket endpoint for the match play client to send control commands and receive status updates.
func MatchPlayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	robotStatusListener := mainArena.robotStatusNotifier.Listen()
	defer close(robotStatusListener)

	// Send the various notifications immediately upon connection.
	err = websocket.Write("status", mainArena)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data := MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)}
	err = websocket.Write("matchTime", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
			case matchTimeSec, ok := <-matchTimeListener:
				if !ok {
					return
				}
				messageType = "matchTime"
				message = MatchTimeMessage{mainArena.MatchState, matchTimeSec.(int)}
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = mainArena
			}
			err = websocket.Write(messageType, message)
			if err != nil {
				// The client has probably closed the connection; nothing to do here.
				return
			}
		}
	}()

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}

		switch messageType {
		case "substituteTeam":
			args := struct {
				Team     int
				Position string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = mainArena.SubstituteTeam(args.Team, args.Position)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "toggleBypass":
			station, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			if _, ok := mainArena.AllianceStations[station]; !ok {
				websocket.WriteError(fmt.Sprintf("Invalid alliance station '%s'.", station))
				continue
			}
			mainArena.AllianceStations[station].Bypass = !mainArena.AllianceStations[station].Bypass
		case "startMatch":
			err = mainArena.StartMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "abortMatch":
			err = mainArena.AbortMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
		case "commitResults":
			// TODO(pat): Deal with scoring here. For now, use an empty match result set for a 0-0 tie.
			err = CommitMatchScore(mainArena.currentMatch, &MatchResult{MatchId: mainArena.currentMatch.Id})
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = mainArena.ResetMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = mainArena.LoadNextMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = websocket.Write("reload", nil)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		case "discardResults":
			err = mainArena.ResetMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = mainArena.LoadNextMatch()
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			err = websocket.Write("reload", nil)
			if err != nil {
				log.Printf("Websocket error: %s", err)
				return
			}
			continue // Skip sending the status update, as the client is about to terminate and reload.
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Send out the status again after handling the command, as it most likely changed as a result.
		err = websocket.Write("status", mainArena)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

func CommitMatchScore(match *Match, matchResult *MatchResult) error {
	if match.Type == "test" {
		// Do nothing since this is a test match and doesn't exist in the database.
		return nil
	}

	if matchResult.PlayNumber == 0 {
		// Determine the play number for this new match result.
		prevMatchResult, err := db.GetMatchResultForMatch(match.Id)
		if err != nil {
			return err
		}
		if prevMatchResult != nil {
			matchResult.PlayNumber = prevMatchResult.PlayNumber + 1
		} else {
			matchResult.PlayNumber = 1
		}

		// Save the match result record to the database.
		err = db.CreateMatchResult(matchResult)
		if err != nil {
			return err
		}
	} else {
		// We are updating a match result record that already exists.
		err := db.SaveMatchResult(matchResult)
		if err != nil {
			return err
		}
	}

	// Update and save the match record to the database.
	match.Status = "complete"
	redScore := matchResult.RedScoreSummary()
	blueScore := matchResult.BlueScoreSummary()
	if redScore.Score > blueScore.Score {
		match.Winner = "R"
	} else if redScore.Score < blueScore.Score {
		match.Winner = "B"
	} else {
		match.Winner = "T"
	}
	err := db.SaveMatch(match)
	if err != nil {
		return err
	}

	// Recalculate all the rankings.
	err = db.CalculateRankings()
	if err != nil {
		return err
	}

	return nil
}

func (list MatchPlayList) Len() int {
	return len(list)
}

func (list MatchPlayList) Less(i, j int) bool {
	return list[i].Status != "complete" && list[j].Status == "complete"
}

func (list MatchPlayList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func buildMatchPlayList(matchType string) (MatchPlayList, error) {
	matches, err := db.GetMatchesByType(matchType)
	if err != nil {
		return MatchPlayList{}, err
	}

	prefix := ""
	if matchType == "practice" {
		prefix = "P"
	} else if matchType == "qualification" {
		prefix = "Q"
	}
	matchPlayList := make(MatchPlayList, len(matches))
	for i, match := range matches {
		matchPlayList[i].Id = match.Id
		matchPlayList[i].DisplayName = prefix + match.DisplayName
		matchPlayList[i].Time = match.Time.Format("3:04 PM")
		matchPlayList[i].Status = match.Status
		switch match.Winner {
		case "R":
			matchPlayList[i].ColorClass = "danger"
		case "B":
			matchPlayList[i].ColorClass = "info"
		case "T":
			matchPlayList[i].ColorClass = "warning"
		default:
			matchPlayList[i].ColorClass = ""
		}
		if mainArena.currentMatch != nil && matchPlayList[i].Id == mainArena.currentMatch.Id {
			matchPlayList[i].ColorClass = "success"
		}
	}

	// Sort the list to put all completed matches at the bottom.
	sort.Stable(matchPlayList)

	return matchPlayList, nil
}
