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

var tbaBaseUrl = "http://www.thebluealliance.com"

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
	Auto       int `json:"auto"`
	Assist     int `json:"assist"`
	TrussCatch int `json:"truss+catch"`
	GoalFoul   int `json:"teleop_goal+foul"`
}

type TbaRanking struct {
	TeamKey    string `json:"team_key"`
	Rank       int    `json:"rank"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Ties       int    `json:"ties"`
	Played     int    `json:"played"`
	Dqs        int    `json:"dqs"`
	QS         int
	Assist     int
	Auto       int
	TrussCatch int `json:"T&C"`
	GoalFoul   int `json:"G&F"`
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

type TbaAward struct {
  Name       string `json:"name"`
  EventKey   string `json:"event_key"`
  Year       int    `json:"year"`
  AwardType  int    `json:"award_type"`
}

// DATA RETRIEVAL
func getTeamFromTba(teamNumber int) (*TbaTeam) {
  url := fmt.Sprint("/api/v2/team/", string(getTbaTeam(teamNumber)))
  resp, _ := getTbaRequest(url);
  

  // Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	
	var teamData TbaTeam
  json.Unmarshal(body, &teamData)
  
  return &teamData
}

func getTeamAwardsFromTba(teamNumber int) ([]TbaAward) {
  url := fmt.Sprint("/api/v2/team/", string(getTbaTeam(teamNumber)), "/history/awards")
  resp, _ := getTbaRequest(url);
  

  // Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	
	var awardData []TbaAward
  json.Unmarshal(body, &awardData)
  
  return awardData
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

	resp, err := postTbaRequest("team_list", jsonBody)
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
				scoreBreakdown["red"] = TbaScoreBreakdown{redScoreSummary.AutoPoints, redScoreSummary.AssistPoints,
					redScoreSummary.TrussCatchPoints, redScoreSummary.GoalPoints + redScoreSummary.FoulPoints}
				scoreBreakdown["blue"] = TbaScoreBreakdown{blueScoreSummary.AutoPoints, blueScoreSummary.AssistPoints,
					blueScoreSummary.TrussCatchPoints, blueScoreSummary.GoalPoints + blueScoreSummary.FoulPoints}
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

	resp, err := postTbaRequest("matches", jsonBody)
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
	breakdowns := []string{"QS", "Assist", "Auto", "T&C", "G&F"}
	tbaRankings := make([]TbaRanking, len(rankings))
	for i, ranking := range rankings {
		tbaRankings[i] = TbaRanking{getTbaTeam(ranking.TeamId), ranking.Rank, ranking.Wins, ranking.Losses,
			ranking.Ties, ranking.Played, ranking.Disqualifications, ranking.QualificationScore,
			ranking.AssistPoints, ranking.AutoPoints, ranking.TrussCatchPoints, ranking.GoalFoulPoints}
	}
	jsonBody, err := json.Marshal(map[string]interface{}{"breakdowns": breakdowns, "rankings": tbaRankings})
	if err != nil {
		return err
	}

	resp, err := postTbaRequest("rankings", jsonBody)
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

	resp, err := postTbaRequest("alliance_selections", jsonBody)
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
func postTbaRequest(resource string, body []byte) (*http.Response, error) {
	path := fmt.Sprintf("/api/trusted/v1/event/%s/%s/update", eventSettings.TbaEventCode, resource)
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
func getTbaRequest(path string) (*http.Response, error) {
  // Make an HTTP GET request with the TBA auth headers
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprint(tbaBaseUrl, path), nil)
	if err != nil {
	  return nil, err
	}
	req.Header.Set("X-TBA-App-Id", "cheesy-arena:cheesy-fms:v0.1")
  return client.Do(req)
}
