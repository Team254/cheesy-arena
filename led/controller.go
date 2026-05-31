// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents an E1.31 sACN (DMX over Ethernet) LED controller for the 2026 hub lights.

package led

import (
	"fmt"
	"net"
)

const (
	port              = 5568
	sourceName        = "Cheesy Arena"
	pixelDataOffset   = 126
	redStripUniverse  = 1
	blueStripUniverse = 2
)

type Controller struct {
	redStrip  strip
	blueStrip strip
	conn      net.Conn
	packet    []byte
}

// NewController creates a controller with both alliance LED strips off.
func NewController() *Controller {
	return &Controller{
		redStrip:  strip{currentMode: OffMode},
		blueStrip: strip{currentMode: OffMode},
	}
}

// SetAddress sets the controller address, or disables output if the address is blank.
func (controller *Controller) SetAddress(address string) error {
	if controller.conn != nil {
		_ = controller.conn.Close()
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

// SetMode sets the current LED sequence mode and resets the intra-sequence counter to the beginning if the new mode
// is different from the current mode.
func (controller *Controller) SetMode(redMode, blueMode Mode) {
	if redMode != controller.redStrip.currentMode {
		controller.redStrip.currentMode = redMode
		controller.redStrip.counter = 0
	}
	if blueMode != controller.blueStrip.currentMode {
		controller.blueStrip.currentMode = blueMode
		controller.blueStrip.counter = 0
	}
}

// GetModes returns the current mode for each alliance side.
func (controller *Controller) GetModes() (Mode, Mode) {
	return controller.redStrip.currentMode, controller.blueStrip.currentMode
}

// Update advances the pixel values through the current sequence and sends a packet if necessary. Should be called from
// a timed loop.
func (controller *Controller) Update() error {
	if controller.conn == nil {
		// This controller is not configured; do nothing.
		return nil
	}

	controller.redStrip.updatePixels()
	controller.blueStrip.updatePixels()

	// Create the template packet if it doesn't already exist.
	if len(controller.packet) == 0 {
		controller.packet = createBlankPacket(numPixels)
	}

	// Send packets if the pixel values have changed.
	if controller.redStrip.shouldSendPacket() {
		controller.redStrip.populatePacketPixels(controller.packet[pixelDataOffset:])
		if err := controller.sendPacket(redStripUniverse); err != nil {
			return err
		}
	}
	if controller.blueStrip.shouldSendPacket() {
		controller.blueStrip.populatePacketPixels(controller.packet[pixelDataOffset:])
		if err := controller.sendPacket(blueStripUniverse); err != nil {
			return err
		}
	}

	return nil
}

// createBlankPacket constructs the structure of an E1.31 data packet that can be re-used indefinitely by updating the
// pixel data and re-sending it.
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

// sendPacket sends the current packet buffer to the given DMX universe.
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
