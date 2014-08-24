// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for controlling the field LED lighting.

package main

import (
	"log"
	"net"
)

type LightPacket [24]byte

type Lights struct {
	connections    map[string]*net.Conn
	packets        map[string]*LightPacket
	oldPackets     map[string]*LightPacket
	hotGoal        string
	newConnections bool
}

func (lightPacket *LightPacket) setColor(channel int, color string) {
	switch color {
	case "red":
		lightPacket.setRgb(channel, 16, 0, 0)
	case "blue":
		lightPacket.setRgb(channel, 0, 0, 16)
	case "yellow":
		lightPacket.setRgb(channel, 16, 16, 0)
	default:
		lightPacket.setRgb(channel, 13, 12, 11)
	}
}

func (lightPacket *LightPacket) setRgb(channel int, red byte, green byte, blue byte) {
	lightPacket[channel*3] = red
	lightPacket[channel*3+1] = green
	lightPacket[channel*3+2] = blue
}

func (lights *Lights) Setup() error {
	err := lights.SetupConnections()
	if err != nil {
		return err
	}

	lights.packets = make(map[string]*LightPacket)
	lights.packets["red"] = &LightPacket{}
	lights.packets["blue"] = &LightPacket{}
	lights.oldPackets = make(map[string]*LightPacket)
	lights.oldPackets["red"] = &LightPacket{}
	lights.oldPackets["blue"] = &LightPacket{}

	lights.sendLights()
	return nil
}

func (lights *Lights) SetupConnections() error {
	lights.connections = make(map[string]*net.Conn)

	// Don't enable lights for a side if the controller address is not configured.
	if len(eventSettings.RedGoalLightsAddress) != 0 {
		conn, err := net.Dial("udp4", eventSettings.RedGoalLightsAddress)
		lights.connections["red"] = &conn
		if err != nil {
			return err
		}
	}
	if len(eventSettings.BlueGoalLightsAddress) != 0 {
		conn, err := net.Dial("udp4", eventSettings.BlueGoalLightsAddress)
		lights.connections["blue"] = &conn
		if err != nil {
			return err
		}
	}
	lights.newConnections = true
	return nil
}

func (lights *Lights) SetHotGoal(alliance string, leftSide bool) {
	if leftSide {
		lights.packets[alliance].setRgb(0, 0, 0, 0)
		lights.packets[alliance].setRgb(1, 0, 0, 0)
		lights.packets[alliance].setRgb(2, 0, 0, 0)
		lights.packets[alliance].setColor(3, "yellow")
		lights.packets[alliance].setColor(4, "yellow")
		lights.packets[alliance].setColor(5, "yellow")
		lights.hotGoal = "left"
	} else {
		lights.packets[alliance].setColor(0, "yellow")
		lights.packets[alliance].setColor(1, "yellow")
		lights.packets[alliance].setColor(2, "yellow")
		lights.packets[alliance].setRgb(3, 0, 0, 0)
		lights.packets[alliance].setRgb(4, 0, 0, 0)
		lights.packets[alliance].setRgb(5, 0, 0, 0)
		lights.hotGoal = "right"
	}
	lights.sendLights()
}

func (lights *Lights) SetAssistGoal(alliance string, numAssists int) {
	lights.packets[alliance].setRgb(0, 0, 0, 0)
	lights.packets[alliance].setRgb(1, 0, 0, 0)
	lights.packets[alliance].setRgb(2, 0, 0, 0)
	lights.packets[alliance].setRgb(3, 0, 0, 0)
	lights.packets[alliance].setRgb(4, 0, 0, 0)
	lights.packets[alliance].setRgb(5, 0, 0, 0)
	if numAssists > 0 {
		lights.packets[alliance].setColor(2, alliance)
		lights.packets[alliance].setColor(3, alliance)
	}
	if numAssists > 1 {
		lights.packets[alliance].setColor(1, alliance)
		lights.packets[alliance].setColor(4, alliance)
	}
	if numAssists > 2 {
		lights.packets[alliance].setColor(0, alliance)
		lights.packets[alliance].setColor(5, alliance)
	}
	lights.hotGoal = ""
	lights.sendLights()
}

func (lights *Lights) ClearGoal(alliance string) {
	lights.SetAssistGoal(alliance, 0)
}

func (lights *Lights) SetPedestal(alliance string) {
	if alliance == "red" {
		lights.packets["blue"].setColor(6, alliance)
	} else {
		lights.packets["red"].setColor(6, alliance)
	}
	lights.sendLights()
}

func (lights *Lights) ClearPedestal(alliance string) {
	if alliance == "red" {
		lights.packets["blue"].setRgb(6, 0, 0, 0)
	} else {
		lights.packets["red"].setRgb(6, 0, 0, 0)
	}
	lights.sendLights()
}

func (lights *Lights) sendLights() {
	for alliance, connections := range lights.connections {
		if lights.newConnections || *lights.packets[alliance] != *lights.oldPackets[alliance] {
			_, err := (*connections).Write(lights.packets[alliance][:])
			if err != nil {
				log.Printf("Failed to send %s light packet.", alliance)
			}
			mainArena.hotGoalLightNotifier.Notify(lights.hotGoal)
		}
		*lights.oldPackets[alliance] = *lights.packets[alliance]
	}
	lights.newConnections = false
}
