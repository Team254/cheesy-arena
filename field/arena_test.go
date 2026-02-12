// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package field

import (
	"testing"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestAssignTeam(t *testing.T) {
	arena := setupTestArena(t)

	team254 := &model.Team{Id: 254, Name: "The Cheesy Poofs"}
	err := arena.Database.CreateTeam(team254)
	assert.Nil(t, err)

	team1114 := &model.Team{Id: 1114, Name: "Simbotics"}
	err = arena.Database.CreateTeam(team1114)
	assert.Nil(t, err)

	// 測試指派
	err = arena.assignTeam(254, "R1")
	assert.Nil(t, err)
	assert.Equal(t, team254.Id, arena.AllianceStations["R1"].Team.Id)

	err = arena.assignTeam(1114, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team1114.Id, arena.AllianceStations["B1"].Team.Id)

	// 測試更換
	err = arena.assignTeam(1114, "R1")
	assert.Nil(t, err)
	assert.Equal(t, team1114.Id, arena.AllianceStations["R1"].Team.Id)

	// 測試清空
	err = arena.assignTeam(0, "R1")
	assert.Nil(t, err)
	assert.Nil(t, arena.AllianceStations["R1"].Team)
}

func TestTrussLightWarningSequence(t *testing.T) {
	// 模擬比賽時間
	warmup := 0.0
	auto := 20.0
	pause := 3.0
	teleop := 140.0
	warning := 20.0

	totalTime := warmup + auto + pause + teleop
	startTime := totalTime - warning

	// 1. 在警告時間之前 -> 燈光應為 False
	active, _ := trussLightWarningSequence(startTime - 1.0)
	assert.False(t, active)

	// 2. 剛進入警告時間 -> 燈光應為 True (開始閃爍序列)
	active, lights := trussLightWarningSequence(startTime + 0.1)
	assert.True(t, active)
	// 序列的第一步應該是開啟某些燈
	assert.True(t, lights[0] || lights[1] || lights[2])
}
