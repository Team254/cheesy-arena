// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for managing awards and their associated lower thirds.

package tournament

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
)

// Creates or updates the given award, depending on whether or not it already exists.
func CreateOrUpdateAward(database *model.Database, award *model.Award, createIntroLowerThird bool) error {
	// Validate the award data.
	if award.AwardName == "" {
		return fmt.Errorf("Award name cannot be blank.")
	}
	var team *model.Team
	if award.TeamId > 0 {
		team, _ = database.GetTeamById(award.TeamId)
		if team == nil {
			return fmt.Errorf("Team %d is not present at this event.", award.TeamId)
		}
	}

	var err error
	if award.Id == 0 {
		err = database.CreateAward(award)
	} else {
		err = database.SaveAward(award)
	}
	if err != nil {
		return err
	}

	// Create or update associated lower thirds.
	awardIntroLowerThird := model.LowerThird{TopText: award.AwardName, AwardId: award.Id}
	awardWinnerLowerThird := model.LowerThird{TopText: award.AwardName, BottomText: award.PersonName,
		AwardId: award.Id}
	if team != nil {
		if award.PersonName == "" {
			awardWinnerLowerThird.BottomText = fmt.Sprintf("Team %d, %s", team.Id, team.Nickname)
		} else {
			awardWinnerLowerThird.BottomText = fmt.Sprintf("%s &ndash; Team %d, %s", award.PersonName, team.Id,
				team.Nickname)
		}
	}
	if awardWinnerLowerThird.BottomText == "" {
		awardWinnerLowerThird.BottomText = "(No awardee assigned yet)"
	}
	lowerThirds, err := database.GetLowerThirdsByAwardId(award.Id)
	if err != nil {
		return err
	}
	bottomIndex := 0
	if createIntroLowerThird {
		if err = createOrUpdateAwardLowerThird(database, &awardIntroLowerThird, lowerThirds, 0); err != nil {
			return err
		}
		bottomIndex++
	}
	if err = createOrUpdateAwardLowerThird(database, &awardWinnerLowerThird, lowerThirds, bottomIndex); err != nil {
		return err
	}

	return nil
}

// Deletes the given award and any associated lower thirds.
func DeleteAward(database *model.Database, awardId int) error {
	var award *model.Award
	award, err := database.GetAwardById(awardId)
	if err != nil {
		return err
	}
	if award == nil {
		return fmt.Errorf("Award with ID %d does not exist.", awardId)
	}
	if err = database.DeleteAward(award); err != nil {
		return err
	}

	// Delete lower thirds.
	lowerThirds, err := database.GetLowerThirdsByAwardId(award.Id)
	if err != nil {
		return err
	}
	for _, lowerThird := range lowerThirds {
		if err = database.DeleteLowerThird(&lowerThird); err != nil {
			return err
		}
	}

	return nil
}

// Generates awards and lower thirds for the tournament winners and finalists.
func CreateOrUpdateWinnerAndFinalistAwards(database *model.Database, winnerAllianceId, finalistAllianceId int) error {
	var winnerAllianceTeams, finalistAllianceTeams []model.AllianceTeam
	var err error
	if winnerAllianceTeams, err = database.GetTeamsByAlliance(winnerAllianceId); err != nil {
		return err
	}
	if finalistAllianceTeams, err = database.GetTeamsByAlliance(finalistAllianceId); err != nil {
		return err
	}
	if len(winnerAllianceTeams) == 0 || len(finalistAllianceTeams) == 0 {
		return fmt.Errorf("Input alliances do not contain any teams.")
	}

	// Clear out any awards that may exist if the final match was scored more than once.
	winnerAwards, err := database.GetAwardsByType(model.WinnerAward)
	if err != nil {
		return err
	}
	finalistAwards, err := database.GetAwardsByType(model.FinalistAward)
	if err != nil {
		return err
	}
	for _, award := range append(winnerAwards, finalistAwards...) {
		if err = DeleteAward(database, award.Id); err != nil {
			return err
		}
	}

	// Create the finalist awards first since they're usually presented first.
	finalistAward := model.Award{AwardName: "Finalist", Type: model.FinalistAward,
		TeamId: finalistAllianceTeams[0].TeamId}
	if err = CreateOrUpdateAward(database, &finalistAward, true); err != nil {
		return err
	}
	for _, allianceTeam := range finalistAllianceTeams[1:] {
		finalistAward.Id = 0
		finalistAward.TeamId = allianceTeam.TeamId
		if err = CreateOrUpdateAward(database, &finalistAward, false); err != nil {
			return err
		}
	}

	// Create the winner awards.
	winnerAward := model.Award{AwardName: "Winner", Type: model.WinnerAward,
		TeamId: winnerAllianceTeams[0].TeamId}
	if err = CreateOrUpdateAward(database, &winnerAward, true); err != nil {
		return err
	}
	for _, allianceTeam := range winnerAllianceTeams[1:] {
		winnerAward.Id = 0
		winnerAward.TeamId = allianceTeam.TeamId
		if err = CreateOrUpdateAward(database, &winnerAward, false); err != nil {
			return err
		}
	}

	return nil
}

func createOrUpdateAwardLowerThird(database *model.Database, lowerThird *model.LowerThird,
	existingLowerThirds []model.LowerThird, index int) error {
	if index < len(existingLowerThirds) {
		lowerThird.Id = existingLowerThirds[index].Id
		lowerThird.DisplayOrder = existingLowerThirds[index].DisplayOrder
		return database.SaveLowerThird(lowerThird)
	} else {
		lowerThird.DisplayOrder = database.GetNextLowerThirdDisplayOrder()
		return database.CreateLowerThird(lowerThird)
	}
}
