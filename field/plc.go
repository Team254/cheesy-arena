// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with the field PLC.

package field

import (
	"fmt"
	"github.com/goburrow/modbus"
	"log"
	"time"
)

type Plc struct {
	IsHealthy        bool
	address          string
	handler          *modbus.TCPClientHandler
	client           modbus.Client
	Inputs           [inputCount]bool
	Registers        [registerCount]uint16
	Coils            [coilCount]bool
	oldInputs        [inputCount]bool
	oldRegisters     [registerCount]uint16
	oldCoils         [coilCount]bool
	IoChangeNotifier *Notifier
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
}

// Loops indefinitely to read inputs from and write outputs to PLC.
func (plc *Plc) Run() {
	plc.IoChangeNotifier = NewNotifier()

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
		if plc.Inputs != plc.oldInputs || plc.Registers != plc.oldRegisters || plc.Coils != plc.oldCoils {
			plc.IoChangeNotifier.Notify(nil)
			plc.oldInputs = plc.Inputs
			plc.oldRegisters = plc.Registers
			plc.oldCoils = plc.Coils
		}

		time.Sleep(time.Until(startTime.Add(time.Millisecond * plcLoopPeriodMs)))
	}
}

// Returns the state of the field emergency stop button (true if e-stop is active).
func (plc *Plc) GetFieldEstop() bool {
	return plc.address != "" && !plc.Inputs[fieldEstop]
}

// Returns the state of the red and blue driver station emergency stop buttons (true if e-stop is active).
func (plc *Plc) GetTeamEstops() ([3]bool, [3]bool) {
	var redEstops, blueEstops [3]bool
	if plc.address != "" {
		redEstops[0] = !plc.Inputs[redEstop1]
		redEstops[1] = !plc.Inputs[redEstop2]
		redEstops[2] = !plc.Inputs[redEstop3]
		blueEstops[0] = !plc.Inputs[blueEstop1]
		blueEstops[1] = !plc.Inputs[blueEstop2]
		blueEstops[2] = !plc.Inputs[blueEstop3]
	}
	return redEstops, blueEstops
}

// Returns the state of the scale and the red and blue switches.
func (plc *Plc) GetScaleAndSwitches() ([2]bool, [2]bool, [2]bool) {
	var scale, redSwitch, blueSwitch [2]bool

	scale[0] = plc.Inputs[scaleNear]
	scale[1] = plc.Inputs[scaleFar]
	redSwitch[0] = plc.Inputs[redSwitchNear]
	redSwitch[1] = plc.Inputs[redSwitchFar]
	blueSwitch[0] = plc.Inputs[blueSwitchNear]
	blueSwitch[1] = plc.Inputs[blueSwitchFar]

	return scale, redSwitch, blueSwitch
}

// Returns the state of the red and blue vault power cube sensors.
func (plc *Plc) GetVaults() (uint16, uint16, uint16, uint16, uint16, uint16) {
	return plc.Registers[redForceDistance], plc.Registers[redLevitateDistance], plc.Registers[redBoostDistance],
		plc.Registers[blueForceDistance], plc.Registers[blueLevitateDistance], plc.Registers[blueBoostDistance]
}

// Returns the state of the red and blue power up buttons on the vaults.
func (plc *Plc) GetPowerUpButtons() (bool, bool, bool, bool, bool, bool) {
	return plc.Inputs[redForceActivate], plc.Inputs[redLevitateActivate], plc.Inputs[redBoostActivate],
		plc.Inputs[blueForceActivate], plc.Inputs[blueLevitateActivate], plc.Inputs[blueBoostActivate]
}

// Set the on/off state of the stack lights on the scoring table.
func (plc *Plc) SetStackLights(red, blue, green bool) {
	plc.Coils[stackLightRed] = red
	plc.Coils[stackLightBlue] = blue
	plc.Coils[stackLightGreen] = green
}

// Set the on/off state of the stack lights on the scoring table.
func (plc *Plc) SetStackBuzzer(state bool) {
	plc.Coils[stackLightBuzzer] = state
}

func (plc *Plc) GetCycleState(max, index, duration int) bool {
	return plc.cycleCounter/duration%max == index
}

func (plc *Plc) GetInputNames() []string {
	inputNames := make([]string, inputCount)
	for i := range plc.Inputs {
		inputNames[i] = input(i).String()
	}
	return inputNames
}

func (plc *Plc) GetRegisterNames() []string {
	registerNames := make([]string, registerCount)
	for i := range plc.Registers {
		registerNames[i] = register(i).String()
	}
	return registerNames
}

func (plc *Plc) GetCoilNames() []string {
	coilNames := make([]string, coilCount)
	for i := range plc.Coils {
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
	if len(plc.Inputs) == 0 {
		return true
	}

	inputs, err := plc.client.ReadDiscreteInputs(0, uint16(len(plc.Inputs)))
	if err != nil {
		log.Printf("PLC error reading inputs: %v", err)
		return false
	}
	if len(inputs)*8 < len(plc.Inputs) {
		log.Printf("Insufficient length of PLC inputs: got %d bytes, expected %d bits.", len(inputs), len(plc.Inputs))
		return false
	}

	copy(plc.Inputs[:], byteToBool(inputs, len(plc.Inputs)))
	return true
}

func (plc *Plc) readCounters() bool {
	if len(plc.Registers) == 0 {
		return true
	}

	registers, err := plc.client.ReadHoldingRegisters(0, uint16(len(plc.Registers)))
	if err != nil {
		log.Printf("PLC error reading registers: %v", err)
		return false
	}
	if len(registers)/2 < len(plc.Registers) {
		log.Printf("Insufficient length of PLC counters: got %d bytes, expected %d words.", len(registers),
			len(plc.Registers))
		return false
	}

	copy(plc.Registers[:], byteToUint(registers, len(plc.Registers)))
	return true
}

func (plc *Plc) writeCoils() bool {
	// Send a heartbeat to the PLC so that it can disable outputs if the connection is lost.
	plc.Coils[heartbeat] = true

	coils := boolToByte(plc.Coils[:])
	_, err := plc.client.WriteMultipleCoils(0, uint16(len(plc.Coils)), coils)
	if err != nil {
		log.Printf("PLC error writing coils: %v", err)
		return false
	}

	return true
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
