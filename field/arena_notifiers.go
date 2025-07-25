// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains configuration of the publish-subscribe notifiers that allow the arena to push updates to websocket clients.

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/playoff"
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
	arena.AllianceStationDisplayModeNotifier = websocket.NewNotifier(
		"allianceStationDisplayMode", arena.generateAllianceStationDisplayModeMessage,
	)
	arena.ArenaStatusNotifier = websocket.NewNotifier("arenaStatus", arena.generateArenaStatusMessage)
	arena.AudienceDisplayModeNotifier = websocket.NewNotifier(
		"audienceDisplayMode", arena.generateAudienceDisplayModeMessage,
	)
	arena.DisplayConfigurationNotifier = websocket.NewNotifier(
		"displayConfiguration", arena.generateDisplayConfigurationMessage,
	)
	arena.EventStatusNotifier = websocket.NewNotifier("eventStatus", arena.generateEventStatusMessage)
	arena.LowerThirdNotifier = websocket.NewNotifier("lowerThird", arena.generateLowerThirdMessage)
	arena.MatchLoadNotifier = websocket.NewNotifier("matchLoad", arena.GenerateMatchLoadMessage)
	arena.MatchTimeNotifier = websocket.NewNotifier("matchTime", arena.generateMatchTimeMessage)
	arena.MatchTimingNotifier = websocket.NewNotifier("matchTiming", arena.generateMatchTimingMessage)
	arena.PlaySoundNotifier = websocket.NewNotifier("playSound", nil)
	arena.RealtimeScoreNotifier = websocket.NewNotifier("realtimeScore", arena.generateRealtimeScoreMessage)
	arena.ReloadDisplaysNotifier = websocket.NewNotifier("reload", nil)
	arena.ScorePostedNotifier = websocket.NewNotifier("scorePosted", arena.GenerateScorePostedMessage)
	arena.ScoringStatusNotifier = websocket.NewNotifier("scoringStatus", arena.generateScoringStatusMessage)
}

func (arena *Arena) generateAllianceSelectionMessage() any {
	return &struct {
		Alliances        []model.Alliance
		ShowTimer        bool
		TimeRemainingSec int
		RankedTeams      []model.AllianceSelectionRankedTeam
	}{
		arena.AllianceSelectionAlliances,
		arena.AllianceSelectionShowTimer,
		arena.AllianceSelectionTimeRemainingSec,
		arena.AllianceSelectionRankedTeams,
	}
}

func (arena *Arena) generateAllianceStationDisplayModeMessage() any {
	return arena.AllianceStationDisplayMode
}

func (arena *Arena) generateArenaStatusMessage() any {
	return &struct {
		MatchId          int
		AllianceStations map[string]*AllianceStation
		MatchState
		CanStartMatch         bool
		AccessPointStatus     string
		SwitchStatus          string
		RedSCCStatus          string
		BlueSCCStatus         string
		PlcIsHealthy          bool
		FieldEStop            bool
		PlcArmorBlockStatuses map[string]bool
	}{
		arena.CurrentMatch.Id,
		arena.AllianceStations,
		arena.MatchState,
		arena.checkCanStartMatch() == nil,
		arena.accessPoint.Status,
		arena.networkSwitch.Status,
		arena.redSCC.Status,
		arena.blueSCC.Status,
		arena.Plc.IsHealthy(),
		arena.Plc.GetFieldEStop(),
		arena.Plc.GetArmorBlockStatuses(),
	}
}

func (arena *Arena) generateAudienceDisplayModeMessage() any {
	return arena.AudienceDisplayMode
}

func (arena *Arena) generateDisplayConfigurationMessage() any {
	// Notify() for this notifier must always called from a method that has a lock on the display mutex.
	// Make a copy of the map to avoid potential data races; otherwise the same map would get iterated through as it is
	// serialized to JSON, outside the mutex lock.
	displaysCopy := make(map[string]Display)
	for displayId, display := range arena.Displays {
		displaysCopy[displayId] = *display
	}
	return displaysCopy
}

func (arena *Arena) generateEventStatusMessage() any {
	return arena.EventStatus
}

func (arena *Arena) generateLowerThirdMessage() any {
	return &struct {
		LowerThird     *model.LowerThird
		ShowLowerThird bool
	}{arena.LowerThird, arena.ShowLowerThird}
}

