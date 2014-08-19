// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for controlling the field LED lighting.

package main

import (
	"fmt"
)

var hotGoalLights map[string]bool
var assistLights map[string]int
var pedestalLights map[string]bool

func SetupLights() {
	hotGoalLights = make(map[string]bool)
	assistLights = make(map[string]int)
	pedestalLights = make(map[string]bool)
}

func SetHotGoalLights(alliance string, leftSide bool) {
	if hotGoalLights[alliance] == leftSide {
		return
	}
	hotGoalLights[alliance] = leftSide
	if leftSide {
		fmt.Printf("Setting left %s goal hot\n", alliance)
	} else {
		fmt.Printf("Setting right %s goal hot\n", alliance)
	}
}

func SetAssistGoalLights(alliance string, numAssists int) {
	if assistLights[alliance] == numAssists {
		return
	}
	assistLights[alliance] = numAssists
	if numAssists <= 0 {
		fmt.Printf("Clearing %s goal lights\n", alliance)
	} else if numAssists < 3 {
		fmt.Printf("Setting %s goal to %d assists\n", alliance, numAssists)
	} else {
		fmt.Printf("Setting %s goal to 3 assists\n", alliance)
	}
}

func ClearGoalLights(alliance string) {
	SetAssistGoalLights(alliance, 0)
}

func SetPedestalLight(alliance string) {
	if pedestalLights[alliance] == false {
		pedestalLights[alliance] = true
		fmt.Printf("Setting %s pedestal\n", alliance)
	}
}

func ClearPedestalLight(alliance string) {
	if pedestalLights[alliance] == true {
		pedestalLights[alliance] = false
		fmt.Printf("Clearing %s pedestal\n", alliance)
	}
}
