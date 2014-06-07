// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"log"
	"math/rand"
	"time"
)

const eventDbPath = "./event.db"

var db *Database
var eventSettings *EventSettings

func main() {
	rand.Seed(time.Now().UnixNano())
	initDb()

	ServeWebInterface()
}

func initDb() {
	var err error
	db, err = OpenDatabase(eventDbPath)
	checkErr(err)
	eventSettings, err = db.GetEventSettings()
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
