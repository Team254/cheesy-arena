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
	database.CreateTeam(&model.Team{Id: 101})
	database.CreateTeam(&model.Team{Id: 102})
	database.CreateTeam(&model.Team{Id: 103})
	database.CreateTeam(&model.Team{Id: 104})
	database.CreateTeam(&model.Team{Id: 201})
	database.CreateTeam(&model.Team{Id: 202})
	database.CreateTeam(&model.Team{Id: 203})
	database.CreateTeam(&model.Team{Id: 204})

	err := CreateOrUpdateWinnerAndFinalistAwards(database, 2, 1)
	assert.Nil(t, err)
	awards, _ := database.GetAllAwards()
	if assert.Equal(t, 8, len(awards)) {
		assert.Equal(t, model.Award{1, model.FinalistAward, "Finalist", 101, ""}, awards[0])
		assert.Equal(t, model.Award{2, model.FinalistAward, "Finalist", 102, ""}, awards[1])
		assert.Equal(t, model.Award{3, model.FinalistAward, "Finalist", 103, ""}, awards[2])
		assert.Equal(t, model.Award{4, model.FinalistAward, "Finalist", 104, ""}, awards[3])
		assert.Equal(t, model.Award{5, model.WinnerAward, "Winner", 201, ""}, awards[4])
		assert.Equal(t, model.Award{6, model.WinnerAward, "Winner", 202, ""}, awards[5])
		assert.Equal(t, model.Award{7, model.WinnerAward, "Winner", 203, ""}, awards[6])
		assert.Equal(t, model.Award{8, model.WinnerAward, "Winner", 204, ""}, awards[7])
	}
	lowerThirds, _ := database.GetAllLowerThirds()
	if assert.Equal(t, 10, len(lowerThirds)) {
		assert.Equal(t, "Finalist", lowerThirds[0].TopText)
		assert.Equal(t, "", lowerThirds[0].BottomText)
		assert.Equal(t, "Finalist", lowerThirds[1].TopText)
		assert.Equal(t, "Team 101, ", lowerThirds[1].BottomText)
		assert.Equal(t, "Winner", lowerThirds[5].TopText)
		assert.Equal(t, "", lowerThirds[5].BottomText)
		assert.Equal(t, "Winner", lowerThirds[6].TopText)
		assert.Equal(t, "Team 201, ", lowerThirds[6].BottomText)
	}

	err = CreateOrUpdateWinnerAndFinalistAwards(database, 1, 2)
	assert.Nil(t, err)
	awards, _ = database.GetAllAwards()
	if assert.Equal(t, 8, len(awards)) {
		assert.Equal(t, model.Award{9, model.FinalistAward, "Finalist", 201, ""}, awards[0])
		assert.Equal(t, model.Award{10, model.FinalistAward, "Finalist", 202, ""}, awards[1])
		assert.Equal(t, model.Award{11, model.FinalistAward, "Finalist", 203, ""}, awards[2])
		assert.Equal(t, model.Award{12, model.FinalistAward, "Finalist", 204, ""}, awards[3])
		assert.Equal(t, model.Award{13, model.WinnerAward, "Winner", 101, ""}, awards[4])
		assert.Equal(t, model.Award{14, model.WinnerAward, "Winner", 102, ""}, awards[5])
		assert.Equal(t, model.Award{15, model.WinnerAward, "Winner", 103, ""}, awards[6])
		assert.Equal(t, model.Award{16, model.WinnerAward, "Winner", 104, ""}, awards[7])
	}
	lowerThirds, _ = database.GetAllLowerThirds()
	if assert.Equal(t, 10, len(lowerThirds)) {
		assert.Equal(t, "Finalist", lowerThirds[0].TopText)
		assert.Equal(t, "", lowerThirds[0].BottomText)
		assert.Equal(t, "Finalist", lowerThirds[1].TopText)
		assert.Equal(t, "Team 201, ", lowerThirds[1].BottomText)
		assert.Equal(t, "Winner", lowerThirds[5].TopText)
		assert.Equal(t, "", lowerThirds[5].BottomText)
		assert.Equal(t, "Winner", lowerThirds[6].TopText)
		assert.Equal(t, "Team 101, ", lowerThirds[6].BottomText)
	}
}
