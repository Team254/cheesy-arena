// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package plc

import (
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlcInitialization(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	var notifier websocket.Notifier
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &notifier

	assert.Equal(t, false, plc.IsEnabled())
	plc.SetAddress("dummy")
	assert.Equal(t, true, plc.IsEnabled())
	assert.Equal(t, &notifier, plc.IoChangeNotifier())
}

func TestPlcGetCycleState(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()

	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, true, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
	plc.update()
	assert.Equal(t, false, plc.GetCycleState(3, 1, 2))
}

func TestPlcGetNames(t *testing.T) {
	var plc ModbusPlc

	assert.Equal(
		t,
		[]string{
			"fieldEStop",
			"red1EStop",
			"red1AStop",
			"red2EStop",
			"red2AStop",
			"red3EStop",
			"red3AStop",
			"blue1EStop",
			"blue1AStop",
			"blue2EStop",
			"blue2AStop",
			"blue3EStop",
			"blue3AStop",
			"redConnected1",
			"redConnected2",
			"redConnected3",
			"blueConnected1",
			"blueConnected2",
			"blueConnected3",
		},
		plc.GetInputNames(),
	)

	assert.Equal(
		t,
		[]string{
			"fieldIoConnection",
			"redProcessor",
			"blueProcessor",
		},
		plc.GetRegisterNames(),
	)

	assert.Equal(
		t,
		[]string{
			"heartbeat",
			"matchReset",
			"stackLightGreen",
			"stackLightOrange",
			"stackLightRed",
			"stackLightBlue",
			"stackLightBuzzer",
			"fieldResetLight",
			"redTrussLightOuter",
			"redTrussLightMiddle",
			"redTrussLightInner",
			"blueTrussLightOuter",
			"blueTrussLightMiddle",
			"blueTrussLightInner",
		},
		plc.GetCoilNames(),
	)
}

func TestPlcInputs(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	client.inputs[0] = true
	plc.update()
	assert.Equal(t, false, plc.GetFieldEStop())
	client.inputs[0] = false
	plc.update()
	assert.Equal(t, true, plc.GetFieldEStop())

	client.inputs[1] = true
	client.inputs[2] = true
	client.inputs[3] = true
	client.inputs[4] = true
	client.inputs[5] = true
	client.inputs[6] = true
	client.inputs[7] = true
	client.inputs[8] = true
	client.inputs[9] = true
	client.inputs[10] = true
	client.inputs[11] = true
	client.inputs[12] = true
	plc.update()
	redEStops, blueEStops := plc.GetTeamEStops()
	redAStops, blueAStops := plc.GetTeamAStops()
	assert.Equal(t, [3]bool{false, false, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{false, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[1] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, false, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{false, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[2] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, false, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[3] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, false, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[4] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, false}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[5] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, false}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[6] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{false, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[7] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{false, false, false}, blueAStops)
	client.inputs[8] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, false, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, false, false}, blueAStops)
	client.inputs[9] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, false, false}, blueAStops)
	client.inputs[10] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, false}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, true, false}, blueAStops)
	client.inputs[11] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, true}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, true, false}, blueAStops)
	client.inputs[12] = false
	plc.update()
	redEStops, blueEStops = plc.GetTeamEStops()
	redAStops, blueAStops = plc.GetTeamAStops()
	assert.Equal(t, [3]bool{true, true, true}, redEStops)
	assert.Equal(t, [3]bool{true, true, true}, blueEStops)
	assert.Equal(t, [3]bool{true, true, true}, redAStops)
	assert.Equal(t, [3]bool{true, true, true}, blueAStops)

	client.inputs[13] = false
	client.inputs[14] = false
	client.inputs[15] = false
	client.inputs[16] = false
	client.inputs[17] = false
	client.inputs[18] = false
	plc.update()
	redConnected, blueConnected := plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{false, false, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[13] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, false, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[14] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[15] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[16] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, false, false}, blueConnected)
	client.inputs[17] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, true, false}, blueConnected)
	client.inputs[18] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, true, true}, blueConnected)
}

func TestPlcInputsGameSpecific(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	// None in 2025.
}

func TestPlcRegisters(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	testCases := map[uint16][4]bool{
		0:  {false, false, false, false},
		1:  {true, false, false, false},
		2:  {false, true, false, false},
		3:  {true, true, false, false},
		4:  {false, false, true, false},
		5:  {true, false, true, false},
		6:  {false, true, true, false},
		7:  {true, true, true, false},
		8:  {false, false, false, true},
		9:  {true, false, false, true},
		10: {false, true, false, true},
		11: {true, true, false, true},
		12: {false, false, true, true},
		13: {true, false, true, true},
		14: {false, true, true, true},
		15: {true, true, true, true},
	}

	for value, bits := range testCases {
		client.registers[0] = value
		plc.update()
		assert.Equal(
			t,
			map[string]bool{"RedDs": bits[0], "BlueDs": bits[1], "RedIoLink": bits[2], "BlueIoLink": bits[3]},
			plc.GetArmorBlockStatuses(),
		)
	}
}

