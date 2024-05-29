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
	"github.com/mitchellh/mapstructure"
	"io"
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
	CompLevel      string                    `json:"comp_level"`
	SetNumber      int                       `json:"set_number"`
	MatchNumber    int                       `json:"match_number"`
	Alliances      map[string]*TbaAlliance   `json:"alliances"`
	ScoreBreakdown map[string]map[string]any `json:"score_breakdown"`
	TimeString     string                    `json:"time_string"`
	TimeUtc        string                    `json:"time_utc"`
	DisplayName    string                    `json:"display_name"`
}

type TbaAlliance struct {
	Teams      []string `json:"teams"`
	Surrogates []string `json:"surrogates"`
	Dqs        []string `json:"dqs"`
	Score      *int     `json:"score"`
}

type TbaScoreBreakdown struct {
	AutoLineRobot1                   string `mapstructure:"autoLineRobot1"`
	AutoLineRobot2                   string `mapstructure:"autoLineRobot2"`
	AutoLineRobot3                   string `mapstructure:"autoLineRobot3"`
	AutoLeavePoints                  int    `mapstructure:"autoLeavePoints"`
	AutoAmpNoteCount                 int    `mapstructure:"autoAmpNoteCount"`
	AutoAmpNotePoints                int    `mapstructure:"autoAmpNotePoints"`
	AutoSpeakerNoteCount             int    `mapstructure:"autoSpeakerNoteCount"`
	AutoSpeakerNotePoints            int    `mapstructure:"autoSpeakerNotePoints"`
	AutoTotalNotePoints              int    `mapstructure:"autoTotalNotePoints"`
	AutoPoints                       int    `mapstructure:"autoPoints"`
	TeleopAmpNoteCount               int    `mapstructure:"teleopAmpNoteCount"`
	TeleopAmpNotePoints              int    `mapstructure:"teleopAmpNotePoints"`
	TeleopSpeakerNoteCount           int    `mapstructure:"teleopSpeakerNoteCount"`
	TeleopSpeakerNotePoints          int    `mapstructure:"teleopSpeakerNotePoints"`
	TeleopSpeakerNoteAmplifiedCount  int    `mapstructure:"teleopSpeakerNoteAmplifiedCount"`
	TeleopSpeakerNoteAmplifiedPoints int    `mapstructure:"teleopSpeakerNoteAmplifiedPoints"`
	TeleopTotalNotePoints            int    `mapstructure:"teleopTotalNotePoints"`
	EndGameRobot1                    string `mapstructure:"endGameRobot1"`
	EndGameRobot2                    string `mapstructure:"endGameRobot2"`
	EndGameRobot3                    string `mapstructure:"endGameRobot3"`
	EndGameParkPoints                int    `mapstructure:"endGameParkPoints"`
	EndGameOnStagePoints             int    `mapstructure:"endGameOnStagePoints"`
	EndGameHarmonyPoints             int    `mapstructure:"endGameHarmonyPoints"`
	MicStageLeft                     bool   `mapstructure:"micStageLeft"`
	MicCenterStage                   bool   `mapstructure:"micCenterStage"`
	MicStageRight                    bool   `mapstructure:"micStageRight"`
	EndGameSpotLightBonusPoints      int    `mapstructure:"endGameSpotLightBonusPoints"`
	TrapStageLeft                    bool   `mapstructure:"trapStageLeft"`
	TrapCenterStage                  bool   `mapstructure:"trapCenterStage"`
	TrapStageRight                   bool   `mapstructure:"trapStageRight"`
	EndGameNoteInTrapPoints          int    `mapstructure:"endGameNoteInTrapPoints"`
	EndGameTotalStagePoints          int    `mapstructure:"endGameTotalStagePoints"`
	TeleopPoints                     int    `mapstructure:"teleopPoints"`
	CoopertitionCriteriaMet          bool   `mapstructure:"coopertitionCriteriaMet"`
	MelodyBonusAchieved              bool   `mapstructure:"melodyBonusAchieved"`
	EnsembleBonusAchieved            bool   `mapstructure:"ensembleBonusAchieved"`
	FoulCount                        int    `mapstructure:"foulCount"`
	TechFoulCount                    int    `mapstructure:"techFoulCount"`
	G424Penalty                      bool   `mapstructure:"g424Penalty"`
	FoulPoints                       int    `mapstructure:"foulPoints"`
	TotalPoints                      int    `mapstructure:"totalPoints"`
	RP                               int    `mapstructure:"rp"`
}

