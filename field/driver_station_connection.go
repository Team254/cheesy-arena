// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and methods for interacting with a team's Driver Station.

package field

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/network"
)

// For the old NI DS, tags 24 and 25 are used for the initial connection.
// FMS uses 1121 for sending UDP packets, and FMS Lite uses 1120. Using 1121
// seems to work just fine and doesn't prompt to let FMS take control.
//
// For the new DS, tags 30 and 31 are used for the initial connection. FMS vs FMS Lite
// comes over that port, and the main difference is that the initial tag contains the
// UDP port for the FMS to reply to.
const (
	driverStationTcpListenPort      = 1750
	driverStationRoboRioUdpPort     = 1121
	driverStationRoboRioUdpPortLite = 1120
	driverStationUdpReceivePort     = 1160
	driverStationTcpLinkTimeoutSec  = 5
	driverStationUdpLinkTimeoutSec  = 1
	maxTcpPacketBytes               = 65537 // 2 for size, then 2^16-1 for data.
)

type DriverStationConnection struct {
	TeamId                    int
	AllianceStation           string
	Auto                      bool
	Enabled                   bool
	EStop                     bool
	AStop                     bool
	DsLinked                  bool
	RadioLinked               bool
	RioLinked                 bool
	RobotLinked               bool
	BatteryVoltage            float64
	DsRobotTripTimeMs         int
	MissedPacketCount         int
	DsReportedStatusValid     bool
	DsReportedAuto            bool
	DsReportedTeleop          bool
	DsReportedDisabled        bool
	DsReportedEnabled         bool
	SecondsSinceLastRobotLink float64
	lastPacketTime            time.Time
	lastRobotLinkedTime       time.Time
	packetCount               int
	tcpConn                   net.Conn
	udpSendPacket             [1500]byte
	SentGameData              string
	udpAddrPort               netip.AddrPort
	newDs                     bool

	// WrongStation indicates if the team in the station is the incorrect team
	// by being non-empty. If the team is in the correct station, or no team is
	// connected, this is empty.
	WrongStation string
}

var allianceStationPositionMap = map[string]byte{"R1": 0, "R2": 1, "R3": 2, "B1": 3, "B2": 4, "B3": 5}

func driverStationTeamIdFromRemoteAddr(addr net.Addr) (int, string, bool) {
	ipAddress, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return 0, "", false
	}

	// Driver stations use team-specific 10.TE.AM.X addresses on a field network.
	ipAddressBytes := net.ParseIP(ipAddress).To4()
	if ipAddressBytes == nil || ipAddressBytes[0] != 10 {
		return 0, ipAddress, false
	}

	return int(ipAddressBytes[1])*100 + int(ipAddressBytes[2]), ipAddress, true
}

// Creates a driver station object to represent a new inbound connection.
func newDriverStationConnection(
	teamId int,
	allianceStation string,
	tcpConn net.Conn,
	udpSendPort int,
	newDs bool,
) (*DriverStationConnection, error) {
	ipAddress, _, err := net.SplitHostPort(tcpConn.RemoteAddr().String())
	if err != nil {
		return nil, err
	}
	log.Printf("Driver station for Team %d connected from %s\n", teamId, ipAddress)

	udpAddr, err := netip.ParseAddr(ipAddress)
	if err != nil {
		return nil, err
	}

	return &DriverStationConnection{
		TeamId:          teamId,
		AllianceStation: allianceStation,
		tcpConn:         tcpConn,
		udpAddrPort:     netip.AddrPortFrom(udpAddr, uint16(udpSendPort)),
		newDs:           newDs,
	}, nil
}

func (arena *Arena) initializeUdpListener() {
	bindAddress := listenAddress(driverStationUdpReceivePort)
	udpAddress, err := net.ResolveUDPAddr("udp4", bindAddress)
	if err != nil {
		log.Fatalf(
			"Error resolving driver station UDP address: %v. Use the -dev flag to unrestrict server IP address for "+
				"development, or change IP address to %s.",
			err,
			network.ServerIpAddress,
		)
	}
	listener, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		log.Fatalf("Error opening driver station UDP socket: %v", err)
	}
	log.Printf("Listening for driver stations on UDP address %s\n", bindAddress)
	arena.DriverStationUdpSocket = listener
}

