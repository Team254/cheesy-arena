// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestEncodeControlPacket(t *testing.T) {
	arena := setupTestArena(t)

	tcpConn := setupFakeTcpConnection(t)
	defer tcpConn.Close()
	dsConn, err := newDriverStationConnection(254, "R1", tcpConn)
	assert.Nil(t, err)
	defer dsConn.close()

	data := dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(0), data[5])
	assert.Equal(t, byte(0), data[6])
	assert.Equal(t, byte(0), data[20])
	assert.Equal(t, byte(15), data[21])

	// Check the different alliance station values.
	dsConn.AllianceStation = "R2"
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(1), data[5])
	dsConn.AllianceStation = "R3"
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(2), data[5])
	dsConn.AllianceStation = "B1"
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(3), data[5])
	dsConn.AllianceStation = "B2"
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(4), data[5])
	dsConn.AllianceStation = "B3"
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(5), data[5])

	// Check packet count rollover.
	dsConn.packetCount = 255
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(0), data[0])
	assert.Equal(t, byte(255), data[1])
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(1), data[0])
	assert.Equal(t, byte(0), data[1])
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(1), data[0])
	assert.Equal(t, byte(1), data[1])
	dsConn.packetCount = 65535
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(255), data[0])
	assert.Equal(t, byte(255), data[1])
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(0), data[0])
	assert.Equal(t, byte(0), data[1])

	// Check different robot statuses.
	dsConn.Auto = true
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(2), data[3])

	dsConn.Enabled = true
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(6), data[3])

	dsConn.Auto = false
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(4), data[3])

	dsConn.Estop = true
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(132), data[3])

	// Check different match types.
	arena.CurrentMatch.Type = model.Practice
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(1), data[6])
	arena.CurrentMatch.Type = model.Qualification
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(2), data[6])
	arena.CurrentMatch.Type = model.Playoff
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(3), data[6])

	// Check match numbers.
	arena.CurrentMatch.Type = model.Practice
	arena.CurrentMatch.TypeOrder = 42
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(0), data[7])
	assert.Equal(t, byte(42), data[8])
	arena.CurrentMatch.Type = model.Qualification
	arena.CurrentMatch.TypeOrder = 258
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(1), data[7])
	assert.Equal(t, byte(2), data[8])
	arena.CurrentMatch.Type = model.Playoff
	arena.CurrentMatch.TypeOrder = 13
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(0), data[7])
	assert.Equal(t, byte(13), data[8])

	// Check the countdown at different points during the match.
	arena.MatchState = AutoPeriod
	arena.MatchStartTime = time.Now().Add(-time.Duration(4 * time.Second))
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(11), data[21])
	arena.MatchState = PausePeriod
	arena.MatchStartTime = time.Now().Add(-time.Duration(16 * time.Second))
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(135), data[21])
	arena.MatchState = TeleopPeriod
	arena.MatchStartTime = time.Now().Add(-time.Duration(33 * time.Second))
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(119), data[21])
	arena.MatchStartTime = time.Now().Add(-time.Duration(150 * time.Second))
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(2), data[21])
	arena.MatchState = PostMatch
	arena.MatchStartTime = time.Now().Add(-time.Duration(180 * time.Second))
	data = dsConn.encodeControlPacket(arena)
	assert.Equal(t, byte(0), data[21])
}

func TestSendControlPacket(t *testing.T) {
	arena := setupTestArena(t)

	tcpConn := setupFakeTcpConnection(t)
	defer tcpConn.Close()
	dsConn, err := newDriverStationConnection(254, "R1", tcpConn)
	assert.Nil(t, err)
	defer dsConn.close()

	// No real way of checking this since the destination IP is remote, so settle for there being no errors.
	err = dsConn.sendControlPacket(arena)
	assert.Nil(t, err)
}

func TestDecodeStatusPacket(t *testing.T) {
	tcpConn := setupFakeTcpConnection(t)
	defer tcpConn.Close()
	dsConn, err := newDriverStationConnection(254, "R1", tcpConn)
	assert.Nil(t, err)
	defer dsConn.close()

	data := [36]byte{22, 28, 103, 19, 192, 0, 246, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0}
	dsConn.decodeStatusPacket(data)
	assert.Equal(t, 103, dsConn.MissedPacketCount)
	assert.Equal(t, 14, dsConn.DsRobotTripTimeMs)
}

func TestListenForDriverStations(t *testing.T) {
	arena := setupTestArena(t)

	oldAddress := network.ServerIpAddress
	network.ServerIpAddress = "127.0.0.1"
	go arena.listenForDriverStations()
	time.Sleep(time.Millisecond * 10)
	network.ServerIpAddress = oldAddress // Put it back to avoid affecting other tests.

	// Connect with an invalid initial packet.
	tcpConn, err := net.Dial("tcp", "127.0.0.1:1750")
	if assert.Nil(t, err) {
		dataSend := [5]byte{0, 3, 29, 0, 0}
		tcpConn.Write(dataSend[:])
		var dataReceived [100]byte
		_, err = tcpConn.Read(dataReceived[:])
		assert.NotNil(t, err)
		tcpConn.Close()
	}

	// Connect as a team not in the current match.
	tcpConn, err = net.Dial("tcp", "127.0.0.1:1750")
	if assert.Nil(t, err) {
		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:])
		var dataReceived [5]byte
		_, err = tcpConn.Read(dataReceived[:])
		assert.NotNil(t, err)
		tcpConn.Close()
	}

	// Connect as a team in the current match.
	arena.assignTeam(1503, "B2")
	tcpConn, err = net.Dial("tcp", "127.0.0.1:1750")
	if assert.Nil(t, err) {
		defer tcpConn.Close()
		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:])
		var dataReceived [5]byte
		_, err = tcpConn.Read(dataReceived[:])
		assert.Nil(t, err)
		assert.Equal(t, [5]byte{0, 3, 25, 4, 0}, dataReceived)

		time.Sleep(time.Millisecond * 10)
		dsConn := arena.AllianceStations["B2"].DsConn
		if assert.NotNil(t, dsConn) {
			assert.Equal(t, 1503, dsConn.TeamId)
			assert.Equal(t, "B2", dsConn.AllianceStation)

			// Check that an unknown packet type gets ignored and a status packet gets decoded.
			dataSend = [5]byte{0, 3, 37, 0, 0}
			tcpConn.Write(dataSend[:])
			time.Sleep(time.Millisecond * 10)
			dataSend2 := [38]byte{0, 36, 22, 28, 103, 19, 192, 0, 246, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
			tcpConn.Write(dataSend2[:])
			time.Sleep(time.Millisecond * 10)
			assert.Equal(t, 103, dsConn.MissedPacketCount)
			assert.Equal(t, 14, dsConn.DsRobotTripTimeMs)
		}
	}
}

func setupFakeTcpConnection(t *testing.T) net.Conn {
	// Set up a fake TCP endpoint and connection to it.
	l, err := net.Listen("tcp", ":9999")
	assert.Nil(t, err)
	defer l.Close()
	tcpConn, err := net.Dial("tcp", "127.0.0.1:9999")
	assert.Nil(t, err)
	return tcpConn
}
