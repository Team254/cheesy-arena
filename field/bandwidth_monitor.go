// Copyright 2015 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for monitoring team bandwidth usage across a managed switch.

package field

import (
	"fmt"
	"github.com/cdevr/WapSNMP"
	"log"
	"time"
)

const (
	monitoringIntervalMs = 1000
	toRobotBytesOid      = ".1.3.6.1.2.1.2.2.1.10"
	fromRobotBytesOid    = ".1.3.6.1.2.1.2.2.1.16"
	red1Port             = 2
	red2Port             = 4
	red3Port             = 6
	blue1Port            = 8
	blue2Port            = 10
	blue3Port            = 12
)

type BandwidthMonitor struct {
	allianceStations   *map[string]*AllianceStation
	snmpClient         *wapsnmp.WapSNMP
	toRobotOid         wapsnmp.Oid
	fromRobotOid       wapsnmp.Oid
	lastToRobotBytes   map[string]interface{}
	lastFromRobotBytes map[string]interface{}
	lastBytesTime      time.Time
}

// Loops indefinitely to query the managed switch via SNMP (Simple Network Management Protocol).
func (arena *Arena) monitorBandwidth() {
	monitor := BandwidthMonitor{allianceStations: &arena.AllianceStations,
		toRobotOid: wapsnmp.MustParseOid(toRobotBytesOid), fromRobotOid: wapsnmp.MustParseOid(fromRobotBytesOid)}
	for {
		if monitor.snmpClient != nil && monitor.snmpClient.Target != arena.EventSettings.SwitchAddress {
			// Switch address has changed; must re-create the SNMP client.
			monitor.snmpClient.Close()
			monitor.snmpClient = nil
		}

		if monitor.snmpClient == nil {
			var err error
			monitor.snmpClient, err = wapsnmp.NewWapSNMP(arena.EventSettings.SwitchAddress,
				arena.EventSettings.SwitchPassword, wapsnmp.SNMPv2c, 2*time.Second, 0)
			if err != nil {
				log.Printf("Error starting bandwidth monitoring: %v", err)
			}
		}

		if arena.EventSettings.NetworkSecurityEnabled && arena.EventSettings.BandwidthMonitoringEnabled {
			err := monitor.updateBandwidth()
			if err != nil {
				log.Printf("Bandwidth monitoring error: %v", err)
			}
		}
		time.Sleep(time.Millisecond * monitoringIntervalMs)
	}
}

func (monitor *BandwidthMonitor) updateBandwidth() error {
	// Retrieve total number of bytes sent/received per port.
	toRobotBytes, err := monitor.snmpClient.GetTable(monitor.toRobotOid)
	if err != nil {
		return err
	}
	fromRobotBytes, err := monitor.snmpClient.GetTable(monitor.fromRobotOid)
	if err != nil {
		return err
	}

	// Calculate the bandwidth usage over time.
	monitor.updateStationBandwidth("R1", red1Port, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("R2", red2Port, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("R3", red3Port, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("B1", blue1Port, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("B2", blue2Port, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("B3", blue3Port, toRobotBytes, fromRobotBytes)

	monitor.lastToRobotBytes = toRobotBytes
	monitor.lastFromRobotBytes = fromRobotBytes
	monitor.lastBytesTime = time.Now()
	return nil
}

func (monitor *BandwidthMonitor) updateStationBandwidth(station string, port int, toRobotBytes map[string]interface{},
	fromRobotBytes map[string]interface{}) {
	dsConn := (*monitor.allianceStations)[station].DsConn
	if dsConn == nil {
		// No team assigned; just skip it.
		return
	}
	secondsSinceLast := time.Now().Sub(monitor.lastBytesTime).Seconds()

	toRobotBytesForPort := uint32(toRobotBytes[fmt.Sprintf("%s.%d", toRobotBytesOid, port)].(wapsnmp.Counter))
	lastToRobotBytesForPort := uint32(monitor.lastToRobotBytes[fmt.Sprintf("%s.%d", toRobotBytesOid, port)].(wapsnmp.Counter))
	dsConn.MBpsToRobot = float64(toRobotBytesForPort-lastToRobotBytesForPort) / 1024 / 1024 / secondsSinceLast

	fromRobotBytesForPort := uint32(fromRobotBytes[fmt.Sprintf("%s.%d", fromRobotBytesOid, port)].(wapsnmp.Counter))
	lastFromRobotBytesForPort := uint32(monitor.lastFromRobotBytes[fmt.Sprintf("%s.%d", fromRobotBytesOid, port)].(wapsnmp.Counter))
	dsConn.MBpsFromRobot = float64(fromRobotBytesForPort-lastFromRobotBytesForPort) / 1024 / 1024 / secondsSinceLast
}
