// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Logic to generate the 2018 game-specific data.

package game

import "math/rand"

var validGameSpecificDatas = []string{"RRR", "LLL", "RLR", "LRL"}

// GenerateGameSpecificData returns a random configuration.
func GenerateGameSpecificData() string {
	return validGameSpecificDatas[rand.Intn(len(validGameSpecificDatas))]
}

// IsValidGameSpecificData returns true if the given game specific data is valid.
func IsValidGameSpecificData(gameSpecificData string) bool {
	for _, data := range validGameSpecificDatas {
		if data == gameSpecificData {
			return true
		}
	}
	return false
}
