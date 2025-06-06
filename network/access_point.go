// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Vivid-Hosting VH-113 access point running OpenWRT for team SSIDs and VLANs.

package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Team254/cheesy-arena/model"
)

const (
	accessPointPollPeriodSec = 1
)

type AccessPoint struct {
	apiUrl                 string
	password               string
	channel                int
	networkSecurityEnabled bool
	Status                 string
	TeamWifiStatuses       [6]*TeamWifiStatus
	lastConfiguredTeams    [6]*model.Team
}

type TeamWifiStatus struct {
	TeamId            int
	RadioLinked       bool
	MBits             float64
	RxRate            float64
	TxRate            float64
	SignalNoiseRatio  int
	ConnectionQuality int
}

type configurationRequest struct {
	Channel               int                             `json:"channel"`
	StationConfigurations map[string]stationConfiguration `json:"stationConfigurations"`
}

type stationConfiguration struct {
	Ssid   string `json:"ssid"`
	WpaKey string `json:"wpaKey"`
}

type accessPointStatus struct {
	Channel         int                       `json:"channel"`
	Status          string                    `json:"status"`
	StationStatuses map[string]*stationStatus `json:"stationStatuses"`
}

type stationStatus struct {
	Ssid              string  `json:"ssid"`
	HashedWpaKey      string  `json:"hashedWpaKey"`
	WpaKeySalt        string  `json:"wpaKeySalt"`
	IsLinked          bool    `json:"isLinked"`
	RxRateMbps        float64 `json:"rxRateMbps"`
	TxRateMbps        float64 `json:"txRateMbps"`
	SignalNoiseRatio  int     `json:"signalNoiseRatio"`
	BandwidthUsedMbps float64 `json:"bandwidthUsedMbps"`
	ConnectionQuality string  `json:"connectionQuality"`
}

var connectionQualityMap = map[string]int{
	"caution":   1,
	"warning":   2,
	"good":      3,
	"excellent": 4,
}

func (ap *AccessPoint) SetSettings(
	address, password string,
	channel int,
	networkSecurityEnabled bool,
	wifiStatuses [6]*TeamWifiStatus,
) {
	ap.apiUrl = fmt.Sprintf("http://%s", address)
	ap.password = password
	ap.channel = channel
	ap.networkSecurityEnabled = networkSecurityEnabled
	ap.Status = "UNKNOWN"
	ap.TeamWifiStatuses = wifiStatuses
}

// Loops indefinitely to read status from the access point.
func (ap *AccessPoint) Run() {
	for {
		time.Sleep(time.Second * accessPointPollPeriodSec)
		if err := ap.updateMonitoring(); err != nil {
			log.Printf("Failed to update access point monitoring: %v", err)
			continue
		}

		// If the access point is in a good state but doesn't match the expected configuration, try again.
		if ap.Status == "ACTIVE" && !ap.statusMatchesLastConfiguration() {
			log.Println("Access point is ACTIVE but does not match expected configuration; retrying configuration.")
			if err := ap.ConfigureTeamWifi(ap.lastConfiguredTeams); err != nil {
				log.Printf("Failed to reconfigure access point: %v", err)
			}
		}
	}
}

// Calls the access point's API to configure the team SSIDs and WPA keys.
func (ap *AccessPoint) ConfigureTeamWifi(teams [6]*model.Team) error {
	if !ap.networkSecurityEnabled {
		return nil
	}

	ap.Status = "CONFIGURING"
	ap.lastConfiguredTeams = teams
	request := configurationRequest{
		Channel:               ap.channel,
		StationConfigurations: make(map[string]stationConfiguration),
	}
	addStation(request.StationConfigurations, "red1", teams[0])
	addStation(request.StationConfigurations, "red2", teams[1])
	addStation(request.StationConfigurations, "red3", teams[2])
	addStation(request.StationConfigurations, "blue1", teams[3])
	addStation(request.StationConfigurations, "blue2", teams[4])
	addStation(request.StationConfigurations, "blue3", teams[5])
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Send the configuration to the access point API.
	url := ap.apiUrl + "/configuration"
	httpRequest, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	if ap.password != "" {
		httpRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ap.password))
	}
	httpClient := http.Client{Timeout: time.Second * 3}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode/100 != 2 {
		body, _ := io.ReadAll(httpResponse.Body)
		return fmt.Errorf("access point returned status %d: %s", httpResponse.StatusCode, string(body))
	}

	log.Println("Access point accepted the new configuration and will apply it asynchronously.")
	return nil
}

