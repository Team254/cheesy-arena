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
	"strings"
	"time"
)

type Plc interface {
	SetAddress(address string)
	IsEnabled() bool
	IsHealthy() bool
	IoChangeNotifier() *websocket.Notifier
	Run()
	GetArmorBlockStatuses() map[string]bool
	GetFieldEstop() bool
	GetTeamEstops() ([3]bool, [3]bool)
	GetEthernetConnected() ([3]bool, [3]bool)
	ResetMatch()
	SetStackLights(red, blue, orange, green bool)
	SetStackBuzzer(state bool)
	SetFieldResetLight(state bool)
	GetCycleState(max, index, duration int) bool
	GetInputNames() []string
	GetRegisterNames() []string
	GetCoilNames() []string
	GetChargeStationsLevel() (bool, bool)
	SetChargeStationLights(redState, blueState bool)
}

type ModbusPlc struct {
	address          string
	handler          *modbus.TCPClientHandler
	client           modbus.Client
	isHealthy        bool
	ioChangeNotifier *websocket.Notifier
	inputs           [inputCount]bool
	registers        [registerCount]uint16
	coils            [coilCount]bool
	oldInputs        [inputCount]bool
	oldRegisters     [registerCount]uint16
	oldCoils         [coilCount]bool
	cycleCounter     int
	matchResetCycles int
}

const (
	modbusPort         = 502
	plcLoopPeriodMs    = 100
	plcRetryIntevalSec = 3
	cycleCounterMax    = 100
)

// Discrete inputs
//
//go:generate stringer -type=input
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
	redChargeStationLevel
	blueChargeStationLevel
	inputCount
)

// 16-bit registers
//
//go:generate stringer -type=register
type register int

const (
	fieldIoConnection register = iota
	registerCount
)

// Coils
//
//go:generate stringer -type=coil
type coil int

const (
	heartbeat coil = iota
	matchReset
	stackLightGreen
	stackLightOrange
	stackLightRed
	stackLightBlue
	stackLightBuzzer
	fieldResetLight
	redChargeStationLight
	blueChargeStationLight
	coilCount
)

// Bitmask for decoding fieldIoConnection into individual ArmorBlock connection statuses.
//
//go:generate stringer -type=armorBlock
type armorBlock int

const (
	redDs armorBlock = iota
	blueDs
	armorBlockCount
)

func (plc *ModbusPlc) SetAddress(address string) {
	plc.address = address
	plc.resetConnection()

	if plc.ioChangeNotifier == nil {
		// Register a notifier that listeners can subscribe to to get websocket updates about I/O value changes.
		plc.ioChangeNotifier = websocket.NewNotifier("plcIoChange", plc.generateIoChangeMessage)
	}
}

// Returns true if the PLC is enabled in the configurations.
func (plc *ModbusPlc) IsEnabled() bool {
	return plc.address != ""
}

// Returns true if the PLC is connected and responding to requests.
func (plc *ModbusPlc) IsHealthy() bool {
	return plc.isHealthy
}

// Returns a notifier which fires whenever the I/O values change.
func (plc *ModbusPlc) IoChangeNotifier() *websocket.Notifier {
	return plc.ioChangeNotifier
}

// Loops indefinitely to read inputs from and write outputs to PLC.
func (plc *ModbusPlc) Run() {
	for {
		if plc.handler == nil {
			if !plc.IsEnabled() {
				// No PLC is configured; just allow the loop to continue to simulate inputs and outputs.
				plc.isHealthy = false
			} else {
				err := plc.connect()
				if err != nil {
					log.Printf("PLC error: %v", err)
					time.Sleep(time.Second * plcRetryIntevalSec)
					plc.isHealthy = false
					continue
				}
			}
		}

		startTime := time.Now()
		plc.update()
		time.Sleep(time.Until(startTime.Add(time.Millisecond * plcLoopPeriodMs)))
	}
}

// Returns a map of ArmorBlocks I/O module names to whether they are connected properly.
func (plc *ModbusPlc) GetArmorBlockStatuses() map[string]bool {
	statuses := make(map[string]bool, armorBlockCount)
	for i := 0; i < int(armorBlockCount); i++ {
		statuses[strings.Title(armorBlock(i).String())] = plc.registers[fieldIoConnection]&(1<<i) > 0
	}
	return statuses
}

// Returns the state of the field emergency stop button (true if e-stop is active).
func (plc *ModbusPlc) GetFieldEstop() bool {
	return !plc.inputs[fieldEstop]
}

// Returns the state of the red and blue driver station emergency stop buttons (true if e-stop is active).
func (plc *ModbusPlc) GetTeamEstops() ([3]bool, [3]bool) {
	var redEstops, blueEstops [3]bool
	redEstops[0] = !plc.inputs[redEstop1]
	redEstops[1] = !plc.inputs[redEstop2]
	redEstops[2] = !plc.inputs[redEstop3]
	blueEstops[0] = !plc.inputs[blueEstop1]
	blueEstops[1] = !plc.inputs[blueEstop2]
	blueEstops[2] = !plc.inputs[blueEstop3]
	return redEstops, blueEstops
}

