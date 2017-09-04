// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for publishing data to and retrieving data from The Blue Alliance.

package partner

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	tbaBaseUrl = "https://www.thebluealliance.com"
	tbaAuthKey = "MAApv9MCuKY9MSFkXLuzTSYBCdosboxDq8Q3ujUE2Mn8PD3Nmv64uczu5Lvy0NQ3"
)

type TbaClient struct {
	BaseUrl         string
	eventCode       string
	secretId        string
	secret          string
	eventNamesCache map[string]string
}

type TbaMatch struct {
	CompLevel      string                        `json:"comp_level"`
	SetNumber      int                           `json:"set_number"`
	MatchNumber    int                           `json:"match_number"`
	Alliances      map[string]interface{}        `json:"alliances"`
	ScoreBreakdown map[string]*TbaScoreBreakdown `json:"score_breakdown"`
	TimeString     string                        `json:"time_string"`
	TimeUtc        string                        `json:"time_utc"`
}

type TbaScoreBreakdown struct {
	AutoFuelHigh              int  `json:"autoFuelHigh"`
	AutoFuelLow               int  `json:"autoFuelLow"`
	AutoFuelPoints            int  `json:"autoFuelPoints"`
	Rotor1Auto                bool `json:"rotor1Auto"`
	Rotor2Auto                bool `json:"rotor2Auto"`
	AutoRotorPoints           int  `json:"autoRotorPoints"`
	AutoMobilityPoints        int  `json:"autoMobilityPoints"`
	AutoPoints                int  `json:"autoPoints"`
	TeleopFuelHigh            int  `json:"teleopFuelHigh"`
	TeleopFuelLow             int  `json:"teleopFuelLow"`
	TeleopFuelPoints          int  `json:"teleopFuelPoints"`
	Rotor1Engaged             bool `json:"rotor1Engaged"`
	Rotor2Engaged             bool `json:"rotor2Engaged"`
	Rotor3Engaged             bool `json:"rotor3Engaged"`
	Rotor4Engaged             bool `json:"rotor4Engaged"`
	TeleopRotorPoints         int  `json:"teleopRotorPoints"`
	TeleopTakeoffPoints       int  `json:"teleopTakeoffPoints"`
	TeleopPoints              int  `json:"teleopPoints"`
	KPaRankingPointAchieved   bool `json:"kPaRankingPointAchieved"`
	KPaBonusPoints            int  `json:"kPaBonusPoints"`
	RotorRankingPointAchieved bool `json:"rotorRankingPointAchieved"`
	RotorBonusPoints          int  `json:"rotorBonusPoints"`
	FoulPoints                int  `json:"foulPoints"`
	TotalPoints               int  `json:"totalPoints"`
}

type TbaRanking struct {
	TeamKey    string  `json:"team_key"`
	Rank       int     `json:"rank"`
	RP         float32 `json:"RP"`
	Match      int     `json:"Match"`
	Auto       int     `json:"Auto"`
	Rotor      int     `json:"Rotor"`
	Touchpad   int     `json:"Touchpad"`
	Pressure   int     `json:"Pressure"`
	WinLossTie string  `json:"W-L-T"`
	Dqs        int     `json:"dqs"`
	Played     int     `json:"played"`
}

type TbaRankings struct {
	Breakdowns []string     `json:"breakdowns"`
	Rankings   []TbaRanking `json:"rankings"`
}

type TbaTeam struct {
	TeamNumber int    `json:"team_number"`
	Name       string `json:"name"`
	Nickname   string `json:"nickname"`
	City       string `json:"city"`
	StateProv  string `json:"state_prov"`
	Country    string `json:"country"`
	RookieYear int    `json:"rookie_year"`
}

type TbaRobot struct {
	RobotName string `json:"robot_name"`
	Year      int    `json:"year"`
}

type TbaAward struct {
	Name      string `json:"name"`
	EventKey  string `json:"event_key"`
	Year      int    `json:"year"`
	EventName string
}

type TbaEvent struct {
	Name string `json:"name"`
}

func NewTbaClient(eventCode, secretId, secret string) *TbaClient {
	return &TbaClient{BaseUrl: tbaBaseUrl, eventCode: eventCode, secretId: secretId, secret: secret,
		eventNamesCache: make(map[string]string)}
}

