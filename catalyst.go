// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Cisco Catalyst 3750 switch for team VLANs.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
)

const catalystTelnetPort = 23
const eventServerAddress = "10.0.100.50"

var catalystMutex sync.Mutex

// Sets up wired networks for the given set of teams.
func ConfigureTeamEthernet(red1, red2, red3, blue1, blue2, blue3 *Team) error {
	// Make sure multiple configurations aren't being set at the same time.
	catalystMutex.Lock()
	defer catalystMutex.Unlock()

	// Determine what new team VLANs are needed and build the commands to set them up.
	oldTeamVlans, err := getTeamVlans()
	if err != nil {
		return err
	}
	addTeamVlansCommand := ""
	replaceTeamVlan := func(team *Team, vlan int) {
		if team == nil {
			return
		}
		if oldTeamVlans[team.Id] == vlan {
			delete(oldTeamVlans, team.Id)
		} else {
			addTeamVlansCommand += fmt.Sprintf("no access-list 1%d\naccess-list 1%d permit ip "+
				"10.%d.%d.0 0.0.0.255 host %s\ninterface Vlan%d\nip address 10.%d.%d.61 255.255.255.0\n", vlan,
				vlan, team.Id/100, team.Id%100, eventServerAddress, vlan, team.Id/100, team.Id%100)
		}
	}
	replaceTeamVlan(red1, red1Vlan)
	replaceTeamVlan(red2, red2Vlan)
	replaceTeamVlan(red3, red3Vlan)
	replaceTeamVlan(blue1, blue1Vlan)
	replaceTeamVlan(blue2, blue2Vlan)
	replaceTeamVlan(blue3, blue3Vlan)

	// Build the command to remove the team VLANs that are no longer needed.
	removeTeamVlansCommand := ""
	for _, vlan := range oldTeamVlans {
		removeTeamVlansCommand += fmt.Sprintf("interface Vlan%d\nno ip address\nno access-list 1%d\n", vlan, vlan)
	}

	command := removeTeamVlansCommand + addTeamVlansCommand
	if len(command) > 0 {
		_, err = runCatalystConfigCommand(removeTeamVlansCommand + addTeamVlansCommand)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns a map of currently-configured teams to VLANs.
func getTeamVlans() (map[int]int, error) {
	// Get the entire config dump.
	config, err := runCatalystCommand("show running-config\n")
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

// Logs into the Catalyst via Telnet and runs the given command in user exec mode. Reads the output and
// returns it as a string.
func runCatalystCommand(command string) (string, error) {
	// Open a Telnet connection to the switch.
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", eventSettings.SwitchAddress, catalystTelnetPort))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Login to the AP, send the command, and log out all at once.
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(fmt.Sprintf("%s\nenable\n%s\nterminal length 0\n%sexit\n",
		eventSettings.SwitchPassword, eventSettings.SwitchPassword, command))
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

// Logs into the Catalyst via Telnet and runs the given command in global configuration mode. Reads the output
// and returns it as a string.
func runCatalystConfigCommand(command string) (string, error) {
	return runCatalystCommand(fmt.Sprintf("config terminal\n%send\ncopy running-config startup-config\n\n",
		command))
}
