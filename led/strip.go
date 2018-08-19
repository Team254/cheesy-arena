// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents an independently controlled LED strip making up one half of a scale or switch LED array.

package led

import (
	"math/rand"
	"time"
)

type strip struct {
	currentMode    Mode
	isRed          bool
	pixels         [numPixels][3]byte
	oldPixels      [numPixels][3]byte
	counter        int
	lastPacketTime time.Time
}

// Calculates the current pixel values depending on the mode and elapsed counter cycles.
func (strip *strip) updatePixels() {
	switch strip.currentMode {
	case RedMode:
		strip.updateSingleColorMode(Red)
	case GreenMode:
		strip.updateSingleColorMode(Green)
	case BlueMode:
		strip.updateSingleColorMode(Blue)
	case WhiteMode:
		strip.updateSingleColorMode(White)
	case PurpleMode:
		strip.updateSingleColorMode(Purple)
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
	case NotOwnedMode:
		strip.updateNotOwnedMode()
	case ForceMode:
		strip.updateForceMode()
	case BoostMode:
		strip.updateBoostMode()
	case RandomMode:
		strip.updateRandomMode()
	case FadeRedBlueMode:
		strip.updateFadeRedBlueMode()
	case FadeSingleMode:
		strip.updateFadeSingleMode()
	case GradientMode:
		strip.updateGradientMode()
	case BlinkMode:
		strip.updateBlinkMode()
	default:
		strip.updateOffMode()
	}
	strip.counter++
}

// Returns true if the pixel data has changed or it has been too long since the last packet was sent.
func (strip *strip) shouldSendPacket() bool {
	for i := 0; i < numPixels; i++ {
		if strip.pixels[i] != strip.oldPixels[i] {
			return true
		}
	}
	return time.Since(strip.lastPacketTime).Seconds() > packetTimeoutSec
}

// Writes the pixel RGB values into the given packet in preparation for sending.
func (strip *strip) populatePacketPixels(pixelData []byte) {
	for i, pixel := range strip.pixels {
		pixelData[3*i] = pixel[0]
		pixelData[3*i+1] = pixel[1]
		pixelData[3*i+2] = pixel[2]
	}

	// Keep a record of the pixel values in order to detect future changes.
	strip.oldPixels = strip.pixels
	strip.lastPacketTime = time.Now()
}

// Returns the primary color (red or blue) of this strip.
func (strip *strip) getColor() Color {
	if strip.isRed {
		return Red
	}
	return Blue
}

// Returns the opposite primary color (red or blue) to this strip.
func (strip *strip) getOppositeColor() Color {
	if strip.isRed {
		return Blue
	}
	return Red
}

// Returns a color partway between purple and the primary color (red or blue) of this strip.
func (strip *strip) getMidColor() Color {
	if strip.isRed {
		return PurpleBlue
	}
	return PurpleRed
}

// Returns a dim version of the primary color (red or blue) of this strip.
func (strip *strip) getDimColor() Color {
	if strip.isRed {
		return DimRed
	}
	return DimBlue
}

// Returns a dim version of the opposite primary color (red or blue) of this strip.
func (strip *strip) getDimOppositeColor() Color {
	if strip.isRed {
		return DimBlue
	}
	return DimRed
}

// Returns the starting offset for the gradient mode for this strip.
func (strip *strip) getGradientStartOffset() int {
	if strip.isRed {
		return numPixels / 3
	}
	return 2 * numPixels / 3
}

func (strip *strip) updateOffMode() {
	for i := 0; i < numPixels; i++ {
		strip.pixels[i] = Colors[Black]
	}
}

func (strip *strip) updateSingleColorMode(color Color) {
	for i := 0; i < numPixels; i++ {
		strip.pixels[i] = Colors[color]
	}
}

func (strip *strip) updateChaseMode() {
	if strip.counter == int(Black)*numPixels { // Ignore colors listed after white.
		strip.counter = 0
	}
	color := Color(strip.counter / numPixels)
	pixelIndex := strip.counter % numPixels
	strip.pixels[pixelIndex] = Colors[color]
}

