// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRankingsCsvReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	ranking1 := Ranking{1114, 2, 18, 1100, 625, 90, 554, 0.254, 9, 1, 0, 0, 10}
	ranking2 := Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10}
	db.CreateRanking(&ranking1)
	db.CreateRanking(&ranking2)

	recorder := getHttpResponse("/reports/csv/rankings")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/plain", recorder.HeaderMap["Content-Type"][0])
	expectedBody := "Rank,TeamId,QualificationScore,AssistPoints,AutoPoints,TrussCatchPoints,GoalFoulPoints," +
		"Wins,Losses,Ties,Disqualifications,Played\n1,254,20,1100,625,90,554,10,0,0,0,10\n2,1114,18,1100,625," +
		"90,554,9,1,0,0,10\n\n"
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestRankingsPdfReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	ranking1 := Ranking{1114, 2, 18, 1100, 625, 90, 554, 0.254, 9, 1, 0, 0, 10}
	ranking2 := Ranking{254, 1, 20, 1100, 625, 90, 554, 0.254, 10, 0, 0, 0, 10}
	db.CreateRanking(&ranking1)
	db.CreateRanking(&ranking2)

	// Can't really parse the PDF content and check it, so just check that what's sent back is a PDF.
	recorder := getHttpResponse("/reports/pdf/rankings")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/pdf", recorder.HeaderMap["Content-Type"][0])
}

func TestScheduleCsvReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	match1 := Match{Type: "qualification", DisplayName: "1", Time: time.Unix(0, 0), Red1: 1, Red2: 2, Red3: 3,
		Blue1: 4, Blue2: 5, Blue3: 6, Blue1IsSurrogate: true, Blue2IsSurrogate: true, Blue3IsSurrogate: true}
	match2 := Match{Type: "qualification", DisplayName: "2", Time: time.Unix(600, 0), Red1: 7, Red2: 8, Red3: 9,
		Blue1: 10, Blue2: 11, Blue3: 12, Red1IsSurrogate: true, Red2IsSurrogate: true, Red3IsSurrogate: true}
	match3 := Match{Type: "practice", DisplayName: "1", Time: time.Now(), Red1: 6, Red2: 5, Red3: 4,
		Blue1: 3, Blue2: 2, Blue3: 1}
	db.CreateMatch(&match1)
	db.CreateMatch(&match2)
	db.CreateMatch(&match3)

	recorder := getHttpResponse("/reports/csv/schedule/qualification")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "text/plain", recorder.HeaderMap["Content-Type"][0])
	expectedBody := "Match,Type,Time,Red1,Red1IsSurrogate,Red2,Red2IsSurrogate,Red3,Red3IsSurrogate,Blue1," +
		"Blue1IsSurrogate,Blue2,Blue2IsSurrogate,Blue3,Blue3IsSurrogate\n1,qualification," +
		"1969-12-31 16:00:00 -0800 PST,1,false,2,false,3,false,4,true,5,true,6,true\n" +
		"2,qualification,1969-12-31 16:10:00 -0800 PST,7,true,8,true,9,true,10,false,11,false,12,false\n\n"
	assert.Equal(t, expectedBody, recorder.Body.String())
}

func TestSchedulePdfReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	match := Match{Type: "practice", DisplayName: "1", Time: time.Unix(0, 0), Red1: 1, Red2: 2, Red3: 3,
		Blue1: 4, Blue2: 5, Blue3: 6, Blue1IsSurrogate: true, Blue2IsSurrogate: true, Blue3IsSurrogate: true}
	db.CreateMatch(&match)
	team := Team{254, "NASA", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)

	// Can't really parse the PDF content and check it, so just check that what's sent back is a PDF.
	recorder := getHttpResponse("/reports/pdf/schedule/practice")
	assert.Equal(t, 200, recorder.Code)
	assert.Equal(t, "application/pdf", recorder.HeaderMap["Content-Type"][0])
}

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
	team := Team{254, "NASA", "The Cheesy Poofs", "San Jose", "CA", "USA", 1999, "Barrage"}
	db.CreateTeam(&team)

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