type TbaRanking struct {
	TeamKey string `json:"team_key"`
	Rank    int    `json:"rank"`
	RP      float32
	Coop    float32
	Match   float32
	Auto    float32
	Stage   float32
	Wins    int `json:"wins"`
	Losses  int `json:"losses"`
	Ties    int `json:"ties"`
	Dqs     int `json:"dqs"`
	Played  int `json:"played"`
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
	Details map[string]any `json:"details"`
	Type    string         `json:"type"`
}

type TbaPublishedAward struct {
	Name    string `json:"name_str"`
	TeamKey string `json:"team_key"`
	Awardee string `json:"awardee"`
}

var leaveMapping = map[bool]string{false: "No", true: "Yes"}
var endGameStatusMapping = map[game.EndgameStatus]string{
	game.EndgameNone:        "None",
	game.EndgameParked:      "Parked",
	game.EndgameStageLeft:   "StageLeft",
	game.EndgameCenterStage: "CenterStage",
	game.EndgameStageRight:  "StageRight",
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
	body, err := io.ReadAll(resp.Body)
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
	body, err := io.ReadAll(resp.Body)
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
	body, err := io.ReadAll(resp.Body)
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
	body, err := io.ReadAll(resp.Body)
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}
	return nil
}

// Uploads the qualification and playoff match schedule and results to The Blue Alliance.
func (client *TbaClient) PublishMatches(database *model.Database) error {
	qualMatches, err := database.GetMatchesByType(model.Qualification, false)
	if err != nil {
		return err
	}
	playoffMatches, err := database.GetMatchesByType(model.Playoff, false)
	if err != nil {
		return err
	}
	eventSettings, err := database.GetEventSettings()
	if err != nil {
		return err
	}
	matches := append(qualMatches, playoffMatches...)
	tbaMatches := make([]TbaMatch, len(matches))

	// Build a JSON array of TBA-format matches.
	for i, match := range matches {
		// Fill in scores if the match has been played.
		var scoreBreakdown map[string]map[string]any
		var redScore, blueScore *int
		var redCards, blueCards map[string]string
		if match.IsComplete() {
			matchResult, err := database.GetMatchResultForMatch(match.Id)
			if err != nil {
				return err
			}
			if matchResult != nil {
				scoreBreakdown = make(map[string]map[string]any)
				scoreBreakdown["red"] = createTbaScoringBreakdown(eventSettings, &match, matchResult, "red")
				scoreBreakdown["blue"] = createTbaScoringBreakdown(eventSettings, &match, matchResult, "blue")
				redScoreValue := scoreBreakdown["red"]["totalPoints"].(int)
				blueScoreValue, _ := scoreBreakdown["blue"]["totalPoints"].(int)
				redScore = &redScoreValue
				blueScore = &blueScoreValue
				redCards = matchResult.RedCards
				blueCards = matchResult.BlueCards
			}
		}
		alliances := make(map[string]*TbaAlliance)
		alliances["red"] = createTbaAlliance([3]int{match.Red1, match.Red2, match.Red3}, [3]bool{match.Red1IsSurrogate,
			match.Red2IsSurrogate, match.Red3IsSurrogate}, redScore, redCards)
		alliances["blue"] = createTbaAlliance([3]int{match.Blue1, match.Blue2, match.Blue3},
			[3]bool{match.Blue1IsSurrogate, match.Blue2IsSurrogate, match.Blue3IsSurrogate}, blueScore, blueCards)

		tbaMatches[i] = TbaMatch{
			CompLevel:      match.TbaMatchKey.CompLevel,
			SetNumber:      match.TbaMatchKey.SetNumber,
			MatchNumber:    match.TbaMatchKey.MatchNumber,
			Alliances:      alliances,
			ScoreBreakdown: scoreBreakdown,
			TimeString:     match.Time.Local().Format("3:04 PM"),
			TimeUtc:        match.Time.UTC().Format("2006-01-02T15:04:05"),
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
		body, _ := io.ReadAll(resp.Body)
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
	breakdowns := []string{"RP", "Coop", "Match", "Auto", "Stage"}
	tbaRankings := make([]TbaRanking, len(rankings))
	for i, ranking := range rankings {
		tbaRankings[i] = TbaRanking{
			TeamKey: getTbaTeam(ranking.TeamId),
			Rank:    ranking.Rank,
			RP:      float32(ranking.RankingPoints) / float32(ranking.Played),
			Coop:    float32(ranking.CoopertitionPoints) / float32(ranking.Played),
			Match:   float32(ranking.MatchPoints) / float32(ranking.Played),
			Auto:    float32(ranking.AutoPoints) / float32(ranking.Played),
			Stage:   float32(ranking.StagePoints) / float32(ranking.Played),
			Wins:    ranking.Wins,
			Losses:  ranking.Losses,
			Ties:    ranking.Ties,
			Dqs:     ranking.Disqualifications,
			Played:  ranking.Played,
		}
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
		body, _ := io.ReadAll(resp.Body)
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
		for _, allianceTeamId := range alliance.TeamIds {
			tbaAlliances[i] = append(tbaAlliances[i], getTbaTeam(allianceTeamId))
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}

	// Also set the playoff type so that TBA renders the correct bracket.
	eventSettings, err := database.GetEventSettings()
	if err != nil {
		return err
	}
	playoffType := 0
	if eventSettings.PlayoffType == model.DoubleEliminationPlayoff {
		playoffType = 10
	}
	resp, err = client.postRequest("info", "update", []byte(fmt.Sprintf("{\"playoff_type\":%d}", playoffType)))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Got status code %d from TBA: %s", resp.StatusCode, body)
	}

	return nil
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
		body, _ := io.ReadAll(resp.Body)
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
		body, _ := io.ReadAll(resp.Body)
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
	body, err := io.ReadAll(resp.Body)
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
	response, err := httpClient.Do(request)
	if client.BaseUrl == tbaBaseUrl && err == nil && response.StatusCode == 200 {
		// Send a non-blocking ping to track usage.
		pingRequest, _ := http.NewRequest(
			"POST", fmt.Sprintf("https://cheesyarena.com/events/%s/%s", client.eventCode, resource), nil,
		)
		_, _ = httpClient.Do(pingRequest)
	}
	return response, err
}

func createTbaAlliance(teamIds [3]int, surrogates [3]bool, score *int, cards map[string]string) *TbaAlliance {
	alliance := TbaAlliance{Teams: []string{}, Surrogates: []string{}, Dqs: []string{}, Score: score}
	for i, teamId := range teamIds {
		if teamId == 0 {
			continue
		}
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

func createTbaScoringBreakdown(
	eventSettings *model.EventSettings,
	match *model.Match,
	matchResult *model.MatchResult,
	alliance string,
) map[string]any {
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

	breakdown.AutoLineRobot1 = leaveMapping[score.LeaveStatuses[0]]
	breakdown.AutoLineRobot2 = leaveMapping[score.LeaveStatuses[1]]
	breakdown.AutoLineRobot3 = leaveMapping[score.LeaveStatuses[2]]
	breakdown.AutoLeavePoints = scoreSummary.LeavePoints
	breakdown.AutoAmpNoteCount = score.AmpSpeaker.AutoAmpNotes
	breakdown.AutoAmpNotePoints = 2 * breakdown.AutoAmpNoteCount
	breakdown.AutoSpeakerNoteCount = score.AmpSpeaker.AutoSpeakerNotes
	breakdown.AutoSpeakerNotePoints = 5 * breakdown.AutoSpeakerNoteCount
	breakdown.AutoTotalNotePoints = breakdown.AutoAmpNotePoints + breakdown.AutoSpeakerNotePoints
	breakdown.AutoPoints = scoreSummary.AutoPoints
	breakdown.TeleopAmpNoteCount = score.AmpSpeaker.TeleopAmpNotes
	breakdown.TeleopAmpNotePoints = 1 * breakdown.TeleopAmpNoteCount
	breakdown.TeleopSpeakerNoteCount = score.AmpSpeaker.TeleopUnamplifiedSpeakerNotes
	breakdown.TeleopSpeakerNotePoints = 2 * breakdown.TeleopSpeakerNoteCount
	breakdown.TeleopSpeakerNoteAmplifiedCount = score.AmpSpeaker.TeleopAmplifiedSpeakerNotes
	breakdown.TeleopSpeakerNoteAmplifiedPoints = 5 * breakdown.TeleopSpeakerNoteAmplifiedCount
	breakdown.TeleopTotalNotePoints = breakdown.TeleopAmpNotePoints + breakdown.TeleopSpeakerNotePoints +
		breakdown.TeleopSpeakerNoteAmplifiedPoints
	breakdown.EndGameRobot1 = endGameStatusMapping[score.EndgameStatuses[0]]
	breakdown.EndGameRobot2 = endGameStatusMapping[score.EndgameStatuses[1]]
	breakdown.EndGameRobot3 = endGameStatusMapping[score.EndgameStatuses[2]]
	breakdown.EndGameParkPoints = scoreSummary.ParkPoints
	breakdown.EndGameOnStagePoints = scoreSummary.OnStagePoints
	breakdown.EndGameHarmonyPoints = scoreSummary.HarmonyPoints
	breakdown.MicStageLeft = score.MicrophoneStatuses[0]
	breakdown.MicCenterStage = score.MicrophoneStatuses[1]
	breakdown.MicStageRight = score.MicrophoneStatuses[2]
	breakdown.EndGameSpotLightBonusPoints = scoreSummary.SpotlightPoints
	breakdown.TrapStageLeft = score.TrapStatuses[0]
	breakdown.TrapCenterStage = score.TrapStatuses[1]
	breakdown.TrapStageRight = score.TrapStatuses[2]
	breakdown.EndGameNoteInTrapPoints = scoreSummary.TrapPoints
	breakdown.EndGameTotalStagePoints = scoreSummary.StagePoints
	breakdown.TeleopPoints = breakdown.TeleopTotalNotePoints + breakdown.EndGameTotalStagePoints
	breakdown.CoopertitionCriteriaMet = scoreSummary.CoopertitionCriteriaMet
	breakdown.MelodyBonusAchieved = scoreSummary.MelodyBonusRankingPoint
	breakdown.EnsembleBonusAchieved = scoreSummary.EnsembleBonusRankingPoint
	for _, foul := range score.Fouls {
		if foul.IsTechnical {
			breakdown.TechFoulCount++
		} else {
			breakdown.FoulCount++
		}
		if foul.Rule() != nil && foul.Rule().IsRankingPoint {
			breakdown.G424Penalty = true
		}
	}
	breakdown.FoulPoints = scoreSummary.FoulPoints
	breakdown.TotalPoints = scoreSummary.Score

	if match.ShouldUpdateRankings() {
		// Calculate and set the ranking points for the match.
		var ranking game.Ranking
		ranking.AddScoreSummary(scoreSummary, opponentScoreSummary, false)
		breakdown.RP = ranking.RankingPoints
	}

	// Turn the breakdown struct into a map in order to be able to remove any fields that are disabled based on the
	// event settings.
	breakdownMap := make(map[string]any)
	_ = mapstructure.Decode(breakdown, &breakdownMap)
	if eventSettings.MelodyBonusThresholdWithCoop == 0 {
		delete(breakdownMap, "coopertitionCriteriaMet")
	}

	return breakdownMap
}
