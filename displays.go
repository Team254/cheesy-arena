// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for displays.

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"io"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

var rules = []string{"G3", "G5", "G10", "G11", "G12", "G14", "G15", "G16", "G17", "G18", "G19", "G21", "G22",
	"G23", "G24", "G25", "G26", "G26-1", "G27", "G28", "G29", "G30", "G31", "G32", "G34", "G35", "G36", "G37",
	"G38", "G39", "G40", "G41", "G42"}

// Renders the audience display to be chroma keyed over the video feed.
func AudienceDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/audience_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.ExecuteTemplate(w, "audience_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the audience display client to receive status updates.
func AudienceDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	audienceDisplayListener := mainArena.audienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	scorePostedListener := mainArena.scorePostedNotifier.Listen()
	defer close(scorePostedListener)
	playSoundListener := mainArena.playSoundNotifier.Listen()
	defer close(playSoundListener)
	allianceSelectionListener := mainArena.allianceSelectionNotifier.Listen()
	defer close(allianceSelectionListener)
	lowerThirdListener := mainArena.lowerThirdNotifier.Listen()
	defer close(lowerThirdListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
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
		Match     *Match
		MatchName string
	}{mainArena.currentMatch, mainArena.currentMatch.CapitalizedType()}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		RedCycle  Cycle
		BlueScore int
		BlueCycle Cycle
	}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
		mainArena.redRealtimeScore.CurrentCycle,
		mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls),
		mainArena.blueRealtimeScore.CurrentCycle}
	err = websocket.Write("realtimeScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		Match     *Match
		MatchName string
		RedScore  *ScoreSummary
		BlueScore *ScoreSummary
	}{mainArena.savedMatch, mainArena.savedMatch.CapitalizedType(),
		mainArena.savedMatchResult.RedScoreSummary(), mainArena.savedMatchResult.BlueScoreSummary()}
	err = websocket.Write("setFinalScore", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("allianceSelection", cachedAlliances)
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
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = mainArena.audienceDisplayScreen
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				message = struct {
					Match     *Match
					MatchName string
				}{mainArena.currentMatch, mainArena.currentMatch.CapitalizedType()}
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
					RedCycle  Cycle
					BlueScore int
					BlueCycle Cycle
				}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
					mainArena.redRealtimeScore.CurrentCycle,
					mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls),
					mainArena.blueRealtimeScore.CurrentCycle}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "setFinalScore"
				message = struct {
					Match     *Match
					MatchName string
					RedScore  *ScoreSummary
					BlueScore *ScoreSummary
				}{mainArena.savedMatch, mainArena.savedMatch.CapitalizedType(),
					mainArena.savedMatchResult.RedScoreSummary(), mainArena.savedMatchResult.BlueScoreSummary()}
			case sound, ok := <-playSoundListener:
				if !ok {
					return
				}
				messageType = "playSound"
				message = sound
			case _, ok := <-allianceSelectionListener:
				if !ok {
					return
				}
				messageType = "allianceSelection"
				message = cachedAlliances
			case lowerThird, ok := <-lowerThirdListener:
				if !ok {
					return
				}
				messageType = "lowerThird"
				message = lowerThird
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		_, _, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