func (arena *Arena) GenerateMatchLoadMessage() any {
	teams := make(map[string]*model.Team)
	var allTeamIds []int
	for station, allianceStation := range arena.AllianceStations {
		teams[station] = allianceStation.Team
		if allianceStation.Team != nil {
			allTeamIds = append(allTeamIds, allianceStation.Team.Id)
		}
	}

	matchResult, _ := arena.Database.GetMatchResultForMatch(arena.CurrentMatch.Id)
	isReplay := matchResult != nil

	var matchup *playoff.Matchup
	redOffFieldTeams := []*model.Team{}
	blueOffFieldTeams := []*model.Team{}
	if arena.CurrentMatch.Type == model.Playoff {
		matchGroup := arena.PlayoffTournament.MatchGroups()[arena.CurrentMatch.PlayoffMatchGroupId]
		matchup, _ = matchGroup.(*playoff.Matchup)
		redOffFieldTeamIds, blueOffFieldTeamIds, _ := arena.Database.GetOffFieldTeamIds(arena.CurrentMatch)
		for _, teamId := range redOffFieldTeamIds {
			team, _ := arena.Database.GetTeamById(teamId)
			redOffFieldTeams = append(redOffFieldTeams, team)
			allTeamIds = append(allTeamIds, teamId)
		}
		for _, teamId := range blueOffFieldTeamIds {
			team, _ := arena.Database.GetTeamById(teamId)
			blueOffFieldTeams = append(blueOffFieldTeams, team)
			allTeamIds = append(allTeamIds, teamId)
		}
	}

	rankings := make(map[string]int)
	for _, teamId := range allTeamIds {
		ranking, _ := arena.Database.GetRankingForTeam(teamId)
		if ranking != nil {
			rankings[strconv.Itoa(teamId)] = ranking.Rank
		}
	}

	return &struct {
		Match             *model.Match
		AllowSubstitution bool
		IsReplay          bool
		Teams             map[string]*model.Team
		Rankings          map[string]int
		Matchup           *playoff.Matchup
		RedOffFieldTeams  []*model.Team
		BlueOffFieldTeams []*model.Team
		BreakDescription  string
	}{
		arena.CurrentMatch,
		arena.CurrentMatch.ShouldAllowSubstitution(),
		isReplay,
		teams,
		rankings,
		matchup,
		redOffFieldTeams,
		blueOffFieldTeams,
		arena.breakDescription,
	}
}

func (arena *Arena) generateMatchTimeMessage() any {
	return MatchTimeMessage{arena.MatchState, int(arena.MatchTimeSec())}
}

func (arena *Arena) generateMatchTimingMessage() any {
	return &game.MatchTiming
}

func (arena *Arena) generateRealtimeScoreMessage() any {
	fields := struct {
		Red       *audienceAllianceScoreFields
		Blue      *audienceAllianceScoreFields
		RedCards  map[string]string
		BlueCards map[string]string
		MatchState
	}{
		getAudienceAllianceScoreFields(arena.RedRealtimeScore, arena.RedScoreSummary()),
		getAudienceAllianceScoreFields(arena.BlueRealtimeScore, arena.BlueScoreSummary()),
		arena.RedRealtimeScore.Cards,
		arena.BlueRealtimeScore.Cards,
		arena.MatchState,
	}
	return &fields
}

