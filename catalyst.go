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
	"sync"
)

const catalystTelnetPort = 23
const eventServerAddress = "10.0.0.50"

var catalystMutex sync.Mutex

// Sets up wired networks for the given set of teams.
func ConfigureTeamEthernet(red1, red2, red3, blue1, blue2, blue3 *Team) error {
	command := setupVlan(red1, red1Vlan) + setupVlan(red2, red2Vlan) + setupVlan(red3, red3Vlan) +
		setupVlan(blue1, blue1Vlan) + setupVlan(blue2, blue2Vlan) + setupVlan(blue3, blue3Vlan)
	_, err := runCatalystConfigCommand(command)
	return err
}

func setupVlan(team *Team, vlan int) string {
	if team == nil {
		return ""
	}
	return fmt.Sprintf("no access-list 1%d\naccess-list 1%d permit ip 10.%d.%d.0 0.0.0.255 host %s\n"+
		"interface Vlan%d\nip address 10.%d.%d.1 255.255.255.0\n", vlan, vlan, team.Id/100, team.Id%100,
		eventServerAddress, vlan, team.Id/100, team.Id%100)
}

// Logs into the Catalyst via Telnet and runs the given command in user exec mode. Reads the output and
// returns it as a string.
func runCatalystCommand(command string) (string, error) {
	// Make sure multiple commands aren't being run at the same time.
	catalystMutex.Lock()
	defer catalystMutex.Unlock()

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