// Renders the pit display which shows scrolling rankings.
func PitDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("templates/pit_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.Execute(w, data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the pit display, used only to force reloads remotely.
func PitDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		_, _, err := websocket.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

// Renders the announcer display which shows team info and scores for the current match.
func AnnouncerDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/announcer_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the announcer display client to send control commands and receive status updates.
func AnnouncerDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	scorePostedListener := mainArena.scorePostedNotifier.Listen()
	defer close(scorePostedListener)
	audienceDisplayListener := mainArena.audienceDisplayNotifier.Listen()
	defer close(audienceDisplayListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	data = struct {
		MatchType        string
		MatchDisplayName string
		Red1             *Team
		Red2             *Team
		Red3             *Team
		Blue1            *Team
		Blue2            *Team
		Blue3            *Team
	}{mainArena.currentMatch.CapitalizedType(), mainArena.currentMatch.DisplayName,
		mainArena.AllianceStations["R1"].team, mainArena.AllianceStations["R2"].team,
		mainArena.AllianceStations["R3"].team, mainArena.AllianceStations["B1"].team,
		mainArena.AllianceStations["B2"].team, mainArena.AllianceStations["B3"].team}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		BlueScore int
	}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
		mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls)}
	err = websocket.Write("realtimeScore", data)
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
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				message = struct {
					MatchType        string
					MatchDisplayName string
					Red1             *Team
					Red2             *Team
					Red3             *Team
					Blue1            *Team
					Blue2            *Team
					Blue3            *Team
				}{mainArena.currentMatch.CapitalizedType(), mainArena.currentMatch.DisplayName,
					mainArena.AllianceStations["R1"].team, mainArena.AllianceStations["R2"].team,
					mainArena.AllianceStations["R3"].team, mainArena.AllianceStations["B1"].team,
					mainArena.AllianceStations["B2"].team, mainArena.AllianceStations["B3"].team}
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
				}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
					mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls)}
			case _, ok := <-scorePostedListener:
				if !ok {
					return
				}
				messageType = "setFinalScore"
				message = struct {
					MatchType        string
					MatchDisplayName string
					RedScoreSummary  *ScoreSummary
					BlueScoreSummary *ScoreSummary
					RedFouls         []Foul
					BlueFouls        []Foul
				}{mainArena.savedMatch.CapitalizedType(), mainArena.savedMatch.DisplayName,
					mainArena.savedMatchResult.RedScoreSummary(), mainArena.savedMatchResult.BlueScoreSummary(),
					mainArena.savedMatchResult.RedFouls, mainArena.savedMatchResult.BlueFouls}
			case _, ok := <-audienceDisplayListener:
				if !ok {
					return
				}
				messageType = "setAudienceDisplay"
				message = mainArena.audienceDisplayScreen
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		case "setAudienceDisplay":
			screen, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			mainArena.audienceDisplayScreen = screen
			mainArena.audienceDisplayNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}

