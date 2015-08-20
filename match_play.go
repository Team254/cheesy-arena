// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for controlling match play.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"text/template"
	"time"
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
	allowSubstitution := mainArena.currentMatch.Type != "qualification"
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

// Loads the given match onto the arena in preparation for playing it.
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

// Loads the results for the given match into the display buffer.
func MatchPlayShowResultHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchId, _ := strconv.Atoi(vars["matchId"])
	match, err := db.GetMatchById(matchId)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if match == nil {
		handleWebErr(w, fmt.Errorf("Invalid match ID %d.", matchId))
		return
	}
	matchResult, err := db.GetMatchResultForMatch(match.Id)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if matchResult == nil {
		handleWebErr(w, fmt.Errorf("No result found for match ID %d.", matchId))
		return
	}
	mainArena.savedMatch = match
	mainArena.savedMatchResult = matchResult
	mainArena.scorePostedNotifier.Notify(nil)

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
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	robotStatusListener := mainArena.robotStatusNotifier.Listen()
	defer close(robotStatusListener)
	audienceDisplayListener := mainArena.audienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	scoringStatusListener := mainArena.scoringStatusNotifier.Listen()
	defer close(scoringStatusListener)
	allianceStationDisplayListener := mainArena.allianceStationDisplayNotifier.Listen()
	defer close(allianceStationDisplayListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
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
	data = MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)}
	err = websocket.Write("matchTime", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		BlueScore int
	}{mainArena.redRealtimeScore.Score(), mainArena.blueRealtimeScore.Score()}
	err = websocket.Write("realtimeScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("setAudienceDisplay", mainArena.audienceDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RefereeScoreReady bool
		RedScoreReady     bool
		BlueScoreReady    bool
	}{mainArena.redRealtimeScore.FoulsCommitted && mainArena.blueRealtimeScore.FoulsCommitted,
		mainArena.redRealtimeScore.TeleopCommitted, mainArena.blueRealtimeScore.TeleopCommitted}
	err = websocket.Write("scoringStatus", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("setAllianceStationDisplay", mainArena.allianceStationDisplayScreen)
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
			case _, ok := <-realtimeScoreListener:
				if !ok {
					return
				}
				messageType = "realtimeScore"
				message = struct {
					RedScore  int
					BlueScore int
				}{mainArena.redRealtimeScore.Score(), mainArena.blueRealtimeScore.Score()}
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = mainArena
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = mainArena.audienceDisplayScreen
			case _, ok := <-scoringStatusListener:
				if !ok {
					return
				}
				messageType = "scoringStatus"
				message = struct {
					RefereeScoreReady bool
					RedScoreReady     bool
					BlueScoreReady    bool
				}{mainArena.redRealtimeScore.FoulsCommitted && mainArena.blueRealtimeScore.FoulsCommitted,
					mainArena.redRealtimeScore.TeleopCommitted, mainArena.blueRealtimeScore.TeleopCommitted}
			case _, ok := <-allianceStationDisplayListener:
				if !ok {
					return
				}
				messageType = "setAllianceStationDisplay"
				message = mainArena.allianceStationDisplayScreen
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
			args := struct {
				MuteMatchSounds bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}
			mainArena.muteMatchSounds = args.MuteMatchSounds
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
			err = CommitCurrentMatchScore()
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
		case "setAudienceDisplay":
			screen, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			mainArena.audienceDisplayScreen = screen
			mainArena.audienceDisplayNotifier.Notify(nil)
			continue
		case "setAllianceStationDisplay":
			screen, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			mainArena.allianceStationDisplayScreen = screen
			mainArena.allianceStationDisplayNotifier.Notify(nil)
			continue
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

// Saves the given match and result to the database, supplanting any previous result for the match.
func CommitMatchScore(match *Match, matchResult *MatchResult, loadToShowBuffer bool) error {
	if matchResult.RedScore.CoopertitionSet != matchResult.BlueScore.CoopertitionSet ||
		matchResult.RedScore.CoopertitionStack != matchResult.BlueScore.CoopertitionStack {
		// Don't accept the score if the red and blue co-opertition points don't match up.
		return fmt.Errorf("Red and blue co-opertition points don't match.")
	}

	// Remove empty stacks to make the results more concise.
	matchResult.RedScore.Stacks = stripEmptyStacks(matchResult.RedScore.Stacks)
	matchResult.BlueScore.Stacks = stripEmptyStacks(matchResult.BlueScore.Stacks)

	if match.Type == "elimination" {
		// Adjust the score if necessary for an elimination DQ or tie.
		matchResult.CorrectEliminationScore()
	}

	if loadToShowBuffer {
		// Store the result in the buffer to be shown in the audience display.
		mainArena.savedMatch = match
		mainArena.savedMatchResult = matchResult
		mainArena.scorePostedNotifier.Notify(nil)
	}

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

	if match.Type != "practice" {
		// Regenerate the residual yellow cards that teams may carry.
		db.CalculateTeamCards(match.Type)
	}

	if match.Type == "qualification" {
		// Recalculate all the rankings.
		err = db.CalculateRankings()
		if err != nil {
			return err
		}
	}

	if match.Type == "elimination" {
		// Generate any subsequent elimination matches.
		_, err = db.UpdateEliminationSchedule(time.Now().Add(time.Second * elimMatchSpacingSec))
		if err != nil {
			return err
		}
	}

	if eventSettings.TbaPublishingEnabled && match.Type != "practice" {
		// Publish asynchronously to The Blue Alliance.
		go func() {
			err = PublishMatches()
			if err != nil {
				log.Printf("Failed to publish matches: %s", err.Error())
			}
			if match.Type == "qualification" {
				err = PublishRankings()
				if err != nil {
					log.Printf("Failed to publish rankings: %s", err.Error())
				}
			}
		}()
	}

	// Back up the database, but don't error out if it fails.
	err = db.Backup(fmt.Sprintf("post_%s_match_%s", match.Type, match.DisplayName))
	if err != nil {
		log.Println(err)
	}

	return nil
}

func GetCurrentMatchResult() *MatchResult {
	return &MatchResult{MatchId: mainArena.currentMatch.Id,
		RedScore: mainArena.redRealtimeScore.CurrentScore, BlueScore: mainArena.blueRealtimeScore.CurrentScore,
		RedCards: mainArena.redRealtimeScore.Cards, BlueCards: mainArena.blueRealtimeScore.Cards}
}

// Saves the realtime result as the final score for the match currently loaded into the arena.
func CommitCurrentMatchScore() error {
	return CommitMatchScore(mainArena.currentMatch, GetCurrentMatchResult(), true)
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Len() int {
	return len(list)
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Less(i, j int) bool {
	return list[i].Status != "complete" && list[j].Status == "complete"
}

// Helper function to implement the required interface for Sort.
func (list MatchPlayList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// Constructs the list of matches to display on the side of the match play interface.
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
		matchPlayList[i].Time = match.Time.Local().Format("3:04 PM")
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

func stripEmptyStacks(stacks []Stack) []Stack {
	var filteredStacks []Stack
	for _, stack := range stacks {
		if stack.Totes > 0 || stack.Container || stack.Litter {
			filteredStacks = append(filteredStacks, stack)
		}
	}
	return filteredStacks
}
