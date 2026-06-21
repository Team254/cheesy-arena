// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents one alliance's independently controlled Hub LED zone.

package led

import "time"

const (
	numSides           = 4
	fixturesPerSide    = 2
	pixelsPerFixture   = 8
	numPixels          = numSides * fixturesPerSide * pixelsPerFixture
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

type zone struct {
	currentMode Mode
	pixels      [numPixels]Color
	counter     int
}

// updatePixels calculates the current pixel values depending on the mode and elapsed counter cycles.
func (zone *zone) updatePixels() {
	switch zone.currentMode {
	case RedMode, BlueMode, GreenMode, PurpleMode, WhiteMode, OffMode:
		zone.updateSingleColorMode(colorForMode(zone.currentMode))
	case RedPulseMode:
		zone.updatePulseMode(Red, zone.counter)
	case BluePulseMode:
		zone.updatePulseMode(Blue, zone.counter)
	case RedStartupMode:
		zone.updateStartupMode(Red, zone.counter)
	case BlueStartupMode:
		zone.updateStartupMode(Blue, zone.counter)
	case RedAdvantageMode:
		zone.updateAdvantageMode(Red, zone.counter)
	case BlueAdvantageMode:
		zone.updateAdvantageMode(Blue, zone.counter)
	case RainbowMode:
		zone.updateRainbowMode(zone.counter)
	default:
		zone.updateSingleColorMode(Black)
	}
	zone.counter++
}

// updateSingleColorMode renders a solid color across the zone.
func (zone *zone) updateSingleColorMode(color Color) {
	for i := range zone.pixels {
		zone.pixels[i] = color
	}
}

// updatePulseMode renders the alliance warning pulse sequence.
func (zone *zone) updatePulseMode(color Color, counter int) {
	phase := counter % (2 * pulseHalfPeriod)
	if phase > pulseHalfPeriod {
		phase = 2*pulseHalfPeriod - phase
	}
	zone.updateSingleColorMode(color.Scale(float64(phase) / float64(pulseHalfPeriod)))
}

// updateStartupMode renders the match start fill sequence.
func (zone *zone) updateStartupMode(color Color, counter int) {
	zone.updateSingleColorMode(Black)
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
		zone.fillSide(side, color, counter-delay, direction)
	}
}

// fillSide renders one side's startup fill sequence.
func (zone *zone) fillSide(side int, color Color, counter int, direction fillDirection) {
	if counter <= 0 {
		return
	}
	fillCycles := startupCycles / 3
	percentage := float64(counter) / float64(fillCycles)
	if percentage > 1 {
		percentage = 1
	}
	for fixture := 0; fixture < fixturesPerSide; fixture++ {
		start := ((side-1)*fixturesPerSide + fixture) * pixelsPerFixture
		zone.fillFixture(start, color, percentage, direction)
	}
}

// fillFixture renders one fixture's startup fill sequence.
func (zone *zone) fillFixture(start int, color Color, percentage float64, direction fillDirection) {
	nodesToFill := percentage * pixelsPerFixture
	order := fillOrder(direction)
	for rank := 0; rank < pixelsPerFixture; rank++ {
		brightness := nodesToFill - float64(rank)
		if brightness > 1 {
			brightness = 1
		}
		if brightness < 0 {
			brightness = 0
		}

		zone.pixels[start+order[rank]] = color.Scale(brightness)
	}
}

// fillOrder returns the fixture fill order for a startup direction.
func fillOrder(direction fillDirection) [pixelsPerFixture]int {
	switch direction {
	case fillLeftToRight:
		return [pixelsPerFixture]int{0, 1, 2, 3, 4, 5, 6, 7}
	case fillRightToLeft:
		return [pixelsPerFixture]int{7, 6, 5, 4, 3, 2, 1, 0}
	case fillEdgesIn:
		return [pixelsPerFixture]int{0, 7, 1, 6, 2, 5, 3, 4}
	default:
		return [pixelsPerFixture]int{3, 4, 2, 5, 1, 6, 0, 7}
	}
}

// updateAdvantageMode renders the transition period sweep sequence.
func (zone *zone) updateAdvantageMode(baseColor Color, counter int) {
	zone.updateSingleColorMode(baseColor)
	for side := 1; side <= numSides; side++ {
		direction := fillLeftToRight
		if side == 2 || side == 4 {
			direction = fillRightToLeft
		}
		for fixture := 0; fixture < fixturesPerSide; fixture++ {
			start := ((side-1)*fixturesPerSide + fixture) * pixelsPerFixture
			zone.sweepFixture(start, counter, direction)
		}
	}
}

// sweepFixture renders one fixture's white sweep over the alliance base color.
func (zone *zone) sweepFixture(start, counter int, direction fillDirection) {
	cycleLength := pixelsPerFixture*2 + 2
	position := (counter / advantageStepCycle) % cycleLength
	if direction == fillRightToLeft {
		position = cycleLength - 1 - position
	}
	head := position - 1
	for trail := 0; trail < pixelsPerFixture; trail++ {
		pixel := head - trail
		if direction == fillRightToLeft {
			pixel = head + trail - pixelsPerFixture
		}
		if pixel < 0 || pixel >= pixelsPerFixture {
			continue
		}
		brightness := 1 - float64(trail)/float64(pixelsPerFixture)
		zone.pixels[start+pixel] = White.Scale(brightness)
	}
}

// updateRainbowMode renders a rotating rainbow pattern counter-clockwise.
func (zone *zone) updateRainbowMode(counter int) {
	logicalPixels := numSides * pixelsPerFixture

	for i := 0; i < logicalPixels; i++ {
		pos := (i - (counter / 2)) % logicalPixels
		if pos < 0 {
			pos += logicalPixels
		}

		h := float64(pos) / float64(logicalPixels) * 6
		idx := int(h)
		f := h - float64(idx)

		q := byte(255 * (1 - f))
		t := byte(255 * f)

		var color Color
		switch idx % 6 {
		case 0:
			color = Color{255, t, 0}
		case 1:
			color = Color{q, 255, 0}
		case 2:
			color = Color{0, 255, t}
		case 3:
			color = Color{0, q, 255}
		case 4:
			color = Color{t, 0, 255}
		case 5:
			color = Color{255, 0, q}
		}

		side := i/pixelsPerFixture + 1
		pixel := i % pixelsPerFixture

		for fixture := 0; fixture < fixturesPerSide; fixture++ {
			start := ((side-1)*fixturesPerSide + fixture) * pixelsPerFixture
			zone.pixels[start+pixel] = color
		}
	}
}
