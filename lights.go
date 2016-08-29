// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for controlling the field LED lighting.

package main

import (
	"log"
	"net"
	"time"
)

const (
	RED_DEFENSE  = "redDefense"
	BLUE_DEFENSE = "blueDefense"
)

type LightPacket [32]byte

type Lights struct {
	connections    map[string]*net.Conn
	packets        map[string]*LightPacket
	oldPackets     map[string]*LightPacket
	newConnections bool
	currentMode    string
	animationCount int
}

// Sets the color by name and transition time for the given LED channel.
func (lightPacket *LightPacket) setColorFade(channel int, color string, fade byte) {
	switch color {
	case "off":
		lightPacket.setRgbFade(channel, 0, 0, 0, fade)
	case "white":
		lightPacket.setRgbFade(channel, 15, 15, 15, fade)
	case "red":
		lightPacket.setRgbFade(channel, 15, 0, 0, fade)
	case "blue":
		lightPacket.setRgbFade(channel, 0, 0, 15, fade)
	case "green":
		lightPacket.setRgbFade(channel, 0, 15, 0, fade)
	case "yellow":
		lightPacket.setRgbFade(channel, 15, 13, 5, fade)
	case "darkred":
		lightPacket.setRgbFade(channel, 1, 0, 0, fade)
	case "darkblue":
		lightPacket.setRgbFade(channel, 0, 0, 1, fade)
	}
}

// Sets the color by name with instant transition for the given LED channel.
func (lightPacket *LightPacket) setColor(channel int, color string) {
	lightPacket.setColorFade(channel, color, 0)
}

// Sets the color by RGB values and transition time for the given LED channel.
func (lightPacket *LightPacket) setRgbFade(channel int, red byte, green byte, blue byte, fade byte) {
	lightPacket[channel*4] = red
	lightPacket[channel*4+1] = green
	lightPacket[channel*4+2] = blue
	lightPacket[channel*4+3] = fade
}

// Sets the color by name with instant transition for all LED channels.
func (lightPacket *LightPacket) setAllColor(color string) {
	lightPacket.setAllColorFade(color, 0)
}

// Sets the color by name and transition time for all LED channels.
func (lightPacket *LightPacket) setAllColorFade(color string, fade byte) {
	for i := 0; i < 8; i++ {
		lightPacket.setColorFade(i, color, fade)
	}
}

func (lights *Lights) Setup() error {
	lights.currentMode = "off"

	err := lights.SetupConnections()
	if err != nil {
		return err
	}

	lights.packets = make(map[string]*LightPacket)
	lights.packets[RED_DEFENSE] = &LightPacket{}
	lights.packets[BLUE_DEFENSE] = &LightPacket{}
	lights.oldPackets = make(map[string]*LightPacket)
	lights.oldPackets[RED_DEFENSE] = &LightPacket{}
	lights.oldPackets[BLUE_DEFENSE] = &LightPacket{}

	lights.sendLights()

	// Set up a goroutine to animate the lights when necessary.
	ticker := time.NewTicker(time.Millisecond * 50)
	go func() {
		for _ = range ticker.C {
			lights.animate()
		}
	}()
	return nil
}

func (lights *Lights) SetupConnections() error {
	lights.connections = make(map[string]*net.Conn)
	if err := lights.connect(RED_DEFENSE, eventSettings.RedDefenseLightsAddress); err != nil {
		return err
	}
	if err := lights.connect(BLUE_DEFENSE, eventSettings.BlueDefenseLightsAddress); err != nil {
		return err
	}
	lights.newConnections = true
	return nil
}

func (lights *Lights) connect(controller, address string) error {
	// Don't enable lights for a side if the controller address is not configured.
	if len(address) != 0 {
		conn, err := net.Dial("udp4", address)
		lights.connections[controller] = &conn
		if err != nil {
			return err
		}
	} else {
		lights.connections[controller] = nil
	}
	return nil
}

func (lights *Lights) ClearAll() {
	lights.packets[RED_DEFENSE].setAllColorFade("off", 10)
	lights.packets[BLUE_DEFENSE].setAllColorFade("off", 10)
	lights.sendLights()
}

// Turns all lights green to signal that the field is safe to enter.
func (lights *Lights) SetFieldReset() {
	lights.packets[RED_DEFENSE].setAllColor("green")
	lights.packets[BLUE_DEFENSE].setAllColor("green")
	lights.sendLights()
}

