// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and methods for interacting with a team's Driver Station.

package field

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
)

// FMS uses 1121 for sending UDP packets, and FMS Lite uses 1120. Using 1121
// seems to work just fine and doesn't prompt to let FMS take control.
const (
	driverStationTcpListenPort     = 1750
	driverStationUdpSendPort       = 1121
	driverStationUdpReceivePort    = 1160
	driverStationTcpLinkTimeoutSec = 5
	driverStationUdpLinkTimeoutSec = 1
	maxTcpPacketBytes              = 4096
)

type DriverStationConnection struct {
	TeamId                    int
	AllianceStation           string
	Auto                      bool
	Enabled                   bool
	Estop                     bool
	DsLinked                  bool
	RadioLinked               bool
	RioLinked                 bool
	RobotLinked               bool
	BatteryVoltage            float64
	DsRobotTripTimeMs         int
	MissedPacketCount         int
	SecondsSinceLastRobotLink float64
	lastPacketTime            time.Time
	lastRobotLinkedTime       time.Time
	packetCount               int
	missedPacketOffset        int
	tcpConn                   net.Conn
	udpConn                   net.Conn
	log                       *TeamMatchLog

	// WrongStation indicates if the team in the station is the incorrect team
	// by being non-empty. If the team is in the correct station, or no team is
	// connected, this is empty.
	WrongStation string
}

var allianceStationPositionMap = map[string]byte{"R1": 0, "R2": 1, "R3": 2, "B1": 3, "B2": 4, "B3": 5}

// Opens a UDP connection for communicating to the driver station.
func newDriverStationConnection(teamId int, allianceStation string, tcpConn net.Conn) (*DriverStationConnection, error) {
	ipAddress, _, err := net.SplitHostPort(tcpConn.RemoteAddr().String())
	if err != nil {
		return nil, err
	}
	log.Printf("Driver station for Team %d connected from %s\n", teamId, ipAddress)

	udpConn, err := net.Dial("udp4", fmt.Sprintf("%s:%d", ipAddress, driverStationUdpSendPort))
	if err != nil {
		return nil, err
	}
	return &DriverStationConnection{TeamId: teamId, AllianceStation: allianceStation, tcpConn: tcpConn, udpConn: udpConn}, nil
}

// Loops indefinitely to read packets and update connection status.
func (arena *Arena) listenForDsUdpPackets() {
	udpAddress, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", driverStationUdpReceivePort))
	listener, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		log.Fatalf("Error opening driver station UDP socket: %v", err)
	}
	log.Printf("Listening for driver stations on UDP port %d\n", driverStationUdpReceivePort)

	var data [50]byte
	for {
		listener.Read(data[:])

		teamId := int(data[4])<<8 + int(data[5])

		var dsConn *DriverStationConnection
		for _, allianceStation := range arena.AllianceStations {
			if allianceStation.Team != nil && allianceStation.Team.Id == teamId {
				dsConn = allianceStation.DsConn
				break
			}
		}

		if dsConn != nil {
			dsConn.DsLinked = true
			dsConn.lastPacketTime = time.Now()

			dsConn.RioLinked = data[3]&0x08 != 0
			dsConn.RadioLinked = data[3]&0x10 != 0
			dsConn.RobotLinked = data[3]&0x20 != 0
			if dsConn.RobotLinked {
				dsConn.lastRobotLinkedTime = time.Now()

				// Robot battery voltage, stored as volts * 256.
				dsConn.BatteryVoltage = float64(data[6]) + float64(data[7])/256
			}
		}
	}
}

// Sends a control packet to the Driver Station and checks for timeout conditions.
func (dsConn *DriverStationConnection) update(arena *Arena) error {
	err := dsConn.sendControlPacket(arena)
	if err != nil {
		return err
	}

	if time.Since(dsConn.lastPacketTime).Seconds() > driverStationUdpLinkTimeoutSec {
		dsConn.DsLinked = false
		dsConn.RadioLinked = false
		dsConn.RioLinked = false
		dsConn.RobotLinked = false
		dsConn.BatteryVoltage = 0
	}
	dsConn.SecondsSinceLastRobotLink = time.Since(dsConn.lastRobotLinkedTime).Seconds()

	return nil
}

func (dsConn *DriverStationConnection) close() {
	if dsConn.log != nil {
		dsConn.log.Close()
	}
	if dsConn.udpConn != nil {
		dsConn.udpConn.Close()
	}
	if dsConn.tcpConn != nil {
		dsConn.tcpConn.Close()
	}
}

