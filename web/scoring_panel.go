// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Web handlers for scoring interface.

package web

import (
	"fmt"
	"io"

	//"log"
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
				return
			}
			//log.Printf("[Scoring] WebSocket Read Error: %v", err)
			return
		}

		// [Debug] 印出收到的原始指令
		//log.Printf("[Scoring] received: Command=%s Data=%+v", command, data)

		// 2026 Fix: 處理前端傳來的 "type: score" 包裝層
		// 如果指令是 "score"，我們需要拆開 data 裡面的內容
		if command == "score" {
			if dataMap, ok := data.(map[string]interface{}); ok {
				// 嘗試提取內層的 command
				if innerCmd, ok := dataMap["command"].(string); ok {
					command = innerCmd
					//log.Printf("[Scoring] unpacked Command: %s", command)
				}
				// 嘗試提取內層的 data
				if innerData, ok := dataMap["data"]; ok {
					data = innerData
				}
			}
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
				//log.Printf("[Scoring] Fuel Decode Error: %v", err)
				ws.WriteError(err.Error())
				continue
			}

			//log.Printf("[Scoring] update Fuel: Auto=%v Adj=%d", args.Autonomous, args.Adjustment)

			// --- [NEW] 防呆檢查: 只有進攻方 (HubActive) 才能加分 ---
			// 注意: 如果 Adjustment 是負數(扣分修正)，通常還是允許的，即使 Hub 關閉
			if args.Adjustment > 0 && !args.Autonomous && !score.HubActive {
				// 如果這不是自動階段，且 Hub 是關閉的，且裁判嘗試加分 -> 拒絕
				ws.WriteError("Hub is INACTIVE! Cannot score Fuel.")
				//log.Printf("[Scoring] 拒絕加分: Hub Inactive")
				continue
			}
			// -----------------------------------------------------

			if args.Autonomous {
				score.AutoFuelCount = max(0, score.AutoFuelCount+args.Adjustment)
			} else {
				score.TeleopFuelCount = max(0, score.TeleopFuelCount+args.Adjustment)
			}
			scoreChanged = true

		} else if command == "climb" {
			// 2026: 處理 Endgame 爬升
			args := struct {
				RobotIndex int
				Level      int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				//log.Printf("[Scoring] Climb Decode Error: %v", err)
				ws.WriteError(err.Error())
				continue
			}

			if args.RobotIndex >= 0 && args.RobotIndex < 3 {
				var status game.EndgameStatus
				switch args.Level {
				case 1:
					status = game.EndgameLevel1
				case 2:
					status = game.EndgameLevel2
				case 3:
					status = game.EndgameLevel3
				default:
					status = game.EndgameNone
				}
				score.EndgameStatuses[args.RobotIndex] = status
				//log.Printf("[Scoring] update Climb: Robot=%d Status=%v", args.RobotIndex, status)
				scoreChanged = true
			}

		} else if command == "auto_tower" {
			// 2026: 處理 Auto 爬升
			args := struct {
				RobotIndex int
				Adjustment int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				//log.Printf("[Scoring] AutoTower Decode Error: %v", err)
				ws.WriteError(err.Error())
				continue
			}

			if args.RobotIndex >= 0 && args.RobotIndex < 3 {
				score.AutoTowerLevel1[args.RobotIndex] = (args.Adjustment > 0)
				//log.Printf("[Scoring] update AutoTower: Robot=%d Active=%v", args.RobotIndex, score.AutoTowerLevel1[args.RobotIndex])
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
				//log.Printf("[Scoring] Foul Decode Error: %v", err)
				ws.WriteError(err.Error())
				continue
			}

			foul := game.Foul{
				FoulId:  web.arena.NextFoulId,
				RuleId:  args.RuleId,
				IsMajor: args.IsMajor,
			}
			web.arena.NextFoulId++

			//log.Printf("[Scoring] add foul: %s Major=%v", args.Alliance, args.IsMajor)

			if args.Alliance == "red" {
				web.arena.RedRealtimeScore.CurrentScore.Fouls =
					append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				web.arena.BlueRealtimeScore.CurrentScore.Fouls =
					append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
			}
			web.arena.RedRealtimeScore.FoulsCommitted = true
			web.arena.BlueRealtimeScore.FoulsCommitted = true
			web.arena.RealtimeScoreNotifier.Notify()
		}

		if scoreChanged {
			//log.Println("[Scoring] score changed, sending notification")
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
