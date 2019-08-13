// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Represents an independently controlled LED strip.

package led

import (
	"math/rand"
	"time"
)

type advatekStrip struct {
	currentMode    StripMode
	isRed          bool
	pixels         [advatekNumPixels][3]byte
	oldPixels      [advatekNumPixels][3]byte
	counter        int
	lastPacketTime time.Time
}

// Calculates the current pixel values depending on the mode and elapsed counter cycles.
func (strip *advatekStrip) updatePixels() {
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
func (strip *advatekStrip) shouldSendPacket() bool {
	for i := 0; i < advatekNumPixels; i++ {
		if strip.pixels[i] != strip.oldPixels[i] {
			return true
		}
	}
	return time.Since(strip.lastPacketTime).Seconds() > advatekPacketTimeoutSec
}

// Writes the pixel RGB values into the given packet in preparation for sending.
func (strip *advatekStrip) populatePacketPixels(pixelData []byte) {
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
func (strip *advatekStrip) getColor() Color {
	if strip.isRed {
		return Red
	}
	return Blue
}

// Returns the opposite primary color (red or blue) to this strip.
func (strip *advatekStrip) getOppositeColor() Color {
	if strip.isRed {
		return Blue
	}
	return Red
}

// Returns a color partway between purple and the primary color (red or blue) of this strip.
func (strip *advatekStrip) getMidColor() Color {
	if strip.isRed {
		return PurpleBlue
	}
	return PurpleRed
}

// Returns a dim version of the primary color (red or blue) of this strip.
func (strip *advatekStrip) getDimColor() Color {
	if strip.isRed {
		return DimRed
	}
	return DimBlue
}

// Returns a dim version of the opposite primary color (red or blue) of this strip.
func (strip *advatekStrip) getDimOppositeColor() Color {
	if strip.isRed {
		return DimBlue
	}
	return DimRed
}

// Returns the starting offset for the gradient mode for this strip.
func (strip *advatekStrip) getGradientStartOffset() int {
	if strip.isRed {
		return advatekNumPixels / 3
	}
	return 2 * advatekNumPixels / 3
}

func (strip *advatekStrip) updateOffMode() {
	for i := 0; i < advatekNumPixels; i++ {
		strip.pixels[i] = Colors[Black]
	}
}

func (strip *advatekStrip) updateSingleColorMode(color Color) {
	for i := 0; i < advatekNumPixels; i++ {
		strip.pixels[i] = Colors[color]
	}
}

func (strip *advatekStrip) updateChaseMode() {
	if strip.counter == int(Black)*advatekNumPixels { // Ignore colors listed after white.
		strip.counter = 0
	}
	color := Color(strip.counter / advatekNumPixels)
	pixelIndex := strip.counter % advatekNumPixels
	strip.pixels[pixelIndex] = Colors[color]
}

func (strip *advatekStrip) updateRandomMode() {
	if strip.counter%10 != 0 {
		return
	}
	for i := 0; i < advatekNumPixels; i++ {
		strip.pixels[i] = Colors[Color(rand.Intn(int(Black)))] // Ignore colors listed after white.
	}
}

func (strip *advatekStrip) updateFadeRedBlueMode() {
	fadeCycles := 40
	holdCycles := 10
	if strip.counter == 4*holdCycles+4*fadeCycles {
		strip.counter = 0
	}

	for i := 0; i < advatekNumPixels; i++ {
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

func (strip *advatekStrip) updateFadeSingleMode() {
	offCycles := 50
	fadeCycles := 100
	if strip.counter == offCycles+2*fadeCycles {
		strip.counter = 0
	}

	for i := 0; i < advatekNumPixels; i++ {
		if strip.counter < offCycles {
			strip.pixels[i] = Colors[Black]
		} else if strip.counter < offCycles+fadeCycles {
			strip.pixels[i] = getFadeColor(Black, strip.getColor(), strip.counter-offCycles, fadeCycles)
		} else if strip.counter < offCycles+2*fadeCycles {
			strip.pixels[i] = getFadeColor(strip.getColor(), Black, strip.counter-offCycles-fadeCycles, fadeCycles)
		}
	}
}

func (strip *advatekStrip) updateGradientMode() {
	for i := 0; i < advatekNumPixels; i++ {
		strip.pixels[advatekNumPixels-i-1] = getGradientColor(i+strip.counter, 75)
	}
}

func (strip *advatekStrip) updateBlinkMode() {
	divisor := 10
	for i := 0; i < advatekNumPixels; i++ {
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
