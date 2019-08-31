// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Linksys WRT1900ACS access point running OpenWRT for team SSIDs and VLANs.

package network

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"golang.org/x/crypto/ssh"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	accessPointSshPort                = 22
	accessPointConnectTimeoutSec      = 1
	accessPointCommandTimeoutSec      = 5
	accessPointPollPeriodSec          = 3
	accessPointRequestBufferSize      = 10
	accessPointConfigRetryIntervalSec = 5
)

type AccessPoint struct {
	address                string
	username               string
	password               string
	teamChannel            int
	adminChannel           int
	adminWpaKey            string
	networkSecurityEnabled bool
	configRequestChan      chan [6]*model.Team
	TeamWifiStatuses       [6]TeamWifiStatus
	initialStatusesFetched bool
}

type TeamWifiStatus struct {
	TeamId      int
	RadioLinked bool
}

type sshOutput struct {
	output string
	err    error
}

func (ap *AccessPoint) SetSettings(address, username, password string, teamChannel, adminChannel int,
	adminWpaKey string, networkSecurityEnabled bool) {
	ap.address = address
	ap.username = username
	ap.password = password
	ap.teamChannel = teamChannel
	ap.adminChannel = adminChannel
	ap.adminWpaKey = adminWpaKey
	ap.networkSecurityEnabled = networkSecurityEnabled

	// Create config channel the first time this method is called.
	if ap.configRequestChan == nil {
		ap.configRequestChan = make(chan [6]*model.Team, accessPointRequestBufferSize)
	}
}

// Loops indefinitely to read status from and write configurations to the access point.
func (ap *AccessPoint) Run() {
	for {
		// Check if there are any pending configuration requests; if not, periodically poll wifi status.
		select {
		case request := <-ap.configRequestChan:
			// If there are multiple requests queued up, only consider the latest one.
			numExtraRequests := len(ap.configRequestChan)
			for i := 0; i < numExtraRequests; i++ {
				request = <-ap.configRequestChan
			}
			ap.handleTeamWifiConfiguration(request)
		case <-time.After(time.Second * accessPointPollPeriodSec):
			ap.updateTeamWifiStatuses()
		}
	}
}

// Adds a request to set up wireless networks for the given set of teams to the asynchronous queue.
func (ap *AccessPoint) ConfigureTeamWifi(teams [6]*model.Team) error {
	// Use a channel to serialize configuration requests; the monitoring goroutine will service them.
	select {
	case ap.configRequestChan <- teams:
		return nil
	default:
		return fmt.Errorf("WiFi config request buffer full")
	}
}

func (ap *AccessPoint) ConfigureAdminWifi() error {
	if !ap.networkSecurityEnabled {
		return nil
	}

	disabled := 0
	if ap.adminChannel == 0 {
		disabled = 1
	}
	commands := []string{
		fmt.Sprintf("set wireless.radio0.channel='%d'", ap.teamChannel),
		fmt.Sprintf("set wireless.radio1.disabled='%d'", disabled),
		fmt.Sprintf("set wireless.radio1.channel='%d'", ap.adminChannel),
		fmt.Sprintf("set wireless.@wifi-iface[0].key='%s'", ap.adminWpaKey),
		"commit wireless",
	}
	command := fmt.Sprintf("uci batch <<ENDCONFIG && wifi radio1\n%s\nENDCONFIG\n", strings.Join(commands, "\n"))
	_, err := ap.runCommand(command)
	return err
}

func (ap *AccessPoint) handleTeamWifiConfiguration(teams [6]*model.Team) {
	if !ap.networkSecurityEnabled {
		return
	}

	if ap.configIsCorrectForTeams(teams) {
		return
	}

	// Generate the configuration command.
	config, err := generateAccessPointConfig(teams)
	if err != nil {
		fmt.Printf("Failed to configure team WiFi: %v", err)
		return
	}
	command := fmt.Sprintf("uci batch <<ENDCONFIG && wifi radio0\n%s\nENDCONFIG\n", config)

	// Loop indefinitely at writing the configuration and reading it back until it is successfully applied.
	attemptCount := 1
	for {
		_, err := ap.runCommand(command)

		// Wait before reading the config back on write success as it doesn't take effect right away, or before retrying
		// on failure.
		time.Sleep(time.Second * accessPointConfigRetryIntervalSec)

		if err == nil {
			err = ap.updateTeamWifiStatuses()
			if err == nil && ap.configIsCorrectForTeams(teams) {
				log.Printf("Successfully configured WiFi after %d attempts.", attemptCount)
				return
			}
		}

		if err != nil {
			log.Printf("Error configuring WiFi: %v", err)
		}

		log.Printf("WiFi configuration still incorrect after %d attempts; trying again.", attemptCount)
		attemptCount++
	}
}

