// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"net/http"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

type ScoringPosition struct {
	Title         string
	Alliance      string
	NearSide      bool
	ScoresAuto    bool
	ScoresEndgame bool
}

// 2026 Configuration: Removed Barge/Processor specific flags
var positionParameters = map[string]ScoringPosition{
	"red_near": {
		Title:         "Red Near",
		Alliance:      "red",
		NearSide:      true,
		ScoresAuto:    true,
		ScoresEndgame: true,
	},
	"red_far": {
		Title:         "Red Far",
		Alliance:      "red",
		NearSide:      false,
		ScoresAuto:    true,
		ScoresEndgame: true,
	},
	"blue_near": {
		Title:         "Blue Near",
		Alliance:      "blue",
		NearSide:      true,
		ScoresAuto:    true,
		ScoresEndgame: true,
	},
	"blue_far": {
		Title:         "Blue Far",
		Alliance:      "blue",
		NearSide:      false,
		ScoresAuto:    true,
		ScoresEndgame: true,
	},
}

func (web *Web) scoringGetHandler(w http.ResponseWriter, r *http.Request) {
	position := r.FormValue("position")
	if _, ok := positionParameters[position]; !ok {
		http.Error(w, fmt.Sprintf("Invalid position: %s", position), 404)
		return
	}

	template, err := web.parseFiles("templates/scoring.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	err = template.ExecuteTemplate(w, "base", positionParameters[position])
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

func (web *Web) scoringHandler(ws *websocket.WebSocket) {
	position := ws.Request.FormValue("position")
	params, ok := positionParameters[position]
	if !ok {
		ws.WriteError(fmt.Sprintf("Invalid position: %s", position))
		return
	}

	// Listen for score updates.
	for {
		message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		if message.Type == "subscribe" {
			web.arena.RealtimeScoreNotifier.Register(ws)
			web.arena.ArenaStatusNotifier.Register(ws)
		} else if message.Type == "score" {
			if web.arena.MatchState == field.PreMatch || web.arena.MatchState == field.PostMatch ||
				web.arena.MatchState == field.TimeoutActive || web.arena.MatchState == field.PostTimeout {
				// Don't allow score updates when a match is not in progress.
				continue
			}

			// Determine which alliance's score to update.
			var score *game.Score
			if params.Alliance == "red" {
				score = &web.arena.RedRealtimeScore.CurrentScore
			} else {
				score = &web.arena.BlueRealtimeScore.CurrentScore
			}

			command := message.Data.(map[string]interface{})["command"].(string)
			data := message.Data.(map[string]interface{})["data"]
			scoreChanged := false

			if command == "foul" {
				// Handle fouls (unchanged from 2025 logic mostly)
				args := struct {
					Alliance string
					RuleId   int
					IsMajor  bool
				}{}
				err = mapstructure.Decode(data, &args)
				if err != nil {
					ws.WriteError(err.Error())
					continue
				}

				foul := game.Foul{
					FoulId:  web.arena.NextFoulId,
					RuleId:  args.RuleId,
					IsMajor: args.IsMajor,
				}
				web.arena.NextFoulId++

				// Invert logic: Panel sends who committed the foul, we add it to THAT alliance's foul list.
				// (Note: score.go logic calculates points FROM opponent's fouls)
				if args.Alliance == "red" {
					web.arena.RedRealtimeScore.CurrentScore.Fouls = append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
				} else {
					web.arena.BlueRealtimeScore.CurrentScore.Fouls = append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
				}
				web.arena.RedRealtimeScore.FoulsCommitted = true
				web.arena.BlueRealtimeScore.FoulsCommitted = true
				web.arena.RealtimeScoreNotifier.Notify()
				continue
			} else if command == "commit" {
				// Commit logic
				if params.Alliance == "red" {
					web.arena.SetNumScoreCommitted(position, web.arena.GetNumScoreCommitted(position)+1)
				} else {
					web.arena.SetNumScoreCommitted(position, web.arena.GetNumScoreCommitted(position)+1)
				}
				web.arena.ScoringStatusNotifier.Notify()
				continue
			}

			// --- 2026 Game Logic ---
			args := struct {
				Adjustment int
				RobotIndex int // 0, 1, 2
				Level      int // For Climb
				Autonomous bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			switch command {
			case "fuel":
				// Add or subtract fuel
				if args.Autonomous {
					score.AutoFuelCount = max(0, score.AutoFuelCount+args.Adjustment)
				} else {
					score.TeleopFuelCount = max(0, score.TeleopFuelCount+args.Adjustment)
				}
				scoreChanged = true

			case "climb":
				// Set climb status for a specific robot
				if args.RobotIndex >= 0 && args.RobotIndex < 3 {
					// Map integer level to Enum
					var status game.EndgameStatus
					switch args.Level {
					case 0:
						status = game.EndgameNone
					case 2:
						status = game.EndgameLevel2
					case 3:
						status = game.EndgameLevel3
					default:
						status = game.EndgameNone
					}
					score.EndgameStatuses[args.RobotIndex] = status
					scoreChanged = true
				}

			case "auto_tower":
				// Set Auto Tower Level 1 status
				if args.RobotIndex >= 0 && args.RobotIndex < 3 {
					// Using Adjustment as boolean (1=true, 0=false)
					score.AutoTowerLevel1[args.RobotIndex] = (args.Adjustment > 0)
					scoreChanged = true
				}
			}

			if scoreChanged {
				web.arena.RealtimeScoreNotifier.Notify()
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
