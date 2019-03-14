// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents an E1.31 sACN (DMX over Ethernet) LED controller with four outputs.

package led

import (
	"fmt"
	"net"
)

const (
	port              = 5568
	sourceName        = "Cheesy Arena"
	packetTimeoutSec  = 1
	numPixels         = 114
	pixelDataOffset   = 126
	nearStripUniverse = 1
	farStripUniverse  = 2
)

type Controller struct {
	nearStrip strip
	farStrip  strip
	conn      net.Conn
	packet    []byte
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

// Sets the current LED sequence mode and resets the intra-sequence counter to the beginning.
func (controller *Controller) SetMode(nearMode, farMode Mode) {
	if nearMode != controller.nearStrip.currentMode {
		controller.nearStrip.currentMode = nearMode
		controller.nearStrip.counter = 0
	}
	if farMode != controller.farStrip.currentMode {
		controller.farStrip.currentMode = farMode
		controller.farStrip.counter = 0
	}
}

// GetCurrentMode returns the current mode if both sides are in the same mode, or off otherwise.
func (controller *Controller) GetCurrentMode() Mode {
	if controller.nearStrip.currentMode == controller.farStrip.currentMode {
		return controller.nearStrip.currentMode
	} else {
		return OffMode
	}
}

// Sets which side of the scale or switch belongs to which alliance. A value of true indicates that the side nearest the
// scoring table is red.
func (controller *Controller) SetSidedness(nearIsRed bool) {
	controller.nearStrip.isRed = nearIsRed
	controller.farStrip.isRed = !nearIsRed
}

// Advances the pixel values through the current sequence and sends a packet if necessary. Should be called from a timed
// loop.
func (controller *Controller) Update() error {
	if controller.conn == nil {
		// This controller is not configured; do nothing.
		return nil
	}

	controller.nearStrip.updatePixels()
	controller.farStrip.updatePixels()

	// Create the template packet if it doesn't already exist.
	if len(controller.packet) == 0 {
		controller.packet = createBlankPacket(numPixels)
	}

	// Send packets if the pixel values have changed.
	if controller.nearStrip.shouldSendPacket() {
		controller.nearStrip.populatePacketPixels(controller.packet[pixelDataOffset:])
		controller.sendPacket(nearStripUniverse)
	}
	if controller.farStrip.shouldSendPacket() {
		controller.farStrip.populatePacketPixels(controller.packet[pixelDataOffset:])
		controller.sendPacket(farStripUniverse)
	}

	return nil
}

// Constructs the structure of an E1.31 data packet that can be re-used indefinitely by updating the pixel data and
// re-sending it.
func createBlankPacket(numPixels int) []byte {
	size := pixelDataOffset + 3*numPixels
	packet := make([]byte, size)

	// Preamble size
	packet[0] = 0x00
	packet[1] = 0x10

	// Postamble size
	packet[2] = 0x00
	packet[3] = 0x00

	// ACN packet identifier
	packet[4] = 0x41
	packet[5] = 0x53
	packet[6] = 0x43
	packet[7] = 0x2d
	packet[8] = 0x45
	packet[9] = 0x31
	packet[10] = 0x2e
	packet[11] = 0x31
	packet[12] = 0x37
	packet[13] = 0x00
	packet[14] = 0x00
	packet[15] = 0x00

	// Root PDU length and flags
	rootPduLength := size - 16
	packet[16] = 0x70 | byte(rootPduLength>>8)
	packet[17] = byte(rootPduLength & 0xff)

	// E1.31 vector indicating that this is a data packet
	packet[18] = 0x00
	packet[19] = 0x00
	packet[20] = 0x00
	packet[21] = 0x04

	// Component ID
	for i, b := range []byte(sourceName) {
		packet[22+i] = b
	}

	// Framing PDU length and flags
	framingPduLength := size - 38
	packet[38] = 0x70 | byte(framingPduLength>>8)
	packet[39] = byte(framingPduLength & 0xff)

	// E1.31 vector indicating that this is a data packet
	packet[40] = 0x00
	packet[41] = 0x00
	packet[42] = 0x00
	packet[43] = 0x02

	// Source name
	for i, b := range []byte(sourceName) {
		packet[44+i] = b
	}

	// Priority
	packet[108] = 100

	// Universe for synchronization packets
	packet[109] = 0x00
	packet[110] = 0x00

	// Sequence number (initial value; will be updated whenever packet is sent)
	packet[111] = 0x00

	// Options flags
	packet[112] = 0x00

	// DMX universe (will be populated whenever packet is sent)
	packet[113] = 0x00
	packet[114] = 0x00

	// DMP layer PDU length
	dmpPduLength := size - 115
	packet[115] = 0x70 | byte(dmpPduLength>>8)
	packet[116] = byte(dmpPduLength & 0xff)

	// E1.31 vector indicating set property
	packet[117] = 0x02

	// Address and data type
	packet[118] = 0xa1

	// First property address
	packet[119] = 0x00
	packet[120] = 0x00

	// Address increment
	packet[121] = 0x00
	packet[122] = 0x01

	// Property value count
	count := 1 + 3*numPixels
	packet[123] = byte(count >> 8)
	packet[124] = byte(count & 0xff)

	// DMX start code
	packet[125] = 0

	// Remainder of packet is pixel data which will be populated whenever packet is sent.
	return packet
}

func (controller *Controller) sendPacket(dmxUniverse int) error {
	// Update non-static packet fields.
	controller.packet[111]++
	controller.packet[113] = byte(dmxUniverse >> 8)
	controller.packet[114] = byte(dmxUniverse & 0xff)

	_, err := controller.conn.Write(controller.packet)
	if err != nil {
		return err
	}

	return nil
}
