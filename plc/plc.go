// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with the field PLC.

package plc

import (
	"fmt"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
	"log"
	"time"
)

type Plc struct {
	IsHealthy        bool
	IoChangeNotifier *websocket.Notifier
	address          string
	handler          *modbus.TCPClientHandler
	client           modbus.Client
	inputs           [inputCount]bool
	registers        [registerCount]uint16
	coils            [coilCount]bool
	oldInputs        [inputCount]bool
	oldRegisters     [registerCount]uint16
	oldCoils         [coilCount]bool
	cycleCounter     int
}

const (
	modbusPort         = 502
	plcLoopPeriodMs    = 100
	plcRetryIntevalSec = 3
	cycleCounterMax    = 100
)

// Discrete inputs
type input int

const (
	fieldEstop input = iota
	redEstop1
	redEstop2
	redEstop3
	blueEstop1
	blueEstop2
	blueEstop3
	redConnected1
	redConnected2
	redConnected3
	blueConnected1
	blueConnected2
	blueConnected3
	scaleNear
	scaleFar
	redSwitchNear
	redSwitchFar
	blueSwitchNear
	blueSwitchFar
	redForceActivate
	redLevitateActivate
	redBoostActivate
	blueForceActivate
	blueLevitateActivate
	blueBoostActivate
	inputCount
)

// 16-bit registers
type register int

const (
	red1Bandwidth register = iota
	red2Bandwidth
	red3Bandwidth
	blue1Bandwidth
	blue2Bandwidth
	blue3Bandwidth
	redForceDistance
	redLevitateDistance
	redBoostDistance
	blueForceDistance
	blueLevitateDistance
	blueBoostDistance
	registerCount
)

// Coils
type coil int

const (
	heartbeat coil = iota
	matchReset
	stackLightGreen
	stackLightOrange
	stackLightRed
	stackLightBlue
	stackLightBuzzer
	red1EthernetDisable
	red2EthernetDisable
	red3EthernetDisable
	blue1EthernetDisable
	blue2EthernetDisable
	blue3EthernetDisable
	coilCount
)

func (plc *Plc) SetAddress(address string) {
	plc.address = address
	plc.resetConnection()

	if plc.IoChangeNotifier == nil {
		// Register a notifier that listeners can subscribe to to get websocket updates about I/O value changes.
		plc.IoChangeNotifier = websocket.NewNotifier("plcIoChange", plc.generateIoChangeMessage)
	}
}

// Loops indefinitely to read inputs from and write outputs to PLC.
func (plc *Plc) Run() {
	for {
		if plc.handler == nil {
			if plc.address == "" {
				time.Sleep(time.Second * plcRetryIntevalSec)
				plc.IsHealthy = false
				continue
			}

			err := plc.connect()
			if err != nil {
				log.Printf("PLC error: %v", err)
				time.Sleep(time.Second * plcRetryIntevalSec)
				plc.IsHealthy = false
				continue
			}
		}

		startTime := time.Now()
		isHealthy := true
		isHealthy = isHealthy && plc.writeCoils()
		isHealthy = isHealthy && plc.readInputs()
		isHealthy = isHealthy && plc.readCounters()
		if !isHealthy {
			plc.resetConnection()
		}
		plc.IsHealthy = isHealthy
		plc.cycleCounter++
		if plc.cycleCounter == cycleCounterMax {
			plc.cycleCounter = 0
		}

		// Detect any changes in input or output and notify listeners if so.
		if plc.inputs != plc.oldInputs || plc.registers != plc.oldRegisters || plc.coils != plc.oldCoils {
			plc.IoChangeNotifier.Notify()
			plc.oldInputs = plc.inputs
			plc.oldRegisters = plc.registers
			plc.oldCoils = plc.coils
		}

		time.Sleep(time.Until(startTime.Add(time.Millisecond * plcLoopPeriodMs)))
	}
}

// GetFieldEstop returns the state of the field emergency stop button (true if e-stop is active).
func (plc *Plc) GetFieldEstop() bool {
	return plc.address != "" && !plc.inputs[fieldEstop]
}

// GetTeamEstops returns the state of the red and blue driver station emergency stop buttons (true if e-stop is active).
func (plc *Plc) GetTeamEstops() ([3]bool, [3]bool) {
	var redEstops, blueEstops [3]bool
	if plc.address != "" {
		redEstops[0] = !plc.inputs[redEstop1]
		redEstops[1] = !plc.inputs[redEstop2]
		redEstops[2] = !plc.inputs[redEstop3]
		blueEstops[0] = !plc.inputs[blueEstop1]
		blueEstops[1] = !plc.inputs[blueEstop2]
		blueEstops[2] = !plc.inputs[blueEstop3]
	}
	return redEstops, blueEstops
}

// GetScaleAndSwitches returns the state of the scale and the red and blue switches.
func (plc *Plc) GetScaleAndSwitches() ([2]bool, [2]bool, [2]bool) {
	var scale, redSwitch, blueSwitch [2]bool

	scale[0] = plc.inputs[scaleNear]
	scale[1] = plc.inputs[scaleFar]
	redSwitch[0] = plc.inputs[redSwitchNear]
	redSwitch[1] = plc.inputs[redSwitchFar]
	blueSwitch[0] = plc.inputs[blueSwitchNear]
	blueSwitch[1] = plc.inputs[blueSwitchFar]

	return scale, redSwitch, blueSwitch
}

