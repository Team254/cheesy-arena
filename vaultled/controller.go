// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents a Philips Color Kinetics LED controller with one output, as used in the 2018 vault.

package vaultled

import (
	"fmt"
	"github.com/Team254/cheesy-arena/led"
	"net"
	"time"
)

const (
	port             = 6038
	packetTimeoutSec = 1
	numPixels        = 17
	pixelDataOffset  = 21
)

type Controller struct {
	CurrentForceMode    Mode
	CurrentLevitateMode Mode
	CurrentBoostMode    Mode
	pixels              [numPixels][3]byte
	oldPixels           [numPixels][3]byte
	conn                net.Conn
	packet              []byte
	lastPacketTime      time.Time
}

func (controller *Controller) SetAddress(address string) error {
	if controller.conn != nil {
		controller.conn.Close()
		controller.conn = nil
	}

	if address != "" {
		var err error
		if controller.conn, err = net.Dial("udp4", fmt.Sprintf("%s:%d", address, port)); err != nil {
			return err
		}
	}

	return nil
}

// Sets the current mode for the section of LEDs corresponding to the force powerup.
func (controller *Controller) SetForceMode(mode Mode) {
	controller.CurrentForceMode = mode
	controller.setPixels(0, mode)
}

// Sets the current mode for the section of LEDs corresponding to the levitate powerup.
func (controller *Controller) SetLevitateMode(mode Mode) {
	controller.CurrentLevitateMode = mode
	controller.setPixels(6, mode)
}

// Sets the current mode for the section of LEDs corresponding to the boost powerup.
func (controller *Controller) SetBoostMode(mode Mode) {
	controller.CurrentBoostMode = mode
	controller.setPixels(12, mode)
}

// Sets the current mode for all sections of LEDs to the same value.
func (controller *Controller) SetAllModes(mode Mode) {
	controller.SetForceMode(mode)
	controller.SetLevitateMode(mode)
	controller.SetBoostMode(mode)
}

// Sends a packet if necessary. Should be called from a timed loop.
func (controller *Controller) Update() error {
	if controller.conn == nil {
		// This controller is not configured; do nothing.
		return nil
	}

	// Create the template packet if it doesn't already exist.
	if len(controller.packet) == 0 {
		controller.packet = createBlankPacket(numPixels)
	}

	// Send packets if the pixel values have changed.
	if controller.shouldSendPacket() {
		controller.populatePacketPixels(controller.packet[pixelDataOffset:])
		controller.sendPacket()
	}

	return nil
}

func (controller *Controller) setPixels(offset int, mode Mode) {
	for i := 0; i < 5; i++ {
		controller.pixels[offset+i] = led.Colors[led.Black]
	}

	switch mode {
	case ThreeCubeMode:
		controller.pixels[offset+3] = led.Colors[led.Yellow]
		fallthrough
	case TwoCubeMode:
		controller.pixels[offset+2] = led.Colors[led.Yellow]
		fallthrough
	case OneCubeMode:
		controller.pixels[offset+1] = led.Colors[led.Yellow]
	case RedPlayedMode:
		for i := 0; i < 5; i++ {
			controller.pixels[offset+i] = led.Colors[led.Red]
		}
	case BluePlayedMode:
		for i := 0; i < 5; i++ {
			controller.pixels[offset+i] = led.Colors[led.Blue]
		}
	}
}

// Constructs the structure of a KiNET data packet that can be re-used indefinitely by updating the pixel data and
// re-sending it.
func createBlankPacket(numPixels int) []byte {
	size := pixelDataOffset + 3*numPixels
	packet := make([]byte, size)

	// Magic sequence
	packet[0] = 0x04
	packet[1] = 0x01
	packet[2] = 0xdc
	packet[3] = 0x4a

	// Version
	packet[4] = 0x01
	packet[5] = 0x00

	// Type
	packet[6] = 0x01
	packet[7] = 0x01

	// Sequence
	packet[8] = 0x00
	packet[9] = 0x00
	packet[10] = 0x00
	packet[11] = 0x00

	// Port
	packet[12] = 0x00

	// Padding
	packet[13] = 0x00

	// Flags
	packet[14] = 0x00
	packet[15] = 0x00

	// Timer
	packet[16] = 0xff
	packet[17] = 0xff
	packet[18] = 0xff
	packet[19] = 0xff

	// Universe
	packet[20] = 0x00

	// Remainder of packet is pixel data which will be populated whenever packet is sent.
	return packet
}

// Returns true if the pixel data has changed.
func (controller *Controller) shouldSendPacket() bool {
	for i := 0; i < numPixels; i++ {
		if controller.pixels[i] != controller.oldPixels[i] {
			return true
		}
	}
	return time.Since(controller.lastPacketTime).Seconds() > packetTimeoutSec
}

// Writes the pixel RGB values into the given packet in preparation for sending.
func (controller *Controller) populatePacketPixels(pixelData []byte) {
	for i, pixel := range controller.pixels {
		pixelData[3*i] = pixel[0]
		pixelData[3*i+1] = pixel[1]
		pixelData[3*i+2] = pixel[2]
	}

	// Keep a record of the pixel values in order to detect future changes.
	controller.oldPixels = controller.pixels
	controller.lastPacketTime = time.Now()
}

func (controller *Controller) sendPacket() error {
	_, err := controller.conn.Write(controller.packet)
	if err != nil {
		return err
	}

	return nil
}
