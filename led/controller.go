// Copyright 2025 Team 254. All Rights Reserved.
// Author: kwaremburg
//
// DMX controller interface for controlling RGB light bars in the hubs.

package led

import (
	"fmt"
	"log"
	"net"
	"time"
)

const (
	sACNPort          = 5568
	sACNSourceName    = "Cheesy Arena"
	sourceName        = "Cheesy Arena"
	pixelDataOffset   = 126
	heartbeatInterval = 1 * time.Second
)

// Color represents an RGB color value.
type Color struct {
	R, G, B uint8
}

func (c Color) Equals(other Color) bool {
	return c.R == other.R && c.G == other.G && c.B == other.B
}

// Predefined colors for different states
var (
	ColorOff    = Color{0, 0, 0}     // Off (inactive hub)
	ColorGreen  = Color{0, 255, 0}   // Green (field safe)
	ColorPurple = Color{128, 0, 128} // Purple (counting after match)
	ColorRed    = Color{255, 0, 0}   // Red (red alliance active hub)
	ColorBlue   = Color{0, 0, 255}   // Blue (blue alliance active hub)
)

// Controller implements the DMX light control for sACN E1.31 over Ethernet.
type Controller struct {
	Address      string
	Universe     int
	conn         net.Conn
	color        Color
	lastColor    Color
	lastSend     time.Time
	StartChannel int // Starting channel for the hub (7 channels)
	packet       []byte
}

func (dmx *Controller) SetAddress(address string) error {
	if dmx.conn != nil {
		dmx.conn.Close()
		dmx.conn = nil
	}

	dmx.Address = address
	if address != "" {
		var err error
		if dmx.conn, err = net.Dial("udp4", fmt.Sprintf("%s:%d", address, sACNPort)); err != nil {
			return err
		}
	}

	return nil
}

func (dmx *Controller) SetColor(color Color) {
	dmx.color = color
}

func (dmx *Controller) GetColor() Color {
	return dmx.color
}

func (dmx *Controller) Close() {
	if dmx.conn != nil {
		dmx.conn.Close()
		dmx.conn = nil
	}
}

func (dmx *Controller) Update() {
	color := dmx.color

	if dmx.conn == nil {
		// This controller is not configured; do nothing.
		return
	}

	// Create the template packet if it doesn't already exist.
	if len(dmx.packet) == 0 {
		dmx.packet = createBlankPacket(3)
	}

	// Send packets if the pixel values have changed.
	if dmx.shouldSendPacket(color) {
		dmx.populatePacket(color, dmx.StartChannel)
		if err := dmx.sendPacket(dmx.Universe); err != nil {
			log.Printf("sACN error writing data to universe %d: %v", dmx.Universe, err)
			return
		}
		dmx.lastColor = color
		dmx.lastSend = time.Now()
	}
}

func (dmx *Controller) shouldSendPacket(color Color) bool {
	if !color.Equals(dmx.lastColor) {
		return true
	}
	return time.Since(dmx.lastSend) >= heartbeatInterval
}

func (dmx *Controller) populatePacket(color Color, startChannel int) {
	// Clear DMX data area
	for i := pixelDataOffset; i < len(dmx.packet); i++ {
		dmx.packet[i] = 0
	}

	// Light mapping (3 channels):
	// 1: Red
	// 2: Green
	// 3: Blue
	dmx.packet[pixelDataOffset+startChannel-1+0] = color.R
	dmx.packet[pixelDataOffset+startChannel-1+1] = color.G
	dmx.packet[pixelDataOffset+startChannel-1+2] = color.B
}

func (dmx *Controller) sendPacket(universe int) error {
	dmx.packet[111]++ // Sequence number
	dmx.packet[113] = byte(universe >> 8)
	dmx.packet[114] = byte(universe & 0xff)
	_, err := dmx.conn.Write(dmx.packet)
	return err
}

func putFlagsLength(pkt []byte, off int, pduLen int) {
	fl := 0x7000 | (pduLen & 0x0FFF)
	pkt[off] = byte(fl >> 8)
	pkt[off+1] = byte(fl)
}

func createBlankPacket(numChannels int) []byte {
	size := pixelDataOffset + numChannels + 3
	packet := make([]byte, size)

	// Preamble size
	packet[0] = 0x00
	packet[1] = 0x10

	// Postamble size
	packet[2] = 0x00
	packet[3] = 0x00

	// ACN packet identifier
	// copy(packet[4:16], []byte("ACN-E1.17\x00\x00\x00"))

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
	// rootPduLength := size - 16
	// packet[16] = 0x70 | byte(rootPduLength>>8)
	// packet[17] = byte(rootPduLength & 0xff)

	rootPduLength := size - 16
	putFlagsLength(packet, 16, rootPduLength)

	// E1.31 vector indicating that this is a data packet
	packet[21] = 0x04

	// CID
	// copy(packet[22:38], []byte("CheesyArena-LEDs"))

	copy(
		packet[22:38], []byte{
			0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0,
			0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
		},
	)

	// Framing PDU length and flags
	// framingPduLength := size - 38
	// packet[38] = 0x70 | byte(framingPduLength>>8)
	// packet[39] = byte(framingPduLength & 0xff)

	framingPduLength := size - 38
	putFlagsLength(packet, 38, framingPduLength)

	// E1.31 vector indicating that this is a data packet
	packet[43] = 0x02

	// Source name
	copy(packet[44:108], []byte(sACNSourceName))

	// Priority
	packet[108] = 100

	// Reserved
	packet[109] = 0x00
	packet[110] = 0x00

	// Sequence number
	packet[111] = 0x00

	// Options flags
	packet[112] = 0x00

	// DMX universe
	packet[113] = 0x00
	packet[114] = 0x00

	// DMP layer PDU length
	// dmpPduLength := size - 115
	// packet[115] = 0x70 | byte(dmpPduLength>>8)
	// packet[116] = byte(dmpPduLength & 0xff)

	dmpPduLength := size - 115
	putFlagsLength(packet, 115, dmpPduLength)

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
	count := 1 + numChannels + 3 // Extra 3 channels to avoid malformed packet errors in wireshark.
	packet[123] = byte(count >> 8)
	packet[124] = byte(count & 0xff)

	// DMX start code
	packet[125] = 0

	return packet
}
