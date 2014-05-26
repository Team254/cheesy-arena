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

func TestScheduleCsvReport(t *testing.T) {
	clearDb()
	defer clearDb()
	db, _ = OpenDatabase(testDbPath)
	match1 := Match{1, "qualification", "1", time.Unix(0, 0), 1, false, 2, false, 3, false, 4, true, 5, true, 6,
		true, "", time.Now()}
	match2 := Match{2, "qualification", "2", time.Unix(600, 0), 7, true, 8, true, 9, true, 10, false, 11, false,
		12, false, "", time.Now()}
	match3 := Match{3, "practice", "1", time.Now(), 6, false, 5, false, 4, false, 3, false, 2, false, 1, false,
		"", time.Now()}
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
	match := Match{1, "practice", "1", time.Unix(0, 0), 1, false, 2, false, 3, false, 4, true, 5,
		true, 6, true, "", time.Now().UTC()}
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
