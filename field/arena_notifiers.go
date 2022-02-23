// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains configuration of the publish-subscribe notifiers that allow the arena to push updates to websocket clients.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
	"github.com/Team254/cheesy-arena/websocket"
	"strconv"
)

type ArenaNotifiers struct {
	AllianceSelectionNotifier          *websocket.Notifier
	AllianceStationDisplayModeNotifier *websocket.Notifier
	ArenaStatusNotifier                *websocket.Notifier
	AudienceDisplayModeNotifier        *websocket.Notifier
	DisplayConfigurationNotifier       *websocket.Notifier
	EventStatusNotifier                *websocket.Notifier
	LowerThirdNotifier                 *websocket.Notifier
	MatchLoadNotifier                  *websocket.Notifier
	MatchTimeNotifier                  *websocket.Notifier
	MatchTimingNotifier                *websocket.Notifier
	PlaySoundNotifier                  *websocket.Notifier
	RealtimeScoreNotifier              *websocket.Notifier
	ReloadDisplaysNotifier             *websocket.Notifier
	ScorePostedNotifier                *websocket.Notifier
	ScoringStatusNotifier              *websocket.Notifier
}

type MatchTimeMessage struct {
	MatchState
	MatchTimeSec int
}

type audienceAllianceScoreFields struct {
	Score        *game.Score
	ScoreSummary *game.ScoreSummary
}

// Instantiates notifiers and configures their message producing methods.
func (arena *Arena) configureNotifiers() {
	arena.AllianceSelectionNotifier = websocket.NewNotifier("allianceSelection", arena.generateAllianceSelectionMessage)
	arena.AllianceStationDisplayModeNotifier = websocket.NewNotifier("allianceStationDisplayMode",
		arena.generateAllianceStationDisplayModeMessage)
	arena.ArenaStatusNotifier = websocket.NewNotifier("arenaStatus", arena.generateArenaStatusMessage)
	arena.AudienceDisplayModeNotifier = websocket.NewNotifier("audienceDisplayMode",
		arena.generateAudienceDisplayModeMessage)
	arena.DisplayConfigurationNotifier = websocket.NewNotifier("displayConfiguration",
		arena.generateDisplayConfigurationMessage)
	arena.EventStatusNotifier = websocket.NewNotifier("eventStatus", arena.generateEventStatusMessage)
	arena.LowerThirdNotifier = websocket.NewNotifier("lowerThird", arena.generateLowerThirdMessage)
	arena.MatchLoadNotifier = websocket.NewNotifier("matchLoad", arena.generateMatchLoadMessage)
	arena.MatchTimeNotifier = websocket.NewNotifier("matchTime", arena.generateMatchTimeMessage)
	arena.MatchTimingNotifier = websocket.NewNotifier("matchTiming", arena.generateMatchTimingMessage)
	arena.PlaySoundNotifier = websocket.NewNotifier("playSound", nil)
	arena.RealtimeScoreNotifier = websocket.NewNotifier("realtimeScore", arena.generateRealtimeScoreMessage)
	arena.ReloadDisplaysNotifier = websocket.NewNotifier("reload", nil)
	arena.ScorePostedNotifier = websocket.NewNotifier("scorePosted", arena.generateScorePostedMessage)
	arena.ScoringStatusNotifier = websocket.NewNotifier("scoringStatus", arena.generateScoringStatusMessage)
}

func (arena *Arena) generateAllianceSelectionMessage() interface{} {
	return &arena.AllianceSelectionAlliances
}

func (arena *Arena) generateAllianceStationDisplayModeMessage() interface{} {
	return arena.AllianceStationDisplayMode
}