// Called at the start of the match to allow for driver station initialization.
func (dsConn *DriverStationConnection) signalMatchStart(match *model.Match, wifiStatus *network.TeamWifiStatus) error {
	// Zero out missed packet count and begin logging.
	dsConn.missedPacketOffset = dsConn.MissedPacketCount
	var err error
	dsConn.log, err = NewTeamMatchLog(dsConn.TeamId, match, wifiStatus)
	return err
}

// Serializes the control information into a packet.
func (dsConn *DriverStationConnection) encodeControlPacket(arena *Arena) [22]byte {
	var packet [22]byte

	// Packet number, stored big-endian in two bytes.
	packet[0] = byte((dsConn.packetCount >> 8) & 0xff)
	packet[1] = byte(dsConn.packetCount & 0xff)

	// Protocol version.
	packet[2] = 0

	// Robot status byte.
	packet[3] = 0
	if dsConn.Auto {
		packet[3] |= 0x02
	}
	if dsConn.Enabled {
		packet[3] |= 0x04
	}
	if dsConn.Estop {
		packet[3] |= 0x80
	}

	// Unknown or unused.
	packet[4] = 0

	// Alliance station.
	packet[5] = allianceStationPositionMap[dsConn.AllianceStation]

	// Match type.
	match := arena.CurrentMatch
	switch match.Type {
	case model.Practice:
		packet[6] = 1
	case model.Qualification:
		packet[6] = 2
	case model.Playoff:
		packet[6] = 3
	default:
		packet[6] = 0
	}

	// Match number.
	packet[7] = byte(match.TypeOrder >> 8)
	packet[8] = byte(match.TypeOrder & 0xff)
	packet[9] = 1 // Match repeat number

	// Current time.
	currentTime := time.Now()
	packet[10] = byte(((currentTime.Nanosecond() / 1000) >> 24) & 0xff)
	packet[11] = byte(((currentTime.Nanosecond() / 1000) >> 16) & 0xff)
	packet[12] = byte(((currentTime.Nanosecond() / 1000) >> 8) & 0xff)
	packet[13] = byte((currentTime.Nanosecond() / 1000) & 0xff)
	packet[14] = byte(currentTime.Second())
	packet[15] = byte(currentTime.Minute())
	packet[16] = byte(currentTime.Hour())
	packet[17] = byte(currentTime.Day())
	packet[18] = byte(currentTime.Month())
	packet[19] = byte(currentTime.Year() - 1900)

	// Remaining number of seconds in match.
	var matchSecondsRemaining int
	switch arena.MatchState {
	case PreMatch, TimeoutActive, PostTimeout:
		matchSecondsRemaining = game.MatchTiming.AutoDurationSec
	case StartMatch, AutoPeriod:
		matchSecondsRemaining = game.MatchTiming.AutoDurationSec - int(arena.MatchTimeSec())
	case PausePeriod:
		matchSecondsRemaining = game.MatchTiming.TeleopDurationSec
	case TeleopPeriod:
		matchSecondsRemaining = game.MatchTiming.AutoDurationSec + game.MatchTiming.TeleopDurationSec +
			game.MatchTiming.PauseDurationSec - int(arena.MatchTimeSec())
	default:
		matchSecondsRemaining = 0
	}
	packet[20] = byte(matchSecondsRemaining >> 8 & 0xff)
	packet[21] = byte(matchSecondsRemaining & 0xff)

	// Increment the packet count for next time.
	dsConn.packetCount++

	return packet
}

// Builds and sends the next control packet to the Driver Station.
func (dsConn *DriverStationConnection) sendControlPacket(arena *Arena) error {
	packet := dsConn.encodeControlPacket(arena)
	if dsConn.udpConn != nil {
		_, err := dsConn.udpConn.Write(packet[:])
		if err != nil {
			return err
		}
	}

	return nil
}

// Deserializes a packet from the DS into a structure representing the DS/robot status.
func (dsConn *DriverStationConnection) decodeStatusPacket(data [36]byte) {
	// Average DS-robot trip time in milliseconds.
	dsConn.DsRobotTripTimeMs = int(data[1]) / 2

	// Number of missed packets sent from the DS to the robot.
	dsConn.MissedPacketCount = int(data[2]) - dsConn.missedPacketOffset

}

