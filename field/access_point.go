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
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"
)

const accessPointSshPort = 22
const accessPointConnectTimeoutSec = 1
const accessPointCommandTimeoutSec = 3
const accessPointRetryCount = 2

const (
	red1Vlan  = 10
	red2Vlan  = 20
	red3Vlan  = 30
	blue1Vlan = 40
	blue2Vlan = 50
	blue3Vlan = 60
)

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
	command := fmt.Sprintf("cat <<ENDCONFIG > /etc/config/wireless && wifi radio0\n%sENDCONFIG\n", config)
	return ap.runCommand(command)
}

func (ap *AccessPoint) ConfigureAdminWifi() error {
	config, err := ap.generateAccessPointConfig(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("cat <<ENDCONFIG > /etc/config/wireless && wifi radio1\n%sENDCONFIG\n", config)
	return ap.runCommand(command)
}

// Logs into the access point via SSH and runs the given shell command.
func (ap *AccessPoint) runCommand(command string) error {
	// Open an SSH connection to the AP.
	config := &ssh.ClientConfig{User: ap.username,
		Auth:            []ssh.AuthMethod{ssh.Password(ap.password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         accessPointConnectTimeoutSec * time.Second}

	doSsh := func() error {
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

	// Retry the SSH connection if it fails, sleeping between attempts.
	attempts := accessPointRetryCount + 1
	for {
		err := doSsh()
		if err == nil {
			return nil
		}
		attempts--
		if attempts <= 0 {
			return err
		}
		log.Printf("WiFi configuration failed; retrying %d more times: %v", attempts, err)
	}
}

func (ap *AccessPoint) generateAccessPointConfig(red1, red2, red3, blue1, blue2, blue3 *model.Team) (string, error) {
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
	data := struct {
		Networks     map[int]*model.Team
		TeamChannel  int
		AdminChannel int
		AdminWpaKey  string
	}{networks, ap.teamChannel, ap.adminChannel, ap.adminWpaKey}
	var configFile bytes.Buffer
	err = template.Execute(&configFile, data)
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