func (arena *Arena) generateArenaStatusMessage() interface{} {
	// Convert AP team wifi network status array to a map by station for ease of client use.
	teamWifiStatuses := make(map[string]network.TeamWifiStatus)
	for i, station := range []string{"R1", "R2", "R3", "B1", "B2", "B3"} {
		if arena.EventSettings.Ap2TeamChannel == 0 || i < 3 {
			teamWifiStatuses[station] = arena.accessPoint.TeamWifiStatuses[i]
		} else {
			teamWifiStatuses[station] = arena.accessPoint2.TeamWifiStatuses[i]
		}
	}

	return &struct {
		MatchId          int
		AllianceStations map[string]*AllianceStation
		TeamWifiStatuses map[string]network.TeamWifiStatus
		MatchState
		CanStartMatch         bool
		PlcIsHealthy          bool
		FieldEstop            bool
		PlcArmorBlockStatuses map[string]bool
	}{arena.CurrentMatch.Id, arena.AllianceStations, teamWifiStatuses, arena.MatchState,
		arena.checkCanStartMatch() == nil, arena.Plc.IsHealthy, arena.Plc.GetFieldEstop(),
		arena.Plc.GetArmorBlockStatuses()}
}

func (arena *Arena) generateAudienceDisplayModeMessage() interface{} {
	return arena.AudienceDisplayMode
}

func (arena *Arena) generateDisplayConfigurationMessage() interface{} {
	// Notify() for this notifier must always called from a method that has a lock on the display mutex.
	// Make a copy of the map to avoid potential data races; otherwise the same map would get iterated through as it is
	// serialized to JSON, outside the mutex lock.
	displaysCopy := make(map[string]Display)
	for displayId, display := range arena.Displays {
		displaysCopy[displayId] = *display
	}
	return displaysCopy
}

func (arena *Arena) generateEventStatusMessage() interface{} {
	return arena.EventStatus
}

func (arena *Arena) generateLowerThirdMessage() interface{} {
	return &struct {
		LowerThird     *model.LowerThird
		ShowLowerThird bool
	}{arena.LowerThird, arena.ShowLowerThird}
}

func (arena *Arena) generateMatchLoadMessage() interface{} {
	teams := make(map[string]*model.Team)
	for station, allianceStation := range arena.AllianceStations {
		teams[station] = allianceStation.Team
	}

	rankings := make(map[string]*game.Ranking)
	for _, allianceStation := range arena.AllianceStations {
		if allianceStation.Team != nil {
			rankings[strconv.Itoa(allianceStation.Team.Id)], _ =
				arena.Database.GetRankingForTeam(allianceStation.Team.Id)
		}
	}

	return &struct {
		MatchType string
		Match     *model.Match
		Teams     map[string]*model.Team
		Rankings  map[string]*game.Ranking
	}{arena.CurrentMatch.CapitalizedType(), arena.CurrentMatch, teams, rankings}
}

func (arena *Arena) generateMatchTimeMessage() interface{} {
	return MatchTimeMessage{arena.MatchState, int(arena.MatchTimeSec())}
}

func (arena *Arena) generateMatchTimingMessage() interface{} {
	return &game.MatchTiming
}

func (arena *Arena) generateRealtimeScoreMessage() interface{} {
	fields := struct {
		Red  *audienceAllianceScoreFields
		Blue *audienceAllianceScoreFields
		MatchState
	}{}
	fields.Red = getAudienceAllianceScoreFields(arena.RedRealtimeScore, arena.RedScoreSummary())
	fields.Blue = getAudienceAllianceScoreFields(arena.BlueRealtimeScore, arena.BlueScoreSummary())
	fields.MatchState = arena.MatchState
	return &fields
}