func TestPlcRegistersGameSpecific(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	client.registers[1] = 0
	client.registers[2] = 0
	plc.update()
	redProcessor, blueProcessor := plc.GetProcessorCounts()
	assert.Equal(t, 0, redProcessor)
	assert.Equal(t, 0, blueProcessor)
	client.registers[1] = 12
	plc.update()
	redProcessor, blueProcessor = plc.GetProcessorCounts()
	assert.Equal(t, 12, redProcessor)
	assert.Equal(t, 0, blueProcessor)
	client.registers[2] = 34
	plc.update()
	redProcessor, blueProcessor = plc.GetProcessorCounts()
	assert.Equal(t, 12, redProcessor)
	assert.Equal(t, 34, blueProcessor)
}

func TestPlcCoils(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	assert.Equal(t, false, client.coils[0])
	plc.update()
	assert.Equal(t, true, client.coils[0])

	assert.Equal(t, false, client.coils[1])
	client.registers[fieldIoConnection] = 31
	plc.registers[fieldIoConnection] = 31
	plc.registers[redProcessor] = 1
	plc.registers[blueProcessor] = 2
	plc.ResetMatch()
	plc.update()
	assert.Equal(t, true, client.coils[1])
	assert.Equal(t, 31, int(plc.registers[fieldIoConnection]))
	assert.Equal(t, 0, int(plc.registers[redProcessor]))
	assert.Equal(t, 0, int(plc.registers[blueProcessor]))

	plc.SetStackLights(false, false, false, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, false, client.coils[3])
	assert.Equal(t, false, client.coils[4])
	assert.Equal(t, false, client.coils[5])
	plc.SetStackLights(true, false, false, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, false, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, false, client.coils[5])
	plc.SetStackLights(true, true, false, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, false, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, true, client.coils[5])
	plc.SetStackLights(true, true, true, false)
	plc.update()
	assert.Equal(t, false, client.coils[2])
	assert.Equal(t, true, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, true, client.coils[5])
	plc.SetStackLights(true, true, true, true)
	plc.update()
	assert.Equal(t, true, client.coils[2])
	assert.Equal(t, true, client.coils[3])
	assert.Equal(t, true, client.coils[4])
	assert.Equal(t, true, client.coils[5])

	plc.SetStackBuzzer(false)
	plc.update()
	assert.Equal(t, false, client.coils[6])
	plc.SetStackBuzzer(true)
	plc.update()
	assert.Equal(t, true, client.coils[6])

	plc.SetFieldResetLight(false)
	plc.update()
	assert.Equal(t, false, client.coils[7])
	plc.SetFieldResetLight(true)
	plc.update()
	assert.Equal(t, true, client.coils[7])
}

func TestPlcCoilsGameSpecific(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	plc.SetTrussLights([3]bool{false, false, false}, [3]bool{false, false, false})
	plc.update()
	assert.Equal(t, []bool{false, false, false, false, false, false}, client.coils[8:14])
	plc.SetTrussLights([3]bool{true, false, false}, [3]bool{false, false, false})
	plc.update()
	assert.Equal(t, []bool{true, false, false, false, false, false}, client.coils[8:14])
	plc.SetTrussLights([3]bool{true, true, false}, [3]bool{false, false, false})
	plc.update()
	assert.Equal(t, []bool{true, true, false, false, false, false}, client.coils[8:14])
	plc.SetTrussLights([3]bool{true, true, true}, [3]bool{false, false, false})
	plc.update()
	assert.Equal(t, []bool{true, true, true, false, false, false}, client.coils[8:14])
	plc.SetTrussLights([3]bool{true, true, true}, [3]bool{true, false, false})
	plc.update()
	assert.Equal(t, []bool{true, true, true, true, false, false}, client.coils[8:14])
	plc.SetTrussLights([3]bool{true, true, true}, [3]bool{true, true, false})
	plc.update()
	assert.Equal(t, []bool{true, true, true, true, true, false}, client.coils[8:14])
	plc.SetTrussLights([3]bool{true, true, true}, [3]bool{true, true, true})
	plc.update()
	assert.Equal(t, []bool{true, true, true, true, true, true}, client.coils[8:14])
}

func TestPlcIsHealthy(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	assert.Equal(t, false, plc.IsHealthy())
	plc.update()
	assert.Equal(t, true, plc.IsHealthy())

	client.returnError = true
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())

	client.returnError = false
	plc.update()
	assert.Equal(t, false, plc.IsHealthy())
}

func TestByteToBool(t *testing.T) {
	bytes := []byte{7, 254, 3}
	bools := byteToBool(bytes, 17)
	if assert.Equal(t, 17, len(bools)) {
		expectedBools := []bool{
			true, true, true, false, false, false, false, false, false, true, true, true, true, true, true, true, true,
		}
		assert.Equal(t, expectedBools, bools)
	}
}

func TestByteToUint(t *testing.T) {
	bytes := []byte{1, 77, 2, 253, 21, 179}
	uints := byteToUint(bytes, 3)
	if assert.Equal(t, 3, len(uints)) {
		assert.Equal(t, []uint16{333, 765, 5555}, uints)
	}
}

func TestBoolToByte(t *testing.T) {
	bools := []bool{true, true, false, false, true, false, false, false, false, true}
	bytes := boolToByte(bools)
	if assert.Equal(t, 2, len(bytes)) {
		assert.Equal(t, []byte{19, 2}, bytes)
		assert.Equal(t, bools, byteToBool(bytes, len(bools)))
	}
}