// Loops indefinitely to read packets and update connection status.
func (arena *Arena) listenForDsUdpPackets() {
	if arena.DriverStationUdpSocket == nil {
		return
	}

	listener := arena.DriverStationUdpSocket

	data := make([]byte, 1500)
	for {
		count, err := listener.Read(data[:])
		if err != nil {
			log.Printf("Error reading driver station UDP packet: %v", err)
			continue
		}
		if count < 8 {
			log.Printf("Received packet with insufficient length: %d", count)
			continue
		}

		teamId := int(data[4])<<8 + int(data[5])

		var dsConn *DriverStationConnection
		for _, allianceStation := range arena.AllianceStations {
			if allianceStation.Team != nil && allianceStation.Team.Id == teamId {
				dsConn = allianceStation.DsConn
				break
			}
		}

		if dsConn != nil {
			// Search through tags looking for tag 1
			index := 8
			for index < count {
				length := data[index]
				index++
				if length == 0 {
					continue
				}
				if index+int(length) > count {
					log.Printf("Unable to finish parsing UDP packet")
					break
				}
				tag := data[index]
				if tag == 1 && length == 6 {
					lost := (int(data[index+1]) << 8) + int(data[index+2])
					ping := int(data[index+5])
					dsConn.MissedPacketCount = lost
					dsConn.DsRobotTripTimeMs = ping
				}
				index += int(length)
			}

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
		} else {
			log.Printf("Failed to find DS for UDP packet with teamid %d", teamId)
		}
	}
}

// Sends a control packet to the Driver Station and checks for timeout conditions.
func (dsConn *DriverStationConnection) update(arena *Arena, gameData string) error {
	err := dsConn.sendControlPacket(arena, gameData)
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
	if dsConn.tcpConn != nil {
		if err := dsConn.tcpConn.Close(); err != nil {
			log.Printf("Error closing TCP connection for Team %d: %v", dsConn.TeamId, err)
		}
	}
}

// Serializes the control information into a packet.
func (dsConn *DriverStationConnection) encodeControlPacket(arena *Arena, gameData string) []byte {
	packet := dsConn.udpSendPacket
	packetLength := 22

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
	if dsConn.EStop {
		packet[3] |= 0x80
	}
	if dsConn.AStop {
		packet[3] |= 0x40
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
		matchSecondsRemaining = game.GetTeleopDurationSec()
	case TeleopPeriod:
		matchSecondsRemaining = game.MatchTiming.AutoDurationSec + game.GetTeleopDurationSec() +
			game.MatchTiming.PauseDurationSec - int(arena.MatchTimeSec())
	default:
		matchSecondsRemaining = 0
	}
	packet[20] = byte(matchSecondsRemaining >> 8 & 0xff)
	packet[21] = byte(matchSecondsRemaining & 0xff)

	// We need to include game data in the new ds packet
	if dsConn.newDs {
		gameDataLen := min(len(gameData), 8)
		if gameDataLen > 0 {
			packet[22] = byte(gameDataLen) + 1 // Length of the tag data, including the tag byte
			packet[23] = 32                    // Tag 32 is for game data
			for i := range gameDataLen {
				packet[24+i] = gameData[i]
			}
			packetLength += 2 + gameDataLen
		}
	}

	// Increment the packet count for next time.
	dsConn.packetCount++

	return packet[:packetLength]
}

// Builds and sends the next control packet to the Driver Station.
func (dsConn *DriverStationConnection) sendControlPacket(arena *Arena, gameData string) error {
	gameDataErr := dsConn.checkGameData(gameData)
	packet := dsConn.encodeControlPacket(arena, gameData)

	// Skip if UDP listener has not been started, or addr is invalid
	if arena.DriverStationUdpSocket == nil || !dsConn.udpAddrPort.IsValid() {
		return nil
	}

	_, err := arena.DriverStationUdpSocket.WriteToUDPAddrPort(packet[:], dsConn.udpAddrPort)
	if err != nil {
		log.Printf("Error sending control packet to Team %d: %v", dsConn.TeamId, err)
		return err
	}

	return gameDataErr
}

func listenAddress(port int) string {
	if network.DevMode {
		return fmt.Sprintf(":%d", port)
	}
	return fmt.Sprintf("%s:%d", network.ServerIpAddress, port)
}

