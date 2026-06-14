// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with the Nexus for FRC API.

package partner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const nexusBaseUrl = "https://frc.nexus"
const nexusApiKey = "Vn6D9y80kQcNijDItKOJHg8yYEk"

type NexusClient struct {
	BaseUrl      string
	apiKey       string
	autoQueueKey string
	eventCode    string
}

type nexusLineup struct {
	Red  [3]string `json:"red"`
	Blue [3]string `json:"blue"`
}

type matchWinner string

const (
	Red  matchWinner = "red"
	Blue matchWinner = "blue"
	Tie  matchWinner = "tie"
)

type autoQueueEventType string

const (
	PostScores autoQueueEventType = "post-scores"
	MatchStart autoQueueEventType = "match-start"
	BreakStart autoQueueEventType = "break-start"
	BreakEnd   autoQueueEventType = "break-end"
)

type autoQueueEvent struct {
	Event       autoQueueEventType `json:"event"`
	MatchName   string             `json:"match"`
	MatchNumber int                `json:"matchNumber"`
	Winner      matchWinner        `json:"winner"`
	DurationSec int                `json:"duration"`
}

type autoQueueResponse struct {
	StatusText string `json:"response"`
}

func NewNexusClient(eventCode string, autoQueueKey string) *NexusClient {
	return &NexusClient{BaseUrl: nexusBaseUrl, apiKey: nexusApiKey, autoQueueKey: autoQueueKey, eventCode: eventCode}
}

// Gets the team lineup for a given match from the Nexus API. Returns nil and an error if the lineup is not available.
func (client *NexusClient) GetLineup(tbaMatchKey model.TbaMatchKey) (*[6]int, error) {
	path := fmt.Sprintf(
		"/api/v1/event/%s/match/%s/lineups?key=%s",
		client.eventCode,
		tbaMatchKey.String(),
		client.apiKey,
	)
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
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error getting lineup from Nexus: %d, %s", resp.StatusCode, string(body))
	}

	var nexusLineup nexusLineup
	if err = json.Unmarshal(body, &nexusLineup); err != nil {
		return nil, err
	}

	var lineup [6]int
	for i, teamString := range []string{
		nexusLineup.Red[0], nexusLineup.Red[1], nexusLineup.Red[2],
		nexusLineup.Blue[0], nexusLineup.Blue[1], nexusLineup.Blue[2],
	} {
		if teamString == "" {
			continue
		}
		lineup[i], err = strconv.Atoi(teamString)
		if err != nil {
			return nil, err
		}
	}

	// Check that at least one spot is filled with a valid team number; otherwise return an error.
	for _, team := range lineup {
		if team > 0 {
			return &lineup, nil
		}
	}
	return nil, fmt.Errorf("Lineup not yet submitted")
}

// Notifies Nexus that this match has been completed and updates the queuing status.
func (client *NexusClient) AutoQueue(matchName string, typeOrder int, matchStatus game.MatchStatus) error {
	var winner matchWinner
	switch matchStatus {
	case game.RedWonMatch:
		winner = Red
	case game.BlueWonMatch:
		winner = Blue
	case game.TieMatch:
		winner = Tie
	}

	_, err := client.postAutoQueueRequest(autoQueueEvent{Event: "post-scores", MatchName: matchName, MatchNumber: typeOrder, Winner: winner})
	return err
}

// Notifies Nexus that the match has started.
func (client *NexusClient) MatchStarted(matchName string, typeOrder int) error {
	_, err := client.postAutoQueueRequest(autoQueueEvent{Event: "match-start", MatchName: matchName, MatchNumber: typeOrder})
	return err
}

// Notifies Nexus that a field break has started.
func (client *NexusClient) BreakStarted(durationSec int) error {
	_, err := client.postAutoQueueRequest(autoQueueEvent{Event: "break-start", DurationSec: durationSec})
	return err
}

// Notifies Nexus that a field break has ended.
func (client *NexusClient) BreakEnded() error {
	_, err := client.postAutoQueueRequest(autoQueueEvent{Event: "break-end"})
	return err
}

// Sends a POST request to the Nexus AutoQueue API.
func (client *NexusClient) postAutoQueueRequest(body autoQueueEvent) (*autoQueueResponse, error) {
	if client.autoQueueKey == "" {
		log.Printf("Nexus AutoQueue error: AutoQueue is enabled but \"Nexus AutoQueue key\" setting is not set")
		return nil, fmt.Errorf("Nexus AutoQueue error: AutoQueue is enabled but \"Nexus AutoQueue key\" setting is not set")
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/event/%s/auto-queue?key=%s", client.BaseUrl, client.eventCode, client.autoQueueKey)
	httpClient := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonResponse autoQueueResponse
	err = json.Unmarshal(respBody, &jsonResponse)
	if err != nil {
		log.Printf("Nexus AutoQueue error: %s", string(respBody))
		return nil, fmt.Errorf("Unable to parse Nexus AutoQueue response: %d, %s", resp.StatusCode, string(respBody))
	}

	if resp.StatusCode != 200 {
		log.Printf("Nexus AutoQueue error: %d, %s", resp.StatusCode, jsonResponse.StatusText)
		return nil, fmt.Errorf("Error calling Nexus AutoQueue API: %d, %s", resp.StatusCode, jsonResponse.StatusText)
	}

	return &jsonResponse, err
}

// Sends a GET request to the Nexus API.
func (client *NexusClient) getRequest(path string) (*http.Response, error) {
	url := client.BaseUrl + path
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return httpClient.Do(req)
}