// Renders the scoring interface which enables input of scores in real-time.
func ScoringDisplayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}

	template, err := template.ParseFiles("templates/scoring_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
		Alliance string
	}{eventSettings, alliance}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func ScoringDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" && alliance != "blue" {
		handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		return
	}
	var score **RealtimeScore
	if alliance == "red" {
		score = &mainArena.redRealtimeScore
	} else {
		score = &mainArena.blueRealtimeScore
	}

	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	err = websocket.Write("score", *score)
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
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "score"
				message = *score
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		case "preload":
			if !(*score).AutoCommitted {
				preloadedBallsStr, ok := data.(string)
				if !ok {
					websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
					continue
				}
				preloadedBalls, err := strconv.Atoi(preloadedBallsStr)
				(*score).AutoPreloadedBalls = preloadedBalls
				if err != nil {
					websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
					continue
				}
			}
		case "mobility":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoMobilityBonuses += 1
			}
		case "scoredHighHot":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoHighHot += 1
			}
		case "scoredHigh":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoHigh += 1
			} else if !(*score).TeleopCommitted && !(*score).CurrentCycle.ScoredHigh {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.ScoredHigh = true
				(*score).CurrentCycle.ScoredLow = false
				(*score).CurrentCycle.DeadBall = false
			}
		case "scoredLowHot":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoLowHot += 1
			}
		case "scoredLow":
			if !(*score).AutoCommitted {
				(*score).undoAutoScores = append((*score).undoAutoScores, (*score).CurrentScore)
				(*score).CurrentScore.AutoLow += 1
			} else if !(*score).TeleopCommitted && !(*score).CurrentCycle.ScoredLow {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.ScoredHigh = false
				(*score).CurrentCycle.ScoredLow = true
				(*score).CurrentCycle.DeadBall = false
			}
		case "assist":
			if (*score).AutoCommitted && !(*score).TeleopCommitted && (*score).AutoLeftoverBalls == 0 &&
				(*score).CurrentCycle.Assists < 3 {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.Assists += 1
			}
		case "truss":
			if (*score).AutoCommitted && !(*score).TeleopCommitted && (*score).AutoLeftoverBalls == 0 &&
				!(*score).CurrentCycle.Truss {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.Truss = true
			}
		case "catch":
			if (*score).AutoCommitted && !(*score).TeleopCommitted && (*score).AutoLeftoverBalls == 0 &&
				!(*score).CurrentCycle.Catch && (*score).CurrentCycle.Truss {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.Catch = true
			}
		case "deadBall":
			if !(*score).TeleopCommitted && !(*score).CurrentCycle.DeadBall {
				(*score).undoCycles = append((*score).undoCycles, (*score).CurrentCycle)
				(*score).CurrentCycle.ScoredHigh = false
				(*score).CurrentCycle.ScoredLow = false
				(*score).CurrentCycle.DeadBall = true
			}
		case "commit":
			if !(*score).AutoCommitted {
				(*score).AutoLeftoverBalls = (*score).AutoPreloadedBalls - (*score).CurrentScore.AutoHighHot -
					(*score).CurrentScore.AutoHigh - (*score).CurrentScore.AutoLowHot -
					(*score).CurrentScore.AutoLow
				(*score).AutoCommitted = true
			} else if !(*score).TeleopCommitted {
				if (*score).CurrentCycle.ScoredHigh || (*score).CurrentCycle.ScoredLow ||
					(*score).CurrentCycle.DeadBall {
					// Check whether this is a leftover ball from autonomous.
					if (*score).AutoLeftoverBalls > 0 {
						if (*score).CurrentCycle.ScoredHigh {
							(*score).CurrentScore.AutoClearHigh += 1
						} else if (*score).CurrentCycle.ScoredLow {
							(*score).CurrentScore.AutoClearLow += 1
						} else {
							(*score).CurrentScore.AutoClearDead += 1
						}
						(*score).AutoLeftoverBalls -= 1
					} else {
						(*score).CurrentScore.Cycles = append((*score).CurrentScore.Cycles, (*score).CurrentCycle)
					}
					(*score).CurrentCycle = Cycle{}
					(*score).undoCycles = []Cycle{}
				}
			}
		case "commitMatch":
			if mainArena.MatchState != POST_MATCH {
				// Don't allow committing the score until the match is over.
				continue
			}
			(*score).AutoCommitted = true
			(*score).TeleopCommitted = true
			if (*score).CurrentCycle != (Cycle{}) {
				// Commit last cycle.
				(*score).CurrentScore.Cycles = append((*score).CurrentScore.Cycles, (*score).CurrentCycle)
			}
			mainArena.scoringStatusNotifier.Notify(nil)
		case "undo":
			if !(*score).AutoCommitted && len((*score).undoAutoScores) > 0 {
				(*score).CurrentScore = (*score).undoAutoScores[len((*score).undoAutoScores)-1]
				(*score).undoAutoScores = (*score).undoAutoScores[0 : len((*score).undoAutoScores)-1]
			} else if !(*score).TeleopCommitted && len((*score).undoCycles) > 0 {
				(*score).CurrentCycle = (*score).undoCycles[len((*score).undoCycles)-1]
				(*score).undoCycles = (*score).undoCycles[0 : len((*score).undoCycles)-1]
			}
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		mainArena.realtimeScoreNotifier.Notify(nil)

		// Send out the score again after handling the command, as it most likely changed as a result.
		err = websocket.Write("score", *score)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

