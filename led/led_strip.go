// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents a LED strip attached to an E1.31 sACN (DMX over Ethernet) controller.

package led

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	controllerPort   = 5568
	sourceName       = "Cheesy Arena"
	packetTimeoutSec = 1
)

// LED sequence modes
type Mode int

const (
	OffMode Mode = iota
	RedMode
	GreenMode
	BlueMode
	WhiteMode
	ChaseMode
	WarmupMode
	Warmup2Mode
	Warmup3Mode
	Warmup4Mode
	OwnedMode
	ForceMode
	BoostMode
	RandomMode
	FadeMode
	GradientMode
	BlinkMode
)

var ModeNames = map[Mode]string{
	OffMode:      "Off",
	RedMode:      "Red",
	GreenMode:    "Green",
	BlueMode:     "Blue",
	WhiteMode:    "White",
	ChaseMode:    "Chase",
	WarmupMode:   "Warmup",
	Warmup2Mode:  "Warmup Purple",
	Warmup3Mode:  "Warmup Sneaky",
	Warmup4Mode:  "Warmup Gradient",
	OwnedMode:    "Owned",
	ForceMode:    "Force",
	BoostMode:    "Boost",
	RandomMode:   "Random",
	FadeMode:     "Fade",
	GradientMode: "Gradient",
	BlinkMode: "Blink",
}

// Color RGB mappings
type color int

const (
	red color = iota
	orange
	yellow
	green
	teal
	blue
	purple
	white
	black
	purpleRed
	purpleBlue
	dimRed
	dimBlue
)

var colors = map[color][3]byte{
	red:        {255, 0, 0},
	orange:     {255, 50, 0},
	yellow:     {255, 255, 0},
	green:      {0, 255, 0},
	teal:       {0, 100, 100},
	blue:       {0, 0, 255},
	purple:     {100, 0, 100},
	white:      {255, 255, 255},
	black:      {0, 0, 0},
	purpleRed:  {200, 0, 50},
	purpleBlue: {50, 0, 200},
	dimRed:     {100, 0, 0},
	dimBlue:    {0, 0, 100},
}

type LedStrip struct {
	CurrentMode    Mode
	conn           net.Conn
	pixels         [][3]byte
	oldPixels      [][3]byte
	packet         []byte
	counter        int
	lastPacketTime time.Time
}

func NewLedStrip(controllerAddress string, dmxUniverse int, numPixels int) (*LedStrip, error) {
	ledStrip := new(LedStrip)

	var err error
	ledStrip.conn, err = net.Dial("udp4", fmt.Sprintf("%s:%d", controllerAddress, controllerPort))
	if err != nil {
		return nil, err
	}

	ledStrip.pixels = make([][3]byte, numPixels)
	ledStrip.oldPixels = make([][3]byte, numPixels)
	ledStrip.packet = createBlankPacket(dmxUniverse, numPixels)

	return ledStrip, nil
}

// Sets the current LED sequence mode and resets the intra-sequence counter to the beginning.
func (strip *LedStrip) SetMode(mode Mode) {
	strip.CurrentMode = mode
	strip.counter = 0
}

// Advances the pixel values through the current sequence and sends a packet if necessary. Should be called from a timed
// loop.
func (strip *LedStrip) Update() error {
	// Determine the pixel values.
	switch strip.CurrentMode {
	case RedMode:
		strip.updateSingleColorMode(red)
	case GreenMode:
		strip.updateSingleColorMode(green)
	case BlueMode:
		strip.updateSingleColorMode(blue)
	case WhiteMode:
		strip.updateSingleColorMode(white)
	case ChaseMode:
		strip.updateChaseMode()
	case WarmupMode:
		strip.updateWarmupMode()
	case Warmup2Mode:
		strip.updateWarmup2Mode()
	case Warmup3Mode:
		strip.updateWarmup3Mode()
	case Warmup4Mode:
		strip.updateWarmup4Mode()
	case OwnedMode:
		strip.updateOwnedMode()
	case ForceMode:
		strip.updateForceMode()
	case BoostMode:
		strip.updateBoostMode()
	case RandomMode:
		strip.updateRandomMode()
	case FadeMode:
		strip.updateFadeMode()
	case GradientMode:
		strip.updateGradientMode()
	case BlinkMode:
		strip.updateBlinkMode()
	default:
		strip.updateOffMode()
	}
	strip.counter++

	// Update non-static packet fields.
	strip.packet[111]++
	for i, pixel := range strip.pixels {
		strip.packet[126+3*i] = pixel[0]
		strip.packet[127+3*i] = pixel[1]
		strip.packet[128+3*i] = pixel[2]
	}

	// Send the packet if it hasn't changed.
	if strip.shouldSendPacket() {
		_, err := strip.conn.Write(strip.packet)
		if err != nil {
			return err
		}
		strip.lastPacketTime = time.Now()

		// Keep a record of the pixel values in order to detect future changes.
		copy(strip.oldPixels, strip.pixels)
	}

	return nil
}