// Sets the lights to the given non-match mode for show or testing.
func (lights *Lights) SetMode(mode string) {
	lights.currentMode = mode
	lights.animationCount = 0

	switch mode {
	case "off":
		lights.packets[RED_DEFENSE].setAllColor("off")
		lights.packets[BLUE_DEFENSE].setAllColor("off")
	case "all_white":
		lights.packets[RED_DEFENSE].setAllColor("white")
		lights.packets[BLUE_DEFENSE].setAllColor("white")
	case "all_red":
		lights.packets[RED_DEFENSE].setAllColor("red")
		lights.packets[BLUE_DEFENSE].setAllColor("red")
	case "all_green":
		lights.packets[RED_DEFENSE].setAllColor("green")
		lights.packets[BLUE_DEFENSE].setAllColor("green")
	case "all_blue":
		lights.packets[RED_DEFENSE].setAllColor("blue")
		lights.packets[BLUE_DEFENSE].setAllColor("blue")
	}
	lights.sendLights()
}

// Sends a control packet to the LED controllers only if their state needs to be updated.
func (lights *Lights) sendLights() {
	for controller, connection := range lights.connections {
		if lights.newConnections || *lights.packets[controller] != *lights.oldPackets[controller] {
			if connection != nil {
				_, err := (*connection).Write(lights.packets[controller][:])
				if err != nil {
					log.Printf("Failed to send %s light packet.", controller)
				}
			}
		}
		*lights.oldPackets[controller] = *lights.packets[controller]
	}
	lights.newConnections = false
}

// State machine for controlling light sequences in the non-match modes.
func (lights *Lights) animate() {
	lights.animationCount += 1

	switch lights.currentMode {
	case "strobe":
		switch lights.animationCount {
		case 1:
			lights.packets[RED_DEFENSE].setAllColor("white")
			lights.packets[BLUE_DEFENSE].setAllColor("off")
		case 2:
			lights.packets[RED_DEFENSE].setAllColor("off")
			lights.packets[BLUE_DEFENSE].setAllColor("white")
			fallthrough
		default:
			lights.animationCount = 0
		}
		lights.sendLights()
	case "fade_red":
		if lights.animationCount == 1 {
			lights.packets[RED_DEFENSE].setAllColorFade("red", 18)
			lights.packets[BLUE_DEFENSE].setAllColorFade("red", 18)
		} else if lights.animationCount == 61 {
			lights.packets[RED_DEFENSE].setAllColorFade("darkred", 18)
			lights.packets[BLUE_DEFENSE].setAllColorFade("darkred", 18)
		} else if lights.animationCount > 120 {
			lights.animationCount = 0
		}
		lights.sendLights()
	case "fade_blue":
		if lights.animationCount == 1 {
			lights.packets[RED_DEFENSE].setAllColorFade("blue", 18)
			lights.packets[BLUE_DEFENSE].setAllColorFade("blue", 18)
		} else if lights.animationCount == 61 {
			lights.packets[RED_DEFENSE].setAllColorFade("darkblue", 18)
			lights.packets[BLUE_DEFENSE].setAllColorFade("darkblue", 18)
		} else if lights.animationCount > 120 {
			lights.animationCount = 0
		}
		lights.sendLights()
	case "fade_red_blue":
		if lights.animationCount == 1 {
			lights.packets[RED_DEFENSE].setAllColorFade("blue", 18)
			lights.packets[BLUE_DEFENSE].setAllColorFade("darkred", 18)
		} else if lights.animationCount == 61 {
			lights.packets[RED_DEFENSE].setAllColorFade("darkblue", 18)
			lights.packets[BLUE_DEFENSE].setAllColorFade("red", 18)
		} else if lights.animationCount > 120 {
			lights.animationCount = 0
		}
		lights.sendLights()
	}
}

// Turns on the lights below the defenses, with one channel per defense.
func (lights *Lights) SetDefenses(redDefensesStrength, blueDefensesStrength [5]int) {
	for i := 0; i < 5; i++ {
		if redDefensesStrength[i] == 0 {
			lights.packets[RED_DEFENSE].setColorFade(i, "off", 10)
		} else if redDefensesStrength[i] == 1 {
			lights.packets[RED_DEFENSE].setColorFade(i, "yellow", 10)
		} else {
			lights.packets[RED_DEFENSE].setColorFade(i, "red", 10)
		}

		if blueDefensesStrength[i] == 0 {
			lights.packets[BLUE_DEFENSE].setColorFade(i, "off", 10)
		} else if blueDefensesStrength[i] == 1 {
			lights.packets[BLUE_DEFENSE].setColorFade(i, "yellow", 10)
		} else {
			lights.packets[BLUE_DEFENSE].setColorFade(i, "blue", 10)
		}
	}
	lights.sendLights()
}
