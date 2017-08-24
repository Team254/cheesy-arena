// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/model"
	"log"
	"math/rand"
	"time"
)

const eventDbPath = "./event.db"

var db *model.Database
var eventSettings *model.EventSettings

// Main entry point for the application.
func main() {
	rand.Seed(time.Now().UnixNano())
	initDb()

	// Run the webserver and DS packet listener in goroutines and use the main one for the arena state machine.
	go ServeWebInterface()
	go ListenForDriverStations()
	go ListenForDsUdpPackets()
	go MonitorBandwidth()
	mainArena.Setup()
	mainArena.Run()
}

// Opens the database and stores a handle to it in a global variable.
func initDb() {
	var err error
	db, err = model.OpenDatabase(".", eventDbPath)
	checkErr(err)
	eventSettings, err = db.GetEventSettings()
	checkErr(err)
}

// Logs and exits the application if the given error is not nil.
func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