// Constructs the structure of an E1.31 data packet that can be re-used indefinitely by updating the pixel data and
// re-sending it.
func createBlankPacket(dmxUniverse, numPixels int) []byte {
	size := 126 + 3*numPixels
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

	// Sequence number (initial value)
	packet[111] = 0x00

	// Options flags
	packet[112] = 0x00

	// Universe
	packet[113] = byte(dmxUniverse >> 8)
	packet[114] = byte(dmxUniverse & 0xff)

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

	return packet
}

// Returns true if the pixel data has changed or it has been too long since the last packet was sent.
func (strip *LedStrip) shouldSendPacket() bool {
	for i := 0; i < len(strip.pixels); i++ {
		if strip.pixels[i] != strip.oldPixels[i] {
			return true
		}
	}
	return time.Since(strip.lastPacketTime).Seconds() > packetTimeoutSec
}

func (strip *LedStrip) updateOffMode() {
	for i := 0; i < len(strip.pixels); i++ {
		strip.pixels[i] = colors[black]
	}
}

func (strip *LedStrip) updateSingleColorMode(color color) {
	for i := 0; i < len(strip.pixels); i++ {
		strip.pixels[i] = colors[color]
	}
}

func (strip *LedStrip) updateChaseMode() {
	if strip.counter == len(colors)*len(strip.pixels) {
		strip.counter = 0
	}
	color := color(strip.counter / len(strip.pixels))
	pixelIndex := strip.counter % len(strip.pixels)
	strip.pixels[pixelIndex] = colors[color]
}

func (strip *LedStrip) updateWarmupMode() {
	endCounter := 250
	if strip.counter == 0 {
		// Show solid white to start.
		for i := 0; i < len(strip.pixels); i++ {
			strip.pixels[i] = colors[white]
		}
	} else if strip.counter <= endCounter {
		// Build to the alliance color from each side.
		numLitPixels := len(strip.pixels) / 2 * strip.counter / endCounter
		for i := 0; i < numLitPixels; i++ {
			strip.pixels[i] = colors[red]
			strip.pixels[len(strip.pixels)-i-1] = colors[red]
		}
	} else {
		// MaintainPrevent the counter from rolling over.
		strip.counter = endCounter
	}
}

func (strip *LedStrip) updateWarmup2Mode() {
	startCounter := 100
	endCounter := 250
	if strip.counter < startCounter {
		// Show solid purple to start.
		for i := 0; i < len(strip.pixels); i++ {
			strip.pixels[i] = colors[purple]
		}
	} else if strip.counter <= endCounter {
		for i := 0; i < len(strip.pixels); i++ {
			strip.pixels[i] = getFadeColor(purple, red, strip.counter-startCounter, endCounter-startCounter)
		}
	} else {
		// MaintainPrevent the counter from rolling over.
		strip.counter = endCounter
	}
}

func (strip *LedStrip) updateWarmup3Mode() {
	startCounter := 50
	middleCounter := 225
	endCounter := 250
	if strip.counter < startCounter {
		// Show solid purple to start.
		for i := 0; i < len(strip.pixels); i++ {
			strip.pixels[i] = colors[purple]
		}
	} else if strip.counter < middleCounter {
		for i := 0; i < len(strip.pixels); i++ {
			strip.pixels[i] = getFadeColor(purple, purpleBlue, strip.counter-startCounter, middleCounter-startCounter)
		}
	} else if strip.counter <= endCounter {
		for i := 0; i < len(strip.pixels); i++ {
			strip.pixels[i] = getFadeColor(purpleBlue, red, strip.counter-middleCounter, endCounter-middleCounter)
		}
	} else {
		// Maintain the current value and prevent the counter from rolling over.
		strip.counter = endCounter
	}
}

func (strip *LedStrip) updateWarmup4Mode() {
	startOffset := 50
	middleCounter := 100
	for i := 0; i < len(strip.pixels); i++ {
		strip.pixels[len(strip.pixels) - i - 1] = getGradientColor(i+strip.counter+startOffset, 75)
	}
	if strip.counter >= middleCounter {
		for i := 0; i < len(strip.pixels); i++ {
			if i < strip.counter - middleCounter {
				strip.pixels[i] = colors[red]
			}
		}
	}
}

