// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"log"
	"math/rand"
	"time"
)

var db *Database
var eventSettings *EventSettings

func main() {
	rand.Seed(time.Now().UnixNano())
	var err error
	db, err = OpenDatabase("test.db")
	checkErr(err)
	eventSettings, err = db.GetEventSettings()
	checkErr(err)

	ServeWebInterface()
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
