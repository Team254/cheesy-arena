// Copyright 2026 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package led

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeConn struct {
	writes [][]byte
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

	if assert.Len(t, conn.writes, 2) {
		redPacket := conn.writes[0]
		bluePacket := conn.writes[1]
		assert.Equal(t, byte(100), redPacket[108])
		assert.Equal(t, byte(redStripUniverse), redPacket[114])
		assert.Equal(t, byte(blueStripUniverse), bluePacket[114])
		assert.Equal(t, byte(0), redPacket[123])
		assert.Equal(t, byte(1+3*numPixels), redPacket[124])
		assert.Equal(t, []byte{255, 0, 0}, redPacket[pixelDataOffset:pixelDataOffset+3])
		assert.Equal(t, []byte{0, 0, 255}, bluePacket[pixelDataOffset:pixelDataOffset+3])
	}
}

func TestControllerUpdateSendsOnChangeAndHeartbeat(t *testing.T) {
	conn := &fakeConn{}
	controller := NewController()
	controller.conn = conn
	controller.SetMode(RedMode, BlueMode)

	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 2)

	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 2)

	controller.redStrip.lastPacketTime = time.Now().Add(-heartbeatInterval)
	controller.blueStrip.lastPacketTime = time.Now().Add(-heartbeatInterval)
	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 4)

	controller.SetMode(OffMode, BlueMode)
	assert.Nil(t, controller.Update())
	assert.Len(t, conn.writes, 5)
}

func TestStartupModeFillsSidesInFmsOrder(t *testing.T) {
	controller := NewController()
	controller.SetMode(RedStartupMode, OffMode)

	controller.redStrip.counter = 50
	controller.redStrip.updatePixels()

	assert.Equal(t, Red, controller.redStrip.pixels[3])
	assert.Equal(t, Red, controller.redStrip.pixels[4])
	assert.NotEqual(t, Black, controller.redStrip.pixels[16])
	assert.Equal(t, Black, controller.redStrip.pixels[32])
	assert.NotEqual(t, Black, controller.redStrip.pixels[55])
}

func TestAdvantageModeSweepsInOppositeDirectionsBySide(t *testing.T) {
	controller := NewController()
	controller.SetMode(RedAdvantageMode, OffMode)

	controller.redStrip.counter = advantageStepCycle
	controller.redStrip.updatePixels()

	assert.Equal(t, White, controller.redStrip.pixels[0])
	assert.Equal(t, White, controller.redStrip.pixels[31])
}

func TestPulseModeScalesAllianceColor(t *testing.T) {
	controller := NewController()
	controller.SetMode(RedPulseMode, OffMode)

	controller.redStrip.updatePixels()
	assert.Equal(t, Black, controller.redStrip.pixels[0])

	controller.redStrip.counter = pulseHalfPeriod
	controller.redStrip.updatePixels()
	assert.Equal(t, Red, controller.redStrip.pixels[0])
}
