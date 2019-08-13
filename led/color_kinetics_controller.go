// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents a Philips Color Kinetics LED controller with one output, as used in the 2018 vault.

package led

import (
	"fmt"
	"net"
	"time"
)

const (
	colorKineticsPort             = 6038
	colorKineticsPacketTimeoutSec = 1
	colorKineticsNumPixels        = 17
	colorKineticsPixelDataOffset  = 21
)

type ColorKineticsController struct {
	CurrentForceMode    VaultMode
	CurrentLevitateMode VaultMode
	CurrentBoostMode    VaultMode
	pixels              [colorKineticsNumPixels][3]byte
	oldPixels           [colorKineticsNumPixels][3]byte
	conn                net.Conn
	packet              []byte
	lastPacketTime      time.Time
}

func (controller *ColorKineticsController) SetAddress(address string) error {
	if controller.conn != nil {
		controller.conn.Close()
		controller.conn = nil
	}

	if address != "" {
		var err error
		if controller.conn, err = net.Dial("udp4", fmt.Sprintf("%s:%d", address, colorKineticsPort)); err != nil {
			return err
		}
	}

	return nil
}

// Sets the current mode for the section of LEDs corresponding to the force powerup.
func (controller *ColorKineticsController) SetForceMode(mode VaultMode) {
	controller.CurrentForceMode = mode
	controller.setPixels(0, mode)
}

// Sets the current mode for the section of LEDs corresponding to the levitate powerup.
func (controller *ColorKineticsController) SetLevitateMode(mode VaultMode) {
	controller.CurrentLevitateMode = mode
	controller.setPixels(6, mode)
}

// Sets the current mode for the section of LEDs corresponding to the boost powerup.
func (controller *ColorKineticsController) SetBoostMode(mode VaultMode) {
	controller.CurrentBoostMode = mode
	controller.setPixels(12, mode)
}

// Sets the current mode for all sections of LEDs to the same value.
func (controller *ColorKineticsController) SetAllModes(mode VaultMode) {
	controller.SetForceMode(mode)
	controller.SetLevitateMode(mode)
	controller.SetBoostMode(mode)
}

// Sends a packet if necessary. Should be called from a timed loop.
func (controller *ColorKineticsController) Update() error {
	if controller.conn == nil {
		// This controller is not configured; do nothing.
		return nil
	}

	// Create the template packet if it doesn't already exist.
	if len(controller.packet) == 0 {
		controller.packet = createBlankColorKineticsPacket(colorKineticsPort)
	}

	// Send packets if the pixel values have changed.
	if controller.shouldSendPacket() {
		controller.populatePacketPixels(controller.packet[colorKineticsPixelDataOffset:])
		controller.sendPacket()
	}

	return nil
}

func (controller *ColorKineticsController) setPixels(offset int, mode VaultMode) {
	for i := 0; i < 5; i++ {
		controller.pixels[offset+i] = Colors[Black]
	}

	switch mode {
	case ThreeCubeMode:
		controller.pixels[offset+3] = Colors[Yellow]
		fallthrough
	case TwoCubeMode:
		controller.pixels[offset+2] = Colors[Yellow]
		fallthrough
	case OneCubeMode:
		controller.pixels[offset+1] = Colors[Yellow]
	case RedPlayedMode:
		for i := 0; i < 5; i++ {
			controller.pixels[offset+i] = Colors[Red]
		}
	case BluePlayedMode:
		for i := 0; i < 5; i++ {
			controller.pixels[offset+i] = Colors[Blue]
		}
	}
}

// Constructs the structure of a KiNET data packet that can be re-used indefinitely by updating the pixel data and
// re-sending it.
func createBlankColorKineticsPacket(numPixels int) []byte {
	size := colorKineticsPixelDataOffset + 3*numPixels
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
func (controller *ColorKineticsController) shouldSendPacket() bool {
	for i := 0; i < colorKineticsNumPixels; i++ {
		if controller.pixels[i] != controller.oldPixels[i] {
			return true
		}
	}
	return time.Since(controller.lastPacketTime).Seconds() > colorKineticsPacketTimeoutSec
}

// Writes the pixel RGB values into the given packet in preparation for sending.
func (controller *ColorKineticsController) populatePacketPixels(pixelData []byte) {
	for i, pixel := range controller.pixels {
		pixelData[3*i] = pixel[0]
		pixelData[3*i+1] = pixel[1]
		pixelData[3*i+2] = pixel[2]
	}

	// Keep a record of the pixel values in order to detect future changes.
	controller.oldPixels = controller.pixels
	controller.lastPacketTime = time.Now()
}

func (controller *ColorKineticsController) sendPacket() error {
	_, err := controller.conn.Write(controller.packet)
	if err != nil {
		return err
	}

	return nil
}