// Listens for TCP connection requests to Cheesy Arena from driver stations.
func (arena *Arena) listenForDriverStations() {
	bindAddress := listenAddress(driverStationTcpListenPort)
	l, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatalf(
			"Error opening driver station TCP socket: %v. Use the -dev flag to unrestrict server IP address for "+
				"development, or change IP address to %s.",
			err,
			network.ServerIpAddress,
		)
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Printf("Error closing driver station TCP listener: %v", err)
		}
	}()

	log.Printf("Listening for driver stations on TCP address %s\n", bindAddress)
	arena.serveDriverStations(l)
}

func (arena *Arena) serveDriverStations(listener net.Listener) {
	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		log.Printf("Listening for driver stations on TCP port %d\n", tcpAddr.Port)
	} else {
		log.Printf("Listening for driver stations on TCP address %s\n", listener.Addr())
	}
	fullPacket := make([]byte, 1500)

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Println("Error accepting driver station connection: ", err.Error())
			continue
		}

		// Read the team number back and start tracking the driver station.
		count, err := readTaggedTcpPacket(tcpConn, fullPacket[:])
		if err != nil {
			log.Println("Error reading initial packet: ", err.Error())
			tcpConn.Close()
			continue
		}
		packet := fullPacket[:count]

		if len(packet) < 5 {
			log.Println("Invalid initial packet received: ", packet)
			tcpConn.Close()
			continue
		}

		udpSendPort := driverStationRoboRioUdpPort
		if arena.EventSettings.UseLiteUdpPort {
			udpSendPort = driverStationRoboRioUdpPortLite
		}

		isNewDs := false

		teamId := 0

		if packet[0] == 0 && packet[1] == 3 && packet[2] == 24 {
			log.Printf("Received NI DS Connection")
			teamId = int(packet[3])<<8 + int(packet[4])
		} else if packet[0] == 0 && packet[1] >= 5 && packet[2] == 30 {
			if len(packet) < 7 {
				log.Printf("Invalid initial packet of length %d received: %v", len(packet), packet)
				tcpConn.Close()
				continue
			}
			log.Printf("Received New DS Connection")
			isNewDs = true
			packenLen := int(packet[0])<<8 + int(packet[1])
			udpSendPort = int(packet[3])<<8 + int(packet[4])
			// Skip 5, its flags currently
			// Try to parse the team number in ASCII
			teamNumberLen := int(packet[6])
			if packenLen < 5+teamNumberLen || len(packet) < 7+teamNumberLen {
				log.Printf("Invalid initial packet of length %d received with team number length %d: %v", packenLen, teamNumberLen, packet)
				tcpConn.Close()
				continue
			}
			teamIdStr := string(packet[7 : 7+teamNumberLen])
			teamId, err = strconv.Atoi(teamIdStr)
			if err != nil {
				log.Printf("Error parsing team number from new DS connection: %v", err)
				go handleInvalidTcpConnection(tcpConn, 3, 0, isNewDs)
				continue
			} else if teamId < 0 || teamId > 65535 {
				log.Printf("Team number from new DS connection out of range: %d", teamId)
				go handleInvalidTcpConnection(tcpConn, 3, 0, isNewDs)
				continue
			}
		} else {
			log.Printf("Invalid initial packet received: %v", packet)
			closeTcpConn(tcpConn, "invalid initial packet")
			continue
		}

		// Check to see if the team is supposed to be on the field, and notify the DS accordingly.
		assignedStation := arena.getAssignedAllianceStation(teamId)
		if assignedStation == "" {
			log.Printf("Rejecting connection from Team %d, who is not in the current match, soon.", teamId)
			go handleInvalidTcpConnection(tcpConn, 2, 0, isNewDs)
			continue
		}

		// Read the team number from the IP address to check for a station mismatch.
		stationStatus := byte(0)
		wrongAssignedStation := ""
		if arena.EventSettings.NetworkSecurityEnabled {
			stationTeamId, ipAddress, ok := driverStationTeamIdFromRemoteAddr(tcpConn.RemoteAddr())
			if ok && stationTeamId != teamId {
				wrongAssignedStation = arena.getAssignedAllianceStation(stationTeamId)
				// The team is supposed to be in this match, but is plugged into the wrong station.
				if wrongAssignedStation != "" {
					log.Printf("Team %d is in incorrect station %s.", teamId, wrongAssignedStation)
					stationStatus = 1
				} else {
					log.Printf("Team %d is in unknown station with IP address %s.", teamId, ipAddress)
					stationStatus = 1
				}
			}
		}

		flags := 0
		if arena.EventSettings.UseLiteUdpPort {
			flags |= 0x01
		}

		sendLength := 8

		var assignmentPacket [8]byte
		assignmentPacket[0] = 0  // Packet size
		assignmentPacket[1] = 6  // Packet size
		assignmentPacket[2] = 31 // Packet type
		log.Printf("Accepting connection from Team %d in station %s with port %d", teamId, assignedStation, udpSendPort)
		assignmentPacket[3] = allianceStationPositionMap[assignedStation]
		assignmentPacket[4] = stationStatus
		assignmentPacket[5] = byte(flags)
		assignmentPacket[6] = byte(teamId >> 8)
		assignmentPacket[7] = byte(teamId & 0xFF)

		if !isNewDs {
			assignmentPacket[2] = 25 // Packet type
			assignmentPacket[1] = 3  // Packet size
			sendLength = 5
		}

		_, err = tcpConn.Write(assignmentPacket[:sendLength])
		if err != nil {
			log.Printf("Error sending driver station assignment packet: %v", err)
			closeTcpConn(tcpConn, "driver station assignment packet error")
			continue
		}

		// Write event code here. We need to strip any numbers off the front if it has it.
		// We also need to limit to 62 characters.
		eventName := arena.EventSettings.TbaEventCode
		if len(eventName) > 0 {
			trimIndex := 0
			for trimIndex < len(eventName) && eventName[trimIndex] >= '0' && eventName[trimIndex] <= '9' {
				trimIndex++
			}
			eventName = eventName[trimIndex:]
			if len(eventName) > 62 {
				eventName = eventName[:62]
			}
			if len(eventName) > 0 {
				eventNamePacket := make([]byte, 4+len(eventName))
				eventNamePacket[0] = 0
				eventNamePacket[1] = byte(len(eventName) + 2)
				eventNamePacket[2] = 20 // Packet type for event name
				eventNamePacket[3] = byte(len(eventName))
				copy(eventNamePacket[4:], []byte(eventName))
				_, err = tcpConn.Write(eventNamePacket)
				if err != nil {
					log.Printf("Error sending event name packet: %v", err)
					closeTcpConn(tcpConn, "event name packet error")
					continue
				}
			}
		}

		dsConn, err := newDriverStationConnection(teamId, assignedStation, tcpConn, udpSendPort, isNewDs)
		if err != nil {
			log.Printf("Error registering driver station connection: %v", err)
			closeTcpConn(tcpConn, "driver station registration error")
			continue
		}
		allianceStation := arena.AllianceStations[assignedStation]
		if previousDsConn := allianceStation.DsConn; previousDsConn != nil {
			dsConn.copyDsReportedStatus(previousDsConn)
			previousDsConn.close()
		}
		allianceStation.DsConn = dsConn

		if wrongAssignedStation != "" {
			dsConn.WrongStation = wrongAssignedStation
		}

		// Spin up a goroutine to handle further TCP communication with this driver station.
		go dsConn.handleTcpConnection(arena)
	}
}