func (arena *Arena) GenerateScorePostedMessage() any {
	redScoreSummary := arena.SavedMatchResult.RedScoreSummary()
	blueScoreSummary := arena.SavedMatchResult.BlueScoreSummary()
	redRankingPoints := redScoreSummary.BonusRankingPoints
	blueRankingPoints := blueScoreSummary.BonusRankingPoints
	switch arena.SavedMatch.Status {
	case game.RedWonMatch:
		redRankingPoints += 3
	case game.BlueWonMatch:
		blueRankingPoints += 3
	case game.TieMatch:
		redRankingPoints++
		blueRankingPoints++
	}

	// For playoff matches, summarize the state of the series.
	var redWins, blueWins int
	var redDestination, blueDestination string
	redOffFieldTeamIds := []int{}
	blueOffFieldTeamIds := []int{}
	if arena.SavedMatch.Type == model.Playoff {
		matchGroup := arena.PlayoffTournament.MatchGroups()[arena.SavedMatch.PlayoffMatchGroupId]
		if matchup, ok := matchGroup.(*playoff.Matchup); ok {
			redWins = matchup.RedAllianceWins
			blueWins = matchup.BlueAllianceWins
			redDestination = matchup.RedAllianceDestination()
			blueDestination = matchup.BlueAllianceDestination()
		}
		redOffFieldTeamIds, blueOffFieldTeamIds, _ = arena.Database.GetOffFieldTeamIds(arena.SavedMatch)
	}

	redRankings := map[int]*game.Ranking{
		arena.SavedMatch.Red1: nil, arena.SavedMatch.Red2: nil, arena.SavedMatch.Red3: nil,
	}
	blueRankings := map[int]*game.Ranking{
		arena.SavedMatch.Blue1: nil, arena.SavedMatch.Blue2: nil, arena.SavedMatch.Blue3: nil,
	}
	for index, ranking := range arena.SavedRankings {
		if _, ok := redRankings[ranking.TeamId]; ok {
			redRankings[ranking.TeamId] = &arena.SavedRankings[index]
		}
		if _, ok := blueRankings[ranking.TeamId]; ok {
			blueRankings[ranking.TeamId] = &arena.SavedRankings[index]
		}
	}

	return &struct {
		Match               *model.Match
		RedScoreSummary     *game.ScoreSummary
		BlueScoreSummary    *game.ScoreSummary
		RedRankingPoints    int
		BlueRankingPoints   int
		RedFouls            []game.Foul
		BlueFouls           []game.Foul
		RulesViolated       map[int]*game.Rule
		RedCards            map[string]string
		BlueCards           map[string]string
		RedRankings         map[int]*game.Ranking
		BlueRankings        map[int]*game.Ranking
		RedOffFieldTeamIds  []int
		BlueOffFieldTeamIds []int
		RedWon              bool
		BlueWon             bool
		RedWins             int
		BlueWins            int
		RedDestination      string
		BlueDestination     string
		CoopertitionEnabled bool
	}{
		arena.SavedMatch,
		redScoreSummary,
		blueScoreSummary,
		redRankingPoints,
		blueRankingPoints,
		arena.SavedMatchResult.RedScore.Fouls,
		arena.SavedMatchResult.BlueScore.Fouls,
		getRulesViolated(arena.SavedMatchResult.RedScore.Fouls, arena.SavedMatchResult.BlueScore.Fouls),
		arena.SavedMatchResult.RedCards,
		arena.SavedMatchResult.BlueCards,
		redRankings,
		blueRankings,
		redOffFieldTeamIds,
		blueOffFieldTeamIds,
		arena.SavedMatch.Status == game.RedWonMatch,
		arena.SavedMatch.Status == game.BlueWonMatch,
		redWins,
		blueWins,
		redDestination,
		blueDestination,
		game.CoralBonusCoopEnabled,
	}
}

func (arena *Arena) generateScoringStatusMessage() any {
	type positionStatus struct {
		Ready          bool
		NumPanels      int
		NumPanelsReady int
	}
	getStatusForPosition := func(position string) positionStatus {
		return positionStatus{
			Ready:          arena.positionPostMatchScoreReady(position),
			NumPanels:      arena.ScoringPanelRegistry.GetNumPanels(position),
			NumPanelsReady: arena.GetNumScoreCommitted(position),
		}
	}

	return &struct {
		RefereeScoreReady bool
		PositionStatuses  map[string]positionStatus
	}{
		arena.RedRealtimeScore.FoulsCommitted && arena.BlueRealtimeScore.FoulsCommitted,
		map[string]positionStatus{
			"red_near":  getStatusForPosition("red_near"),
			"red_far":   getStatusForPosition("red_far"),
			"blue_near": getStatusForPosition("blue_near"),
			"blue_far":  getStatusForPosition("blue_far"),
		},
	}
}

// Constructs the data object for one alliance sent to the audience display for the realtime scoring overlay.
func getAudienceAllianceScoreFields(
	allianceScore *RealtimeScore,
	allianceScoreSummary *game.ScoreSummary,
) *audienceAllianceScoreFields {
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
