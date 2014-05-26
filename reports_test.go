// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTeamsCsvReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	team1 := Team{254, "NASA", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	team2 := Team{1114, "GM", "Simbotics", "St. Catharines", "ON", "Canada", 2003, "Simbot Evolution"}
	db.CreateTeam(&team1)
	db.CreateTeam(&team2)

	recorder := getHttpResponse("/reports/csv/teams")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/plain", recorder.HeaderMap["Content-Type"][0])
	expectedBody := "Number,Name,Nickname,City,StateProv,Country,RookieYear,RobotName\n254,\"NASA\"," +
		"\"The Cheesy Poofs\",\"San Jose\",\"CA\",\"USA\",1999,\"Barrage\"\n1114,\"GM\",\"Simbotics\"," +
		"\"St. Catharines\",\"ON\",\"Canada\",2003,\"Simbot Evolution\"\n\n"
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestTeamsPdfReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	team1 := Team{254, "NASA", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team1)

	// Can't really parse the PDF content and check it, so just check that what's sent back is a PDF.
	recorder := getHttpResponse("/reports/pdf/teams")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/pdf", recorder.HeaderMap["Content-Type"][0])
}

func getHttpResponse(path string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	newHandler().ServeHTTP(recorder, req)
	return recorder
}