func readTaggedTcpPacket(tcpConn net.Conn, buffer []byte) (int, error) {
	if len(buffer) < 2 {
		return 0, fmt.Errorf("buffer too small to read TCP packet")
	}

	if err := tcpConn.SetReadDeadline(time.Now().Add(time.Second * driverStationTcpLinkTimeoutSec)); err != nil {
		return 0, err
	}
	_, err := io.ReadFull(tcpConn, buffer[:2])
	if err != nil {
		return 0, err
	}

	packetLength := int(buffer[0])<<8 + int(buffer[1])

	if len(buffer) < 2+packetLength {
		return 0, fmt.Errorf("buffer too small to read full TCP packet")
	}

	_, err = io.ReadFull(tcpConn, buffer[2:2+packetLength])
	if err != nil {
		return 0, err
	}

	return 2 + packetLength, nil
}

func (dsConn *DriverStationConnection) handleTcpConnection(arena *Arena) {
	buffer := make([]byte, maxTcpPacketBytes)
	for {
		count, err := readTaggedTcpPacket(dsConn.tcpConn, buffer)
		if err != nil {
			log.Printf("Error reading from connection for Team %d: %v", dsConn.TeamId, err)
			dsConn.close()
			if arena.AllianceStations[dsConn.AllianceStation].DsConn == dsConn {
				arena.AllianceStations[dsConn.AllianceStation].DsConn = nil
			}
			break
		}

		packetType := int(buffer[2])
		switch packetType {
		case 29:
			// DS keepalive packet; do nothing.
			continue
		case 22:
			dsConn.parseDsLogPacket(buffer[:count])
		default:
			log.Printf("Received unknown packet type %d from Team %d", packetType, dsConn.TeamId)
		}
	}
}