func (client *TbaClient) GetTeam(teamNumber int) (*TbaTeam, error) {
	path := fmt.Sprintf("/api/v3/team/%s", getTbaTeam(teamNumber))
	resp, err := client.getRequest(path)
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

func (client *TbaClient) GetRobotName(teamNumber int, year int) (string, error) {
	path := fmt.Sprintf("/api/v3/team/%s/robots", getTbaTeam(teamNumber))
	resp, err := client.getRequest(path)
	if err != nil {
		return "", err
	}

	// Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var robots []*TbaRobot
	err = json.Unmarshal(body, &robots)
	if err != nil {
		return "", err
	}
	for _, robot := range robots {
		if robot.Year == year {
			return robot.RobotName, nil
		}
	}
	return "", nil
}

func (client *TbaClient) GetTeamAwards(teamNumber int) ([]*TbaAward, error) {
	path := fmt.Sprintf("/api/v3/team/%s/awards", getTbaTeam(teamNumber))
	resp, err := client.getRequest(path)
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
		if _, ok := client.eventNamesCache[award.EventKey]; !ok {
			client.eventNamesCache[award.EventKey], err = client.getEventName(award.EventKey)
			if err != nil {
				return nil, err
			}
		}
		award.EventName = client.eventNamesCache[award.EventKey]
	}

	return awards, nil
}

