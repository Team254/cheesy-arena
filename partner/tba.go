// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for publishing data to and retrieving data from The Blue Alliance.

package partner

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
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
	AvatarsDir = "static/img/avatars"
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
	Alliances      map[string]*TbaAlliance       `json:"alliances"`
	ScoreBreakdown map[string]*TbaScoreBreakdown `json:"score_breakdown"`
	TimeString     string                        `json:"time_string"`
	TimeUtc        string                        `json:"time_utc"`
}

type TbaAlliance struct {
	Teams      []string `json:"teams"`
	Surrogates []string `json:"surrogates"`
	Dqs        []string `json:"dqs"`
	Score      *int     `json:"score"`
}

type TbaScoreBreakdown struct {
	PreMatchLevelRobot1        string `json:"preMatchLevelRobot1"`
	PreMatchLevelRobot2        string `json:"preMatchLevelRobot2"`
	PreMatchLevelRobot3        string `json:"preMatchLevelRobot3"`
	PreMatchBay1               string `json:"preMatchBay1"`
	PreMatchBay2               string `json:"preMatchBay2"`
	PreMatchBay3               string `json:"preMatchBay3"`
	PreMatchBay6               string `json:"preMatchBay6"`
	PreMatchBay7               string `json:"preMatchBay7"`
	PreMatchBay8               string `json:"preMatchBay8"`
	HabLineRobot1              string `json:"habLineRobot1"`
	HabLineRobot2              string `json:"habLineRobot2"`
	HabLineRobot3              string `json:"habLineRobot3"`
	SandstormBonusPoints       int    `json:"sandStormBonusPoints"`
	Bay1                       string `json:"bay1"`
	Bay2                       string `json:"bay2"`
	Bay3                       string `json:"bay3"`
	Bay4                       string `json:"bay4"`
	Bay5                       string `json:"bay5"`
	Bay6                       string `json:"bay6"`
	Bay7                       string `json:"bay7"`
	Bay8                       string `json:"bay8"`
	LowLeftRocketNear          string `json:"lowLeftRocketNear"`
	MidLeftRocketNear          string `json:"midLeftRocketNear"`
	TopLeftRocketNear          string `json:"topLeftRocketNear"`
	LowRightRocketNear         string `json:"lowRightRocketNear"`
	MidRightRocketNear         string `json:"midRightRocketNear"`
	TopRightRocketNear         string `json:"topRightRocketNear"`
	LowLeftRocketFar           string `json:"lowLeftRocketFar"`
	MidLeftRocketFar           string `json:"midLeftRocketFar"`
	TopLeftRocketFar           string `json:"topLeftRocketFar"`
	LowRightRocketFar          string `json:"lowRightRocketFar"`
	MidRightRocketFar          string `json:"midRightRocketFar"`
	TopRightRocketFar          string `json:"topRightRocketFar"`
	CargoPoints                int    `json:"cargoPoints"`
	HatchPanelPoints           int    `json:"hatchPanelPoints"`
	EndgameRobot1              string `json:"endgameRobot1"`
	EndgameRobot2              string `json:"endgameRobot2"`
	EndgameRobot3              string `json:"endgameRobot3"`
	HabClimbPoints             int    `json:"habClimbPoints"`
	TeleopPoints               int    `json:"teleopPoints"`
	CompleteRocketRankingPoint bool   `json:"completeRocketRankingPoint"`
	HabDockingRankingPoint     bool   `json:"habDockingRankingPoint"`
	FoulPoints                 int    `json:"foulPoints"`
	TotalPoints                int    `json:"totalPoints"`
	RP                         int    `json:"rp"`
}

