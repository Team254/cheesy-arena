// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package field

import (
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestDriverStationListenAddress(t *testing.T) {
	oldDevMode := network.DevMode
	t.Cleanup(
		func() {
			network.DevMode = oldDevMode
		},
	)

	network.DevMode = false
	assert.Equal(t, network.ServerIpAddress+":1750", listenAddress(1750))

	network.DevMode = true
	assert.Equal(t, ":1750", listenAddress(1750))
}

func TestEncodeControlPacket(t *testing.T) {
	arena := setupTestArena(t)

	tcpConn := setupFakeTcpConnection(t)
	defer tcpConn.Close()
	dsConn, err := newDriverStationConnection(254, "R1", tcpConn, driverStationRoboRioUdpPort, false)
	assert.Nil(t, err)
	defer dsConn.close()

	data := dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(0), data[5])
	assert.Equal(t, byte(0), data[6])
	assert.Equal(t, byte(0), data[20])
	assert.Equal(t, byte(20), data[21])

	// Check the different alliance station values.
	dsConn.AllianceStation = "R2"
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(1), data[5])
	dsConn.AllianceStation = "R3"
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(2), data[5])
	dsConn.AllianceStation = "B1"
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(3), data[5])
	dsConn.AllianceStation = "B2"
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(4), data[5])
	dsConn.AllianceStation = "B3"
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(5), data[5])

	// Check packet count rollover.
	dsConn.packetCount = 255
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(0), data[0])
	assert.Equal(t, byte(255), data[1])
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(1), data[0])
	assert.Equal(t, byte(0), data[1])
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(1), data[0])
	assert.Equal(t, byte(1), data[1])
	dsConn.packetCount = 65535
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(255), data[0])
	assert.Equal(t, byte(255), data[1])
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(0), data[0])
	assert.Equal(t, byte(0), data[1])

	// Check different robot statuses.
	dsConn.Auto = true
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(2), data[3])

	dsConn.Enabled = true
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(6), data[3])

	dsConn.Auto = false
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(4), data[3])

	dsConn.EStop = true
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(132), data[3])

	dsConn.AStop = true
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(196), data[3])

	// Check different match types.
	arena.CurrentMatch.Type = model.Practice
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(1), data[6])
	arena.CurrentMatch.Type = model.Qualification
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(2), data[6])
	arena.CurrentMatch.Type = model.Playoff
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(3), data[6])

	// Check match numbers.
	arena.CurrentMatch.Type = model.Practice
	arena.CurrentMatch.TypeOrder = 42
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(0), data[7])
	assert.Equal(t, byte(42), data[8])
	arena.CurrentMatch.Type = model.Qualification
	arena.CurrentMatch.TypeOrder = 258
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(1), data[7])
	assert.Equal(t, byte(2), data[8])
	arena.CurrentMatch.Type = model.Playoff
	arena.CurrentMatch.TypeOrder = 13
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(0), data[7])
	assert.Equal(t, byte(13), data[8])

	// Check the countdown at different points during the match.
	arena.MatchState = AutoPeriod
	arena.MatchStartTime = time.Now().Add(-9 * time.Second)
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(11), data[21])
	arena.MatchState = PausePeriod
	arena.MatchStartTime = time.Now().Add(-21 * time.Second)
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(140), data[21])
	arena.MatchState = TeleopPeriod
	arena.MatchStartTime = time.Now().Add(-33 * time.Second)
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(129), data[21])
	arena.MatchStartTime = time.Now().Add(-160 * time.Second)
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(2), data[21])
	arena.MatchState = PostMatch
	arena.MatchStartTime = time.Now().Add(-180 * time.Second)
	data = dsConn.encodeControlPacket(arena, "")
	assert.Equal(t, byte(0), data[21])
}

