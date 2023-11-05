// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Linksys WRT1900ACS or Vivid-Hosting VH-109 access point running OpenWRT for team SSIDs and
// VLANs.

package network

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"golang.org/x/crypto/ssh"
)

const (
	accessPointSshPort                = 22
	accessPointConnectTimeoutSec      = 1
	accessPointCommandTimeoutSec      = 30
	accessPointPollPeriodSec          = 3
	accessPointRequestBufferSize      = 10
	accessPointConfigRetryIntervalSec = 30
	accessPointConfigBackoffSec       = 5
)

var accessPointInfoLines = []string{"ESSID: ", "Mode: ", "Tx-Power: ", "Signal: ", "Bit Rate: "}

type AccessPoint struct {
	apNumber               int
	isVividType            bool
	address                string
	username               string
	password               string
	teamChannel            int
	networkSecurityEnabled bool
	configRequestChan      chan [6]*model.Team
	TeamWifiStatuses       [6]*TeamWifiStatus
	initialStatusesFetched bool
}

type TeamWifiStatus struct {
	TeamId           int
	RadioLinked      bool
	MBits            float64
	RxRate           float64
	TxRate           float64
	SignalNoiseRatio int
}

type sshOutput struct {
	output string
	err    error
}

func (ap *AccessPoint) SetSettings(
	apNumber int,
	isVividType bool,
	address, username, password string,
	teamChannel int,
	networkSecurityEnabled bool,
	wifiStatuses [6]*TeamWifiStatus,
) {
	ap.apNumber = apNumber
	ap.isVividType = isVividType
	ap.address = address
	ap.username = username
	ap.password = password
	ap.teamChannel = teamChannel
	ap.networkSecurityEnabled = networkSecurityEnabled
	ap.TeamWifiStatuses = wifiStatuses

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
			ap.updateTeamWifiBTU()
		}
	}
}

