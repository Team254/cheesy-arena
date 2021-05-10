// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package tournament

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateOrUpdateAwardWithIntro(t *testing.T) {
	database := setupTestDb(t)
	database.CreateTeam(&model.Team{Id: 254, Nickname: "Teh Chezy Pofs"})

	award := model.Award{0, model.JudgedAward, "Safety Award", 0, ""}
	err := CreateOrUpdateAward(database, &award, true)
	assert.Nil(t, err)
	award2, _ := database.GetAwardById(award.Id)
	assert.Equal(t, award, *award2)
	lowerThirds, _ := database.GetAllLowerThirds()
	if assert.Equal(t, 2, len(lowerThirds)) {
		assert.Equal(t, "Safety Award", lowerThirds[0].TopText)
		assert.Equal(t, "", lowerThirds[0].BottomText)
		assert.Equal(t, "Safety Award", lowerThirds[1].TopText)
		assert.Equal(t, "(No awardee assigned yet)", lowerThirds[1].BottomText)
	}

	award.AwardName = "Saftey Award"
	award.TeamId = 254
	err = CreateOrUpdateAward(database, &award, true)
	assert.Nil(t, err)
	award2, _ = database.GetAwardById(award.Id)
	assert.Equal(t, award, *award2)
	lowerThirds, _ = database.GetAllLowerThirds()
	if assert.Equal(t, 2, len(lowerThirds)) {
		assert.Equal(t, "Saftey Award", lowerThirds[0].TopText)
		assert.Equal(t, "", lowerThirds[0].BottomText)
		assert.Equal(t, "Saftey Award", lowerThirds[1].TopText)
		assert.Equal(t, "Team 254, Teh Chezy Pofs", lowerThirds[1].BottomText)
	}

	err = DeleteAward(database, award.Id)
	assert.Nil(t, err)
	award2, _ = database.GetAwardById(award.Id)
	assert.Nil(t, award2)
	lowerThirds, _ = database.GetAllLowerThirds()
	assert.Empty(t, lowerThirds)
}

func TestCreateOrUpdateAwardWithoutIntro(t *testing.T) {
	database := setupTestDb(t)
	database.CreateTeam(&model.Team{Id: 254, Nickname: "Teh Chezy Pofs"})
	otherLowerThird := model.LowerThird{TopText: "Marco", BottomText: "Polo"}
	database.CreateLowerThird(&otherLowerThird)

	award := model.Award{0, model.WinnerAward, "Winner", 0, "Bob Dorough"}
	err := CreateOrUpdateAward(database, &award, false)
	assert.Nil(t, err)
	award2, _ := database.GetAwardById(award.Id)
	assert.Equal(t, award, *award2)
	lowerThirds, _ := database.GetAllLowerThirds()
	if assert.Equal(t, 2, len(lowerThirds)) {
		assert.Equal(t, otherLowerThird, lowerThirds[0])
		assert.Equal(t, "Winner", lowerThirds[1].TopText)
		assert.Equal(t, "Bob Dorough", lowerThirds[1].BottomText)
	}

	award.TeamId = 254
	err = CreateOrUpdateAward(database, &award, false)
	assert.Nil(t, err)
	award2, _ = database.GetAwardById(award.Id)
	assert.Equal(t, award, *award2)
	lowerThirds, _ = database.GetAllLowerThirds()
	if assert.Equal(t, 2, len(lowerThirds)) {
		assert.Equal(t, otherLowerThird, lowerThirds[0])
		assert.Equal(t, "Winner", lowerThirds[1].TopText)
		assert.Equal(t, "Bob Dorough &ndash; Team 254, Teh Chezy Pofs", lowerThirds[1].BottomText)
	}

	err = DeleteAward(database, award.Id)
	assert.Nil(t, err)
	award2, _ = database.GetAwardById(award.Id)
	assert.Nil(t, award2)
	lowerThirds, _ = database.GetAllLowerThirds()
	if assert.Equal(t, 1, len(lowerThirds)) {
		assert.Equal(t, otherLowerThird, lowerThirds[0])
	}
}