func (strip *strip) updateWarmupMode() {
	endCounter := 250
	if strip.counter == 0 {
		// Show solid white to start.
		for i := 0; i < numPixels; i++ {
			strip.pixels[i] = Colors[White]
		}
	} else if strip.counter <= endCounter {
		// Build to the alliance color from each side.
		numLitPixels := numPixels / 2 * strip.counter / endCounter
		for i := 0; i < numLitPixels; i++ {
			strip.pixels[i] = Colors[strip.getColor()]
			strip.pixels[numPixels-i-1] = Colors[strip.getColor()]
		}
	} else {
		// Prevent the counter from rolling over.
		strip.counter = endCounter
	}
}

func (strip *strip) updateWarmup2Mode() {
	startCounter := 100
	endCounter := 250
	if strip.counter < startCounter {
		// Show solid purple to start.
		for i := 0; i < numPixels; i++ {
			strip.pixels[i] = Colors[Purple]
		}
	} else if strip.counter <= endCounter {
		for i := 0; i < numPixels; i++ {
			strip.pixels[i] = getFadeColor(Purple, strip.getColor(), strip.counter-startCounter,
				endCounter-startCounter)
		}
	} else {
		// Prevent the counter from rolling over.
		strip.counter = endCounter
	}
}

func (strip *strip) updateWarmup3Mode() {
	startCounter := 50
	middleCounter := 225
	endCounter := 250
	if strip.counter < startCounter {
		// Show solid purple to start.
		for i := 0; i < numPixels; i++ {
			strip.pixels[i] = Colors[Purple]
		}
	} else if strip.counter < middleCounter {
		for i := 0; i < numPixels; i++ {
			strip.pixels[i] = getFadeColor(Purple, strip.getMidColor(), strip.counter-startCounter,
				middleCounter-startCounter)
		}
	} else if strip.counter <= endCounter {
		for i := 0; i < numPixels; i++ {
			strip.pixels[i] = getFadeColor(strip.getMidColor(), strip.getColor(), strip.counter-middleCounter,
				endCounter-middleCounter)
		}
	} else {
		// Maintain the current value and prevent the counter from rolling over.
		strip.counter = endCounter
	}
}

func (strip *strip) updateWarmup4Mode() {
	middleCounter := 100
	for i := 0; i < numPixels; i++ {
		strip.pixels[numPixels-i-1] = getGradientColor(i+strip.counter+strip.getGradientStartOffset(), numPixels/2)
	}
	if strip.counter >= middleCounter {
		for i := 0; i < numPixels; i++ {
			if i < strip.counter-middleCounter {
				strip.pixels[i] = Colors[strip.getColor()]
			}
		}
	}
}

func (strip *strip) updateOwnedMode() {
	speedDivisor := 30
	pixelSpacing := 4
	if strip.counter%speedDivisor != 0 {
		return
	}
	for i := 0; i < numPixels; i++ {
		switch (i + strip.counter/speedDivisor) % pixelSpacing {
		case 0:
			fallthrough
		case 1:
			strip.pixels[len(strip.pixels)-i-1] = Colors[strip.getColor()]
		default:
			strip.pixels[len(strip.pixels)-i-1] = Colors[Black]
		}
	}
}

func (strip *strip) updateNotOwnedMode() {
	for i := 0; i < numPixels; i++ {
		strip.pixels[i] = Colors[strip.getDimColor()]
	}
}

func (strip *strip) updateForceMode() {
	speedDivisor := 30
	pixelSpacing := 7
	if strip.counter%speedDivisor != 0 {
		return
	}
	for i := 0; i < numPixels; i++ {
		switch (i + strip.counter/speedDivisor) % pixelSpacing {
		case 2:
			fallthrough
		case 4:
			strip.pixels[i] = Colors[strip.getColor()]
		case 3:
			strip.pixels[i] = Colors[strip.getDimOppositeColor()]
		default:
			strip.pixels[i] = Colors[Black]
		}
	}
}

