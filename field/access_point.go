// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Linksys WRT1900ACS access point running OpenWRT for team SSIDs and VLANs.

package field

import (
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
	"text/template"
)

const accessPointSshPort = 22

const (
	red1Vlan  = 10
	red2Vlan  = 20
	red3Vlan  = 30
	blue1Vlan = 40
	blue2Vlan = 50
	blue3Vlan = 60
)

type AccessPoint struct {
	address  string
	port     int
	username string
	password string
}

func NewAccessPoint(address, username, password string) *AccessPoint {
	return &AccessPoint{address: address, port: accessPointSshPort, username: username, password: password}
}

// Sets up wireless networks for the given set of teams.
func (ap *AccessPoint) ConfigureTeamWifi(red1, red2, red3, blue1, blue2, blue3 *model.Team) error {
	config, err := generateAccessPointConfig(red1, red2, red3, blue1, blue2, blue3)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("cat <<ENDCONFIG > /etc/config/wireless && wifi radio0\n%sENDCONFIG\n", config)
	return ap.runCommand(command)
}

// Logs into the access point via SSH and runs the given shell command.
func (ap *AccessPoint) runCommand(command string) error {
	// Open an SSH connection to the AP.
	config := &ssh.ClientConfig{User: ap.username,
		Auth:            []ssh.AuthMethod{ssh.Password(ap.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey()}
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

	// Run the command. An error will be returned if the exit status is non-zero.
	return session.Run(command)
}

func generateAccessPointConfig(red1, red2, red3, blue1, blue2, blue3 *model.Team) (string, error) {
	// Determine what new SSIDs are needed.
	networks := make(map[int]*model.Team)
	var err error
	if err = addTeamNetwork(networks, red1, red1Vlan); err != nil {
		return "", err
	}
	if err = addTeamNetwork(networks, red2, red2Vlan); err != nil {
		return "", err
	}
	if err = addTeamNetwork(networks, red3, red3Vlan); err != nil {
		return "", err
	}
	if err = addTeamNetwork(networks, blue1, blue1Vlan); err != nil {
		return "", err
	}
	if err = addTeamNetwork(networks, blue2, blue2Vlan); err != nil {
		return "", err
	}
	if err = addTeamNetwork(networks, blue3, blue3Vlan); err != nil {
		return "", err
	}

	// Generate the config file to be uploaded to the AP.
	template, err := template.ParseFiles(filepath.Join(model.BaseDir, "templates/access_point.cfg"))
	if err != nil {
		return "", err
	}
	var configFile bytes.Buffer
	err = template.Execute(&configFile, networks)
	if err != nil {
		return "", err
	}

	return configFile.String(), nil
}

// Verifies the validity of the given team's WPA key and adds a network for it to the list to be configured.
func addTeamNetwork(networks map[int]*model.Team, team *model.Team, vlan int) error {
	if team == nil {
		return nil
	}
	if len(team.WpaKey) < 8 || len(team.WpaKey) > 63 {
		return fmt.Errorf("Invalid WPA key '%s' configured for team %d.", team.WpaKey, team.Id)
	}
	networks[vlan] = team
	return nil
}