func TestCreateOrUpdateWinnerAndFinalistAwards(t *testing.T) {
	database := setupTestDb(t)
	CreateTestAlliances(database, 2)
	database.CreateTeam(&model.Team{Id: 1})
	database.CreateTeam(&model.Team{Id: 10})
	database.CreateTeam(&model.Team{Id: 100})
	database.CreateTeam(&model.Team{Id: 2})
	database.CreateTeam(&model.Team{Id: 20})
	database.CreateTeam(&model.Team{Id: 200})

	err := CreateOrUpdateWinnerAndFinalistAwards(database, 2, 1)
	assert.Nil(t, err)
	awards, _ := database.GetAllAwards()
	if assert.Equal(t, 6, len(awards)) {
		assert.Equal(t, model.Award{1, model.FinalistAward, "Finalist", 1, ""}, awards[0])
		assert.Equal(t, model.Award{2, model.FinalistAward, "Finalist", 10, ""}, awards[1])
		assert.Equal(t, model.Award{3, model.FinalistAward, "Finalist", 100, ""}, awards[2])
		assert.Equal(t, model.Award{4, model.WinnerAward, "Winner", 2, ""}, awards[3])
		assert.Equal(t, model.Award{5, model.WinnerAward, "Winner", 20, ""}, awards[4])
		assert.Equal(t, model.Award{6, model.WinnerAward, "Winner", 200, ""}, awards[5])
	}
	lowerThirds, _ := database.GetAllLowerThirds()
	if assert.Equal(t, 8, len(lowerThirds)) {
		assert.Equal(t, "Finalist", lowerThirds[0].TopText)
		assert.Equal(t, "", lowerThirds[0].BottomText)
		assert.Equal(t, "Finalist", lowerThirds[1].TopText)
		assert.Equal(t, "Team 1, ", lowerThirds[1].BottomText)
		assert.Equal(t, "Winner", lowerThirds[4].TopText)
		assert.Equal(t, "", lowerThirds[4].BottomText)
		assert.Equal(t, "Winner", lowerThirds[5].TopText)
		assert.Equal(t, "Team 2, ", lowerThirds[5].BottomText)
	}

	err = CreateOrUpdateWinnerAndFinalistAwards(database, 1, 2)
	assert.Nil(t, err)
	awards, _ = database.GetAllAwards()
	if assert.Equal(t, 6, len(awards)) {
		assert.Equal(t, model.Award{7, model.FinalistAward, "Finalist", 2, ""}, awards[0])
		assert.Equal(t, model.Award{8, model.FinalistAward, "Finalist", 20, ""}, awards[1])
		assert.Equal(t, model.Award{9, model.FinalistAward, "Finalist", 200, ""}, awards[2])
		assert.Equal(t, model.Award{10, model.WinnerAward, "Winner", 1, ""}, awards[3])
		assert.Equal(t, model.Award{11, model.WinnerAward, "Winner", 10, ""}, awards[4])
		assert.Equal(t, model.Award{12, model.WinnerAward, "Winner", 100, ""}, awards[5])
	}
	lowerThirds, _ = database.GetAllLowerThirds()
	if assert.Equal(t, 8, len(lowerThirds)) {
		assert.Equal(t, "Finalist", lowerThirds[0].TopText)
		assert.Equal(t, "", lowerThirds[0].BottomText)
		assert.Equal(t, "Finalist", lowerThirds[1].TopText)
		assert.Equal(t, "Team 2, ", lowerThirds[1].BottomText)
		assert.Equal(t, "Winner", lowerThirds[4].TopText)
		assert.Equal(t, "", lowerThirds[4].BottomText)
		assert.Equal(t, "Winner", lowerThirds[5].TopText)
		assert.Equal(t, "Team 1, ", lowerThirds[5].BottomText)
	}
}