func (strip *strip) updateBoostMode() {
	speedDivisor := 4
	pixelSpacing := 4
	if strip.counter%speedDivisor != 0 {
		return
	}
	for i := 0; i < numPixels; i++ {
		if i%pixelSpacing == strip.counter/speedDivisor%pixelSpacing {
			strip.pixels[i] = Colors[strip.getColor()]
		} else {
			strip.pixels[i] = Colors[Black]
		}
	}
}

func (strip *strip) updateRandomMode() {
	if strip.counter%10 != 0 {
		return
	}
	for i := 0; i < numPixels; i++ {
		strip.pixels[i] = Colors[Color(rand.Intn(int(Black)))] // Ignore colors listed after white.
	}
}

func (strip *strip) updateFadeRedBlueMode() {
	fadeCycles := 40
	holdCycles := 10
	if strip.counter == 4*holdCycles+4*fadeCycles {
		strip.counter = 0
	}

	for i := 0; i < numPixels; i++ {
		if strip.counter < holdCycles {
			strip.pixels[i] = Colors[Black]
		} else if strip.counter < holdCycles+fadeCycles {
			strip.pixels[i] = getFadeColor(Black, Red, strip.counter-holdCycles, fadeCycles)
		} else if strip.counter < 2*holdCycles+fadeCycles {
			strip.pixels[i] = Colors[Red]
		} else if strip.counter < 2*holdCycles+2*fadeCycles {
			strip.pixels[i] = getFadeColor(Red, Black, strip.counter-2*holdCycles-fadeCycles, fadeCycles)
		} else if strip.counter < 3*holdCycles+2*fadeCycles {
			strip.pixels[i] = Colors[Black]
		} else if strip.counter < 3*holdCycles+3*fadeCycles {
			strip.pixels[i] = getFadeColor(Black, Blue, strip.counter-3*holdCycles-2*fadeCycles, fadeCycles)
		} else if strip.counter < 4*holdCycles+3*fadeCycles {
			strip.pixels[i] = Colors[Blue]
		} else if strip.counter < 4*holdCycles+4*fadeCycles {
			strip.pixels[i] = getFadeColor(Blue, Black, strip.counter-4*holdCycles-3*fadeCycles, fadeCycles)
		}
	}
}

func (strip *strip) updateFadeSingleMode() {
	offCycles := 50
	fadeCycles := 100
	if strip.counter == offCycles+2*fadeCycles {
		strip.counter = 0
	}

	for i := 0; i < numPixels; i++ {
		if strip.counter < offCycles {
			strip.pixels[i] = Colors[Black]
		} else if strip.counter < offCycles+fadeCycles {
			strip.pixels[i] = getFadeColor(Black, strip.getColor(), strip.counter-offCycles, fadeCycles)
		} else if strip.counter < offCycles+2*fadeCycles {
			strip.pixels[i] = getFadeColor(strip.getColor(), Black, strip.counter-offCycles-fadeCycles, fadeCycles)
		}
	}
}

func (strip *strip) updateGradientMode() {
	for i := 0; i < numPixels; i++ {
		strip.pixels[numPixels-i-1] = getGradientColor(i+strip.counter, 75)
	}
}

func (strip *strip) updateBlinkMode() {
	divisor := 10
	for i := 0; i < numPixels; i++ {
		if strip.counter%divisor < divisor/2 {
			strip.pixels[i] = Colors[White]
		} else {
			strip.pixels[i] = Colors[Black]
		}
	}
}

// Interpolates between the two colors based on the given fraction.
func getFadeColor(fromColor, toColor Color, numerator, denominator int) [3]byte {
	from := Colors[fromColor]
	to := Colors[toColor]
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
		return getFadeColor(Red, Green, 3*offset, numPixels)
	} else if 3*offset < 2*numPixels {
		return getFadeColor(Green, Blue, 3*offset-numPixels, numPixels)
	} else {
		return getFadeColor(Blue, Red, 3*offset-2*numPixels, numPixels)
	}
}