func TestSendControlPacket(t *testing.T) {
	arena := setupTestArena(t)

	tcpConn := setupFakeTcpConnection(t)
	defer tcpConn.Close()
	dsConn, err := newDriverStationConnection(254, "R1", tcpConn, driverStationRoboRioUdpPort, false)
	assert.Nil(t, err)
	defer dsConn.close()

	// No real way of checking this since the destination IP is remote, so settle for there being no errors.
	err = dsConn.sendControlPacket(arena, "")
	assert.Nil(t, err)
}

func TestParseDsLogPacketUpdatesDsReportedStatus(t *testing.T) {
	dsConn := &DriverStationConnection{TeamId: 254}

	dsConn.parseDsLogPacket([]byte{0, 6, 22, 0, 0, 12, 128, 0x30})
	assert.True(t, dsConn.DsReportedStatusValid)
	assert.True(t, dsConn.DsReportedAuto)
	assert.True(t, dsConn.DsReportedTeleop)
	assert.False(t, dsConn.DsReportedDisabled)
	assert.True(t, dsConn.DsReportedEnabled)

	dsConn.parseDsLogPacket([]byte{0, 6, 22, 0, 0, 12, 128, 0x08})
	assert.True(t, dsConn.DsReportedStatusValid)
	assert.False(t, dsConn.DsReportedAuto)
	assert.False(t, dsConn.DsReportedTeleop)
	assert.True(t, dsConn.DsReportedDisabled)
	assert.False(t, dsConn.DsReportedEnabled)
}

func TestListenForDriverStations(t *testing.T) {
	arena := setupTestArena(t)
	serverAddress := startTestDriverStationServer(t, arena)

	// Connect with an invalid initial packet.
	tcpConn, err := net.Dial("tcp", serverAddress)
	if assert.Nil(t, err) {
		dataSend := [5]byte{0, 3, 29, 0, 0}
		tcpConn.Write(dataSend[:])
		var dataReceived [100]byte
		_, err = readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.NotNil(t, err)
		tcpConn.Close()
	}

	// Connect as a team not in the current match.
	tcpConn, err = net.Dial("tcp", serverAddress)
	if assert.Nil(t, err) {
		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:])
		var dataReceived [5]byte
		count, err := readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.Nil(t, err)
		assert.Equal(t, count, 5)
		assert.Equal(t, [5]byte{0, 3, 25, 0, 2}, dataReceived)
		tcpConn.Close()
	}

	// Connect as a team in the current match.
	arena.assignTeam(1503, "B2")

	// Connect as a team in the current match with a fragmented initial packet.
	tcpConn, err = net.Dial("tcp", serverAddress)
	if assert.Nil(t, err) {
		defer tcpConn.Close()
		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:1])
		tcpConn.Write(dataSend[1:5])
		var dataReceived [5]byte
		count, err := readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.Nil(t, err)
		assert.Equal(t, count, 5)
	}

	// Set event name
	arena.EventSettings.TbaEventCode = "2026CC"
	tcpConn, err = net.Dial("tcp", serverAddress)
	if assert.Nil(t, err) {
		defer tcpConn.Close()
		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:])
		var dataReceived [100]byte
		_, err := readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.Nil(t, err)
		// Read event name
		count, err := readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.Nil(t, err)
		assert.Equal(t, count, 6)
		assert.Equal(t, []byte{0, 4, 20, 2, 67, 67}, dataReceived[:6])
	}

	tcpConn, err = net.Dial("tcp", serverAddress)
	if assert.Nil(t, err) {
		defer tcpConn.Close()
		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:])
		var dataReceived [5]byte
		_, err = readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.Nil(t, err)
		assert.Equal(t, [5]byte{0, 3, 25, 4, 0}, dataReceived)

		dsConn := waitForDriverStationConnection(t, arena, "B2")
		if assert.NotNil(t, dsConn) {
			assert.Equal(t, 1503, dsConn.TeamId)
			assert.Equal(t, "B2", dsConn.AllianceStation)

			// Check that an unknown packet type gets ignored and a status packet gets decoded.
			dataSend = [5]byte{0, 3, 37, 0, 0}
			tcpConn.Write(dataSend[:])
		}
	}
}

