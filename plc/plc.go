// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for interfacing with the field PLC.

package plc

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
	"log"
	"strings"
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
	matchResetCycles int
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
	redRungIsLevel
	blueRungIsLevel
	redPowerPortJam
	bluePowerPortJam
	inputCount
)

// 16-bit registers
type register int

const (
	fieldIoConnection register = iota
	redPowerPortBottom
	redPowerPortOuter
	redPowerPortInner
	bluePowerPortBottom
	bluePowerPortOuter
	bluePowerPortInner
	redControlPanelRed
	redControlPanelGreen
	redControlPanelBlue
	redControlPanelIntensity
	blueControlPanelRed
	blueControlPanelGreen
	blueControlPanelBlue
	blueControlPanelIntensity
	redControlPanelColor
	blueControlPanelColor
	redControlPanelLastColor
	blueControlPanelLastColor
	redControlPanelSegments
	blueControlPanelSegments
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
	fieldResetLight
	powerPortMotors
	redStage1Light
	redStage2Light
	redStage3Light
	blueStage1Light
	blueStage2Light
	blueStage3Light
	redTrussLight
	blueTrussLight
	redControlPanelLight
	blueControlPanelLight
	coilCount
)

// Bitmask for decoding fieldIoConnection into individual ArmorBlock connection statuses.
type armorBlock int

const (
	redDs armorBlock = iota
	blueDs
	shieldGenerator
	controlPanel
	armorBlockCount
)

func (plc *Plc) SetAddress(address string) {
	plc.address = address
	plc.resetConnection()

	if plc.IoChangeNotifier == nil {
		// Register a notifier that listeners can subscribe to to get websocket updates about I/O value changes.
		plc.IoChangeNotifier = websocket.NewNotifier("plcIoChange", plc.generateIoChangeMessage)
	}
}

// Returns true if the PLC is enabled in the configurations.
func (plc *Plc) IsEnabled() bool {
	return plc.address != ""
}