// Returns true if the configured networks as read from the access point match the given teams.
func (ap *AccessPoint) configIsCorrectForTeams(teams [6]*model.Team) bool {
	if !ap.initialStatusesFetched {
		return false
	}

	for i, team := range teams {
		expectedTeamId := 0
		if team != nil {
			expectedTeamId = team.Id
		}
		if ap.TeamWifiStatuses[i].TeamId != expectedTeamId {
			return false
		}
	}

	return true
}

// Fetches the current wifi network status from the access point and updates the status structure.
func (ap *AccessPoint) updateTeamWifiStatuses() error {
	if !ap.networkSecurityEnabled {
		return nil
	}

	output, err := ap.runCommand("iwinfo")
	if err == nil {
		err = decodeWifiInfo(output, ap.TeamWifiStatuses[:])
	}

	if err != nil {
		return fmt.Errorf("Error getting wifi info from AP: %v", err)
	} else {
		if !ap.initialStatusesFetched {
			ap.initialStatusesFetched = true
		}
	}
	return nil
}

// Logs into the access point via SSH and runs the given shell command.
func (ap *AccessPoint) runCommand(command string) (string, error) {
	// Open an SSH connection to the AP.
	config := &ssh.ClientConfig{User: ap.username,
		Auth:            []ssh.AuthMethod{ssh.Password(ap.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         accessPointConnectTimeoutSec * time.Second}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ap.address, accessPointSshPort), config)
	if err != nil {
		return "", err
	}
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	defer conn.Close()

	// Run the command with a timeout.
	commandChan := make(chan sshOutput, 1)
	go func() {
		outputBytes, err := session.Output(command)
		commandChan <- sshOutput{string(outputBytes), err}
	}()
	select {
	case output := <-commandChan:
		return output.output, output.err
	case <-time.After(accessPointCommandTimeoutSec * time.Second):
		return "", fmt.Errorf("WiFi SSH command timed out after %d seconds", accessPointCommandTimeoutSec)
	}
}

// Verifies WPA key validity and produces the configuration command for the given list of teams.
func generateAccessPointConfig(teams [6]*model.Team) (string, error) {
	commands := &[]string{}
	for i, team := range teams {
		position := i + 1
		if team == nil {
			*commands = append(*commands, fmt.Sprintf("set wireless.@wifi-iface[%d].disabled='0'", position),
				fmt.Sprintf("set wireless.@wifi-iface[%d].ssid='no-team-%d'", position, position),
				fmt.Sprintf("set wireless.@wifi-iface[%d].key='no-team-%d'", position, position))
		} else {
			if len(team.WpaKey) < 8 || len(team.WpaKey) > 63 {
				return "", fmt.Errorf("Invalid WPA key '%s' configured for team %d.", team.WpaKey, team.Id)
			}

			*commands = append(*commands, fmt.Sprintf("set wireless.@wifi-iface[%d].disabled='0'", position),
				fmt.Sprintf("set wireless.@wifi-iface[%d].ssid='%d'", position, team.Id),
				fmt.Sprintf("set wireless.@wifi-iface[%d].key='%s'", position, team.WpaKey))
		}
	}
	*commands = append(*commands, "commit wireless")
	return strings.Join(*commands, "\n"), nil
}

// Parses the given output from the "iwinfo" command on the AP and updates the given status structure with the result.
func decodeWifiInfo(wifiInfo string, statuses []TeamWifiStatus) error {
	ssidRe := regexp.MustCompile("ESSID: \"([-\\w ]*)\"")
	ssids := ssidRe.FindAllStringSubmatch(wifiInfo, -1)
	linkQualityRe := regexp.MustCompile("Link Quality: ([-\\w ]+)/([-\\w ]+)")
	linkQualities := linkQualityRe.FindAllStringSubmatch(wifiInfo, -1)

	// There should be at least six networks present -- one for each team on the 5GHz radio, plus one on the 2.4GHz
	// radio if the admin network is enabled.
	if len(ssids) < 6 || len(linkQualities) < 6 {
		return fmt.Errorf("Could not parse wifi info; expected 6 team networks, got %d.", len(ssids))
	}

	for i := range statuses {
		ssid := ssids[i][1]
		statuses[i].TeamId, _ = strconv.Atoi(ssid) // Any non-numeric SSIDs will be represented by a zero.
		linkQualityNumerator := linkQualities[i][1]
		statuses[i].RadioLinked = linkQualityNumerator != "unknown"
	}

	return nil
}
