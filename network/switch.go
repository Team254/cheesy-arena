// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Cisco Switch 3500-series switch for team VLANs.

package network

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"net"
	"regexp"
	"strconv"
	"sync"
)

const switchTelnetPort = 23

const (
	red1Vlan  = 10
	red2Vlan  = 20
	red3Vlan  = 30
	blue1Vlan = 40
	blue2Vlan = 50
	blue3Vlan = 60
)

type Switch struct {
	address  string
	port     int
	password string
	mutex    sync.Mutex
}

var ServerIpAddress = "10.0.100.5" // The DS will try to connect to this address only.

func NewSwitch(address, password string) *Switch {
	return &Switch{address: address, port: switchTelnetPort, password: password}
}

// Sets up wired networks for the given set of teams.
func (sw *Switch) ConfigureTeamEthernet(teams [6]*model.Team) error {
	// Make sure multiple configurations aren't being set at the same time.
	sw.mutex.Lock()
	defer sw.mutex.Unlock()

	// Determine what new team VLANs are needed and build the commands to set them up.
	oldTeamVlans, err := sw.getTeamVlans()
	if err != nil {
		return err
	}
	addTeamVlansCommand := ""
	replaceTeamVlan := func(team *model.Team, vlan int) {
		if team == nil {
			return
		}
		if oldTeamVlans[team.Id] == vlan {
			delete(oldTeamVlans, team.Id)
		} else {
			addTeamVlansCommand += fmt.Sprintf(
				"ip dhcp excluded-address 10.%d.%d.1 10.%d.%d.100\n"+
					"no ip dhcp pool dhcp%d\n"+
					"ip dhcp pool dhcp%d\n"+
					"network 10.%d.%d.0 255.255.255.0\n"+
					"default-router 10.%d.%d.61\n"+
					"lease 7\n"+
					"no access-list 1%d\n"+
					"access-list 1%d permit ip 10.%d.%d.0 0.0.0.255 host %s\n"+
					"access-list 1%d permit udp any eq bootpc any eq bootps\n"+
					"interface Vlan%d\nip address 10.%d.%d.61 255.255.255.0\n",
				team.Id/100, team.Id%100, team.Id/100, team.Id%100, vlan, vlan, team.Id/100, team.Id%100, team.Id/100,
				team.Id%100, vlan, vlan, team.Id/100, team.Id%100, ServerIpAddress, vlan, vlan, team.Id/100,
				team.Id%100)
		}
	}
	replaceTeamVlan(teams[0], red1Vlan)
	replaceTeamVlan(teams[1], red2Vlan)
	replaceTeamVlan(teams[2], red3Vlan)
	replaceTeamVlan(teams[3], blue1Vlan)
	replaceTeamVlan(teams[4], blue2Vlan)
	replaceTeamVlan(teams[5], blue3Vlan)

	// Build the command to remove the team VLANs that are no longer needed.
	removeTeamVlansCommand := ""
	for _, vlan := range oldTeamVlans {
		removeTeamVlansCommand += fmt.Sprintf("interface Vlan%d\nno ip address\nno access-list 1%d\n", vlan, vlan)
	}

	// Build and run the overall command to do everything in a single telnet session.
	command := removeTeamVlansCommand + addTeamVlansCommand
	if len(command) > 0 {
		_, err = sw.runConfigCommand(removeTeamVlansCommand + addTeamVlansCommand)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns a map of currently-configured teams to VLANs.
func (sw *Switch) getTeamVlans() (map[int]int, error) {
	// Get the entire config dump.
	config, err := sw.runCommand("show running-config\n")
	if err != nil {
		return nil, err
	}

	// Parse out the team IDs and VLANs from the config dump.
	re := regexp.MustCompile("(?s)interface Vlan(\\d\\d)\\s+ip address 10\\.(\\d+)\\.(\\d+)\\.61")
	teamVlanMatches := re.FindAllStringSubmatch(config, -1)
	if teamVlanMatches == nil {
		// There are probably no teams currently configured.
		return nil, nil
	}

	// Build the map of team to VLAN.
	teamVlans := make(map[int]int)
	for _, match := range teamVlanMatches {
		team100s, _ := strconv.Atoi(match[2])
		team1s, _ := strconv.Atoi(match[3])
		team := int(team100s)*100 + team1s
		vlan, _ := strconv.Atoi(match[1])
		teamVlans[team] = vlan
	}
	return teamVlans, nil
}

// Logs into the switch via Telnet and runs the given command in user exec mode. Reads the output and
// returns it as a string.
func (sw *Switch) runCommand(command string) (string, error) {
	// Open a Telnet connection to the switch.
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", sw.address, sw.port))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Login to the AP, send the command, and log out all at once.
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(fmt.Sprintf("%s\nenable\n%s\nterminal length 0\n%sexit\n", sw.password, sw.password,
		command))
	if err != nil {
		return "", err
	}
	err = writer.Flush()
	if err != nil {
		return "", err
	}

	// Read the response.
	var reader bytes.Buffer
	_, err = reader.ReadFrom(conn)
	if err != nil {
		return "", err
	}
	return reader.String(), nil
}

// Logs into the switch via Telnet and runs the given command in global configuration mode. Reads the output
// and returns it as a string.
func (sw *Switch) runConfigCommand(command string) (string, error) {
	return sw.runCommand(fmt.Sprintf("config terminal\n%send\ncopy running-config startup-config\n\n", command))
}