// copyDsReportedStatus preserves the last DS-reported mode bits when the same team reconnects mid-match.
func (dsConn *DriverStationConnection) copyDsReportedStatus(previousDsConn *DriverStationConnection) {
	dsConn.DsReportedStatusValid = previousDsConn.DsReportedStatusValid
	dsConn.DsReportedAuto = previousDsConn.DsReportedAuto
	dsConn.DsReportedTeleop = previousDsConn.DsReportedTeleop
	dsConn.DsReportedDisabled = previousDsConn.DsReportedDisabled
	dsConn.DsReportedEnabled = previousDsConn.DsReportedEnabled
}

// parseDsLogPacket updates DS-reported mode and enable state from a driver station TCP log packet.
func (dsConn *DriverStationConnection) parseDsLogPacket(packet []byte) {
	if len(packet) < 8 {
		log.Printf("Received DS log packet with insufficient length from Team %d: %d", dsConn.TeamId, len(packet))
		return
	}

	// Packet type 22 carries the DS-side robot status byte at offset 7.
	statusByte := packet[7]
	dsConn.DsReportedStatusValid = true
	dsConn.DsReportedTeleop = statusByte&0x20 != 0
	dsConn.DsReportedAuto = statusByte&0x10 != 0
	dsConn.DsReportedDisabled = statusByte&0x08 != 0
	dsConn.DsReportedEnabled = !dsConn.DsReportedDisabled
}

func handleInvalidTcpConnection(tcpConn net.Conn, status int, station int, isNewDs bool) {
	log.Printf("Handling invalid TCP connection from %v with status %d and station %d", tcpConn.RemoteAddr(), status, station)
	var assignmentPacket [8]byte
	sendLength := 8

	assignmentPacket[0] = 0  // Packet size
	assignmentPacket[1] = 6  // Packet size
	assignmentPacket[2] = 31 // Packet type
	assignmentPacket[3] = byte(station)
	assignmentPacket[4] = byte(status)
	assignmentPacket[5] = 0
	assignmentPacket[6] = 0
	assignmentPacket[7] = 0
	if !isNewDs {
		assignmentPacket[2] = 25 // Packet type
		assignmentPacket[1] = 3  // Packet size
		sendLength = 5
	}
	_, err := tcpConn.Write(assignmentPacket[:sendLength])
	if err != nil {
		log.Printf("Error sending invalid driver station assignment packet: %v", err)
		closeTcpConn(tcpConn, "invalid driver station assignment packet error")
		return
	}

	buffer := make([]byte, maxTcpPacketBytes)
	for {
		_, err := readTaggedTcpPacket(tcpConn, buffer)
		if err != nil {
			log.Printf("Error reading from connection for invalid driver station: %v", err)
			break
		}
	}

	closeTcpConn(tcpConn, "invalid driver station connection")
}

func closeTcpConn(tcpConn net.Conn, context string) {
	if err := tcpConn.Close(); err != nil {
		log.Printf("Error closing TCP connection after %s: %v", context, err)
	}
}

func (dsConn *DriverStationConnection) checkGameData(gameData string) error {
	if dsConn.newDs {
		return nil
	}

	needsGameDataUpdate := dsConn.SentGameData != gameData
	if needsGameDataUpdate {
		err := dsConn.sendGameDataPacketTcp(gameData)
		if err != nil {
			log.Printf("Error sending game data packet to Team %d: %v", dsConn.TeamId, err)
			return err
		} else {
			dsConn.SentGameData = gameData
		}
	}
	return nil
}

// Sends a TCP packet containing the given game data to the driver station.
func (dsConn *DriverStationConnection) sendGameDataPacketTcp(gameData string) error {
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
