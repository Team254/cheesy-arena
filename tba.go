// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for publishing data to and retrieving data from The Blue Alliance.

package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Distinct endpoints are necessary for testing.
var tbaBaseUrl = "http://www.thebluealliance.com"
var tbaTeamBaseUrl = tbaBaseUrl
var tbaTeamRobotsBaseUrl = tbaBaseUrl
var tbaTeamAwardsBaseUrl = tbaBaseUrl
var tbaEventBaseUrl = tbaBaseUrl

// Cache of event codes to names.
var tbaEventNames = make(map[string]string)

// MODELS

type TbaMatch struct {
	CompLevel      string                       `json:"comp_level"`
	SetNumber      int                          `json:"set_number"`
	MatchNumber    int                          `json:"match_number"`
	Alliances      map[string]interface{}       `json:"alliances"`
	ScoreBreakdown map[string]TbaScoreBreakdown `json:"score_breakdown"`
	TimeString     string                       `json:"time_string"`
	TimeUtc        string                       `json:"time_utc"`
}

type TbaScoreBreakdown struct {
	Coopertition int `json:"coopertition_points"`
	Auto         int `json:"auto_points"`
	Container    int `json:"container_points"`
	Tote         int `json:"tote_points"`
	Litter       int `json:"litter_points"`
	Foul         int `json:"foul_points"`
}

type TbaRanking struct {
	TeamKey      string  `json:"team_key"`
	Rank         int     `json:"rank"`
	QA           float64 `json:"QA"`
	Coopertition int     `json:"Coopertition"`
	Auto         int     `json:"Auto"`
	Container    int     `json:"Container"`
	Tote         int     `json:"Tote"`
	Litter       int     `json:"Litter"`
	Dqs          int     `json:"dqs"`
	Played       int     `json:"played"`
}

type TbaTeam struct {
	Website    string `json:"website"`
	Name       string `json:"name"`
	Locality   string `json:"locality"`
	RookieYear int    `json:"rookie_year"`
	Reigon     string `json:"region"`
	TeamNumber int    `json:"team_number"`
	Location   string `json:"location"`
	Key        string `json:"key"`
	Country    string `json:"country_name"`
	Nickname   string `json:"nickname"`
}

type TbaRobot struct {
	Name string `json:"name"`
}

type TbaAward struct {
	Name      string `json:"name"`
	EventKey  string `json:"event_key"`
	Year      int    `json:"year"`
	AwardType int    `json:"award_type"`
	EventName string
}

type TbaEvent struct {
	Name string `json:"name"`
}

// DATA RETRIEVAL
func getTeamFromTba(teamNumber int) (*TbaTeam, error) {
	url := fmt.Sprintf("%s/api/v2/team/%s", tbaTeamBaseUrl, getTbaTeam(teamNumber))
	resp, err := getTbaRequest(url)
	if err != nil {
		return nil, err
	}

	// Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var teamData TbaTeam
	err = json.Unmarshal(body, &teamData)

	return &teamData, err
}

func getRobotNameFromTba(teamNumber int, year int) (string, error) {
	url := fmt.Sprintf("%s/api/v2/team/frc%d/history/robots", tbaTeamRobotsBaseUrl, teamNumber)
	resp, err := getTbaRequest(url)
	if err != nil {
		return "", err
	}

	// Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var robots map[string]TbaRobot
	err = json.Unmarshal(body, &robots)
	if err != nil {
		return "", err
	}
	if robotName, ok := robots[strconv.Itoa(year)]; ok {
		return robotName.Name, nil
	}
	return "", nil
}

func getTeamAwardsFromTba(teamNumber int) ([]*TbaAward, error) {
	url := fmt.Sprintf("%s/api/v2/team/%s/history/awards", tbaTeamAwardsBaseUrl, getTbaTeam(teamNumber))
	resp, err := getTbaRequest(url)
	if err != nil {
		return nil, err
	}

	// Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var awards []*TbaAward
	err = json.Unmarshal(body, &awards)
	if err != nil {
		return nil, err
	}

	for _, award := range awards {
		if _, ok := tbaEventNames[award.EventKey]; !ok {
			tbaEventNames[award.EventKey], err = getEventNameFromTba(award.EventKey)
			if err != nil {
				return nil, err
			}
		}
		award.EventName = tbaEventNames[award.EventKey]
	}

	return awards, nil
}

func getEventNameFromTba(eventCode string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/event/%s", tbaEventBaseUrl, eventCode)
	resp, err := getTbaRequest(url)
	if err != nil {
		return "", err
	}

	// Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var event TbaEvent
	err = json.Unmarshal(body, &event)
	if err != nil {
		return "", err
	}

	return event.Name, err
}

// PUBLISHING