// Fetches the current access point status from the API and updates the status structure.
func (ap *AccessPoint) updateMonitoring() error {
	if !ap.networkSecurityEnabled {
		return nil
	}

	// Fetch the status from the access point API.
	url := ap.apiUrl + "/status"
	httpRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if ap.password != "" {
		httpRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ap.password))
	}
	var httpClient http.Client
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		ap.Status = "ERROR"
		return fmt.Errorf("failed to fetch access point status: %v", err)
	}
	if httpResponse.StatusCode/100 != 2 {
		ap.Status = "ERROR"
		body, _ := io.ReadAll(httpResponse.Body)
		return fmt.Errorf("access point returned status %d: %s", httpResponse.StatusCode, string(body))
	}

	// Parse the response and populate the status structure.
	var apStatus accessPointStatus
	err = json.NewDecoder(httpResponse.Body).Decode(&apStatus)
	if err != nil {
		ap.Status = "ERROR"
		return fmt.Errorf("failed to parse access point status: %v", err)
	}
	if ap.Status != apStatus.Status {
		log.Printf("Access point status changed from %s to %s.", ap.Status, apStatus.Status)
		ap.Status = apStatus.Status
		if ap.Status == "ACTIVE" {
			log.Printf("Access point detailed status:\n%s", apStatus.toLogString())
		}
	}
	updateTeamWifiStatus(ap.TeamWifiStatuses[0], apStatus.StationStatuses["red1"])
	updateTeamWifiStatus(ap.TeamWifiStatuses[1], apStatus.StationStatuses["red2"])
	updateTeamWifiStatus(ap.TeamWifiStatuses[2], apStatus.StationStatuses["red3"])
	updateTeamWifiStatus(ap.TeamWifiStatuses[3], apStatus.StationStatuses["blue1"])
	updateTeamWifiStatus(ap.TeamWifiStatuses[4], apStatus.StationStatuses["blue2"])
	updateTeamWifiStatus(ap.TeamWifiStatuses[5], apStatus.StationStatuses["blue3"])

	return nil
}

// Returns true if the access point's current status matches the last configuration that was sent to it.
func (ap *AccessPoint) statusMatchesLastConfiguration() bool {
	for i := 0; i < 6; i++ {
		var expectedTeamId, actualTeamId int
		if ap.lastConfiguredTeams[i] != nil {
			expectedTeamId = ap.lastConfiguredTeams[i].Id
		}
		if ap.TeamWifiStatuses[i] != nil {
			actualTeamId = ap.TeamWifiStatuses[i].TeamId
		}
		if expectedTeamId != actualTeamId {
			return false
		}
	}
	return true
}

// Generates the configuration for the given team's station and adds it to the map. If the team is nil, no entry is
// added for the station.
func addStation(stationsConfigurations map[string]stationConfiguration, station string, team *model.Team) {
	if team == nil {
		return
	}
	stationsConfigurations[station] = stationConfiguration{
		Ssid:   strconv.Itoa(team.Id),
		WpaKey: team.WpaKey,
	}
}

// Updates the given team's wifi status structure with the given station status.
func updateTeamWifiStatus(teamWifiStatus *TeamWifiStatus, stationStatus *stationStatus) {
	if stationStatus == nil {
		teamWifiStatus.TeamId = 0
		teamWifiStatus.RadioLinked = false
		teamWifiStatus.MBits = 0
		teamWifiStatus.RxRate = 0
		teamWifiStatus.TxRate = 0
		teamWifiStatus.SignalNoiseRatio = 0
		teamWifiStatus.ConnectionQuality = 0
	} else {
		teamWifiStatus.TeamId, _ = strconv.Atoi(stationStatus.Ssid)
		teamWifiStatus.RadioLinked = stationStatus.IsLinked
		teamWifiStatus.MBits = stationStatus.BandwidthUsedMbps
		teamWifiStatus.RxRate = stationStatus.RxRateMbps
		teamWifiStatus.TxRate = stationStatus.TxRateMbps
		teamWifiStatus.SignalNoiseRatio = stationStatus.SignalNoiseRatio
		if quality, ok := connectionQualityMap[stationStatus.ConnectionQuality]; ok {
			teamWifiStatus.ConnectionQuality = quality
		} else {
			// Default to 0 if there is no mapping for the connection quality string.
			teamWifiStatus.ConnectionQuality = 0
		}
	}
}

// Returns an abbreviated string representation of the access point status for inclusion in the log.
func (apStatus *accessPointStatus) toLogString() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Channel: %d\n", apStatus.Channel))
	for _, station := range []string{"red1", "red2", "red3", "blue1", "blue2", "blue3"} {
		stationStatus := apStatus.StationStatuses[station]
		ssid := "[empty]"
		if stationStatus != nil {
			ssid = stationStatus.Ssid
		}
		buffer.WriteString(fmt.Sprintf("%-6s %s\n", station+":", ssid))
	}
	return buffer.String()
}