// Loops indefinitely to read inputs from and write outputs to PLC.
func (plc *Plc) Run() {
	for {
		if plc.handler == nil {
			if !plc.IsEnabled() {
				// No PLC is configured; just allow the loop to continue to simulate inputs and outputs.
				plc.IsHealthy = false
			} else {
				err := plc.connect()
				if err != nil {
					log.Printf("PLC error: %v", err)
					time.Sleep(time.Second * plcRetryIntevalSec)
					plc.IsHealthy = false
					continue
				}
			}
		}

		startTime := time.Now()

		if plc.handler != nil {
			isHealthy := true
			isHealthy = isHealthy && plc.writeCoils()
			isHealthy = isHealthy && plc.readInputs()
			isHealthy = isHealthy && plc.readRegisters()
			if !isHealthy {
				plc.resetConnection()
			}
			plc.IsHealthy = isHealthy
		}

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

// Returns a map of ArmorBlocks I/O module names to whether they are connected properly.
func (plc *Plc) GetArmorBlockStatuses() map[string]bool {
	statuses := make(map[string]bool, armorBlockCount)
	for i := 0; i < int(armorBlockCount); i++ {
		statuses[strings.Title(armorBlock(i).String())] = plc.registers[fieldIoConnection]&(1<<i) > 0
	}
	return statuses
}

// Returns the state of the field emergency stop button (true if e-stop is active).
func (plc *Plc) GetFieldEstop() bool {
	return plc.IsEnabled() && !plc.inputs[fieldEstop]
}

// Returns the state of the red and blue driver station emergency stop buttons (true if e-stop is active).
func (plc *Plc) GetTeamEstops() ([3]bool, [3]bool) {
	var redEstops, blueEstops [3]bool
	if plc.IsEnabled() {
		redEstops[0] = !plc.inputs[redEstop1]
		redEstops[1] = !plc.inputs[redEstop2]
		redEstops[2] = !plc.inputs[redEstop3]
		blueEstops[0] = !plc.inputs[blueEstop1]
		blueEstops[1] = !plc.inputs[blueEstop2]
		blueEstops[2] = !plc.inputs[blueEstop3]
	}
	return redEstops, blueEstops
}

// Returns whether anything is connected to each station's designated Ethernet port on the SCC.
func (plc *Plc) GetEthernetConnected() ([3]bool, [3]bool) {
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
func (plc *Plc) ResetMatch() {
	plc.coils[matchReset] = true
	plc.matchResetCycles = 0
}

// Returns the total number of power cells scored since match start in each level of the red and blue power ports.
func (plc *Plc) GetPowerPorts() ([3]int, [3]int) {
	return [3]int{
			int(plc.registers[redPowerPortBottom]),
			int(plc.registers[redPowerPortOuter]),
			int(plc.registers[redPowerPortInner]),
		},
		[3]int{
			int(plc.registers[bluePowerPortBottom]),
			int(plc.registers[bluePowerPortOuter]),
			int(plc.registers[bluePowerPortInner]),
		}
}

// Returns whether each of the red and blue power ports are jammed.
func (plc *Plc) GetPowerPortJams() (bool, bool) {
	return plc.inputs[redPowerPortJam], plc.inputs[bluePowerPortJam]
}

// Returns the current color and number of segment transitions for each of the red and blue control panels.
func (plc *Plc) GetControlPanels() (game.ControlPanelColor, int, game.ControlPanelColor, int) {
	return game.ControlPanelColor(plc.registers[redControlPanelColor]), int(plc.registers[redControlPanelSegments]),
		game.ControlPanelColor(plc.registers[blueControlPanelColor]), int(plc.registers[blueControlPanelSegments])
}

// Returns whether each of the red and blue rungs is level.
func (plc *Plc) GetRungs() (bool, bool) {
	return plc.inputs[redRungIsLevel], plc.inputs[blueRungIsLevel]
}

// Sets the on/off state of the stack lights on the scoring table.
func (plc *Plc) SetStackLights(red, blue, orange, green bool) {
	plc.coils[stackLightRed] = red
	plc.coils[stackLightBlue] = blue
	plc.coils[stackLightOrange] = orange
	plc.coils[stackLightGreen] = green
}

// Triggers the "match ready" chime if the state is true.
func (plc *Plc) SetStackBuzzer(state bool) {
	plc.coils[stackLightBuzzer] = state
}

// Sets the on/off state of the field reset light.
func (plc *Plc) SetFieldResetLight(state bool) {
	plc.coils[fieldResetLight] = state
}

// Sets the on/off state of the agitator motors within each power port.
func (plc *Plc) SetPowerPortMotors(state bool) {
	plc.coils[powerPortMotors] = state
}

// Sets the on/off state of the lights mounted within the shield generator trussing.
func (plc *Plc) SetStageActivatedLights(red, blue [3]bool) {
	plc.coils[redStage1Light] = red[0]
	plc.coils[redStage2Light] = red[1]
	plc.coils[redStage3Light] = red[2]
	plc.coils[blueStage1Light] = blue[0]
	plc.coils[blueStage2Light] = blue[1]
	plc.coils[blueStage3Light] = blue[2]
}

// Sets the on/off state of the red and blue alliance stack lights mounted to the control panel.
func (plc *Plc) SetControlPanelLights(red, blue bool) {
	plc.coils[redControlPanelLight] = red
	plc.coils[blueControlPanelLight] = blue
}

// Sets the on/off state of the red and blue alliance stack lights mounted to the top of the shield generator.
func (plc *Plc) SetShieldGeneratorLights(red, blue bool) {
	plc.coils[redTrussLight] = red
	plc.coils[blueTrussLight] = blue
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

func (plc *Plc) readRegisters() bool {
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

func (plc *Plc) writeCoils() bool {
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