// Returns whether anything is connected to each station's designated Ethernet port on the SCC.
func (plc *ModbusPlc) GetEthernetConnected() ([3]bool, [3]bool) {
	return [3]bool{
			plc.inputs[redConnected1],
			plc.inputs[redConnected2],
			plc.inputs[redConnected3],
		},
		[3]bool{
			plc.inputs[blueConnected1],
			plc.inputs[blueConnected2],
			plc.inputs[blueConnected3],
		}
}

// Resets the internal state of the PLC to start a new match.
func (plc *ModbusPlc) ResetMatch() {
	plc.coils[matchReset] = true
	plc.matchResetCycles = 0
}

// Sets the on/off state of the stack lights on the scoring table.
func (plc *ModbusPlc) SetStackLights(red, blue, orange, green bool) {
	plc.coils[stackLightRed] = red
	plc.coils[stackLightBlue] = blue
	plc.coils[stackLightOrange] = orange
	plc.coils[stackLightGreen] = green
}

// Triggers the "match ready" chime if the state is true.
func (plc *ModbusPlc) SetStackBuzzer(state bool) {
	plc.coils[stackLightBuzzer] = state
}

// Sets the on/off state of the field reset light.
func (plc *ModbusPlc) SetFieldResetLight(state bool) {
	plc.coils[fieldResetLight] = state
}

func (plc *ModbusPlc) GetCycleState(max, index, duration int) bool {
	return plc.cycleCounter/duration%max == index
}

func (plc *ModbusPlc) GetInputNames() []string {
	inputNames := make([]string, inputCount)
	for i := range plc.inputs {
		inputNames[i] = input(i).String()
	}
	return inputNames
}

func (plc *ModbusPlc) GetRegisterNames() []string {
	registerNames := make([]string, registerCount)
	for i := range plc.registers {
		registerNames[i] = register(i).String()
	}
	return registerNames
}

func (plc *ModbusPlc) GetCoilNames() []string {
	coilNames := make([]string, coilCount)
	for i := range plc.coils {
		coilNames[i] = coil(i).String()
	}
	return coilNames
}

// Returns the levelness states of the red and blue charge stations, respectively.
func (plc *ModbusPlc) GetChargeStationsLevel() (bool, bool) {
	return plc.inputs[redChargeStationLevel], plc.inputs[blueChargeStationLevel]
}

// Sets the on/off state of the charge station lights.
func (plc *ModbusPlc) SetChargeStationLights(redState, blueState bool) {
	plc.coils[redChargeStationLight] = redState
	plc.coils[blueChargeStationLight] = blueState
}

func (plc *ModbusPlc) connect() error {
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

func (plc *ModbusPlc) resetConnection() {
	if plc.handler != nil {
		plc.handler.Close()
		plc.handler = nil
	}
}

// Performs a single iteration of reading inputs from and writing outputs to the PLC.
func (plc *ModbusPlc) update() {
	if plc.handler != nil {
		isHealthy := true
		isHealthy = isHealthy && plc.writeCoils()
		isHealthy = isHealthy && plc.readInputs()
		isHealthy = isHealthy && plc.readRegisters()
		if !isHealthy {
			plc.resetConnection()
		}
		plc.isHealthy = isHealthy
	}

	plc.cycleCounter++
	if plc.cycleCounter == cycleCounterMax {
		plc.cycleCounter = 0
	}

	// Detect any changes in input or output and notify listeners if so.
	if plc.inputs != plc.oldInputs || plc.registers != plc.oldRegisters || plc.coils != plc.oldCoils {
		plc.ioChangeNotifier.Notify()
		plc.oldInputs = plc.inputs
		plc.oldRegisters = plc.registers
		plc.oldCoils = plc.coils
	}
}

func (plc *ModbusPlc) readInputs() bool {
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

func (plc *ModbusPlc) readRegisters() bool {
	if len(plc.registers) == 0 {
		return true
	}

	registers, err := plc.client.ReadHoldingRegisters(0, uint16(len(plc.registers)))
	if err != nil {
		log.Printf("PLC error reading registers: %v", err)
		return false
	}
	if len(registers)/2 < len(plc.registers) {
		log.Printf("Insufficient length of PLC registers: got %d bytes, expected %d words.", len(registers),
			len(plc.registers))
		return false
	}

	copy(plc.registers[:], byteToUint(registers, len(plc.registers)))
	return true
}

func (plc *ModbusPlc) writeCoils() bool {
	// Send a heartbeat to the PLC so that it can disable outputs if the connection is lost.
	plc.coils[heartbeat] = true

	coils := boolToByte(plc.coils[:])
	_, err := plc.client.WriteMultipleCoils(0, uint16(len(plc.coils)), coils)
	if err != nil {
		log.Printf("PLC error writing coils: %v", err)
		return false
	}

	if plc.matchResetCycles > 5 {
		plc.coils[matchReset] = false // Only need a short pulse to reset the internal state of the PLC.
	} else {
		plc.matchResetCycles++
	}
	return true
}

func (plc *ModbusPlc) generateIoChangeMessage() any {
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