// GetVaults returns the state of the red and blue vault power cube sensors.
func (plc *Plc) GetVaults() (uint16, uint16, uint16, uint16, uint16, uint16) {
	return plc.registers[redForceDistance], plc.registers[redLevitateDistance], plc.registers[redBoostDistance],
		plc.registers[blueForceDistance], plc.registers[blueLevitateDistance], plc.registers[blueBoostDistance]
}

// GetPowerUpButtons returns the state of the red and blue power up buttons on the vaults.
func (plc *Plc) GetPowerUpButtons() (bool, bool, bool, bool, bool, bool) {
	return plc.inputs[redForceActivate], plc.inputs[redLevitateActivate], plc.inputs[redBoostActivate],
		plc.inputs[blueForceActivate], plc.inputs[blueLevitateActivate], plc.inputs[blueBoostActivate]
}

// Set the on/off state of the stack lights on the scoring table.
func (plc *Plc) SetStackLights(red, blue, green bool) {
	plc.coils[stackLightRed] = red
	plc.coils[stackLightBlue] = blue
	plc.coils[stackLightGreen] = green
}

// Set the on/off state of the stack lights on the scoring table.
func (plc *Plc) SetStackBuzzer(state bool) {
	plc.coils[stackLightBuzzer] = state
}

func (plc *Plc) GetCycleState(max, index, duration int) bool {
	return plc.cycleCounter/duration%max == index
}

func (plc *Plc) GetInputNames() []string {
	inputNames := make([]string, inputCount)
	for i := range plc.inputs {
		inputNames[i] = input(i).String()
	}
	return inputNames
}

func (plc *Plc) GetRegisterNames() []string {
	registerNames := make([]string, registerCount)
	for i := range plc.registers {
		registerNames[i] = register(i).String()
	}
	return registerNames
}

func (plc *Plc) GetCoilNames() []string {
	coilNames := make([]string, coilCount)
	for i := range plc.coils {
		coilNames[i] = coil(i).String()
	}
	return coilNames
}

func (plc *Plc) connect() error {
	address := fmt.Sprintf("%s:%d", plc.address, modbusPort)
	handler := modbus.NewTCPClientHandler(address)
	handler.Timeout = 1 * time.Second
	handler.SlaveId = 0xFF
	err := handler.Connect()
	if err != nil {
		return err
	}
	log.Printf("Connected to PLC at %s", address)

	plc.handler = handler
	plc.client = modbus.NewClient(plc.handler)
	plc.writeCoils() // Force initial write of the coils upon connection since they may not be triggered by a change.
	return nil
}

func (plc *Plc) resetConnection() {
	if plc.handler != nil {
		plc.handler.Close()
		plc.handler = nil
	}
}

func (plc *Plc) readInputs() bool {
	if len(plc.inputs) == 0 {
		return true
	}

	inputs, err := plc.client.ReadDiscreteInputs(0, uint16(len(plc.inputs)))
	if err != nil {
		log.Printf("PLC error reading inputs: %v", err)
		return false
	}
	if len(inputs)*8 < len(plc.inputs) {
		log.Printf("Insufficient length of PLC inputs: got %d bytes, expected %d bits.", len(inputs), len(plc.inputs))
		return false
	}

	copy(plc.inputs[:], byteToBool(inputs, len(plc.inputs)))
	return true
}

func (plc *Plc) readCounters() bool {
	if len(plc.registers) == 0 {
		return true
	}

	registers, err := plc.client.ReadHoldingRegisters(0, uint16(len(plc.registers)))
	if err != nil {
		log.Printf("PLC error reading registers: %v", err)
		return false
	}
	if len(registers)/2 < len(plc.registers) {
		log.Printf("Insufficient length of PLC counters: got %d bytes, expected %d words.", len(registers),
			len(plc.registers))
		return false
	}

	copy(plc.registers[:], byteToUint(registers, len(plc.registers)))
	return true
}

func (plc *Plc) writeCoils() bool {
	// Send a heartbeat to the PLC so that it can disable outputs if the connection is lost.
	plc.coils[heartbeat] = true

	coils := boolToByte(plc.coils[:])
	_, err := plc.client.WriteMultipleCoils(0, uint16(len(plc.coils)), coils)
	if err != nil {
		log.Printf("PLC error writing coils: %v", err)
		return false
	}

	return true
}

func (plc *Plc) generateIoChangeMessage() interface{} {
	return &struct {
		Inputs    []bool
		Registers []uint16
		Coils     []bool
	}{plc.inputs[:], plc.registers[:], plc.coils[:]}
}

func byteToBool(bytes []byte, size int) []bool {
	bools := make([]bool, size)
	for i := 0; i < size; i++ {
		byteIndex := i / 8
		bitIndex := uint(i % 8)
		bitMask := byte(1 << bitIndex)
		bools[i] = bytes[byteIndex]&bitMask != 0
	}
	return bools
}

func byteToUint(bytes []byte, size int) []uint16 {
	uints := make([]uint16, size)
	for i := 0; i < size; i++ {
		uints[i] = uint16(bytes[2*i])<<8 + uint16(bytes[2*i+1])
	}
	return uints
}

func boolToByte(bools []bool) []byte {
	bytes := make([]byte, (len(bools)+7)/8)
	for i, bit := range bools {
		if bit {
			bytes[i/8] |= 1 << uint(i%8)
		}
	}
	return bytes
}
