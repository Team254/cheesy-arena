// Copyright 2025 Team 254. All Rights Reserved.
// Author: kwaremburg
//
// DMX controller interface for controlling RGB light bars in the hubs.

package dmx

import (
	"log"
	"sync"
	"time"

	"github.com/tarm/serial"
)

// Color represents an RGB color value.
type Color struct {
	R, G, B uint8
}

// Predefined colors for different states
var (
	ColorOff    = Color{0, 0, 0}     // Off (inactive hub)
	ColorGreen  = Color{0, 255, 0}   // Green (field safe)
	ColorPurple = Color{128, 0, 128} // Purple (counting after match)
	ColorRed    = Color{255, 0, 0}   // Red (red alliance active hub)
	ColorBlue   = Color{0, 0, 255}   // Blue (blue alliance active hub)
)

// Controller interface for DMX light control.
type Controller interface {
	SetAddress(port string)
	IsEnabled() bool
	IsHealthy() bool
	Run()
	SetHubColors(redColor, blueColor Color)
	Close()
}

// OpenDMXController implements the Controller interface for OpenDMX USB dongles.
type OpenDMXController struct {
	port        string
	serialPort  *serial.Port
	isHealthy   bool
	dmxData     [512]byte // DMX universe (512 channels)
	redChannel  int       // Starting channel for red hub (RGB = 3 channels)
	blueChannel int       // Starting channel for blue hub (RGB = 3 channels)
	mutex       sync.Mutex
	stopChan    chan bool
}

// NewOpenDMXController creates a new OpenDMX controller.
func NewOpenDMXController(redChannel, blueChannel int) *OpenDMXController {
	return &OpenDMXController{
		redChannel:  redChannel,
		blueChannel: blueChannel,
		stopChan:    make(chan bool),
	}
}

func (dmx *OpenDMXController) SetAddress(port string) {
	dmx.mutex.Lock()
	defer dmx.mutex.Unlock()

	dmx.port = port
	dmx.resetConnection()
}

func (dmx *OpenDMXController) IsEnabled() bool {
	return dmx.port != ""
}

func (dmx *OpenDMXController) IsHealthy() bool {
	dmx.mutex.Lock()
	defer dmx.mutex.Unlock()
	return dmx.isHealthy
}

func (dmx *OpenDMXController) Run() {
	if !dmx.IsEnabled() {
		return
	}

	// Try to connect initially
	dmx.connect()

	// Main update loop - send DMX data at ~44Hz (standard DMX refresh rate)
	ticker := time.NewTicker(23 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-dmx.stopChan:
			dmx.Close()
			return
		case <-ticker.C:
			if !dmx.IsHealthy() {
				dmx.connect()
			} else {
				dmx.update()
			}
		}
	}
}

func (dmx *OpenDMXController) SetHubColors(redColor, blueColor Color) {
	dmx.mutex.Lock()
	defer dmx.mutex.Unlock()

	// Set red hub RGB channels
	dmx.dmxData[dmx.redChannel] = redColor.R
	dmx.dmxData[dmx.redChannel+1] = redColor.G
	dmx.dmxData[dmx.redChannel+2] = redColor.B

	// Set blue hub RGB channels
	dmx.dmxData[dmx.blueChannel] = blueColor.R
	dmx.dmxData[dmx.blueChannel+1] = blueColor.G
	dmx.dmxData[dmx.blueChannel+2] = blueColor.B
}

func (dmx *OpenDMXController) Close() {
	dmx.mutex.Lock()
	defer dmx.mutex.Unlock()

	if dmx.serialPort != nil {
		dmx.serialPort.Close()
		dmx.serialPort = nil
	}
	dmx.isHealthy = false
}

func (dmx *OpenDMXController) connect() error {
	dmx.mutex.Lock()
	defer dmx.mutex.Unlock()

	if dmx.serialPort != nil {
		dmx.serialPort.Close()
	}

	config := &serial.Config{
		Name: dmx.port,
		Baud: 250000, // DMX512 baud rate
	}

	port, err := serial.OpenPort(config)
	if err != nil {
		log.Printf("DMX error connecting to %s: %v", dmx.port, err)
		dmx.isHealthy = false
		return err
	}

	dmx.serialPort = port
	dmx.isHealthy = true
	log.Printf("Connected to DMX controller at %s", dmx.port)
	return nil
}

func (dmx *OpenDMXController) update() {
	dmx.mutex.Lock()
	defer dmx.mutex.Unlock()

	if dmx.serialPort == nil {
		return
	}

	// OpenDMX packet format:
	// Start byte (0x7E), Label (6 for "Output Only Send DMX Packet Request"),
	// Data length LSB, Data length MSB, DMX data (512 bytes), End byte (0xE7)
	packet := make([]byte, 517)
	packet[0] = 0x7E // Start byte
	packet[1] = 6    // Label: Output Only Send DMX Packet Request
	packet[2] = 0x00 // Data length LSB (512 & 0xFF)
	packet[3] = 0x02 // Data length MSB (512 >> 8)

	// Copy DMX data
	copy(packet[4:516], dmx.dmxData[:])

	packet[516] = 0xE7 // End byte

	_, err := dmx.serialPort.Write(packet)
	if err != nil {
		log.Printf("DMX error writing data: %v", err)
		dmx.isHealthy = false
		dmx.resetConnection()
	}
}

func (dmx *OpenDMXController) resetConnection() {
	if dmx.serialPort != nil {
		dmx.serialPort.Close()
		dmx.serialPort = nil
	}
	dmx.isHealthy = false
}

// Stop signals the Run loop to stop.
func (dmx *OpenDMXController) Stop() {
	close(dmx.stopChan)
}
