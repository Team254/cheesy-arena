// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a Cisco Aironet AP1252AG access point for team SSIDs and VLANs.

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

const aironetTelnetPort = 23
const (
	red1Vlan  = 11
	red2Vlan  = 12
	red3Vlan  = 13
	blue1Vlan = 14
	blue2Vlan = 15
	blue3Vlan = 16
)

var aironetMutex sync.Mutex

// Sets up wireless networks for the given set of teams.
func ConfigureTeamWifi(red1, red2, red3, blue1, blue2, blue3 *Team) error {
	for _, team := range []*Team{red1, red2, red3, blue1, blue2, blue3} {
		if team != nil && (len(team.WpaKey) < 8 || len(team.WpaKey) > 63) {
			return fmt.Errorf("Invalid WPA key '%s' configured for team %d.", team.WpaKey, team.Id)
		}
	}

	// Determine what new SSIDs are needed and build the commands to set them up.
	oldSsids, err := getSsids()
	if err != nil {
		return err
	}
	addSsidsCommand := ""
	associateSsidsCommand := ""
	replaceSsid := func(team *Team, vlan int) {
		if team == nil {
			return
		}
		if oldSsids[strconv.Itoa(team.Id)] == vlan {
			delete(oldSsids, strconv.Itoa(team.Id))
		} else {
			addSsidsCommand += fmt.Sprintf("dot11 ssid %d\nvlan %d\nauthentication open\nauthentication "+
				"key-management wpa version 2\nmbssid guest-mode\nwpa-psk ascii %s\n", team.Id, vlan, team.WpaKey)
			associateSsidsCommand += fmt.Sprintf("ssid %d\n", team.Id)
		}
	}
	replaceSsid(red1, red1Vlan)
	replaceSsid(red2, red2Vlan)
	replaceSsid(red3, red3Vlan)
	replaceSsid(blue1, blue1Vlan)
	replaceSsid(blue2, blue2Vlan)
	replaceSsid(blue3, blue3Vlan)
	if len(addSsidsCommand) != 0 {
		associateSsidsCommand = "interface Dot11Radio1\n" + associateSsidsCommand
	}

	// Build the command to remove the SSIDs that are no longer needed.
	removeSsidsCommand := ""
	for ssid, _ := range oldSsids {
		removeSsidsCommand += fmt.Sprintf("no dot11 ssid %s\n", ssid)
	}

	command := removeSsidsCommand + addSsidsCommand + associateSsidsCommand
	if len(command) > 0 {
		_, err = runAironetConfigCommand(removeSsidsCommand + addSsidsCommand + associateSsidsCommand)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns a map of currently-configured SSIDs to VLANs.
func getSsids() (map[string]int, error) {
	// Get the entire config dump.
	config, err := runAironetCommand("show running-config\n")
	if err != nil {
		return nil, err
	}

	// Parse out the SSIDs and VLANs from the config dump.
	re := regexp.MustCompile("(?s)dot11 ssid (\\w+)\\s+vlan (\\d+)")
	ssidMatches := re.FindAllStringSubmatch(config, -1)
	if ssidMatches == nil {
		// There are probably no SSIDs currently configured.
		return nil, nil
	}

	// Build the map of SSID to VLAN.
	ssids := make(map[string]int)
	for _, match := range ssidMatches {
		vlan, _ := strconv.Atoi(match[2])
		ssids[match[1]] = vlan
	}
	return ssids, nil
}

// Logs into the Aironet via Telnet and runs the given command in user exec mode. Reads the output and returns
// it as a string.
func runAironetCommand(command string) (string, error) {
	// Make sure multiple commands aren't being run at the same time.
	aironetMutex.Lock()
	defer aironetMutex.Unlock()

	// Open a Telnet connection to the AP.
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", eventSettings.ApAddress, aironetTelnetPort))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Login to the AP, send the command, and log out all at once.
	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(fmt.Sprintf("%s\n%s\nterminal length 0\n%sexit\n", eventSettings.ApUsername,
		eventSettings.ApPassword, command))
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

// Logs into the Aironet via Telnet and runs the given command in global configuration mode. Reads the output
// and returns it as a string.
func runAironetConfigCommand(command string) (string, error) {
	return runAironetCommand(fmt.Sprintf("config terminal\n%send\ncopy running-config startup-config\n\n",
		command))
}