func TestListenForDriverStations_NetworkSecurityIgnoresNonFieldIp(t *testing.T) {
	arena := setupTestArena(t)
	arena.EventSettings.NetworkSecurityEnabled = true
	arena.assignTeam(1503, "B2")
	serverAddress := startTestDriverStationServer(t, arena)

	tcpConn, err := net.Dial("tcp", serverAddress)
	if assert.Nil(t, err) {
		defer tcpConn.Close()

		dataSend := [5]byte{0, 3, 24, 5, 223}
		tcpConn.Write(dataSend[:])

		var dataReceived [5]byte
		_, err = readTaggedTcpPacket(tcpConn, dataReceived[:])
		assert.Nil(t, err)
		assert.Equal(t, [5]byte{0, 3, 25, 4, 0}, dataReceived)

		dsConn := waitForDriverStationConnection(t, arena, "B2")
		if assert.NotNil(t, dsConn) {
			assert.Equal(t, "", dsConn.WrongStation)
		}
	}
}

func TestNewDriverStationConnection_UdpPortSelection(t *testing.T) {
	tcpConn := setupFakeTcpConnection(t)
	defer tcpConn.Close()

	// Test with default settings (FMS port).
	dsConn, err := newDriverStationConnection(254, "R1", tcpConn, driverStationRoboRioUdpPort, false)
	assert.Nil(t, err)
	defer dsConn.close()
	assert.Contains(t, dsConn.udpAddrPort.String(), fmt.Sprintf(":%d", driverStationRoboRioUdpPort))

	tcpConnLite := setupFakeTcpConnection(t)
	defer tcpConnLite.Close()

	// Test with FMS Lite port enabled.
	dsConnLite, err := newDriverStationConnection(254, "R1", tcpConnLite, driverStationRoboRioUdpPortLite, false)
	assert.Nil(t, err)
	defer dsConnLite.close()
	assert.Contains(t, dsConnLite.udpAddrPort.String(), fmt.Sprintf(":%d", driverStationRoboRioUdpPortLite))
}

func TestNewDriverStationConnection_Ipv6Address(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	tcpConn := fakeRemoteAddrConn{
		Conn: clientConn,
		remoteAddr: &net.TCPAddr{
			IP:   net.ParseIP("::1"),
			Port: 1750,
		},
	}
	defer tcpConn.Close()

	dsConn, err := newDriverStationConnection(254, "R1", tcpConn, driverStationRoboRioUdpPort, false)
	assert.Nil(t, err)
	defer dsConn.close()
	assert.Equal(t, "::1", dsConn.udpAddrPort.Addr().String())
	assert.Equal(t, uint16(driverStationRoboRioUdpPort), dsConn.udpAddrPort.Port())
}

func setupFakeTcpConnection(t *testing.T) net.Conn {
	// Set up a fake TCP endpoint and connection to it.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(t, err)
	defer l.Close()
	tcpConn, err := net.Dial("tcp", l.Addr().String())
	assert.Nil(t, err)
	return tcpConn
}

type fakeRemoteAddrConn struct {
	net.Conn
	remoteAddr net.Addr
}

func (conn fakeRemoteAddrConn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}

func startTestDriverStationServer(t *testing.T, arena *Arena) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.Nil(t, err)
	t.Cleanup(
		func() {
			listener.Close()
		},
	)

	go arena.serveDriverStations(listener)
	return listener.Addr().String()
}

func waitForDriverStationConnection(t *testing.T, arena *Arena, station string) *DriverStationConnection {
	t.Helper()

	var dsConn *DriverStationConnection
	if !assert.Eventually(
		t,
		func() bool {
			dsConn = arena.AllianceStations[station].DsConn
			return dsConn != nil
		},
		time.Second,
		10*time.Millisecond,
	) {
		return nil
	}

	return dsConn
}