func (strip *LedStrip) updateOwnedMode() {
	speedDivisor := 30
	pixelSpacing := 4
	if strip.counter%speedDivisor != 0 {
		return
	}
	for i := 0; i < len(strip.pixels); i++ {
		if i%pixelSpacing == strip.counter/speedDivisor%pixelSpacing {
			strip.pixels[i] = colors[red]
		} else {
			strip.pixels[i] = colors[black]
		}
	}
}

func (strip *LedStrip) updateForceMode() {
	speedDivisor := 30
	pixelSpacing := 7
	if strip.counter%speedDivisor != 0 {
		return
	}
	for i := 0; i < len(strip.pixels); i++ {
		switch (i + strip.counter/speedDivisor) % pixelSpacing {
		case 2:
			fallthrough
		case 4:
			strip.pixels[i] = colors[red]
		case 3:
			strip.pixels[i] = colors[dimBlue]
		default:
			strip.pixels[i] = colors[black]
		}
	}
}

func (strip *LedStrip) updateBoostMode() {
	speedDivisor := 4
	pixelSpacing := 4
	if strip.counter%speedDivisor != 0 {
		return
	}
	for i := 0; i < len(strip.pixels); i++ {
		if i%pixelSpacing == strip.counter/speedDivisor%pixelSpacing {
			strip.pixels[i] = colors[blue]
		} else {
			strip.pixels[i] = colors[black]
		}
	}
}

func (strip *LedStrip) updateRandomMode() {
	if strip.counter%10 != 0 {
		return
	}
	for i := 0; i < len(strip.pixels); i++ {
		strip.pixels[i] = colors[color(rand.Intn(int(black)))] // Ignore colors listed after white.
	}
}

func (strip *LedStrip) updateFadeMode() {
	fadeCycles := 40
	holdCycles := 10
	if strip.counter == 4*holdCycles+4*fadeCycles {
		strip.counter = 0
	}

	for i := 0; i < len(strip.pixels); i++ {
		if strip.counter < holdCycles {
			strip.pixels[i] = colors[black]
		} else if strip.counter < holdCycles+fadeCycles {
			strip.pixels[i] = getFadeColor(black, red, strip.counter-holdCycles, fadeCycles)
		} else if strip.counter < 2*holdCycles+fadeCycles {
			strip.pixels[i] = colors[red]
		} else if strip.counter < 2*holdCycles+2*fadeCycles {
			strip.pixels[i] = getFadeColor(red, black, strip.counter-2*holdCycles-fadeCycles, fadeCycles)
		} else if strip.counter < 3*holdCycles+2*fadeCycles {
			strip.pixels[i] = colors[black]
		} else if strip.counter < 3*holdCycles+3*fadeCycles {
			strip.pixels[i] = getFadeColor(black, blue, strip.counter-3*holdCycles-2*fadeCycles, fadeCycles)
		} else if strip.counter < 4*holdCycles+3*fadeCycles {
			strip.pixels[i] = colors[blue]
		} else if strip.counter < 4*holdCycles+4*fadeCycles {
			strip.pixels[i] = getFadeColor(blue, black, strip.counter-4*holdCycles-3*fadeCycles, fadeCycles)
		}
	}
}

func (strip *LedStrip) updateGradientMode() {
	for i := 0; i < len(strip.pixels); i++ {
		strip.pixels[len(strip.pixels) - i - 1] = getGradientColor(i+strip.counter, 75)
	}
}

func (strip *LedStrip) updateBlinkMode() {
	divisor := 10
	for i := 0; i < len(strip.pixels); i++ {
		if strip.counter % divisor < divisor / 2 {
			strip.pixels[i] = colors[white]
		} else {
			strip.pixels[i] = colors[black]
		}
	}
}

// Interpolates between the two colors based on the given fraction.
func getFadeColor(fromColor, toColor color, numerator, denominator int) [3]byte {
	from := colors[fromColor]
	to := colors[toColor]
	var fadeColor [3]byte
	for i := 0; i < 3; i++ {
		fadeColor[i] = byte(int(from[i]) + numerator*(int(to[i])-int(from[i]))/denominator)
	}
	return fadeColor
}

// Calculates the value of a single pixel in a gradient.
func getGradientColor(offset, numPixels int) [3]byte {
	offset %= numPixels
	if 3*offset < numPixels {
		return getFadeColor(red, green, 3*offset, numPixels)
	} else if 3*offset < 2*numPixels {
		return getFadeColor(green, blue, 3*offset-numPixels, numPixels)
	} else {
		return getFadeColor(blue, red, 3*offset-2*numPixels, numPixels)
	}
}