// Calls the access point to configure the non-team-related settings.
func (ap *AccessPoint) ConfigureAdminSettings() error {
	if !ap.networkSecurityEnabled {
		return nil
	}

	var device string
	if ap.isVividType {
		device = "wifi1"
	} else {
		device = "radio0"
	}
	command := fmt.Sprintf("uci set wireless.%s.channel=%d && uci commit wireless", device, ap.teamChannel)
	_, err := ap.runCommand(command)
	return err
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

func (ap *AccessPoint) handleTeamWifiConfiguration(teams [6]*model.Team) {
	if !ap.networkSecurityEnabled {
		return
	}

	if !ap.isVividType {
		// Clear the state of the radio before loading teams; the Linksys AP is crash-prone otherwise.
		ap.configureTeams([6]*model.Team{nil, nil, nil, nil, nil, nil})
	}
	ap.configureTeams(teams)
}

func (ap *AccessPoint) configureTeams(teams [6]*model.Team) {
	retryCount := 1

	for {
		teamIndex := 0
		for teamIndex < 6 {
			config, err := ap.generateTeamAccessPointConfig(teams[teamIndex], teamIndex+1)
			if err != nil {
				log.Printf("Failed to generate WiFi configuration for AP %d: %v", ap.apNumber, err)
			}

			command := addConfigurationHeader(config)
			log.Printf("Configuring AP %d with command: %s\n", ap.apNumber, command)

			_, err = ap.runCommand(command)
			if err != nil {
				log.Printf("Error writing team configuration to AP %d: %v", ap.apNumber, err)
				retryCount++
				time.Sleep(time.Second * accessPointConfigRetryIntervalSec)
				continue
			}

			teamIndex++
		}

		_, _ = ap.runCommand("uci commit wireless")
		_, _ = ap.runCommand("wifi reload")
		if !ap.isVividType {
			// The Linksys AP returns immediately after 'wifi reload' but may not have applied the configuration yet;
			// sleep for a bit to compensate. (The Vivid AP waits for the configuration to be applied before returning.)
			time.Sleep(time.Second * accessPointConfigBackoffSec)
		}
		err := ap.updateTeamWifiStatuses()
		if err == nil && ap.configIsCorrectForTeams(teams) {
			log.Printf("Successfully configured AP %d Wi-Fi after %d attempts.", ap.apNumber, retryCount)
			break
		}
		log.Printf(
			"WiFi configuration still incorrect on AP %d after %d attempts; trying again.", ap.apNumber, retryCount,
		)
	}
}

// Returns true if the configured networks as read from the access point match the given teams.
func (ap *AccessPoint) configIsCorrectForTeams(teams [6]*model.Team) bool {
	if !ap.initialStatusesFetched {
		return false
	}

	for i, team := range teams {
		expectedTeamId := 0
		actualTeamId := 0
		if team != nil {
			expectedTeamId = team.Id
		}
		if ap.TeamWifiStatuses[i] != nil {
			actualTeamId = ap.TeamWifiStatuses[i].TeamId
		}
		if actualTeamId != expectedTeamId {
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
		ap.logWifiInfo(output)
		err = ap.decodeWifiInfo(output)
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

func addConfigurationHeader(commandList string) string {
	return fmt.Sprintf("uci batch <<ENDCONFIG\n%s\nENDCONFIG\n", commandList)
}

// Verifies WPA key validity and produces the configuration command for the given team.
func (ap *AccessPoint) generateTeamAccessPointConfig(team *model.Team, position int) (string, error) {
	if position < 1 || position > 6 {
		return "", fmt.Errorf("invalid team position %d", position)
	}

	var ssid, key string
	if team == nil {
		ssid = fmt.Sprintf("no-team-%d", position)
		key = fmt.Sprintf("no-team-%d", position)
	} else {
		if len(team.WpaKey) < 8 || len(team.WpaKey) > 63 {
			return "", fmt.Errorf("invalid WPA key '%s' configured for team %d", team.WpaKey, team.Id)
		}
		ssid = strconv.Itoa(team.Id)
		key = team.WpaKey
	}

	commands := []string{
		fmt.Sprintf("set wireless.@wifi-iface[%d].disabled='0'", position),
		fmt.Sprintf("set wireless.@wifi-iface[%d].ssid='%s'", position, ssid),
		fmt.Sprintf("set wireless.@wifi-iface[%d].key='%s'", position, key),
	}
	if ap.isVividType {
		commands = append(commands, fmt.Sprintf("set wireless.@wifi-iface[%d].sae_password='%s'", position, key))
	}

	return strings.Join(commands, "\n"), nil
}

// Filters the given output from the "iwiinfo" command on the AP and logs the relevant parts.
func (ap *AccessPoint) logWifiInfo(wifiInfo string) {
	lines := strings.Split(wifiInfo, "\n")
	var filteredLines []string
	for _, line := range lines {
		for _, infoLine := range accessPointInfoLines {
			if strings.Contains(line, infoLine) {
				filteredLines = append(filteredLines, line)
				break
			}
		}
	}
	log.Printf("AP %d status:\n%s\n", ap.apNumber, strings.Join(filteredLines, "\n"))
}

// Parses the given output from the "iwinfo" command on the AP and updates the given status structure with the result.
func (ap *AccessPoint) decodeWifiInfo(wifiInfo string) error {
	ssidRe := regexp.MustCompile("ESSID: \"([-\\w ]*)\"")
	ssids := ssidRe.FindAllStringSubmatch(wifiInfo, -1)

	// There should be six networks present -- one for each team on the 5GHz radio.
	if len(ssids) < 6 {
		return fmt.Errorf("Could not parse wifi info; expected 6 team networks, got %d.", len(ssids))
	}

	for i, wifiStatus := range ap.TeamWifiStatuses {
		if wifiStatus != nil {
			ssid := ssids[i][1]
			wifiStatus.TeamId, _ = strconv.Atoi(ssid) // Any non-numeric SSIDs will be represented by a zero.
		}
	}

	return nil
}

// Polls the 6 wlans on the ap for bandwidth use and updates data structure.
func (ap *AccessPoint) updateTeamWifiBTU() error {
	if !ap.networkSecurityEnabled {
		return nil
	}

	var interfaces []string
	if ap.isVividType {
		interfaces = []string{"ath1", "ath11", "ath12", "ath13", "ath14", "ath15"}
	} else {
		interfaces = []string{"wlan0", "wlan0-1", "wlan0-2", "wlan0-3", "wlan0-4", "wlan0-5"}
	}

	for i := range ap.TeamWifiStatuses {
		if ap.TeamWifiStatuses[i] == nil {
			continue
		}
		output, err := ap.runCommand(fmt.Sprintf("luci-bwc -i %s && iwinfo %s assoclist", interfaces[i], interfaces[i]))
		if err == nil {
			ap.TeamWifiStatuses[i].MBits = parseBtu(output)
			ap.TeamWifiStatuses[i].parseAssocList(output)
		}
		if err != nil {
			return fmt.Errorf("Error getting BTU info from AP: %v", err)
		}
	}
	return nil
}

// Parses the given data from the access point's onboard bandwidth monitor and returns five-second average bandwidth in
// megabits per second.
func parseBtu(response string) float64 {
	mBits := 0.0
	btuRe := regexp.MustCompile("\\[ (\\d+), (\\d+), (\\d+), (\\d+), (\\d+) ]")
	btuMatches := btuRe.FindAllStringSubmatch(response, -1)
	if len(btuMatches) >= 7 {
		firstMatch := btuMatches[len(btuMatches)-6]
		lastMatch := btuMatches[len(btuMatches)-1]
		rXBytes, _ := strconv.Atoi(lastMatch[2])
		tXBytes, _ := strconv.Atoi(lastMatch[4])
		rXBytesOld, _ := strconv.Atoi(firstMatch[2])
		tXBytesOld, _ := strconv.Atoi(firstMatch[4])
		mBits = float64(rXBytes-rXBytesOld+tXBytes-tXBytesOld) * 0.000008 / 5.0
	}
	return mBits
}

// Parses the given data from the access point's association list and updates the status structure with the result.
func (wifiStatus *TeamWifiStatus) parseAssocList(response string) {
	radioLinkRe := regexp.MustCompile("((?:[0-9A-F]{2}:){5}(?:[0-9A-F]{2})).*\\(SNR (\\d+)\\)\\s+(\\d+) ms ago")
	rxRateRe := regexp.MustCompile("RX:\\s+(\\d+\\.\\d+)\\s+MBit/s")
	txRateRe := regexp.MustCompile("TX:\\s+(\\d+\\.\\d+)\\s+MBit/s")

	wifiStatus.RadioLinked = false
	wifiStatus.RxRate = 0
	wifiStatus.TxRate = 0
	wifiStatus.SignalNoiseRatio = 0
	for _, radioLinkMatch := range radioLinkRe.FindAllStringSubmatch(response, -1) {
		macAddress := radioLinkMatch[1]
		dataAgeMs, _ := strconv.Atoi(radioLinkMatch[3])
		if macAddress != "00:00:00:00:00:00" && dataAgeMs <= 4000 {
			wifiStatus.RadioLinked = true
			wifiStatus.SignalNoiseRatio, _ = strconv.Atoi(radioLinkMatch[2])
			rxRateMatch := rxRateRe.FindStringSubmatch(response)
			if len(rxRateMatch) > 0 {
				wifiStatus.RxRate, _ = strconv.ParseFloat(rxRateMatch[1], 64)
			}
			txRateMatch := txRateRe.FindStringSubmatch(response)
			if len(txRateMatch) > 0 {
				wifiStatus.TxRate, _ = strconv.ParseFloat(txRateMatch[1], 64)
			}
			break
		}
	}
}
