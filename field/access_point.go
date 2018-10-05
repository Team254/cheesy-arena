// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Linksys WRT1900ACS access point running OpenWRT for team SSIDs and VLANs.

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"sync"
	"time"
)

const accessPointSshPort = 22
const accessPointConnectTimeoutSec = 1
const accessPointCommandTimeoutSec = 3

type AccessPoint struct {
	address      string
	port         int
	username     string
	password     string
	teamChannel  int
	adminChannel int
	adminWpaKey  string
	mutex        sync.Mutex
}

func NewAccessPoint(address, username, password string, teamChannel, adminChannel int, adminWpaKey string) *AccessPoint {
	return &AccessPoint{address: address, port: accessPointSshPort, username: username, password: password,
		teamChannel: teamChannel, adminChannel: adminChannel, adminWpaKey: adminWpaKey}
}

// Sets up wireless networks for the given set of teams.
func (ap *AccessPoint) ConfigureTeamWifi(red1, red2, red3, blue1, blue2, blue3 *model.Team) error {
	// Make sure multiple configurations aren't being set at the same time.
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

	config, err := ap.generateAccessPointConfig(red1, red2, red3, blue1, blue2, blue3)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("uci batch <<ENDCONFIG && wifi radio0\n%s\nENDCONFIG\n", config)
	return ap.runCommand(command)
}

func (ap *AccessPoint) ConfigureAdminWifi() error {
	// Make sure multiple configurations aren't being set at the same time.
	ap.mutex.Lock()
	defer ap.mutex.Unlock()

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
	return ap.runCommand(command)
}

// Logs into the access point via SSH and runs the given shell command.
func (ap *AccessPoint) runCommand(command string) error {
	// Open an SSH connection to the AP.
	config := &ssh.ClientConfig{User: ap.username,
		Auth:            []ssh.AuthMethod{ssh.Password(ap.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         accessPointConnectTimeoutSec * time.Second}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ap.address, ap.port), config)
	if err != nil {
		return err
	}
	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	defer conn.Close()
	session.Stdout = os.Stdout

	// Run the command with a timeout. An error will be returned if the exit status is non-zero.
	commandChan := make(chan error, 1)
	go func() {
		commandChan <- session.Run(command)
	}()
	select {
	case err = <-commandChan:
		return err
	case <-time.After(accessPointCommandTimeoutSec * time.Second):
		return fmt.Errorf("WiFi SSH command timed out after %d seconds", accessPointCommandTimeoutSec)
	}
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
		*commands = append(*commands, fmt.Sprintf("set wireless.@wifi-iface[%d].disabled='1'", position),
			fmt.Sprintf("set wireless.@wifi-iface[%d].ssid=''", position),
			fmt.Sprintf("set wireless.@wifi-iface[%d].key=''", position))
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
