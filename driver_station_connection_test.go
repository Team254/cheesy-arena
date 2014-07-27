// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

func TestEncodeControlPacket(t *testing.T) {
	dsConn, err := NewDriverStationConnection(254, "R1")
	assert.Nil(t, err)
	defer dsConn.Close()

	data := dsConn.encodeControlPacket()
	assert.Equal(t, [74]byte{0, 0, 67, 82, 49, 0, 0, 0, 0, 0, 49, 49, 49, 57, 49, 49, 48, 48, 49, 49, 49, 57,
		49, 49, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 110, 235, 5, 29}, data)

	// Check the different alliance station values as well as the checksums.
	dsConn.AllianceStation = "R2"
	data = dsConn.encodeControlPacket()
	assert.Equal(t, [74]byte{0, 1, 67, 82, 50, 0, 0, 0, 0, 0, 49, 49, 49, 57, 49, 49, 48, 48, 49, 49, 49, 57,
		49, 49, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 114, 141, 17, 174}, data)
	dsConn.AllianceStation = "R3"
	data = dsConn.encodeControlPacket()
	assert.Equal(t, [74]byte{0, 2, 67, 82, 51, 0, 0, 0, 0, 0, 49, 49, 49, 57, 49, 49, 48, 48, 49, 49, 49, 57,
		49, 49, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 232, 206, 203, 150}, data)
	dsConn.AllianceStation = "B1"
	data = dsConn.encodeControlPacket()
	assert.Equal(t, [74]byte{0, 3, 67, 66, 49, 0, 0, 0, 0, 0, 49, 49, 49, 57, 49, 49, 48, 48, 49, 49, 49, 57,
		49, 49, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 57, 55, 68}, data)
	dsConn.AllianceStation = "B2"
	data = dsConn.encodeControlPacket()
	assert.Equal(t, [74]byte{0, 4, 67, 66, 50, 0, 0, 0, 0, 0, 49, 49, 49, 57, 49, 49, 48, 48, 49, 49, 49, 57,
		49, 49, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 34, 101, 225, 16}, data)
	dsConn.AllianceStation = "B3"
	data = dsConn.encodeControlPacket()
	assert.Equal(t, [74]byte{0, 5, 67, 66, 51, 0, 0, 0, 0, 0, 49, 49, 49, 57, 49, 49, 48, 48, 49, 49, 49, 57,
		49, 49, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 140, 207, 133, 117}, data)

	// Check packet count rollover.
	dsConn.packetCount = 255
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(0), data[0])
	assert.Equal(t, byte(255), data[1])
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(1), data[0])
	assert.Equal(t, byte(0), data[1])
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(1), data[0])
	assert.Equal(t, byte(1), data[1])
	dsConn.packetCount = 65535
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(255), data[0])
	assert.Equal(t, byte(255), data[1])
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(0), data[0])
	assert.Equal(t, byte(0), data[1])

	// Check different robot statuses.
	dsConn.Auto = true
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(83), data[2])

	dsConn.Enabled = true
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(115), data[2])

	dsConn.Auto = false
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(99), data[2])

	dsConn.EmergencyStop = true
	data = dsConn.encodeControlPacket()
	assert.Equal(t, byte(35), data[2])
}

func TestSendControlPacket(t *testing.T) {
	dsConn, err := NewDriverStationConnection(254, "R1")
	assert.Nil(t, err)
	defer dsConn.Close()

	// No real way of checking this since the destination IP is remote, so settle for there being no errors.
	err = dsConn.sendControlPacket()
	assert.Nil(t, err)
}

