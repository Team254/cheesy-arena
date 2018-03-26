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
	Inputs           [37]bool
	Counters         [0]uint16
	Coils            [8]bool
	cycleCounter     int
	resetCountCycles int
}

const (
	modbusPort         = 502
	plcLoopPeriodMs    = 100
	plcRetryIntevalSec = 3
	cycleCounterMax    = 96
)

// Discrete inputs
const (
	fieldEstop = iota
	scaleNear
	scaleFar
	redEstop1
	redEstop2
	redEstop3
	redSwitchNear
	redSwitchFar
	redForceCube1
	redForceCube2
	redForceCube3
	redForceButton
	redLevitateCube1
	redLevitateCube2
	redLevitateCube3
	redLevitateButton
	redBoostCube1
	redBoostCube2
	redBoostCube3
	redBoostButton
	blueEstop1
	blueEstop2
	blueEstop3
	blueSwitchNear
	blueSwitchFar
	blueForceCube1
	blueForceCube2
	blueForceCube3
	blueForceButton
	blueLevitate1
	blueLevitate2
	blueLevitate3
	blueLevitateButton
	blueBoostCube1
	blueBoostCube2
	blueBoostCube3
	blueBoostButton
)

// 16-bit registers
const ()

// Coils
const (
	redForceLight = iota
	redLevitateLight
	redBoostLight
	blueForceLight
	blueLevitateLight
	blueBoostLight
	resetCounts
	heartbeat
)

func (plc *Plc) SetAddress(address string) {
	plc.address = address
	plc.resetConnection()
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
func (plc *Plc) GetVaults() ([3]bool, [3]bool, [3]bool, [3]bool, [3]bool, [3]bool) {
	var redForce, redLevitate, redBoost, blueForce, blueLevitate, blueBoost [3]bool

	redForce[0] = plc.Inputs[redForceCube1]
	redForce[1] = plc.Inputs[redForceCube2]
	redForce[2] = plc.Inputs[redForceCube3]
	redLevitate[0] = plc.Inputs[redLevitateCube1]
	redLevitate[1] = plc.Inputs[redLevitateCube2]
	redLevitate[2] = plc.Inputs[redLevitateCube3]
	redBoost[0] = plc.Inputs[redBoostCube1]
	redBoost[1] = plc.Inputs[redBoostCube2]
	redBoost[2] = plc.Inputs[redBoostCube3]
	blueForce[0] = plc.Inputs[blueForceCube1]
	blueForce[1] = plc.Inputs[blueForceCube2]
	blueForce[2] = plc.Inputs[blueForceCube3]
	blueLevitate[0] = plc.Inputs[blueLevitate1]
	blueLevitate[1] = plc.Inputs[blueLevitate2]
	blueLevitate[2] = plc.Inputs[blueLevitate3]
	blueBoost[0] = plc.Inputs[blueBoostCube1]
	blueBoost[1] = plc.Inputs[blueBoostCube2]
	blueBoost[2] = plc.Inputs[blueBoostCube3]

	return redForce, redLevitate, redBoost, blueForce, blueLevitate, blueBoost
}

// Returns the state of the red and blue power up buttons on the vaults.
func (plc *Plc) GetPowerUpButtons() (bool, bool, bool, bool, bool, bool) {
	return plc.Inputs[redForceButton], plc.Inputs[redLevitateButton], plc.Inputs[redBoostButton],
		plc.Inputs[blueForceButton], plc.Inputs[blueLevitateButton], plc.Inputs[blueBoostButton]
}

// Resets the counter counts to zero.
func (plc *Plc) ResetCounts() {
	plc.Coils[resetCounts] = true
	plc.resetCountCycles = 0
}

// Sets the state of the lights inside the power up buttons on the vaults.
func (plc *Plc) SetPowerUpLights(redForce, redLevitate, redBoost, blueForce, blueLevitate, blueBoost bool) {
	plc.Coils[redForceLight] = redForce
	plc.Coils[redLevitateLight] = redLevitate
	plc.Coils[redBoostLight] = redBoost
	plc.Coils[blueForceLight] = blueForce
	plc.Coils[blueLevitateLight] = blueLevitate
	plc.Coils[blueBoostLight] = blueBoost
}

func (plc *Plc) GetCycleState(max, index, duration int) bool {
	return plc.cycleCounter/duration%max == index
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
	if len(plc.Counters) == 0 {
		return true
	}

	registers, err := plc.client.ReadHoldingRegisters(0, uint16(len(plc.Counters)))
	if err != nil {
		log.Printf("PLC error reading registers: %v", err)
		return false
	}
	if len(registers)/2 < len(plc.Counters) {
		log.Printf("Insufficient length of PLC counters: got %d bytes, expected %d words.", len(registers),
			len(plc.Counters))
		return false
	}

	copy(plc.Counters[:], byteToUint(registers, len(plc.Counters)))
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

	if plc.resetCountCycles > 5 {
		plc.Coils[resetCounts] = false // Need to send a short pulse to reset the counters.
	} else {
		plc.resetCountCycles++
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
