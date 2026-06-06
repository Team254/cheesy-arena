// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents an E1.31 sACN (DMX over Ethernet) LED controller for the 2026 hub lights.

package led

import (
	"fmt"
	"net"
	"time"
)

const (
	port                 = 5568
	sourceName           = "Cheesy Arena"
	pixelDataOffset      = 126
	channelsPerPixel     = 3
	universeChannelCount = 512
	channelsPerFixture   = channelsPerPixel * pixelsPerFixture
)

type Controller struct {
	redZone   zone
	blueZone  zone
	conn      net.Conn
	fixtures  fixtureLayout
	universes map[int]*universe
	packet    []byte
}

type universe struct {
	currentData    [universeChannelCount]byte
	oldData        [universeChannelCount]byte
	lastPacketTime time.Time
	sequence       byte
}

// NewController creates a controller with both alliance LED zones off.
func NewController() *Controller {
	return &Controller{
		redZone:   zone{currentMode: OffMode},
		blueZone:  zone{currentMode: OffMode},
		fixtures:  defaultFixtureLayout,
		universes: map[int]*universe{},
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
	if redMode != controller.redZone.currentMode {
		controller.redZone.currentMode = redMode
		controller.redZone.counter = 0
	}
	if blueMode != controller.blueZone.currentMode {
		controller.blueZone.currentMode = blueMode
		controller.blueZone.counter = 0
	}
}

// GetModes returns the current mode for each alliance side.
func (controller *Controller) GetModes() (Mode, Mode) {
	return controller.redZone.currentMode, controller.blueZone.currentMode
}

// Update advances the pixel values through the current sequence and sends a packet if necessary. Should be called from
// a timed loop.
func (controller *Controller) Update() error {
	if controller.conn == nil {
		// This controller is not configured; do nothing.
		return nil
	}

	controller.redZone.updatePixels()
	controller.blueZone.updatePixels()

	// Create the template packet if it doesn't already exist.
	if len(controller.packet) == 0 {
		controller.packet = createBlankPacket(universeChannelCount)
	}

	for _, universe := range controller.universes {
		universe.currentData = [universeChannelCount]byte{}
	}

	if err := controller.populateFixtureData(&controller.redZone, controller.fixtures.red); err != nil {
		return err
	}
	if err := controller.populateFixtureData(&controller.blueZone, controller.fixtures.blue); err != nil {
		return err
	}

	for dmxUniverse, universe := range controller.universes {
		if universe.shouldSendPacket() {
			if err := controller.sendPacket(dmxUniverse, universe); err != nil {
				return err
			}
		}
	}

	return nil
}

func (controller *Controller) populateFixtureData(zone *zone, fixtures []fixture) error {
	for i, fixture := range fixtures {
		if fixture.universe <= 0 {
			return fmt.Errorf("invalid universe %d for fixture %d", fixture.universe, fixture.id)
		}

		startIndex := fixture.startAddress - 1
		if startIndex < 0 || startIndex+channelsPerFixture > universeChannelCount {
			return fmt.Errorf("invalid start address %d for fixture %d", fixture.startAddress, fixture.id)
		}

		universeData, ok := controller.universes[fixture.universe]
		if !ok {
			universeData = &universe{}
			controller.universes[fixture.universe] = universeData
		}

		pixelStart := i * pixelsPerFixture
		for j := 0; j < pixelsPerFixture; j++ {
			pixel := zone.pixels[pixelStart+j]
			channelStart := startIndex + j*channelsPerPixel
			universeData.currentData[channelStart] = pixel.R
			universeData.currentData[channelStart+1] = pixel.G
			universeData.currentData[channelStart+2] = pixel.B
		}
	}
	return nil
}

// shouldSendPacket returns true if the universe data has changed or it has been too long since the last packet was sent.
func (universe *universe) shouldSendPacket() bool {
	if universe.lastPacketTime.IsZero() || time.Since(universe.lastPacketTime) >= heartbeatInterval {
		return true
	}
	return universe.currentData != universe.oldData
}

func (universe *universe) markSent() {
	universe.oldData = universe.currentData
	universe.lastPacketTime = time.Now()
}

// createBlankPacket constructs the structure of an E1.31 data packet that can be re-used indefinitely by updating the
// pixel data and re-sending it.
func createBlankPacket(channelCount int) []byte {
	size := pixelDataOffset + channelCount
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
	count := 1 + channelCount
	packet[123] = byte(count >> 8)
	packet[124] = byte(count & 0xff)

	// DMX start code
	packet[125] = 0

	// Remainder of packet is pixel data which will be populated whenever packet is sent.
	return packet
}

// sendPacket sends the current packet buffer to the given DMX universe.
func (controller *Controller) sendPacket(dmxUniverse int, universe *universe) error {
	// Update non-static packet fields.
	universe.sequence++
	controller.packet[111] = universe.sequence
	controller.packet[113] = byte(dmxUniverse >> 8)
	controller.packet[114] = byte(dmxUniverse & 0xff)
	copy(controller.packet[pixelDataOffset:], universe.currentData[:])

	_, err := controller.conn.Write(controller.packet)
	if err != nil {
		return err
	}

	universe.markSent()
	return nil
}
