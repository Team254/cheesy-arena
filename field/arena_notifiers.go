// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains configuration of the publish-subscribe notifiers that allow the arena to push updates to websocket clients.

package field

import (
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/led"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/vaultled"
	"github.com/Team254/cheesy-arena/websocket"
	"strconv"
	"time"
)

type ArenaNotifiers struct {
	AllianceSelectionNotifier          *websocket.Notifier
	AllianceStationDisplayModeNotifier *websocket.Notifier
	ArenaStatusNotifier                *websocket.Notifier
	AudienceDisplayModeNotifier        *websocket.Notifier
	LedModeNotifier                    *websocket.Notifier
	LowerThirdNotifier                 *websocket.Notifier
	MatchLoadNotifier                  *websocket.Notifier
	MatchTimeNotifier                  *websocket.Notifier
	PlaySoundNotifier                  *websocket.Notifier
	RealtimeScoreNotifier              *websocket.Notifier
	ReloadDisplaysNotifier             *websocket.Notifier
	ScorePostedNotifier                *websocket.Notifier
	ScoringStatusNotifier              *websocket.Notifier
}

type LedModeMessage struct {
	LedMode      led.Mode
	VaultLedMode vaultled.Mode
}

type MatchTimeMessage struct {
	MatchState   int
	MatchTimeSec int
}

type audienceAllianceScoreFields struct {
	Score         int
	RealtimeScore *RealtimeScore
	ForceState    game.PowerUpState
	LevitateState game.PowerUpState
	BoostState    game.PowerUpState
	SwitchOwnedBy game.Alliance
}

// Instantiates notifiers and configures their message producing methods.
func (arena *Arena) configureNotifiers() {
	arena.AllianceSelectionNotifier = websocket.NewNotifier("allianceSelection", nil)
	arena.AllianceStationDisplayModeNotifier = websocket.NewNotifier("allianceStationDisplayMode",
		arena.generateAllianceStationDisplayModeMessage)
	arena.ArenaStatusNotifier = websocket.NewNotifier("arenaStatus", arena.generateArenaStatusMessage)
	arena.AudienceDisplayModeNotifier = websocket.NewNotifier("audienceDisplayMode",
		arena.generateAudienceDisplayModeMessage)
	arena.LedModeNotifier = websocket.NewNotifier("ledMode", arena.generateLedModeMessage)
	arena.LowerThirdNotifier = websocket.NewNotifier("lowerThird", nil)
	arena.MatchLoadNotifier = websocket.NewNotifier("matchLoad", arena.generateMatchLoadMessage)
	arena.MatchTimeNotifier = websocket.NewNotifier("matchTime", arena.generateMatchTimeMessage)
	arena.PlaySoundNotifier = websocket.NewNotifier("playSound", nil)
	arena.RealtimeScoreNotifier = websocket.NewNotifier("realtimeScore", arena.generateRealtimeScoreMessage)
	arena.ReloadDisplaysNotifier = websocket.NewNotifier("reload", nil)
	arena.ScorePostedNotifier = websocket.NewNotifier("scorePosted", arena.generateScorePostedMessage)
	arena.ScoringStatusNotifier = websocket.NewNotifier("scoringStatus", arena.generateScoringStatusMessage)
}

func (arena *Arena) generateAllianceStationDisplayModeMessage() interface{} {
	return arena.AllianceStationDisplayMode
}

func (arena *Arena) generateArenaStatusMessage() interface{} {
	return &struct {
		AllianceStations map[string]*AllianceStation
		MatchState
		CanStartMatch    bool
		PlcIsHealthy     bool
		FieldEstop       bool
		GameSpecificData string
	}{arena.AllianceStations, arena.MatchState, arena.checkCanStartMatch() == nil, arena.Plc.IsHealthy,
		arena.Plc.GetFieldEstop(), arena.CurrentMatch.GameSpecificData}
}