// Renders the referee interface for assigning fouls.
func RefereeDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/referee_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	match := mainArena.currentMatch
	matchType := match.CapitalizedType()
	red1 := mainArena.AllianceStations["R1"].team
	if red1 == nil {
		red1 = &Team{}
	}
	red2 := mainArena.AllianceStations["R2"].team
	if red2 == nil {
		red2 = &Team{}
	}
	red3 := mainArena.AllianceStations["R3"].team
	if red3 == nil {
		red3 = &Team{}
	}
	blue1 := mainArena.AllianceStations["B1"].team
	if blue1 == nil {
		blue1 = &Team{}
	}
	blue2 := mainArena.AllianceStations["B2"].team
	if blue2 == nil {
		blue2 = &Team{}
	}
	blue3 := mainArena.AllianceStations["B3"].team
	if blue3 == nil {
		blue3 = &Team{}
	}
	data := struct {
		*EventSettings
		MatchType        string
		MatchDisplayName string
		Red1             *Team
		Red2             *Team
		Red3             *Team
		Blue1            *Team
		Blue2            *Team
		Blue3            *Team
		RedFouls         []Foul
		BlueFouls        []Foul
		RedCards         map[string]string
		BlueCards        map[string]string
		Rules            []string
		EntryEnabled     bool
	}{eventSettings, matchType, match.DisplayName, red1, red2, red3, blue1, blue2, blue3,
		mainArena.redRealtimeScore.Fouls, mainArena.blueRealtimeScore.Fouls, mainArena.redRealtimeScore.Cards,
		mainArena.blueRealtimeScore.Cards, rules,
		!(mainArena.redRealtimeScore.FoulsCommitted && mainArena.blueRealtimeScore.FoulsCommitted)}
	err = template.ExecuteTemplate(w, "referee_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the refereee interface client to send control commands and receive status updates.
func RefereeDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Spin off a goroutine to listen for notifications and pass them on through the websocket.
	go func() {
		for {
			var messageType string
			var message interface{}
			select {
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		case "addFoul":
			args := struct {
				Alliance    string
				TeamId      int
				Rule        string
				IsTechnical bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := Foul{TeamId: args.TeamId, Rule: args.Rule, IsTechnical: args.IsTechnical,
				TimeInMatchSec: mainArena.MatchTimeSec()}
			if args.Alliance == "red" {
				mainArena.redRealtimeScore.Fouls = append(mainArena.redRealtimeScore.Fouls, foul)
			} else {
				mainArena.blueRealtimeScore.Fouls = append(mainArena.blueRealtimeScore.Fouls, foul)
			}
			mainArena.realtimeScoreNotifier.Notify(nil)
		case "deleteFoul":
			args := struct {
				Alliance       string
				TeamId         int
				Rule           string
				TimeInMatchSec float64
				IsTechnical    bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Remove the foul from the correct alliance's list.
			deleteFoul := Foul{TeamId: args.TeamId, Rule: args.Rule, IsTechnical: args.IsTechnical,
				TimeInMatchSec: args.TimeInMatchSec}
			var fouls *[]Foul
			if args.Alliance == "red" {
				fouls = &mainArena.redRealtimeScore.Fouls
			} else {
				fouls = &mainArena.blueRealtimeScore.Fouls
			}
			for i, foul := range *fouls {
				if foul == deleteFoul {
					*fouls = append((*fouls)[:i], (*fouls)[i+1:]...)
					break
				}
			}
			mainArena.realtimeScoreNotifier.Notify(nil)
		case "card":
			args := struct {
				Alliance string
				TeamId   int
				Card     string
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				websocket.WriteError(err.Error())
				continue
			}

			// Set the card in the correct alliance's score.
			var cards map[string]string
			if args.Alliance == "red" {
				cards = mainArena.redRealtimeScore.Cards
			} else {
				cards = mainArena.blueRealtimeScore.Cards
			}
			cards[strconv.Itoa(args.TeamId)] = args.Card
			continue
		case "commitMatch":
			if mainArena.MatchState != POST_MATCH {
				// Don't allow committing the fouls until the match is over.
				continue
			}
			mainArena.redRealtimeScore.FoulsCommitted = true
			mainArena.blueRealtimeScore.FoulsCommitted = true
			mainArena.scoringStatusNotifier.Notify(nil)
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			continue
		}

		// Force a reload of the client to render the updated foul list.
		err = websocket.Write("reload", nil)
		if err != nil {
			log.Printf("Websocket error: %s", err)
			return
		}
	}
}

// Renders the team number and status display shown above each alliance station.
func AllianceStationDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/alliance_station_display.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	displayId := ""
	if _, ok := r.URL.Query()["displayId"]; ok {
		displayId = r.URL.Query()["displayId"][0]
	}
	data := struct {
		*EventSettings
		DisplayId string
	}{eventSettings, displayId}
	err = template.ExecuteTemplate(w, "alliance_station_display.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the alliance station display client to receive status updates.
func AllianceStationDisplayWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket, err := NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer websocket.Close()

	displayId := r.URL.Query()["displayId"][0]
	station, ok := mainArena.allianceStationDisplays[displayId]
	if !ok {
		station = ""
		mainArena.allianceStationDisplays[displayId] = station
	}

	allianceStationDisplayListener := mainArena.allianceStationDisplayNotifier.Listen()
	defer close(allianceStationDisplayListener)
	matchLoadTeamsListener := mainArena.matchLoadTeamsNotifier.Listen()
	defer close(matchLoadTeamsListener)
	robotStatusListener := mainArena.robotStatusNotifier.Listen()
	defer close(robotStatusListener)
	matchTimeListener := mainArena.matchTimeNotifier.Listen()
	defer close(matchTimeListener)
	realtimeScoreListener := mainArena.realtimeScoreNotifier.Listen()
	defer close(realtimeScoreListener)
	hotGoalLightListener := mainArena.hotGoalLightNotifier.Listen()
	defer close(hotGoalLightListener)
	reloadDisplaysListener := mainArena.reloadDisplaysNotifier.Listen()
	defer close(reloadDisplaysListener)

	// Send the various notifications immediately upon connection.
	var data interface{}
	err = websocket.Write("setAllianceStationDisplay", mainArena.allianceStationDisplayScreen)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTiming", mainArena.matchTiming)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	err = websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		AllianceStation string
		Teams           map[string]*Team
	}{station, map[string]*Team{"R1": mainArena.AllianceStations["R1"].team,
		"R2": mainArena.AllianceStations["R2"].team, "R3": mainArena.AllianceStations["R3"].team,
		"B1": mainArena.AllianceStations["B1"].team, "B2": mainArena.AllianceStations["B2"].team,
		"B3": mainArena.AllianceStations["B3"].team}}
	err = websocket.Write("setMatch", data)
	if err != nil {
		log.Printf("Websocket error: %s", err)
		return
	}
	data = struct {
		RedScore  int
		RedCycle  Cycle
		BlueScore int
		BlueCycle Cycle
	}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
		mainArena.redRealtimeScore.CurrentCycle,
		mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls),
		mainArena.blueRealtimeScore.CurrentCycle}
	err = websocket.Write("realtimeScore", data)
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
			case _, ok := <-allianceStationDisplayListener:
				if !ok {
					return
				}
				websocket.Write("matchTime", MatchTimeMessage{mainArena.MatchState, int(mainArena.lastMatchTimeSec)})
				messageType = "setAllianceStationDisplay"
				message = mainArena.allianceStationDisplayScreen
			case _, ok := <-matchLoadTeamsListener:
				if !ok {
					return
				}
				messageType = "setMatch"
				station = mainArena.allianceStationDisplays[displayId]
				message = struct {
					AllianceStation string
					Teams           map[string]*Team
				}{station, map[string]*Team{station: mainArena.AllianceStations["R1"].team,
					"R2": mainArena.AllianceStations["R2"].team, "R3": mainArena.AllianceStations["R3"].team,
					"B1": mainArena.AllianceStations["B1"].team, "B2": mainArena.AllianceStations["B2"].team,
					"B3": mainArena.AllianceStations["B3"].team}}
			case _, ok := <-robotStatusListener:
				if !ok {
					return
				}
				messageType = "status"
				message = mainArena
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
					RedCycle  Cycle
					BlueScore int
					BlueCycle Cycle
				}{mainArena.redRealtimeScore.Score(mainArena.blueRealtimeScore.Fouls),
					mainArena.redRealtimeScore.CurrentCycle,
					mainArena.blueRealtimeScore.Score(mainArena.redRealtimeScore.Fouls),
					mainArena.blueRealtimeScore.CurrentCycle}
			case side, ok := <-hotGoalLightListener:
				if !ok {
					return
				}
				if !eventSettings.AllianceDisplayHotGoals {
					continue
				}
				messageType = "hotGoalLight"
				message = side
			case _, ok := <-reloadDisplaysListener:
				if !ok {
					return
				}
				messageType = "reload"
				message = nil
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
		case "setAllianceStation":
			station, ok := data.(string)
			if !ok {
				websocket.WriteError(fmt.Sprintf("Failed to parse '%s' message.", messageType))
				continue
			}
			mainArena.allianceStationDisplays[displayId] = station
		default:
			websocket.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
		}
	}
}

// Renders the FTA diagnostic display.
func FtaDisplayHandler(w http.ResponseWriter, r *http.Request) {
	template := template.New("").Funcs(templateHelpers)
	_, err := template.ParseFiles("templates/fta_display.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*EventSettings
	}{eventSettings}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}
