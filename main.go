// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Cheesy Arena")
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
