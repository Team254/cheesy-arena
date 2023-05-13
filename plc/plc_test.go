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
			"fieldEstop",
			"redEstop1",
			"redEstop2",
			"redEstop3",
			"blueEstop1",
			"blueEstop2",
			"blueEstop3",
			"redConnected1",
			"redConnected2",
			"redConnected3",
			"blueConnected1",
			"blueConnected2",
			"blueConnected3",
			"redChargeStationLevel",
			"blueChargeStationLevel",
		},
		plc.GetInputNames(),
	)

	assert.Equal(
		t,
		[]string{
			"fieldIoConnection",
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
			"redChargeStationLight",
			"blueChargeStationLight",
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
	assert.Equal(t, false, plc.GetFieldEstop())
	client.inputs[0] = false
	plc.update()
	assert.Equal(t, true, plc.GetFieldEstop())

	client.inputs[1] = true
	client.inputs[2] = true
	client.inputs[3] = true
	client.inputs[4] = true
	client.inputs[5] = true
	client.inputs[6] = true
	plc.update()
	redEstops, blueEstops := plc.GetTeamEstops()
	assert.Equal(t, [3]bool{false, false, false}, redEstops)
	assert.Equal(t, [3]bool{false, false, false}, blueEstops)
	client.inputs[1] = false
	plc.update()
	redEstops, blueEstops = plc.GetTeamEstops()
	assert.Equal(t, [3]bool{true, false, false}, redEstops)
	assert.Equal(t, [3]bool{false, false, false}, blueEstops)
	client.inputs[2] = false
	plc.update()
	redEstops, blueEstops = plc.GetTeamEstops()
	assert.Equal(t, [3]bool{true, true, false}, redEstops)
	assert.Equal(t, [3]bool{false, false, false}, blueEstops)
	client.inputs[3] = false
	plc.update()
	redEstops, blueEstops = plc.GetTeamEstops()
	assert.Equal(t, [3]bool{true, true, true}, redEstops)
	assert.Equal(t, [3]bool{false, false, false}, blueEstops)
	client.inputs[4] = false
	plc.update()
	redEstops, blueEstops = plc.GetTeamEstops()
	assert.Equal(t, [3]bool{true, true, true}, redEstops)
	assert.Equal(t, [3]bool{true, false, false}, blueEstops)
	client.inputs[5] = false
	plc.update()
	redEstops, blueEstops = plc.GetTeamEstops()
	assert.Equal(t, [3]bool{true, true, true}, redEstops)
	assert.Equal(t, [3]bool{true, true, false}, blueEstops)
	client.inputs[6] = false
	plc.update()
	redEstops, blueEstops = plc.GetTeamEstops()
	assert.Equal(t, [3]bool{true, true, true}, redEstops)
	assert.Equal(t, [3]bool{true, true, true}, blueEstops)

	client.inputs[7] = false
	client.inputs[8] = false
	client.inputs[9] = false
	client.inputs[10] = false
	client.inputs[11] = false
	client.inputs[12] = false
	plc.update()
	redConnected, blueConnected := plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{false, false, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[7] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, false, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[8] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, false}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[9] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{false, false, false}, blueConnected)
	client.inputs[10] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, false, false}, blueConnected)
	client.inputs[11] = true
	plc.update()
	redConnected, blueConnected = plc.GetEthernetConnected()
	assert.Equal(t, [3]bool{true, true, true}, redConnected)
	assert.Equal(t, [3]bool{true, true, false}, blueConnected)
	client.inputs[12] = true
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

	client.inputs[13] = false
	client.inputs[14] = false
	plc.update()
	redLevel, blueLevel := plc.GetChargeStationsLevel()
	assert.Equal(t, false, redLevel)
	assert.Equal(t, false, blueLevel)
	client.inputs[13] = true
	plc.update()
	redLevel, blueLevel = plc.GetChargeStationsLevel()
	assert.Equal(t, true, redLevel)
	assert.Equal(t, false, blueLevel)
	client.inputs[14] = true
	plc.update()
	redLevel, blueLevel = plc.GetChargeStationsLevel()
	assert.Equal(t, true, redLevel)
	assert.Equal(t, true, blueLevel)
}

func TestPlcRegisters(t *testing.T) {
	var client FakeModbusClient
	var plc ModbusPlc
	plc.client = &client
	plc.handler = modbus.NewTCPClientHandler("dummy")
	plc.ioChangeNotifier = &websocket.Notifier{}

	client.registers[0] = 0
	plc.update()
	assert.Equal(t, map[string]bool{"RedDs": false, "BlueDs": false}, plc.GetArmorBlockStatuses())
	client.registers[0] = 1
	plc.update()
	assert.Equal(t, map[string]bool{"RedDs": true, "BlueDs": false}, plc.GetArmorBlockStatuses())
	client.registers[0] = 2
	plc.update()
	assert.Equal(t, map[string]bool{"RedDs": false, "BlueDs": true}, plc.GetArmorBlockStatuses())
	client.registers[0] = 3
	plc.update()
	assert.Equal(t, map[string]bool{"RedDs": true, "BlueDs": true}, plc.GetArmorBlockStatuses())
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
	plc.ResetMatch()
	plc.update()
	assert.Equal(t, true, client.coils[1])

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

	plc.SetChargeStationLights(false, false)
	plc.update()
	assert.Equal(t, false, client.coils[8])
	assert.Equal(t, false, client.coils[9])
	plc.SetChargeStationLights(true, false)
	plc.update()
	assert.Equal(t, true, client.coils[8])
	assert.Equal(t, false, client.coils[9])
	plc.SetChargeStationLights(true, true)
	plc.update()
	assert.Equal(t, true, client.coils[8])
	assert.Equal(t, true, client.coils[9])
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
		expectedBools := []bool{true, true, true, false, false, false, false, false, false, true, true, true, true,
			true, true, true, true}
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
