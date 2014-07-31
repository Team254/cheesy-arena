// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for publishing data to and retrieving data from The Blue Alliance.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var tbaBaseUrl = "http://tbatv-dev-efang.appspot.com"

type TbaMatch struct {
	CompLevel   string                 `json:"comp_level"`
	SetNumber   int                    `json:"set_number"`
	MatchNumber int                    `json:"match_number"`
	Alliances   map[string]interface{} `json:"alliances"`
	TimeString  string                 `json:"time_string"`
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

	resp, err := http.PostForm(getTbaPostUrl("team_list"), getTbaPostBody("team_list", string(jsonBody)))
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

		// Fill in scores if the match has been played.
		if match.Status == "complete" {
			matchResult, err := db.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if matchResult != nil {
				redAlliance["score"] = matchResult.RedScoreSummary().Score
				blueAlliance["score"] = matchResult.BlueScoreSummary().Score
			}
		}

		tbaMatches[i] = TbaMatch{"qm", 0, matchNumber, map[string]interface{}{"red": redAlliance,
			"blue": blueAlliance}, match.Time.Local().Format("3:04 PM")}
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

	resp, err := http.PostForm(getTbaPostUrl("matches"), getTbaPostBody("matches", string(jsonBody)))
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

	resp, err := http.PostForm(getTbaPostUrl("rankings"), getTbaPostBody("rankings", string(jsonBody)))
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

	resp, err := http.PostForm(getTbaPostUrl("alliance_selections"), getTbaPostBody("alliances", string(jsonBody)))
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

func getTbaPostUrl(resource string) string {
	return fmt.Sprintf("%s/api/trusted/v1/event/%s/%s/update", tbaBaseUrl, eventSettings.TbaEventCode,
		resource)
}

func getTbaPostBody(key string, value string) url.Values {
	return url.Values{"secret-id": {eventSettings.TbaSecretId}, "secret": {eventSettings.TbaSecret}, key: {value}}
}

// Converts an integer team number into the "frcXXXX" format TBA expects.
func getTbaTeam(team int) string {
	return fmt.Sprintf("frc%d", team)
}
