// Copyright 2025 Team 254. All Rights Reserved.
// Author: kwaremburg
//
// DMX controller interface for controlling RGB light bars in the hubs.

package led

import (
	"fmt"
	"log"
	"math"
	"net"
	"time"
)

const (
	sACNPort          = 5568
	sACNSourceName    = "Cheesy Arena"
	sourceName        = "Cheesy Arena"
	pixelDataOffset   = 126
	heartbeatInterval = 1 * time.Second
	NumSegments       = 16
)

// Color represents an RGB color value.
type Color struct {
	R, G, B uint8
}

func (c Color) Equals(other Color) bool {
	return c.R == other.R && c.G == other.G && c.B == other.B
}

func (c Color) Scale(multiplier float64) Color {
	return Color{
		R: uint8(float64(c.R) * multiplier),
		G: uint8(float64(c.G) * multiplier),
		B: uint8(float64(c.B) * multiplier),
	}
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
	colors       [NumSegments]Color
	lastColors   [NumSegments]Color
	lastSend     time.Time
	StartChannel int // Starting channel for the hub
	packet       []byte
	chaseIndex   int // Current segment index for the 100% part of the chase
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
	for i := 0; i < NumSegments; i++ {
		dmx.colors[i] = color
	}
}

func (dmx *Controller) SetChase(color Color, matchTimeSec float64) {
	// 1 second to go through all 16 segments.
	// matchTimeSec % 1.0 gives time within the current second [0.0, 1.0)
	timeInCycle := math.Mod(matchTimeSec, 1.0)
	dmx.chaseIndex = int(timeInCycle * NumSegments)

	for i := 0; i < NumSegments; i++ {
		dmx.colors[i] = ColorOff
	}

	// 33, 66, 100, 100, 100, 66, 33 chase pattern
	dmx.colors[(dmx.chaseIndex-3+NumSegments)%NumSegments] = color.Scale(0.33)
	dmx.colors[(dmx.chaseIndex-2+NumSegments)%NumSegments] = color.Scale(0.66)
	dmx.colors[(dmx.chaseIndex-1+NumSegments)%NumSegments] = color
	dmx.colors[dmx.chaseIndex] = color
	dmx.colors[(dmx.chaseIndex+1)%NumSegments] = color
	dmx.colors[(dmx.chaseIndex+2)%NumSegments] = color.Scale(0.66)
	dmx.colors[(dmx.chaseIndex+3)%NumSegments] = color.Scale(0.33)
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
	if dmx.conn == nil {
		// This controller is not configured; do nothing.
		return
	}

	// Create the template packet if it doesn't already exist.
	if len(dmx.packet) == 0 {
		dmx.packet = createBlankPacket(NumSegments * 3)
	}

	// Send packets if the pixel values have changed.
	if dmx.shouldSendPacket() {
		dmx.populatePacket(dmx.StartChannel)
		if err := dmx.sendPacket(dmx.Universe); err != nil {
			log.Printf("sACN error writing data to universe %d: %v", dmx.Universe, err)
			return
		}
		dmx.lastColors = dmx.colors
		dmx.lastSend = time.Now()
	}
}

func (dmx *Controller) shouldSendPacket() bool {
	for i := 0; i < NumSegments; i++ {
		if !dmx.colors[i].Equals(dmx.lastColors[i]) {
			return true
		}
	}
	return time.Since(dmx.lastSend) >= heartbeatInterval
}

func (dmx *Controller) populatePacket(startChannel int) {
	// Clear DMX data area
	for i := pixelDataOffset; i < len(dmx.packet); i++ {
		dmx.packet[i] = 0
	}

	// Light mapping (48 channels):
	// 1-3: Segment 1 (R, G, B)
	// 4-6: Segment 2 (R, G, B)
	// ...
	for i := 0; i < NumSegments; i++ {
		dmx.packet[pixelDataOffset+startChannel-1+i*3+0] = dmx.colors[i].R
		dmx.packet[pixelDataOffset+startChannel-1+i*3+1] = dmx.colors[i].G
		dmx.packet[pixelDataOffset+startChannel-1+i*3+2] = dmx.colors[i].B
	}
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
