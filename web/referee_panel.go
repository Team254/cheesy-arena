// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Web handlers for the referee panel.

package web

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

// Renders the referee interface.
func (web *Web) refereePanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/referee_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// 2026 Fix: 包裝 EventSettings 以符合 base.html 的預期
	data := struct {
		*model.EventSettings
	}{web.arena.EventSettings}

	err = template.ExecuteTemplate(w, "base_no_navbar", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// 2026: 犯規列表的 HTML 片段渲染 (供 AJAX 呼叫)
func (web *Web) refereePanelFoulListHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/referee_panel_foul_list.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	// 準備資料給 Template
	data := struct {
		Match     *model.Match
		RedFouls  []game.Foul
		BlueFouls []game.Foul
		Rules     map[int]*game.Rule
	}{
		web.arena.CurrentMatch,
		web.arena.RedRealtimeScore.CurrentScore.Fouls,
		web.arena.BlueRealtimeScore.CurrentScore.Fouls,
		game.GetAllRules(),
	}

	err = template.ExecuteTemplate(w, "referee_panel_foul_list", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the referee interface.
func (web *Web) refereePanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()

	// 訂閱通知，當比賽狀態改變時推播給裁判面板
	go ws.HandleNotifiers(
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ScoringStatusNotifier,
		web.arena.ReloadDisplaysNotifier,
	)

	for {
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Printf("[Referee] WebSocket Read Error: %v", err)
			return
		}

		// 處理各種裁判指令
		if command == "addFoul" {
			args := struct {
				Alliance string
				RuleId   int
				IsMajor  bool
			}{}
			if err := mapstructure.Decode(data, &args); err != nil {
				log.Printf("[Referee] Decode Error (addFoul): %v", err)
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

			web.arena.RedRealtimeScore.FoulsCommitted = true
			web.arena.BlueRealtimeScore.FoulsCommitted = true
			web.arena.RealtimeScoreNotifier.Notify()

		} else if command == "deleteFoul" || command == "toggleFoulType" || command == "updateFoulTeam" || command == "updateFoulRule" {
			// 處理犯規列表的編輯操作
			args := struct {
				Alliance string
				Index    int
				TeamId   int
				RuleId   int
			}{}
			if err := mapstructure.Decode(data, &args); err != nil {
				log.Printf("[Referee] Decode Error (%s): %v", command, err)
				continue
			}

			// 取得對應聯盟的犯規列表指標
			var fouls *[]game.Foul
			if args.Alliance == "red" {
				fouls = &web.arena.RedRealtimeScore.CurrentScore.Fouls
			} else {
				fouls = &web.arena.BlueRealtimeScore.CurrentScore.Fouls
			}

			if args.Index >= 0 && args.Index < len(*fouls) {
				switch command {
				case "deleteFoul":
					*fouls = append((*fouls)[:args.Index], (*fouls)[args.Index+1:]...)
				case "toggleFoulType":
					(*fouls)[args.Index].IsMajor = !(*fouls)[args.Index].IsMajor
					(*fouls)[args.Index].RuleId = 0 // 重置規則，因為規則通常綁定 Major/Minor
				case "updateFoulTeam":
					if (*fouls)[args.Index].TeamId == args.TeamId {
						(*fouls)[args.Index].TeamId = 0 // Toggle off if same team clicked
					} else {
						(*fouls)[args.Index].TeamId = args.TeamId
					}
				case "updateFoulRule":
					(*fouls)[args.Index].RuleId = args.RuleId
				}
				web.arena.RealtimeScoreNotifier.Notify()
			}

		} else if command == "card" {
			args := struct {
				Alliance string
				TeamId   int
				Card     string
			}{}
			if err := mapstructure.Decode(data, &args); err != nil {
				log.Printf("[Referee] Decode Error (card): %v", err)
				continue
			}

			// 1. 取得對應聯盟的卡片 Map 指標
			var cards map[string]string
			if args.Alliance == "red" {
				cards = web.arena.RedRealtimeScore.Cards
			} else {
				cards = web.arena.BlueRealtimeScore.Cards
			}

			// 2. 判斷比賽類型：季後賽 (Playoff) 卡片對全聯盟生效
			if web.arena.CurrentMatch.Type == model.Playoff {
				// 取得該聯盟所有隊伍 ID
				var teamIds []int
				if args.Alliance == "red" {
					teamIds = []int{web.arena.CurrentMatch.Red1, web.arena.CurrentMatch.Red2, web.arena.CurrentMatch.Red3}
				} else {
					teamIds = []int{web.arena.CurrentMatch.Blue1, web.arena.CurrentMatch.Blue2, web.arena.CurrentMatch.Blue3}
				}

				// 對聯盟內所有隊伍進行操作
				for _, id := range teamIds {
					if id == 0 {
						continue
					}
					teamStr := strconv.Itoa(id)
					if args.Card == "none" {
						delete(cards, teamStr)
					} else {
						cards[teamStr] = args.Card
					}
				}
			} else {
				// 3. 例行賽 (Qualification)：卡片只對該隊伍生效
				teamStr := strconv.Itoa(args.TeamId)
				if args.Card == "none" {
					delete(cards, teamStr)
				} else {
					cards[teamStr] = args.Card
				}
			}

			// 4. 通知更新
			web.arena.RealtimeScoreNotifier.Notify()

		} else if command == "signalVolunteers" {
			if web.arena.MatchState != field.PostMatch {
				continue
			}
			web.arena.FieldVolunteers = true
			web.arena.AllianceStationDisplayMode = "signalCount"
			web.arena.AllianceStationDisplayModeNotifier.Notify()

		} else if command == "signalReset" {
			if web.arena.MatchState != field.PostMatch {
				continue
			}
			web.arena.FieldVolunteers = false
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()

		} else if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				continue
			}
			web.arena.RedRealtimeScore.FoulsCommitted = true
			web.arena.BlueRealtimeScore.FoulsCommitted = true
			web.arena.FieldVolunteers = false
			web.arena.FieldReset = true
			web.arena.AllianceStationDisplayMode = "fieldReset"
			web.arena.AllianceStationDisplayModeNotifier.Notify()
			web.arena.ScoringStatusNotifier.Notify()
		}
	}
}
