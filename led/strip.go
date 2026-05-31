// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents one alliance's independently controlled Hub LED universe.

package led

import "time"

const (
	numSides           = 4
	fixturesPerSide    = 2
	nodesPerFixture    = 8
	numPixels          = numSides * fixturesPerSide * nodesPerFixture
	startupCycles      = 100
	advantageStepCycle = 4
	heartbeatInterval  = time.Second
	pulseHalfPeriod    = 70
	startupSide1Delay  = 0
	startupSide24Delay = 33
	startupSide3Delay  = 66
)

type fillDirection int

const (
	fillCenterOut fillDirection = iota
	fillLeftToRight
	fillRightToLeft
	fillEdgesIn
)

type strip struct {
	currentMode    Mode
	pixels         [numPixels]Color
	oldPixels      [numPixels]Color
	counter        int
	lastPacketTime time.Time
}

// updatePixels calculates the current pixel values depending on the mode and elapsed counter cycles.
func (strip *strip) updatePixels() {
	switch strip.currentMode {
	case RedMode, BlueMode, GreenMode, PurpleMode, WhiteMode, OffMode:
		strip.updateSingleColorMode(colorForMode(strip.currentMode))
	case RedPulseMode:
		strip.updatePulseMode(Red, strip.counter)
	case BluePulseMode:
		strip.updatePulseMode(Blue, strip.counter)
	case RedStartupMode:
		strip.updateStartupMode(Red, strip.counter)
	case BlueStartupMode:
		strip.updateStartupMode(Blue, strip.counter)
	case RedAdvantageMode:
		strip.updateAdvantageMode(Red, strip.counter)
	case BlueAdvantageMode:
		strip.updateAdvantageMode(Blue, strip.counter)
	default:
		strip.updateSingleColorMode(Black)
	}
	strip.counter++
}

// shouldSendPacket returns true if the pixel data has changed or it has been too long since the last packet was sent.
func (strip *strip) shouldSendPacket() bool {
	if strip.lastPacketTime.IsZero() || time.Since(strip.lastPacketTime) >= heartbeatInterval {
		return true
	}
	return strip.pixels != strip.oldPixels
}

// populatePacketPixels writes the pixel RGB values into the given packet in preparation for sending.
func (strip *strip) populatePacketPixels(pixelData []byte) {
	for i, pixel := range strip.pixels {
		pixelData[3*i] = pixel.R
		pixelData[3*i+1] = pixel.G
		pixelData[3*i+2] = pixel.B
	}

	// Keep a record of the pixel values in order to detect future changes.
	strip.oldPixels = strip.pixels
	strip.lastPacketTime = time.Now()
}

// updateSingleColorMode renders a solid color across the strip.
func (strip *strip) updateSingleColorMode(color Color) {
	for i := range strip.pixels {
		strip.pixels[i] = color
	}
}

// updatePulseMode renders the alliance warning pulse sequence.
func (strip *strip) updatePulseMode(color Color, counter int) {
	phase := counter % (2 * pulseHalfPeriod)
	if phase > pulseHalfPeriod {
		phase = 2*pulseHalfPeriod - phase
	}
	strip.updateSingleColorMode(color.Scale(float64(phase) / float64(pulseHalfPeriod)))
}

// updateStartupMode renders the match start fill sequence.
func (strip *strip) updateStartupMode(color Color, counter int) {
	strip.updateSingleColorMode(Black)
	for side := 1; side <= numSides; side++ {
		var delay int
		var direction fillDirection
		switch side {
		case 1:
			delay = startupSide1Delay
			direction = fillCenterOut
		case 2:
			delay = startupSide24Delay
			direction = fillLeftToRight
		case 3:
			delay = startupSide3Delay
			direction = fillEdgesIn
		case 4:
			delay = startupSide24Delay
			direction = fillRightToLeft
		}
		strip.fillSide(side, color, counter-delay, direction)
	}
}

// fillSide renders one side's startup fill sequence.
func (strip *strip) fillSide(side int, color Color, counter int, direction fillDirection) {
	if counter <= 0 {
		return
	}
	fillCycles := startupCycles / 3
	percentage := float64(counter) / float64(fillCycles)
	if percentage > 1 {
		percentage = 1
	}
	for fixture := 0; fixture < fixturesPerSide; fixture++ {
		start := ((side-1)*fixturesPerSide + fixture) * nodesPerFixture
		strip.fillFixture(start, color, percentage, direction)
	}
}

// fillFixture renders one fixture's startup fill sequence.
func (strip *strip) fillFixture(start int, color Color, percentage float64, direction fillDirection) {
	nodesToFill := percentage * nodesPerFixture
	order := fillOrder(direction)
	for rank := 0; rank < nodesPerFixture; rank++ {
		brightness := nodesToFill - float64(rank)
		if brightness > 1 {
			brightness = 1
		}
		if brightness < 0 {
			brightness = 0
		}

		strip.pixels[start+order[rank]] = color.Scale(brightness)
	}
}

// fillOrder returns the fixture fill order for a startup direction.
func fillOrder(direction fillDirection) [nodesPerFixture]int {
	switch direction {
	case fillLeftToRight:
		return [nodesPerFixture]int{0, 1, 2, 3, 4, 5, 6, 7}
	case fillRightToLeft:
		return [nodesPerFixture]int{7, 6, 5, 4, 3, 2, 1, 0}
	case fillEdgesIn:
		return [nodesPerFixture]int{0, 7, 1, 6, 2, 5, 3, 4}
	default:
		return [nodesPerFixture]int{3, 4, 2, 5, 1, 6, 0, 7}
	}
}

// updateAdvantageMode renders the transition period sweep sequence.
func (strip *strip) updateAdvantageMode(baseColor Color, counter int) {
	strip.updateSingleColorMode(baseColor)
	for side := 1; side <= numSides; side++ {
		direction := fillLeftToRight
		if side == 2 || side == 4 {
			direction = fillRightToLeft
		}
		for fixture := 0; fixture < fixturesPerSide; fixture++ {
			start := ((side-1)*fixturesPerSide + fixture) * nodesPerFixture
			strip.sweepFixture(start, counter, direction)
		}
	}
}

// sweepFixture renders one fixture's white sweep over the alliance base color.
func (strip *strip) sweepFixture(start, counter int, direction fillDirection) {
	cycleLength := nodesPerFixture*2 + 2
	position := (counter / advantageStepCycle) % cycleLength
	if direction == fillRightToLeft {
		position = cycleLength - 1 - position
	}
	head := position - 1
	for trail := 0; trail < nodesPerFixture; trail++ {
		pixel := head - trail
		if direction == fillRightToLeft {
			pixel = head + trail - nodesPerFixture
		}
		if pixel < 0 || pixel >= nodesPerFixture {
			continue
		}
		brightness := 1 - float64(trail)/float64(nodesPerFixture)
		strip.pixels[start+pixel] = White.Scale(brightness)
	}
}
