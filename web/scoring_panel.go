// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

type ScoringPosition struct {
	Title         string
	Alliance      string
	NearSide      bool
	ScoresAuto    bool
	ScoresEndgame bool
	// 2026 移除 Barge/Processor 相關欄位
}

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

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	position := r.PathValue("position")
	parameters, ok := positionParameters[position]
	if !ok {
		handleWebErr(w, fmt.Errorf("Invalid position '%s'.", position))
		return
	}

	template, err := web.parseFiles("templates/scoring_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		PlcIsEnabled bool
		PositionName string
		Position     ScoringPosition
	}{web.arena.EventSettings, web.arena.Plc.IsEnabled(), position, parameters}
	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	position := r.PathValue("position")
	if _, ok := positionParameters[position]; !ok {
		handleWebErr(w, fmt.Errorf("Invalid position '%s'.", position))
		return
	}
	alliance := strings.Split(position, "_")[0]

	var realtimeScore **field.RealtimeScore
	if alliance == "red" {
		realtimeScore = &web.arena.RedRealtimeScore
	} else {
		realtimeScore = &web.arena.BlueRealtimeScore
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(position, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(position, ws)

	// Instruct panel to clear any local state in case this is a reconnect
	ws.Write("resetLocalState", nil)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ReloadDisplaysNotifier,
	)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}
		score := &(*realtimeScore).CurrentScore
		scoreChanged := false

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(position, ws)
			web.arena.ScoringStatusNotifier.Notify()

		} else if command == "fuel" {
			// 2026: 處理投球 (Fuel)
			args := struct {
				Adjustment int
				Autonomous bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.Autonomous {
				score.AutoFuelCount = max(0, score.AutoFuelCount+args.Adjustment)
			} else {
				score.TeleopFuelCount = max(0, score.TeleopFuelCount+args.Adjustment)
			}
			scoreChanged = true

		} else if command == "climb" {
			// 2026: 處理 Endgame 爬升 (Level 2/3)
			args := struct {
				RobotIndex int // 0-2 (注意前端可能傳 1-3，需確認)
				Level      int // 0, 2, 3
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// 這裡假設前端傳來的是 0-2 的 RobotIndex，如果是 1-3 請自行減 1
			if args.RobotIndex >= 0 && args.RobotIndex < 3 {
				var status game.EndgameStatus
				switch args.Level {
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

		} else if command == "auto_tower" {
			// 2026: 處理 Auto 爬升 (Level 1)
			args := struct {
				RobotIndex int
				Adjustment int // 使用 +1/-1 來代表 True/False 切換，或直接傳 Bool
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			if args.RobotIndex >= 0 && args.RobotIndex < 3 {
				// 簡單切換：如果 Adjustment > 0 設為 true，否則 false
				// 或者是 toggle 邏輯，視前端實作而定。這裡假設是設定值。
				score.AutoTowerLevel1[args.RobotIndex] = (args.Adjustment > 0)
				scoreChanged = true
			}

		} else if command == "addFoul" {
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
			if args.Alliance == "red" {
				web.arena.RedRealtimeScore.CurrentScore.Fouls =
					append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				web.arena.BlueRealtimeScore.CurrentScore.Fouls =
					append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
			}
			// 更新雙方的 Foul 提交狀態
			web.arena.RedRealtimeScore.FoulsCommitted = true
			web.arena.BlueRealtimeScore.FoulsCommitted = true
			web.arena.RealtimeScoreNotifier.Notify()
		}

		if scoreChanged {
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
