// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"log"
	"math/rand"
	"time"
)

var db *Database

func main() {
	rand.Seed(time.Now().UnixNano())
	db, _ = OpenDatabase("test.db")
	ServeWebInterface()
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