func (arena *Arena) generateAudienceDisplayModeMessage() interface{} {
	return arena.AudienceDisplayMode
}

func (arena *Arena) generateLedModeMessage() interface{} {
	return &LedModeMessage{arena.ScaleLeds.GetCurrentMode(), arena.RedVaultLeds.CurrentForceMode}
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
	return MatchTimeMessage{int(arena.MatchState), int(arena.MatchTimeSec())}
}

func (arena *Arena) generateRealtimeScoreMessage() interface{} {
	fields := struct {
		Red          *audienceAllianceScoreFields
		Blue         *audienceAllianceScoreFields
		ScaleOwnedBy game.Alliance
	}{}
	fields.Red = getAudienceAllianceScoreFields(arena.RedRealtimeScore, arena.RedScoreSummary(),
		arena.RedVault, arena.RedSwitch)
	fields.Blue = getAudienceAllianceScoreFields(arena.BlueRealtimeScore, arena.BlueScoreSummary(),
		arena.BlueVault, arena.BlueSwitch)
	fields.ScaleOwnedBy = arena.Scale.GetOwnedBy()
	return &fields
}

func (arena *Arena) generateScorePostedMessage() interface{} {
	return &struct {
		MatchType        string
		Match            *model.Match
		RedScoreSummary  *game.ScoreSummary
		BlueScoreSummary *game.ScoreSummary
		RedFouls         []game.Foul
		BlueFouls        []game.Foul
		RedCards         map[string]string
		BlueCards        map[string]string
	}{arena.SavedMatch.CapitalizedType(), arena.SavedMatch, arena.SavedMatchResult.RedScoreSummary(),
		arena.SavedMatchResult.BlueScoreSummary(), populateFoulDescriptions(arena.SavedMatchResult.RedScore.Fouls),
		populateFoulDescriptions(arena.SavedMatchResult.BlueScore.Fouls), arena.SavedMatchResult.RedCards,
		arena.SavedMatchResult.BlueCards}
}

func (arena *Arena) generateScoringStatusMessage() interface{} {
	return &struct {
		RefereeScoreReady bool
		RedScoreReady     bool
		BlueScoreReady    bool
	}{arena.RedRealtimeScore.FoulsCommitted && arena.BlueRealtimeScore.FoulsCommitted,
		arena.RedRealtimeScore.TeleopCommitted, arena.BlueRealtimeScore.TeleopCommitted}
}

// Constructs the data object for one alliance sent to the audience display for the realtime scoring overlay.
func getAudienceAllianceScoreFields(allianceScore *RealtimeScore, allianceScoreSummary *game.ScoreSummary,
	allianceVault *game.Vault, allianceSwitch *game.Seesaw) *audienceAllianceScoreFields {
	fields := new(audienceAllianceScoreFields)
	fields.RealtimeScore = allianceScore
	fields.Score = allianceScoreSummary.Score
	if allianceVault.ForcePowerUp != nil {
		fields.ForceState = allianceVault.ForcePowerUp.GetState(time.Now())
	} else {
		fields.ForceState = game.Unplayed
	}
	if allianceVault.LevitatePlayed {
		fields.LevitateState = game.Expired
	} else {
		fields.LevitateState = game.Unplayed
	}
	if allianceVault.BoostPowerUp != nil {
		fields.BoostState = allianceVault.BoostPowerUp.GetState(time.Now())
	} else {
		fields.BoostState = game.Unplayed
	}
	fields.SwitchOwnedBy = allianceSwitch.GetOwnedBy()
	return fields
}

// Copy the description from the rules to the fouls so that they are available to the announcer.
func populateFoulDescriptions(fouls []game.Foul) []game.Foul {
	for i := range fouls {
		for _, rule := range game.Rules {
			if fouls[i].RuleNumber == rule.RuleNumber {
				fouls[i].Description = rule.Description
				break
			}
		}
	}
	return fouls
}
