// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for pulling match lineups from Nexus for FRC.

package partner

import (
	"encoding/json"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"io"
	"net/http"
	"strconv"
)

const nexusBaseUrl = "https://api.frc.nexus"

type NexusClient struct {
	BaseUrl   string
	eventCode string
}

type nexusLineup struct {
	Red  [3]string `json:"red"`
	Blue [3]string `json:"blue"`
}

func NewNexusClient(eventCode string) *NexusClient {
	return &NexusClient{BaseUrl: nexusBaseUrl, eventCode: eventCode}
}

// Gets the team lineup for a given match from the Nexus API. Returns nil and an error if the lineup is not available.
func (client *NexusClient) GetLineup(tbaMatchKey model.TbaMatchKey) (*[6]int, error) {
	path := fmt.Sprintf("/v1/%s/%s/lineup", client.eventCode, tbaMatchKey.String())
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
	lineup[0], _ = strconv.Atoi(nexusLineup.Red[0])
	lineup[1], _ = strconv.Atoi(nexusLineup.Red[1])
	lineup[2], _ = strconv.Atoi(nexusLineup.Red[2])
	lineup[3], _ = strconv.Atoi(nexusLineup.Blue[0])
	lineup[4], _ = strconv.Atoi(nexusLineup.Blue[1])
	lineup[5], _ = strconv.Atoi(nexusLineup.Blue[2])

	// Check that each spot is filled with a valid team number; otherwise return an error.
	for _, team := range lineup {
		if team == 0 {
			return nil, fmt.Errorf("Lineup not yet submitted")
		}
	}

	return &lineup, err
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