type TbaRanking struct {
	TeamKey    string `json:"team_key"`
	Rank       int    `json:"rank"`
	RP         float32
	ParkClimb  int
	Auto       int
	Ownership  int
	Vault      int
	WinLossTie string
	Dqs        int `json:"dqs"`
	Played     int `json:"played"`
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

type TbaMediaItem struct {
	Details map[string]interface{} `json:"details"`
	Type    string                 `json:"type"`
}

type TbaPublishedAward struct {
	Name    string `json:"name_str"`
	TeamKey string `json:"team_key"`
	Awardee string `json:"awardee"`
}

var habLevelMapping = []string{"None", "HabLevel1", "HabLevel2", "HabLevel3"}
var bayStatusMapping = []string{"None", "Panel", "PanelAndCargo", "Cargo"}
var sandstormBonusMapping = map[bool]string{false: "None", true: "CrossedHabLineInSandstorm"}

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

func (client *TbaClient) DownloadTeamAvatar(teamNumber, year int) error {
	path := fmt.Sprintf("/api/v3/team/%s/media/%d", getTbaTeam(teamNumber), year)
	resp, err := client.getRequest(path)
	if err != nil {
		return err
	}

	// Get the response and handle errors
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var mediaItems []*TbaMediaItem
	err = json.Unmarshal(body, &mediaItems)
	if err != nil {
		return err
	}

	for _, item := range mediaItems {
		if item.Type == "avatar" {
			base64String, ok := item.Details["base64Image"].(string)
			if !ok {
				return fmt.Errorf("Could not interpret avatar response from TBA: %v", item)
			}
			avatarBytes, err := base64.StdEncoding.DecodeString(base64String)
			if err != nil {
				return err
			}

			// Store the avatar to disk as a PNG file.
			avatarPath := fmt.Sprintf("%s/%d.png", AvatarsDir, teamNumber)
			ioutil.WriteFile(avatarPath, avatarBytes, 0644)
			return nil
		}
	}

	return fmt.Errorf("No avatar found for team %d in year %d.", teamNumber, year)
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

		// Fill in scores if the match has been played.
		var scoreBreakdown map[string]*TbaScoreBreakdown
		var redScore, blueScore *int
		var redCards, blueCards map[string]string
		if match.Status == "complete" {
			matchResult, err := database.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if matchResult != nil {
				scoreBreakdown = make(map[string]*TbaScoreBreakdown)
				scoreBreakdown["red"] = createTbaScoringBreakdown(&match, matchResult, "red")
				scoreBreakdown["blue"] = createTbaScoringBreakdown(&match, matchResult, "blue")
				redScore = &scoreBreakdown["red"].TotalPoints
				blueScore = &scoreBreakdown["blue"].TotalPoints
				redCards = matchResult.RedCards
				blueCards = matchResult.BlueCards
			}
		}
		alliances := make(map[string]*TbaAlliance)
		alliances["red"] = createTbaAlliance([3]int{match.Red1, match.Red2, match.Red3}, [3]bool{match.Red1IsSurrogate,
			match.Red2IsSurrogate, match.Red3IsSurrogate}, redScore, redCards)
		alliances["blue"] = createTbaAlliance([3]int{match.Blue1, match.Blue2, match.Blue3},
			[3]bool{match.Blue1IsSurrogate, match.Blue2IsSurrogate, match.Blue3IsSurrogate}, blueScore, blueCards)

		tbaMatches[i] = TbaMatch{"qm", 0, matchNumber, alliances, scoreBreakdown, match.Time.Local().Format("3:04 PM"),
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
	breakdowns := []string{"RP", "ParkClimb", "Auto", "Ownership", "Vault", "WinLossTie"}
	tbaRankings := make([]TbaRanking, len(rankings))
	for i, ranking := range rankings {
		tbaRankings[i] = TbaRanking{getTbaTeam(ranking.TeamId), ranking.Rank,
			float32(ranking.RankingPoints) / float32(ranking.Played), ranking.CargoPoints, ranking.HatchPanelPoints,
			ranking.HabClimbPoints, ranking.SandstormBonusPoints,
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

func createTbaAlliance(teamIds [3]int, surrogates [3]bool, score *int, cards map[string]string) *TbaAlliance {
	alliance := TbaAlliance{Surrogates: []string{}, Dqs: []string{}, Score: score}
	for i, teamId := range teamIds {
		teamKey := getTbaTeam(teamId)
		alliance.Teams = append(alliance.Teams, teamKey)
		if surrogates[i] {
			alliance.Surrogates = append(alliance.Surrogates, teamKey)
		}
		if cards != nil {
			if card, ok := cards[strconv.Itoa(teamId)]; ok && card == "red" {
				alliance.Dqs = append(alliance.Dqs, teamKey)
			}
		}
	}

	return &alliance
}

func createTbaScoringBreakdown(match *model.Match, matchResult *model.MatchResult, alliance string) *TbaScoreBreakdown {
	var breakdown TbaScoreBreakdown
	var score *game.Score
	var scoreSummary, opponentScoreSummary *game.ScoreSummary
	if alliance == "red" {
		score = matchResult.RedScore
		scoreSummary = matchResult.RedScoreSummary()
		opponentScoreSummary = matchResult.BlueScoreSummary()
	} else {
		score = matchResult.BlueScore
		scoreSummary = matchResult.BlueScoreSummary()
		opponentScoreSummary = matchResult.RedScoreSummary()
	}

	breakdown.PreMatchLevelRobot1 = habLevelMapping[score.RobotStartLevels[0]]
	breakdown.PreMatchLevelRobot2 = habLevelMapping[score.RobotStartLevels[1]]
	breakdown.PreMatchLevelRobot3 = habLevelMapping[score.RobotStartLevels[2]]
	breakdown.PreMatchBay1 = bayStatusMapping[score.CargoBaysPreMatch[0]]
	breakdown.PreMatchBay2 = bayStatusMapping[score.CargoBaysPreMatch[1]]
	breakdown.PreMatchBay3 = bayStatusMapping[score.CargoBaysPreMatch[2]]
	breakdown.PreMatchBay6 = bayStatusMapping[score.CargoBaysPreMatch[5]]
	breakdown.PreMatchBay7 = bayStatusMapping[score.CargoBaysPreMatch[6]]
	breakdown.PreMatchBay8 = bayStatusMapping[score.CargoBaysPreMatch[7]]
	breakdown.HabLineRobot1 = sandstormBonusMapping[score.SandstormBonuses[0]]
	breakdown.HabLineRobot2 = sandstormBonusMapping[score.SandstormBonuses[1]]
	breakdown.HabLineRobot3 = sandstormBonusMapping[score.SandstormBonuses[2]]
	breakdown.SandstormBonusPoints = scoreSummary.SandstormBonusPoints
	breakdown.Bay1 = bayStatusMapping[score.CargoBays[0]]
	breakdown.Bay2 = bayStatusMapping[score.CargoBays[1]]
	breakdown.Bay3 = bayStatusMapping[score.CargoBays[2]]
	breakdown.Bay4 = bayStatusMapping[score.CargoBays[3]]
	breakdown.Bay5 = bayStatusMapping[score.CargoBays[4]]
	breakdown.Bay6 = bayStatusMapping[score.CargoBays[5]]
	breakdown.Bay7 = bayStatusMapping[score.CargoBays[6]]
	breakdown.Bay8 = bayStatusMapping[score.CargoBays[7]]
	breakdown.LowLeftRocketNear = bayStatusMapping[score.RocketNearLeftBays[0]]
	breakdown.MidLeftRocketNear = bayStatusMapping[score.RocketNearLeftBays[1]]
	breakdown.TopLeftRocketNear = bayStatusMapping[score.RocketNearLeftBays[2]]
	breakdown.LowRightRocketNear = bayStatusMapping[score.RocketNearRightBays[0]]
	breakdown.MidRightRocketNear = bayStatusMapping[score.RocketNearRightBays[1]]
	breakdown.TopRightRocketNear = bayStatusMapping[score.RocketNearRightBays[2]]
	breakdown.LowLeftRocketFar = bayStatusMapping[score.RocketFarLeftBays[0]]
	breakdown.MidLeftRocketFar = bayStatusMapping[score.RocketFarLeftBays[1]]
	breakdown.TopLeftRocketFar = bayStatusMapping[score.RocketFarLeftBays[2]]
	breakdown.LowRightRocketFar = bayStatusMapping[score.RocketFarRightBays[0]]
	breakdown.MidRightRocketFar = bayStatusMapping[score.RocketFarRightBays[1]]
	breakdown.TopRightRocketFar = bayStatusMapping[score.RocketFarRightBays[2]]
	breakdown.CargoPoints = scoreSummary.CargoPoints
	breakdown.HatchPanelPoints = scoreSummary.HatchPanelPoints
	breakdown.EndgameRobot1 = habLevelMapping[score.RobotEndLevels[0]]
	breakdown.EndgameRobot2 = habLevelMapping[score.RobotEndLevels[1]]
	breakdown.EndgameRobot3 = habLevelMapping[score.RobotEndLevels[2]]
	breakdown.HabClimbPoints = scoreSummary.HabClimbPoints
	breakdown.TeleopPoints = scoreSummary.CargoPoints + scoreSummary.HatchPanelPoints + scoreSummary.HabClimbPoints
	breakdown.CompleteRocketRankingPoint = scoreSummary.CompleteRocket
	breakdown.HabDockingRankingPoint = scoreSummary.HabDocking
	breakdown.FoulPoints = scoreSummary.FoulPoints
	breakdown.TotalPoints = scoreSummary.Score
	if match.ShouldUpdateRankings() {
		// Calculate and set the ranking points for the match.
		var ranking game.Ranking
		ranking.AddScoreSummary(scoreSummary, opponentScoreSummary, false)
		breakdown.RP = ranking.RankingPoints
	}

	return &breakdown
}

// Uploads the awards to The Blue Alliance.
func (client *TbaClient) PublishAwards(database *model.Database) error {
	awards, err := database.GetAllAwards()
	if err != nil {
		return err
	}

	// Build a JSON array of TBA-format award models.
	tbaAwards := make([]TbaPublishedAward, len(awards))
	for i, award := range awards {
		tbaAwards[i].Name = award.AwardName
		tbaAwards[i].TeamKey = getTbaTeam(award.TeamId)
		tbaAwards[i].Awardee = award.PersonName
	}
	jsonBody, err := json.Marshal(tbaAwards)
	if err != nil {
		return err
	}

	resp, err := client.postRequest("awards", "update", jsonBody)
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
