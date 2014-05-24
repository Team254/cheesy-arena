// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Cheesy Arena")
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
