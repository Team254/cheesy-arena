// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package field

import (
	"os"
	"testing"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func setupTestArena(t *testing.T) *Arena {
	// 使用暫存資料庫進行測試
	dbPath := "test.db"
	os.Remove(dbPath)
	arena, err := NewArena(dbPath)
	if err != nil {
		t.Fatalf("Failed to create arena: %v", err)
	}
	return arena
}

func TestAssignTeam(t *testing.T) {
	arena := setupTestArena(t)
	defer os.Remove("test.db")

	// 建立測試隊伍
	team254 := &model.Team{Id: 254, Name: "The Cheesy Poofs"}
	err := arena.Database.CreateTeam(team254)
	assert.Nil(t, err)

	team1114 := &model.Team{Id: 1114, Name: "Simbotics"}
	err = arena.Database.CreateTeam(team1114)
	assert.Nil(t, err)

	// 測試指派隊伍到紅藍聯盟
	err = arena.assignTeam(254, "R1")
	assert.Nil(t, err)
	assert.Equal(t, team254.Id, arena.AllianceStations["R1"].Team.Id)

	err = arena.assignTeam(1114, "B1")
	assert.Nil(t, err)
	assert.Equal(t, team1114.Id, arena.AllianceStations["B1"].Team.Id)

	// 測試更換隊伍
	err = arena.assignTeam(1114, "R1")
	assert.Nil(t, err)
	assert.Equal(t, team1114.Id, arena.AllianceStations["R1"].Team.Id)

	// 測試清空位置 (指派 0)
	err = arena.assignTeam(0, "R1")
	assert.Nil(t, err)
	assert.Nil(t, arena.AllianceStations["R1"].Team)
}

func TestTrussLightWarningSequence(t *testing.T) {
	// 測試最後倒數階段的燈光閃爍邏輯 (Sonar Ping)
	// 假設總時間設定與預設相符
	warmup := 0.0
	auto := 20.0
	pause := 3.0
	teleop := 140.0
	warning := 20.0

	totalTime := warmup + auto + pause + teleop
	startTime := totalTime - warning

	// 1. 在警告時間之前 -> 燈光應該不動作 (return false)
	active, _ := trussLightWarningSequence(startTime - 1.0)
	assert.False(t, active)

	// 2. 剛進入警告時間 -> 燈光開始動作 (return true)
	active, lights := trussLightWarningSequence(startTime + 0.1)
	assert.True(t, active)
	// 序列的第一步應該是開啟某些燈 (根據 trussLightWarningSequence 的定義)
	// sequence := []int{1, 2, 3...} -> lights[0] should be true
	assert.True(t, lights[0])
}