// Listens for TCP connection requests to Cheesy Arena from driver stations.
func (arena *Arena) listenForDriverStations() {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", network.ServerIpAddress, driverStationTcpListenPort))
	if err != nil {
		log.Printf("Error opening driver station TCP socket: %v", err.Error())
		log.Printf("Change IP address to %s and restart Cheesy Arena to fix.", network.ServerIpAddress)
		return
	}
	defer l.Close()

	log.Printf("Listening for driver stations on TCP port %d\n", driverStationTcpListenPort)
	for {
		tcpConn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting driver station connection: ", err.Error())
			continue
		}

		// Read the team number back and start tracking the driver station.
		var packet [5]byte
		_, err = tcpConn.Read(packet[:])
		if err != nil {
			log.Println("Error reading initial packet: ", err.Error())
			continue
		}
		if !(packet[0] == 0 && packet[1] == 3 && packet[2] == 24) {
			log.Printf("Invalid initial packet received: %v", packet)
			tcpConn.Close()
			continue
		}
		teamId := int(packet[3])<<8 + int(packet[4])

		// Check to see if the team is supposed to be on the field, and notify the DS accordingly.
		assignedStation := arena.getAssignedAllianceStation(teamId)
		if assignedStation == "" {
			log.Printf("Rejecting connection from Team %d, who is not in the current match, soon.", teamId)
			go func() {
				// Wait a second and then close it so it doesn't chew up bandwidth constantly trying to reconnect.
				time.Sleep(time.Second)
				tcpConn.Close()
			}()
			continue
		}

		// Read the team number from the IP address to check for a station mismatch.
		stationStatus := byte(0)
		teamRe := regexp.MustCompile("\\d+\\.(\\d+)\\.(\\d+)\\.")
		ipAddress, _, err := net.SplitHostPort(tcpConn.RemoteAddr().String())
		teamDigits := teamRe.FindStringSubmatch(ipAddress)
		teamDigit1, _ := strconv.Atoi(teamDigits[1])
		teamDigit2, _ := strconv.Atoi(teamDigits[2])
		stationTeamId := teamDigit1*100 + teamDigit2
		wrongAssignedStation := ""
		if stationTeamId != teamId {
			wrongAssignedStation = arena.getAssignedAllianceStation(stationTeamId)
			if wrongAssignedStation != "" {
				// The team is supposed to be in this match, but is plugged into the wrong station.
				log.Printf("Team %d is in incorrect station %s.", teamId, wrongAssignedStation)
				stationStatus = 1
			}
		}

		var assignmentPacket [5]byte
		assignmentPacket[0] = 0  // Packet size
		assignmentPacket[1] = 3  // Packet size
		assignmentPacket[2] = 25 // Packet type
		log.Printf("Accepting connection from Team %d in station %s.", teamId, assignedStation)
		assignmentPacket[3] = allianceStationPositionMap[assignedStation]
		assignmentPacket[4] = stationStatus
		_, err = tcpConn.Write(assignmentPacket[:])
		if err != nil {
			log.Printf("Error sending driver station assignment packet: %v", err)
			tcpConn.Close()
			continue
		}

		dsConn, err := newDriverStationConnection(teamId, assignedStation, tcpConn)
		if err != nil {
			log.Printf("Error registering driver station connection: %v", err)
			tcpConn.Close()
			continue
		}
		arena.AllianceStations[assignedStation].DsConn = dsConn

		if wrongAssignedStation != "" {
			dsConn.WrongStation = wrongAssignedStation
		}

		// Spin up a goroutine to handle further TCP communication with this driver station.
		go dsConn.handleTcpConnection(arena)
	}
}

func (dsConn *DriverStationConnection) handleTcpConnection(arena *Arena) {
	buffer := make([]byte, maxTcpPacketBytes)
	for {
		dsConn.tcpConn.SetReadDeadline(time.Now().Add(time.Second * driverStationTcpLinkTimeoutSec))
		_, err := dsConn.tcpConn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from connection for Team %d: %v", dsConn.TeamId, err)
			dsConn.close()
			arena.AllianceStations[dsConn.AllianceStation].DsConn = nil
			break
		}

		packetType := int(buffer[2])
		switch packetType {
		case 28:
			// DS keepalive packet; do nothing.
		case 22:
			// Robot status packet.
			var statusPacket [36]byte
			copy(statusPacket[:], buffer[2:38])
			dsConn.decodeStatusPacket(statusPacket)
		}

		// Log the packet if the match is in progress.
		matchTimeSec := arena.MatchTimeSec()
		if matchTimeSec > 0 && dsConn.log != nil {
			dsConn.log.LogDsPacket(matchTimeSec, packetType, dsConn)
		}
	}
}

// Sends a TCP packet containing the given game data to the driver station.
func (dsConn *DriverStationConnection) sendGameDataPacket(gameData string) error {
	byteData := []byte(gameData)
	size := len(byteData)
	packet := make([]byte, size+4)

	packet[0] = 0              // Packet size
	packet[1] = byte(size + 2) // Packet size
	packet[2] = 28             // Packet type
	packet[3] = byte(size)     // Data size

	// Fill the rest of the packet with the data.
	for i, character := range byteData {
		packet[i+4] = character
	}

	if dsConn.tcpConn != nil {
		_, err := dsConn.tcpConn.Write(packet)
		return err
	}
	return nil
}