func TestDecodeStatusPacket(t *testing.T) {
	// Check with no linked robot.
	data := [50]byte{0, 0, 64, 1, 2, 54, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		98, 200, 63, 43, 0, 11, 0, 240, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus := decodeStatusPacket(data)
	assert.Equal(t, 254, dsStatus.TeamId)
	assert.Equal(t, "R1", dsStatus.AllianceStation)
	assert.Equal(t, false, dsStatus.RobotLinked)
	assert.Equal(t, false, dsStatus.Auto)
	assert.Equal(t, false, dsStatus.Enabled)
	assert.Equal(t, false, dsStatus.EmergencyStop)
	assert.Equal(t, 0, dsStatus.BatteryVoltage)
	assert.Equal(t, "02121300", dsStatus.DsVersion)
	assert.Equal(t, 16171, dsStatus.PacketCount)
	assert.Equal(t, 25288, dsStatus.MissedPacketCount)
	assert.Equal(t, 11, dsStatus.DsRobotTripTimeMs)

	// Check different team numbers.
	data = [50]byte{0, 0, 64, 1, 7, 66, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, 766, dsStatus.TeamId)
	data = [50]byte{0, 0, 64, 1, 51, 36, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, 5136, dsStatus.TeamId)

	// Check different alliance stations.
	data = [50]byte{0, 0, 64, 1, 51, 36, 0, 0, 0, 0, 66, 51, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, "B3", dsStatus.AllianceStation)

	// Check different robot statuses.
	data = [50]byte{0, 0, 66, 1, 7, 66, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, true, dsStatus.RobotLinked)
	data = [50]byte{0, 0, 98, 1, 7, 66, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, true, dsStatus.Enabled)
	data = [50]byte{0, 0, 114, 1, 7, 66, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, true, dsStatus.Auto)
	data = [50]byte{0, 0, 50, 1, 7, 66, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, true, dsStatus.EmergencyStop)

	// Check different battery voltages.
	data = [50]byte{0, 0, 64, 1, 7, 66, 0, 0, 0, 0, 82, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 255, 255, 255, 255, 82, 0, 0, 0, 0, 0, 25, 117, 0, 0, 0, 0, 42, 7, 189, 111}
	dsStatus = decodeStatusPacket(data)
	assert.Equal(t, 19.75, dsStatus.BatteryVoltage)
}

func TestListenForDsPackets(t *testing.T) {
	db, _ = OpenDatabase(testDbPath)

	listener, err := DsPacketListener()
	if assert.Nil(t, err) {
		go ListenForDsPackets(listener)
	}
	mainArena.Setup()

	dsConn, err := NewDriverStationConnection(254, "B1")
	defer dsConn.Close()
	assert.Nil(t, err)
	mainArena.AllianceStations["B1"].DsConn = dsConn
	dsConn, err = NewDriverStationConnection(1114, "R3")
	defer dsConn.Close()
	assert.Nil(t, err)
	mainArena.AllianceStations["R3"].DsConn = dsConn

	// Create a socket to send fake DS packets to localhost.
	conn, err := net.Dial("udp4", fmt.Sprintf("127.0.0.1:%d", driverStationReceivePort))
	assert.Nil(t, err)

	// Check receiving a packet from an expected team.
	packet := [50]byte{0, 0, 48, 1, 2, 54, 0, 0, 0, 0, 66, 49, 0, 0, 0, 0, 0, 0, 48, 50, 49, 50, 49, 51, 48, 48,
		152, 160, 152, 160, 1, 0, 255, 255, 82, 0, 0, 0, 0, 0, 25, 117, 0, 0, 0, 0, 42, 7, 189, 111}
	_, err = conn.Write(packet[:])
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 10) // Allow some time for the goroutine to process the incoming packet.
	dsStatus := mainArena.AllianceStations["B1"].DsConn.DriverStationStatus
	if assert.NotNil(t, dsStatus) {
		assert.Equal(t, 254, dsStatus.TeamId)
		assert.Equal(t, "B1", dsStatus.AllianceStation)
		assert.Equal(t, true, dsStatus.DsLinked)
		assert.Equal(t, false, dsStatus.RobotLinked)
		assert.Equal(t, true, dsStatus.Auto)
		assert.Equal(t, true, dsStatus.Enabled)
		assert.Equal(t, true, dsStatus.EmergencyStop)
		assert.Equal(t, 19.75, dsStatus.BatteryVoltage)
		assert.Equal(t, "02121300", dsStatus.DsVersion)
		assert.Equal(t, 39072, dsStatus.PacketCount)
		assert.Equal(t, 39072, dsStatus.MissedPacketCount)
		assert.Equal(t, 256, dsStatus.DsRobotTripTimeMs)
	}
	assert.True(t, time.Since(mainArena.AllianceStations["B1"].DsConn.LastPacketTime).Seconds() < 0.1)
	assert.True(t, time.Since(mainArena.AllianceStations["B1"].DsConn.LastRobotLinkedTime).Seconds() > 100)
	packet[2] = byte(98)
	_, err = conn.Write(packet[:])
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 10)
	dsStatus2 := mainArena.AllianceStations["B1"].DsConn.DriverStationStatus
	if assert.NotNil(t, dsStatus2) {
		assert.Equal(t, true, dsStatus2.RobotLinked)
		assert.Equal(t, false, dsStatus2.Auto)
		assert.Equal(t, true, dsStatus2.Enabled)
		assert.Equal(t, false, dsStatus2.EmergencyStop)
	}
	assert.True(t, time.Since(mainArena.AllianceStations["B1"].DsConn.LastPacketTime).Seconds() < 0.1)
	assert.True(t, time.Since(mainArena.AllianceStations["B1"].DsConn.LastRobotLinkedTime).Seconds() < 0.1)

	// Should ignore a packet coming from an expected team in the wrong position.
	statusBefore := mainArena.AllianceStations["R3"].DsConn.DriverStationStatus
	packet[10] = 'R'
	packet[11] = '3'
	packet[2] = 48
	_, err = conn.Write(packet[:])
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, statusBefore, mainArena.AllianceStations["R3"].DsConn.DriverStationStatus)
	assert.Equal(t, true, mainArena.AllianceStations["B1"].DsConn.DriverStationStatus.RobotLinked)

	// Should ignore a packet coming from an unexpected team.
	packet[4] = byte(15)
	packet[5] = byte(3)
	packet[10] = 'B'
	packet[11] = '1'
	packet[2] = 48
	_, err = conn.Write(packet[:])
	assert.Nil(t, err)
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, true, mainArena.AllianceStations["B1"].DsConn.DriverStationStatus.RobotLinked)

	// Should indicate that the connection has dropped if a response isn't received before the timeout.
	dsConn = mainArena.AllianceStations["B1"].DsConn
	dsConn.Update()
	assert.Equal(t, true, dsConn.DriverStationStatus.DsLinked)
	assert.Equal(t, true, dsConn.DriverStationStatus.RobotLinked)
	assert.NotEqual(t, 0, dsConn.DriverStationStatus.BatteryVoltage)
	dsConn.LastPacketTime = dsConn.LastPacketTime.Add(-1 * time.Second)
	dsConn.Update()
	assert.Equal(t, false, dsConn.DriverStationStatus.DsLinked)
	assert.Equal(t, false, dsConn.DriverStationStatus.RobotLinked)
	assert.Equal(t, 0, dsConn.DriverStationStatus.BatteryVoltage)
}