func (arena *Arena) generateScorePostedMessage() interface{} {
	// For elimination matches, summarize the state of the series.
	var seriesStatus, seriesLeader string
	if arena.SavedMatch.Type == "elimination" {
		matches, _ := arena.Database.GetMatchesByElimRoundGroup(arena.SavedMatch.ElimRound, arena.SavedMatch.ElimGroup)
		var redWins, blueWins int
		for _, match := range matches {
			if match.Status == model.RedWonMatch {
				redWins++
			} else if match.Status == model.BlueWonMatch {
				blueWins++
			}
		}

		if redWins == 2 {
			seriesStatus = fmt.Sprintf("Red Wins Series %d-%d", redWins, blueWins)
			seriesLeader = "red"
		} else if blueWins == 2 {
			seriesStatus = fmt.Sprintf("Blue Wins Series %d-%d", blueWins, redWins)
			seriesLeader = "blue"
		} else if redWins > blueWins {
			seriesStatus = fmt.Sprintf("Red Leads Series %d-%d", redWins, blueWins)
			seriesLeader = "red"
		} else if blueWins > redWins {
			seriesStatus = fmt.Sprintf("Blue Leads Series %d-%d", blueWins, redWins)
			seriesLeader = "blue"
		} else {
			seriesStatus = fmt.Sprintf("Series Tied %d-%d", redWins, blueWins)
		}
	}

	rankings := make(map[int]game.Ranking, len(arena.SavedRankings))
	for _, ranking := range arena.SavedRankings {
		rankings[ranking.TeamId] = ranking
	}

	return &struct {
		MatchType        string
		Match            *model.Match
		RedScoreSummary  *game.ScoreSummary
		BlueScoreSummary *game.ScoreSummary
		Rankings         map[int]game.Ranking
		RedFouls         []game.Foul
		BlueFouls        []game.Foul
		RulesViolated    map[int]*game.Rule
		RedCards         map[string]string
		BlueCards        map[string]string
		SeriesStatus     string
		SeriesLeader     string
	}{arena.SavedMatch.CapitalizedType(), arena.SavedMatch, arena.SavedMatchResult.RedScoreSummary(),
		arena.SavedMatchResult.BlueScoreSummary(), rankings, arena.SavedMatchResult.RedScore.Fouls,
		arena.SavedMatchResult.BlueScore.Fouls,
		getRulesViolated(arena.SavedMatchResult.RedScore.Fouls, arena.SavedMatchResult.BlueScore.Fouls),
		arena.SavedMatchResult.RedCards, arena.SavedMatchResult.BlueCards, seriesStatus, seriesLeader}
}

func (arena *Arena) generateScoringStatusMessage() interface{} {
	return &struct {
		RefereeScoreReady         bool
		RedScoreReady             bool
		BlueScoreReady            bool
		NumRedScoringPanels       int
		NumRedScoringPanelsReady  int
		NumBlueScoringPanels      int
		NumBlueScoringPanelsReady int
	}{arena.RedRealtimeScore.FoulsCommitted && arena.BlueRealtimeScore.FoulsCommitted,
		arena.alliancePostMatchScoreReady("red"), arena.alliancePostMatchScoreReady("blue"),
		arena.ScoringPanelRegistry.GetNumPanels("red"), arena.ScoringPanelRegistry.GetNumScoreCommitted("red"),
		arena.ScoringPanelRegistry.GetNumPanels("blue"), arena.ScoringPanelRegistry.GetNumScoreCommitted("blue")}
}

// Constructs the data object for one alliance sent to the audience display for the realtime scoring overlay.
func getAudienceAllianceScoreFields(allianceScore *RealtimeScore,
	allianceScoreSummary *game.ScoreSummary) *audienceAllianceScoreFields {
	fields := new(audienceAllianceScoreFields)
	fields.Score = &allianceScore.CurrentScore
	fields.ScoreSummary = allianceScoreSummary
	return fields
}

// Produce a map of rules that were violated by either alliance so that they are available to the announcer.
func getRulesViolated(redFouls, blueFouls []game.Foul) map[int]*game.Rule {
	rules := make(map[int]*game.Rule)
	for _, foul := range redFouls {
		rules[foul.RuleId] = game.GetRuleById(foul.RuleId)
	}
	for _, foul := range blueFouls {
		rules[foul.RuleId] = game.GetRuleById(foul.RuleId)
	}
	return rules
}