// Uploads the event team list to The Blue Alliance.
func PublishTeams() error {
	teams, err := db.GetAllTeams()
	if err != nil {
		return err
	}

	// Build a JSON array of TBA-format team keys (e.g. "frc254").
	teamKeys := make([]string, len(teams))
	for i, team := range teams {
		teamKeys[i] = getTbaTeam(team.Id)
	}
	jsonBody, err := json.Marshal(teamKeys)
	if err != nil {
		return err
	}

	resp, err := postTbaRequest("team_list", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

// Uploads the qualification and elimination match schedule and results to The Blue Alliance.
func PublishMatches() error {
	qualMatches, err := db.GetMatchesByType("qualification")
	if err != nil {
		return err
	}
	elimMatches, err := db.GetMatchesByType("elimination")
	if err != nil {
		return err
	}
	matches := append(qualMatches, elimMatches...)
	tbaMatches := make([]TbaMatch, len(matches))

	// Build a JSON array of TBA-format matches.
	for i, match := range matches {
		matchNumber, _ := strconv.Atoi(match.DisplayName)
		redAlliance := map[string]interface{}{"teams": []string{getTbaTeam(match.Red1), getTbaTeam(match.Red2),
			getTbaTeam(match.Red3)}, "score": nil}
		blueAlliance := map[string]interface{}{"teams": []string{getTbaTeam(match.Blue1), getTbaTeam(match.Blue2),
			getTbaTeam(match.Blue3)}, "score": nil}
		var scoreBreakdown map[string]TbaScoreBreakdown

		// Fill in scores if the match has been played.
		if match.Status == "complete" {
			matchResult, err := db.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if matchResult != nil {
				redScoreSummary := matchResult.RedScoreSummary()
				blueScoreSummary := matchResult.BlueScoreSummary()
				redAlliance["score"] = redScoreSummary.Score
				blueAlliance["score"] = blueScoreSummary.Score
				scoreBreakdown = make(map[string]TbaScoreBreakdown)
				scoreBreakdown["red"] = TbaScoreBreakdown{redScoreSummary.CoopertitionPoints,
					redScoreSummary.AutoPoints, redScoreSummary.ContainerPoints, redScoreSummary.TotePoints,
					redScoreSummary.LitterPoints, redScoreSummary.FoulPoints}
				scoreBreakdown["blue"] = TbaScoreBreakdown{blueScoreSummary.CoopertitionPoints,
					blueScoreSummary.AutoPoints, blueScoreSummary.ContainerPoints, blueScoreSummary.TotePoints,
					blueScoreSummary.LitterPoints, blueScoreSummary.FoulPoints}
			}
		}

		tbaMatches[i] = TbaMatch{"qm", 0, matchNumber, map[string]interface{}{"red": redAlliance,
			"blue": blueAlliance}, scoreBreakdown, match.Time.Local().Format("3:04 PM"),
			match.Time.Format("2006-01-02T15:04:05")}
		if match.Type == "elimination" {
			tbaMatches[i].CompLevel = map[int]string{1: "f", 2: "sf", 4: "qf", 8: "ef"}[match.ElimRound]
			tbaMatches[i].SetNumber = match.ElimGroup
			tbaMatches[i].MatchNumber = match.ElimInstance
		}
	}
	jsonBody, err := json.Marshal(tbaMatches)
	if err != nil {
		return err
	}

	resp, err := postTbaRequest("matches", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

// Uploads the team standings to The Blue Alliance.
func PublishRankings() error {
	rankings, err := db.GetAllRankings()
	if err != nil {
		return err
	}

	// Build a JSON object of TBA-format rankings.
	breakdowns := []string{"QA", "Coopertition", "Auto", "Container", "Tote", "Litter"}
	tbaRankings := make([]TbaRanking, len(rankings))
	for i, ranking := range rankings {
		tbaRankings[i] = TbaRanking{getTbaTeam(ranking.TeamId), ranking.Rank, ranking.QualificationAverage,
			ranking.CoopertitionPoints, ranking.AutoPoints, ranking.ContainerPoints, ranking.TotePoints,
			ranking.LitterPoints, ranking.Disqualifications, ranking.Played}
	}
	jsonBody, err := json.Marshal(map[string]interface{}{"breakdowns": breakdowns, "rankings": tbaRankings})
	if err != nil {
		return err
	}

	resp, err := postTbaRequest("rankings", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

// Uploads the alliances selection results to The Blue Alliance.
func PublishAlliances() error {
	alliances, err := db.GetAllAlliances()
	if err != nil {
		return err
	}

	// Build a JSON object of TBA-format alliances.
	tbaAlliances := make([][]string, len(alliances))
	for i, alliance := range alliances {
		for _, team := range alliance {
			tbaAlliances[i] = append(tbaAlliances[i], getTbaTeam(team.TeamId))
		}
	}
	jsonBody, err := json.Marshal(tbaAlliances)
	if err != nil {
		return err
	}

	resp, err := postTbaRequest("alliance_selections", "update", jsonBody)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

// Clears out the existing match data on The Blue Alliance for the event.
func DeletePublishedMatches() error {
	resp, err := postTbaRequest("matches", "delete_all", []byte(eventSettings.TbaEventCode))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

// Converts an integer team number into the "frcXXXX" format TBA expects.
func getTbaTeam(team int) string {
	return fmt.Sprintf("frc%d", team)
}

// HELPERS

// Signs the request and sends it to the TBA API.
func postTbaRequest(resource string, action string, body []byte) (*http.Response, error) {
	path := fmt.Sprintf("/api/trusted/v1/event/%s/%s/%s", eventSettings.TbaEventCode, resource, action)
	signature := fmt.Sprintf("%x", md5.Sum(append([]byte(eventSettings.TbaSecret+path), body...)))

	client := &http.Client{}
	request, err := http.NewRequest("POST", fmt.Sprintf("%s%s", tbaBaseUrl, path), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("X-TBA-Auth-Id", eventSettings.TbaSecretId)
	request.Header.Add("X-TBA-Auth-Sig", signature)
	return client.Do(request)
}

// Sends a GET request to the TBA API
func getTbaRequest(url string) (*http.Response, error) {
	// Make an HTTP GET request with the TBA auth headers
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-TBA-App-Id", "cheesy-arena:cheesy-fms:v0.1")
	return client.Do(req)
}
