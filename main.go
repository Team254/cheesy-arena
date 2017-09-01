// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/web"
	"log"
	"math/rand"
	"time"
)

const eventDbPath = "./event.db"
const httpPort = 8080

// Main entry point for the application.
func main() {
	rand.Seed(time.Now().UnixNano())

	arena, err := field.NewArena(eventDbPath)
	if err != nil {
		log.Fatalln("Error during startup: ", err)
	}

	// Start the web server in a separate goroutine.
	web := web.NewWeb(arena)
	go web.ServeWebInterface(httpPort)

	// Run the arena state machine in the main thread.
	arena.Run()
}
