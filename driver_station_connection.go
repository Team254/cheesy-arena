// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and methods for interacting with a team's Driver Station.

package main

import (
	"fmt"
	"hash/crc32"
	"net"
	"strconv"
	"time"
)

// UDP port numbers that the Driver Station sends and receives on.
const driverStationSendPort = 1120
const driverStationReceivePort = 1160
const driverStationProtocolVersion = "11191100"

type DriverStationStatus struct {
	TeamId            int
	AllianceStation   string
	RobotLinked       bool
	Auto              bool
	Enabled           bool
	EmergencyStop     bool
	BatteryVoltage    float64
	DsVersion         string
	PacketCount       int
	MissedPacketCount int
	DsRobotTripTimeMs int
}

type DriverStationConnection struct {
	TeamId              int
	AllianceStation     string
	Auto                bool
	Enabled             bool
	EmergencyStop       bool
	DriverStationStatus *DriverStationStatus
	LastPacketTime      time.Time
	LastRobotLinkedTime time.Time
	conn                net.Conn
	packetCount         int
}

// Opens a UDP connection for communicating to the driver station.
func NewDriverStationConnection(teamId int, station string) (*DriverStationConnection, error) {
	conn, err := net.Dial("udp4", fmt.Sprintf("10.%d.%d.5:%d", teamId/100, teamId%100, driverStationSendPort))
	if err != nil {
		return nil, err
	}
	return &DriverStationConnection{TeamId: teamId, AllianceStation: station, conn: conn}, nil
}

// Builds and sends the next control packet to the Driver Station.
func (dsConn *DriverStationConnection) SendControlPacket() error {
	packet := dsConn.encodeControlPacket()
	_, err := dsConn.conn.Write(packet[:])
	if err != nil {
		return err
	}

	return nil
}

func (dsConn *DriverStationConnection) Close() error {
	return dsConn.conn.Close()
}

// Sets up a watch on the UDP port that Driver Stations send on.
func DsPacketListener() (*net.UDPConn, error) {
	udpAddress, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", driverStationReceivePort))
	if err != nil {
		return nil, err
	}
	listen, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		return nil, err
	}
	return listen, nil
}

// Loops indefinitely to read packets and update connection status.
func ListenForDsPackets(listener *net.UDPConn) {
	var data [50]byte
	for {
		listener.Read(data[:])
		dsStatus := decodeStatusPacket(data)

		// Update the status and last packet times for this alliance/team in the global struct.
		dsConn := arena.DriverStationConnections[dsStatus.AllianceStation]
		if dsConn != nil && dsConn.TeamId == dsStatus.TeamId {
			dsConn.DriverStationStatus = dsStatus
			dsConn.LastPacketTime = time.Now()
			if dsStatus.RobotLinked {
				dsConn.LastRobotLinkedTime = time.Now()
			}
		}
	}
}

// Serializes the control information into a packet.
func (dsConn *DriverStationConnection) encodeControlPacket() [74]byte {
	var packet [74]byte

	// Packet number, stored big-endian in two bytes.
	packet[0] = byte((dsConn.packetCount >> 8) & 0xff)
	packet[1] = byte(dsConn.packetCount & 0xff)

	// Robot status byte. 0x01=competition mode, 0x02=link, 0x04=check version, 0x08=request DS ID,
	// 0x10=autonomous, 0x20=enable, 0x40=e-stop not on
	packet[2] = 0x03
	if dsConn.Auto {
		packet[2] |= 0x10
	}
	if dsConn.Enabled {
		packet[2] |= 0x20
	}
	if !dsConn.EmergencyStop {
		packet[2] |= 0x40
	}

	// Alliance station, stored as ASCII characters 'R/B' and '1/2/3'.
	packet[3] = dsConn.AllianceStation[0]
	packet[4] = dsConn.AllianceStation[1]

	// Static protocol version repeated twice.
	for i := 0; i < 8; i++ {
		packet[10+i] = driverStationProtocolVersion[i]
	}
	for i := 0; i < 8; i++ {
		packet[18+i] = driverStationProtocolVersion[i]
	}

	// Calculate and store the 4-byte CRC32 checksum.
	checksum := crc32.ChecksumIEEE(packet[:])
	packet[70] = byte((checksum >> 24) & 0xff)
	packet[71] = byte((checksum >> 16) & 0xff)
	packet[72] = byte((checksum >> 8) & 0xff)
	packet[73] = byte((checksum) & 0xff)

	// Increment the packet count for next time.
	dsConn.packetCount++

	return packet
}

// Deserializes a packet from the DS into a structure representing the DS/robot status.
func decodeStatusPacket(data [50]byte) *DriverStationStatus {
	dsStatus := new(DriverStationStatus)

	// Robot status byte.
	dsStatus.RobotLinked = (data[2] & 0x02) != 0
	dsStatus.Auto = (data[2] & 0x10) != 0
	dsStatus.Enabled = (data[2] & 0x20) != 0
	dsStatus.EmergencyStop = (data[2] & 0x40) == 0

	// Team number, stored in two bytes as hundreds and then ones (like the IP address).
	dsStatus.TeamId = int(data[4])*100 + int(data[5])

	// Alliance station, stored as ASCII characters 'R/B' and '1/2/3'.
	dsStatus.AllianceStation = string(data[10:12])

	// Driver Station software version, stored as 8-byte string.
	dsStatus.DsVersion = string(data[18:26])

	// Number of missed packets sent from the DS to the robot, stored in two big-endian bytes.
	dsStatus.MissedPacketCount = int(data[26])*256 + int(data[27])

	// Total number of packets sent from the DS to the robot, stored in two big-endian bytes.
	dsStatus.PacketCount = int(data[28])*256 + int(data[29])

	// Average DS-robot trip time in milliseconds, stored in two big-endian bytes.
	dsStatus.DsRobotTripTimeMs = int(data[29])*256 + int(data[30])

	// Robot battery voltage, stored (bizarrely) what it looks like in decimal but as two hexadecimal numbers.
	dsStatus.BatteryVoltage, _ = strconv.ParseFloat(fmt.Sprintf("%x.%x", data[40], data[41]), 32)

	return dsStatus
}
