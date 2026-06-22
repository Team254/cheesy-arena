// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package led

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

type fakeConn struct {
	writes [][]byte
	err    error
}

func (conn *fakeConn) Read(_ []byte) (int, error)         { return 0, nil }
func (conn *fakeConn) Close() error                       { return nil }
func (conn *fakeConn) LocalAddr() net.Addr                { return nil }
func (conn *fakeConn) RemoteAddr() net.Addr               { return nil }
func (conn *fakeConn) SetDeadline(_ time.Time) error      { return nil }
func (conn *fakeConn) SetReadDeadline(_ time.Time) error  { return nil }
func (conn *fakeConn) SetWriteDeadline(_ time.Time) error { return nil }

func (conn *fakeConn) Write(packet []byte) (int, error) {
	packetCopy := make([]byte, len(packet))
	copy(packetCopy, packet)
	conn.writes = append(conn.writes, packetCopy)
	if conn.err != nil {
		return 0, conn.err
	}
	return len(packet), nil
}

func TestControllerSetAddressBlankDisables(t *testing.T) {
	controller := NewController()

	assert.Nil(t, controller.SetAddress(""))
	assert.Nil(t, controller.conn)
	assert.Nil(t, controller.Update())
}

func TestControllerUpdateSendsSacnPackets(t *testing.T) {
	conn := &fakeConn{}
	controller := NewController()
	controller.conn = conn
	controller.SetMode(RedMode, BlueMode)

	assert.Nil(t, controller.Update())

	if assert.Len(t, conn.writes, 1) {
		packet := conn.writes[0]
		assert.Equal(t, byte(100), packet[108])
		assert.Equal(t, byte(1), packet[114])
		assert.Equal(t, byte(2), packet[123])
		assert.Equal(t, byte(1), packet[124])
		assert.Equal(t, []byte{255, 0, 0}, packet[dmxOffset(1):dmxOffset(1)+3])
		assert.Equal(t, []byte{255, 0, 0}, packet[dmxOffset(49):dmxOffset(49)+3])
		assert.Equal(t, []byte{0, 0, 255}, packet[dmxOffset(193):dmxOffset(193)+3])
		assert.Equal(t, []byte{0, 0, 255}, packet[dmxOffset(337):dmxOffset(337)+3])
	}
}

func TestControllerUpdateSendsOnChangeAndHeartbeat(t *testing.T) {
	conn := &fakeConn{}
	controller := NewController()
	controller.conn = conn
	controller.SetMode(RedMode, BlueMode)

	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 1)

	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 1)

	controller.universes[1].lastPacketTime = time.Now().Add(-heartbeatInterval)
	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 2)

	controller.SetMode(OffMode, BlueMode)
	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 3)
}

func TestControllerUpdateThrottlesAfterWriteError(t *testing.T) {
	writeErr := errors.New("broken pipe")
	conn := &fakeConn{err: writeErr}
	controller := NewController()
	controller.conn = conn
	controller.SetMode(RedMode, BlueMode)

	assert.Equal(t, writeErr, controller.Update())
	assert.Len(t, conn.writes, 1)

	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 1)

	controller.universes[1].lastPacketTime = time.Now().Add(-heartbeatInterval)
	assert.Equal(t, writeErr, controller.Update())
	assert.Len(t, conn.writes, 2)

	controller.SetMode(OffMode, BlueMode)
	assert.Equal(t, writeErr, controller.Update())
	assert.Len(t, conn.writes, 3)
}

func TestControllerUpdateSupportsMultipleUniverses(t *testing.T) {
	conn := &fakeConn{}
	controller := NewController()
	controller.conn = conn
	controller.fixtures = fixtureLayout{
		red:  []fixture{{redGoalSide1Bot, 1, 1}},
		blue: []fixture{{blueGoalSide1Bot, 2, 1}},
	}
	controller.SetMode(RedMode, BlueMode)

	assert.Nil(t, controller.Update())

	if assert.Len(t, conn.writes, 2) {
		universe1Packet := packetByUniverse(conn.writes, 1)
		universe2Packet := packetByUniverse(conn.writes, 2)
		if assert.NotNil(t, universe1Packet) {
			assert.Equal(t, []byte{255, 0, 0}, universe1Packet[dmxOffset(1):dmxOffset(1)+3])
		}
		if assert.NotNil(t, universe2Packet) {
			assert.Equal(t, []byte{0, 0, 255}, universe2Packet[dmxOffset(1):dmxOffset(1)+3])
		}
	}
}

func TestStartupModeFillsSidesInFmsOrder(t *testing.T) {
	controller := NewController()
	controller.SetMode(RedStartupMode, OffMode)

	controller.redZone.counter = 50
	controller.redZone.updatePixels(Red)

	assert.Equal(t, Red, controller.redZone.pixels[3])
	assert.Equal(t, Red, controller.redZone.pixels[4])
	assert.NotEqual(t, Black, controller.redZone.pixels[16])
	assert.Equal(t, Black, controller.redZone.pixels[32])
	assert.NotEqual(t, Black, controller.redZone.pixels[55])
}

func TestAdvantageModeSweepsInOppositeDirectionsBySide(t *testing.T) {
	controller := NewController()
	controller.SetMode(RedAdvantageMode, OffMode)

	controller.redZone.counter = advantageStepCycle
	controller.redZone.updatePixels(Red)

	assert.Equal(t, White, controller.redZone.pixels[0])
	assert.Equal(t, White, controller.redZone.pixels[31])
}

func TestPulseModeScalesAllianceColor(t *testing.T) {
	controller := NewController()
	controller.SetMode(RedPulseMode, OffMode)

	controller.redZone.updatePixels(Red)
	assert.Equal(t, Black, controller.redZone.pixels[0])

	controller.redZone.counter = pulseHalfPeriod
	controller.redZone.updatePixels(Red)
	assert.Equal(t, Red, controller.redZone.pixels[0])
}

func dmxOffset(startAddress int) int {
	return pixelDataOffset + startAddress - 1
}

func packetByUniverse(packets [][]byte, universe int) []byte {
	for _, packet := range packets {
		if int(packet[113])<<8|int(packet[114]) == universe {
			return packet
		}
	}
	return nil
}