// Uploads the event team list to The Blue Alliance.
func (client *TbaClient) PublishTeams(database *model.Database) error {
	teams, err := database.GetAllTeams()
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

	resp, err := client.postRequest("team_list", "update", jsonBody)
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
func (client *TbaClient) PublishMatches(database *model.Database) error {
	qualMatches, err := database.GetMatchesByType("qualification")
	if err != nil {
		return err
	}
	elimMatches, err := database.GetMatchesByType("elimination")
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
		var scoreBreakdown map[string]*TbaScoreBreakdown

		// Fill in scores if the match has been played.
		if match.Status == "complete" {
			matchResult, err := database.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if matchResult != nil {
				scoreBreakdown = make(map[string]*TbaScoreBreakdown)
				scoreBreakdown["red"] = createTbaScoringBreakdown(&match, matchResult, "red")
				scoreBreakdown["blue"] = createTbaScoringBreakdown(&match, matchResult, "blue")
				redAlliance["score"] = scoreBreakdown["red"].TotalPoints
				blueAlliance["score"] = scoreBreakdown["blue"].TotalPoints
			}
		}

		tbaMatches[i] = TbaMatch{"qm", 0, matchNumber, map[string]interface{}{"red": redAlliance,
			"blue": blueAlliance}, scoreBreakdown, match.Time.Local().Format("3:04 PM"),
			match.Time.UTC().Format("2006-01-02T15:04:05")}
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

	resp, err := client.postRequest("matches", "update", jsonBody)
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
func (client *TbaClient) PublishRankings(database *model.Database) error {
	rankings, err := database.GetAllRankings()
	if err != nil {
		return err
	}

	// Build a JSON object of TBA-format rankings.
	breakdowns := []string{"RP", "Match", "Auto", "Rotor", "Touchpad", "Pressure", "W-L-T"}
	tbaRankings := make([]TbaRanking, len(rankings))
	for i, ranking := range rankings {
		tbaRankings[i] = TbaRanking{getTbaTeam(ranking.TeamId), ranking.Rank,
			float32(ranking.RankingPoints) / float32(ranking.Played), ranking.MatchPoints, ranking.AutoPoints,
			ranking.RotorPoints, ranking.TakeoffPoints, ranking.PressurePoints,
			fmt.Sprintf("%d-%d-%d", ranking.Wins, ranking.Losses, ranking.Ties), ranking.Disqualifications,
			ranking.Played}
	}
	jsonBody, err := json.Marshal(TbaRankings{breakdowns, tbaRankings})
	if err != nil {
		return err
	}

	resp, err := client.postRequest("rankings", "update", jsonBody)
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
func (client *TbaClient) PublishAlliances(database *model.Database) error {
	alliances, err := database.GetAllAlliances()
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

	resp, err := client.postRequest("alliance_selections", "update", jsonBody)
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
func (client *TbaClient) DeletePublishedMatches() error {
	resp, err := client.postRequest("matches", "delete_all", []byte(client.eventCode))
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

func (client *TbaClient) getEventName(eventCode string) (string, error) {
	path := fmt.Sprintf("/api/v3/event/%s", eventCode)
	resp, err := client.getRequest(path)
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

// Converts an integer team number into the "frcXXXX" format TBA expects.
func getTbaTeam(team int) string {
	return fmt.Sprintf("frc%d", team)
}

// Sends a GET request to the TBA API.
func (client *TbaClient) getRequest(path string) (*http.Response, error) {
	url := client.BaseUrl + path

	// Make an HTTP GET request with the TBA auth headers.
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-TBA-Auth-Key", tbaAuthKey)
	return httpClient.Do(req)
}

// Signs the request and sends it to the TBA API.
func (client *TbaClient) postRequest(resource string, action string, body []byte) (*http.Response, error) {
	path := fmt.Sprintf("/api/trusted/v1/event/%s/%s/%s", client.eventCode, resource, action)
	signature := fmt.Sprintf("%x", md5.Sum(append([]byte(client.secret+path), body...)))

	httpClient := &http.Client{}
	request, err := http.NewRequest("POST", client.BaseUrl+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("X-TBA-Auth-Id", client.secretId)
	request.Header.Add("X-TBA-Auth-Sig", signature)
	return httpClient.Do(request)
}

func createTbaScoringBreakdown(match *model.Match, matchResult *model.MatchResult, alliance string) *TbaScoreBreakdown {
	var breakdown TbaScoreBreakdown
	var score *game.Score
	var scoreSummary *game.ScoreSummary
	if alliance == "red" {
		score = matchResult.RedScore
		scoreSummary = matchResult.RedScoreSummary()
	} else {
		score = matchResult.BlueScore
		scoreSummary = matchResult.BlueScoreSummary()
	}

	breakdown.AutoFuelHigh = score.AutoFuelHigh
	breakdown.AutoFuelLow = score.AutoFuelLow
	breakdown.AutoFuelPoints = score.AutoFuelHigh + score.AutoFuelLow/3
	breakdown.Rotor1Auto = score.AutoRotors >= 1
	breakdown.Rotor2Auto = score.AutoRotors >= 2
	breakdown.AutoRotorPoints = 60 * score.AutoRotors
	breakdown.AutoMobilityPoints = scoreSummary.AutoMobilityPoints
	breakdown.AutoPoints = scoreSummary.AutoPoints
	breakdown.TeleopFuelHigh = score.FuelHigh
	breakdown.TeleopFuelLow = score.FuelLow
	breakdown.TeleopFuelPoints = scoreSummary.PressurePoints - breakdown.AutoFuelPoints
	totalRotors := score.AutoRotors + score.Rotors
	breakdown.Rotor1Engaged = totalRotors >= 1
	breakdown.Rotor2Engaged = totalRotors >= 2
	breakdown.Rotor3Engaged = totalRotors >= 3
	breakdown.Rotor4Engaged = totalRotors >= 4
	breakdown.TeleopRotorPoints = scoreSummary.RotorPoints - breakdown.AutoRotorPoints
	breakdown.TeleopTakeoffPoints = scoreSummary.TakeoffPoints
	breakdown.TeleopPoints = breakdown.TeleopFuelPoints + breakdown.TeleopRotorPoints +
		breakdown.TeleopTakeoffPoints + scoreSummary.BonusPoints
	if match.Type == "elimination" {
		if scoreSummary.PressureGoalReached {
			breakdown.KPaBonusPoints = 20
		}
		if scoreSummary.RotorGoalReached {
			breakdown.RotorBonusPoints = 100
		}
	} else {
		breakdown.KPaRankingPointAchieved = scoreSummary.PressureGoalReached
		breakdown.RotorRankingPointAchieved = scoreSummary.RotorGoalReached
	}
	breakdown.FoulPoints = scoreSummary.FoulPoints
	breakdown.TotalPoints = scoreSummary.Score

	return &breakdown
}
