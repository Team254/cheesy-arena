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
	Inputs           [15]bool
	Counters         [10]uint16
	Coils            [24]bool
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
	redEstop1
	redEstop2
	redEstop3
	redRotor1
	redTouchpad1
	redTouchpad2
	redTouchpad3
	blueEstop1
	blueEstop2
	blueEstop3
	blueRotor1
	blueTouchpad1
	blueTouchpad2
	blueTouchpad3
)

// 16-bit registers
const (
	redRotor2Count = iota
	redRotor3Count
	redRotor4Count
	redLowBoilerCount
	redHighBoilerCount
	blueRotor2Count
	blueRotor3Count
	blueRotor4Count
	blueLowBoilerCount
	blueHighBoilerCount
)

// Coils
const (
	redSerializer = iota
	redBallLift
	redRotorMotor1
	redRotorMotor2
	redRotorMotor3
	redRotorMotor4
	redAutoLight1
	redAutoLight2
	redTouchpadLight1
	redTouchpadLight2
	redTouchpadLight3
	blueSerializer
	blueBallLift
	blueRotorMotor1
	blueRotorMotor2
	blueRotorMotor3
	blueRotorMotor4
	blueAutoLight1
	blueAutoLight2
	blueTouchpadLight1
	blueTouchpadLight2
	blueTouchpadLight3
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

// Returns the count of the red and blue low and high boilers.
func (plc *Plc) GetBalls() (int, int, int, int) {
	return int(plc.Counters[redLowBoilerCount]), int(plc.Counters[redHighBoilerCount]),
		int(plc.Counters[blueLowBoilerCount]), int(plc.Counters[blueHighBoilerCount])
}

// Returns the state of red and blue activated rotors.
func (plc *Plc) GetRotors() (bool, [3]int, bool, [3]int) {
	var redOtherRotors, blueOtherRotors [3]int

	redOtherRotors[0] = int(plc.Counters[redRotor2Count])
	redOtherRotors[1] = int(plc.Counters[redRotor3Count])
	redOtherRotors[2] = int(plc.Counters[redRotor4Count])
	blueOtherRotors[0] = int(plc.Counters[blueRotor2Count])
	blueOtherRotors[1] = int(plc.Counters[blueRotor3Count])
	blueOtherRotors[2] = int(plc.Counters[blueRotor4Count])

	return plc.Inputs[redRotor1], redOtherRotors, plc.Inputs[blueRotor1], blueOtherRotors
}

func (plc *Plc) GetTouchpads() ([3]bool, [3]bool) {
	var redTouchpads, blueTouchpads [3]bool
	redTouchpads[0] = plc.Inputs[redTouchpad1]
	redTouchpads[1] = plc.Inputs[redTouchpad2]
	redTouchpads[2] = plc.Inputs[redTouchpad3]
	blueTouchpads[0] = plc.Inputs[blueTouchpad1]
	blueTouchpads[1] = plc.Inputs[blueTouchpad2]
	blueTouchpads[2] = plc.Inputs[blueTouchpad3]
	return redTouchpads, blueTouchpads
}

// Resets the ball and rotor gear tooth counts to zero.
func (plc *Plc) ResetCounts() {
	plc.Coils[resetCounts] = true
	plc.resetCountCycles = 0
}

func (plc *Plc) SetBoilerMotors(on bool) {
	plc.Coils[redSerializer] = on
	plc.Coils[redBallLift] = on
	plc.Coils[blueSerializer] = on
	plc.Coils[blueBallLift] = on
}

// Turns on/off the rotor motors based on how many rotors each alliance has.
func (plc *Plc) SetRotorMotors(redRotors, blueRotors int) {
	plc.Coils[redRotorMotor1] = redRotors >= 1
	plc.Coils[redRotorMotor2] = redRotors >= 2
	plc.Coils[redRotorMotor3] = redRotors >= 3
	plc.Coils[redRotorMotor4] = redRotors == 4
	plc.Coils[blueRotorMotor1] = blueRotors >= 1
	plc.Coils[blueRotorMotor2] = blueRotors >= 2
	plc.Coils[blueRotorMotor3] = blueRotors >= 3
	plc.Coils[blueRotorMotor4] = blueRotors == 4
}

// Turns on/off the auto rotor lights based on how many auto rotors each alliance has.
func (plc *Plc) SetRotorLights(redAutoRotors, blueAutoRotors int) {
	plc.Coils[redAutoLight1] = redAutoRotors >= 1
	plc.Coils[redAutoLight2] = redAutoRotors == 2
	plc.Coils[blueAutoLight1] = blueAutoRotors >= 1
	plc.Coils[blueAutoLight2] = blueAutoRotors == 2
}

func (plc *Plc) SetTouchpadLights(redTouchpads, blueTouchpads [3]bool) {
	plc.Coils[redTouchpadLight1] = redTouchpads[0]
	plc.Coils[redTouchpadLight2] = redTouchpads[1]
	plc.Coils[redTouchpadLight3] = redTouchpads[2]
	plc.Coils[blueTouchpadLight1] = blueTouchpads[0]
	plc.Coils[blueTouchpadLight2] = blueTouchpads[1]
	plc.Coils[blueTouchpadLight3] = blueTouchpads[2]
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
