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
	"sync"
	"time"
)

const (
	accessPointSshPort           = 22
	accessPointConnectTimeoutSec = 1
	accessPointPollPeriodSec     = 3
)

type AccessPoint struct {
	address                string
	username               string
	password               string
	teamChannel            int
	adminChannel           int
	adminWpaKey            string
	networkSecurityEnabled bool
	mutex                  sync.Mutex
	TeamWifiStatuses       [6]TeamWifiStatus
}

type TeamWifiStatus struct {
	TeamId      int
	RadioLinked bool
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
}

// Loops indefinitely to read status from the access point.
func (ap *AccessPoint) Run() {
	for {
		if ap.networkSecurityEnabled {
			ap.updateTeamWifiStatuses()
		}

		time.Sleep(time.Second * accessPointPollPeriodSec)
	}
}

// Sets up wireless networks for the given set of teams.
func (ap *AccessPoint) ConfigureTeamWifi(red1, red2, red3, blue1, blue2, blue3 *model.Team) error {
	config, err := ap.generateAccessPointConfig(red1, red2, red3, blue1, blue2, blue3)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("uci batch <<ENDCONFIG && wifi radio0\n%s\nENDCONFIG\n", config)
	_, err = ap.runCommand(command)
	return err
}

func (ap *AccessPoint) ConfigureAdminWifi() error {
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
	command := fmt.Sprintf("uci batch <<ENDCONFIG && wifi\n%s\nENDCONFIG\n", strings.Join(commands, "\n"))
	_, err := ap.runCommand(command)
	return err
}

// Logs into the access point via SSH and runs the given shell command.
func (ap *AccessPoint) runCommand(command string) (string, error) {
	// Make sure multiple commands aren't being run at the same time.
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

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

	outputBytes, err := session.Output(command)
	return string(outputBytes), err
}

func (ap *AccessPoint) generateAccessPointConfig(red1, red2, red3, blue1, blue2, blue3 *model.Team) (string, error) {
	// Determine what new SSIDs are needed.
	commands := &[]string{}
	var err error
	if err = addTeamConfigCommands(1, red1, commands); err != nil {
		return "", err
	}
	if err = addTeamConfigCommands(2, red2, commands); err != nil {
		return "", err
	}
	if err = addTeamConfigCommands(3, red3, commands); err != nil {
		return "", err
	}
	if err = addTeamConfigCommands(4, blue1, commands); err != nil {
		return "", err
	}
	if err = addTeamConfigCommands(5, blue2, commands); err != nil {
		return "", err
	}
	if err = addTeamConfigCommands(6, blue3, commands); err != nil {
		return "", err
	}

	*commands = append(*commands, "commit wireless")

	return strings.Join(*commands, "\n"), nil
}

// Verifies the validity of the given team's WPA key and adds a network for it to the list to be configured.
func addTeamConfigCommands(position int, team *model.Team, commands *[]string) error {
	if team == nil {
		*commands = append(*commands, fmt.Sprintf("set wireless.@wifi-iface[%d].disabled='0'", position),
			fmt.Sprintf("set wireless.@wifi-iface[%d].ssid='no-team-%d'", position, position),
			fmt.Sprintf("set wireless.@wifi-iface[%d].key='no-team-%d'", position, position))
	} else {
		if len(team.WpaKey) < 8 || len(team.WpaKey) > 63 {
			return fmt.Errorf("Invalid WPA key '%s' configured for team %d.", team.WpaKey, team.Id)
		}

		*commands = append(*commands, fmt.Sprintf("set wireless.@wifi-iface[%d].disabled='0'", position),
			fmt.Sprintf("set wireless.@wifi-iface[%d].ssid='%d'", position, team.Id),
			fmt.Sprintf("set wireless.@wifi-iface[%d].key='%s'", position, team.WpaKey))
	}

	return nil
}

// Fetches the current wifi network status from the access point and updates the status structure.
func (ap *AccessPoint) updateTeamWifiStatuses() {
	output, err := ap.runCommand("iwinfo")
	if err != nil {
		log.Printf("Error getting wifi info from AP: %v", err)
		return
	}

	if err := decodeWifiInfo(output, ap.TeamWifiStatuses[:]); err != nil {
		log.Println(err.Error())
	}
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
