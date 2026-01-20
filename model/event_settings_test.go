// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for 2026 REBUILT Game

package model

import (
	"os"
	"testing"

	"github.com/Team254/cheesy-arena/game"
	"github.com/stretchr/testify/assert"
)

func TestEventSettingsDefaults(t *testing.T) {
	// 使用暫存資料庫
	os.Remove("test_settings.db")
	db, err := OpenDatabase("test_settings.db")
	assert.Nil(t, err)
	defer os.Remove("test_settings.db")

	// 1. 測試：取得設定 (如果是空的，應該要自動建立預設值)
	settings, err := db.GetEventSettings()
	assert.Nil(t, err)

	// 驗證是否載入了 2026 的預設值 (來自 game package)
	assert.Equal(t, game.EnergizedFuelThreshold, settings.EnergizedFuelThreshold)
	assert.Equal(t, game.SuperchargedFuelThreshold, settings.SuperchargedFuelThreshold)
	assert.Equal(t, game.TraversalPointThreshold, settings.TraversalPointThreshold)
}

func TestUpdateEventSettings(t *testing.T) {
	os.Remove("test_settings_update.db")
	db, err := OpenDatabase("test_settings_update.db")
	assert.Nil(t, err)
	defer os.Remove("test_settings_update.db")

	settings, _ := db.GetEventSettings()

	// 2. 測試：修改並儲存設定
	settings.Name = "2026 Championship"
	settings.EnergizedFuelThreshold = 999 // 修改 RP 門檻
	settings.SuperchargedFuelThreshold = 1000

	err = db.UpdateEventSettings(settings)
	assert.Nil(t, err)

	// 重新讀取並驗證
	newSettings, _ := db.GetEventSettings()
	assert.Equal(t, "2026 Championship", newSettings.Name)
	assert.Equal(t, 999, newSettings.EnergizedFuelThreshold)
	assert.Equal(t, 1000, newSettings.SuperchargedFuelThreshold)
}
