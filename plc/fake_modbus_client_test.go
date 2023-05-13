// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains a fake implementation of the modbus.Client interface for testing.

package plc

import "errors"

type FakeModbusClient struct {
	inputs      [32]bool
	registers   [32]uint16
	coils       [32]bool
	returnError bool
}

func (client *FakeModbusClient) ReadCoils(address, quantity uint16) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) ReadDiscreteInputs(address, quantity uint16) (results []byte, err error) {
	if address != 0 {
		return nil, errors.New("unexpected address")
	}
	if client.returnError {
		return nil, errors.New("dummy error")
	}
	inputsToRead := client.inputs[0:quantity]
	return boolToByte(inputsToRead), nil
}

func (client *FakeModbusClient) WriteSingleCoil(address, value uint16) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error) {
	if address != 0 {
		return nil, errors.New("unexpected address")
	}
	bools := byteToBool(value, int(quantity))
	for i, b := range bools {
		client.coils[i] = b
	}
	return nil, nil
}

func (client *FakeModbusClient) ReadInputRegisters(address, quantity uint16) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) ReadHoldingRegisters(address, quantity uint16) (results []byte, err error) {
	if address != 0 {
		return nil, errors.New("unexpected address")
	}
	registersToRead := client.registers[0:quantity]
	bytes := make([]byte, len(registersToRead)*2)
	for i, value := range registersToRead {
		bytes[2*i] = byte(value >> 8)
		bytes[2*i+1] = byte(value)
	}
	return bytes, nil
}

func (client *FakeModbusClient) WriteSingleRegister(address, value uint16) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) WriteMultipleRegisters(
	address, quantity uint16, value []byte,
) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) ReadWriteMultipleRegisters(
	readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte,
) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) MaskWriteRegister(address, andMask, orMask uint16) (results []byte, err error) {
	return nil, nil
}

func (client *FakeModbusClient) ReadFIFOQueue(address uint16) (results []byte, err error) {
	return nil, nil
}
