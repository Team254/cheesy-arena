// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game
//
// Web handlers for the referee panel.

package web

import (
	"net/http"
	"strconv"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/mitchellh/mapstructure"
)

func (web *Web) refereePanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/referee_panel.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	// 傳遞 EventSettings 以便前端知道比賽設定
	err = template.ExecuteTemplate(w, "base_no_navbar", web.arena.EventSettings)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

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

	// 註冊 Notifiers，讓裁判能即時看到比賽狀態與分數
	go ws.HandleNotifiers(
		web.arena.MatchLoadNotifier,
		web.arena.MatchTimeNotifier,
		web.arena.RealtimeScoreNotifier,
		web.arena.ArenaStatusNotifier,
	)

	for {
		command, data, err := ws.Read()
		if err != nil {
			break
		}

		if command == "addFoul" {
			// 處理犯規
			args := struct {
				Alliance string
				RuleId   int
				IsMajor  bool
			}{}
			if err := mapstructure.Decode(data, &args); err != nil {
				continue
			}

			foul := game.Foul{
				FoulId:  web.arena.NextFoulId,
				RuleId:  args.RuleId,
				IsMajor: args.IsMajor,
			}
			web.arena.NextFoulId++

			// 根據聯盟寫入犯規
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

		} else if command == "assignCard" {
			// 處理黃牌/紅牌
			args := struct {
				TeamId string
				Type   string // "yellow", "red", or "none"
			}{}
			if err := mapstructure.Decode(data, &args); err != nil {
				continue
			}

			// 檢查該隊伍是在紅隊還是藍隊
			isRed := false
			isBlue := false
			teamIdInt, _ := strconv.Atoi(args.TeamId)

			// 簡單遍歷檢查隊伍歸屬
			if web.arena.CurrentMatch.Red1 == teamIdInt || web.arena.CurrentMatch.Red2 == teamIdInt || web.arena.CurrentMatch.Red3 == teamIdInt {
				isRed = true
			} else if web.arena.CurrentMatch.Blue1 == teamIdInt || web.arena.CurrentMatch.Blue2 == teamIdInt || web.arena.CurrentMatch.Blue3 == teamIdInt {
				isBlue = true
			}

			if isRed {
				if args.Type == "none" {
					delete(web.arena.RedRealtimeScore.Cards, args.TeamId)
				} else {
					web.arena.RedRealtimeScore.Cards[args.TeamId] = args.Type
				}
			} else if isBlue {
				if args.Type == "none" {
					delete(web.arena.BlueRealtimeScore.Cards, args.TeamId)
				} else {
					web.arena.BlueRealtimeScore.Cards[args.TeamId] = args.Type
				}
			}
			web.arena.RealtimeScoreNotifier.Notify()
		}
	}
}
