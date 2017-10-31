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
	red1Port             = 6
	red2Port             = 8
	red3Port             = 10
	blue1Port            = 12
	blue2Port            = 14
	blue3Port            = 16
)

type BandwidthMonitor struct {
	allianceStations   *map[string]*AllianceStation
	snmpClient         *wapsnmp.WapSNMP
	toRobotOids        []wapsnmp.Oid
	fromRobotOids      []wapsnmp.Oid
	lastToRobotBytes   map[string]interface{}
	lastFromRobotBytes map[string]interface{}
	lastBytesTime      time.Time
}

// Loops indefinitely to query the managed switch via SNMP (Simple Network Management Protocol).
func (arena *Arena) monitorBandwidth() {
	monitor := BandwidthMonitor{allianceStations: &arena.AllianceStations}

	for _, port := range []int{red1Port, red2Port, red3Port, blue1Port, blue2Port, blue3Port} {
		toOid := fmt.Sprintf("%s.%d", toRobotBytesOid, 10100+port)
		fromOid := fmt.Sprintf("%s.%d", fromRobotBytesOid, 10100+port)
		monitor.toRobotOids = append(monitor.toRobotOids, wapsnmp.MustParseOid(toOid))
		monitor.fromRobotOids = append(monitor.fromRobotOids, wapsnmp.MustParseOid(fromOid))
	}

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
				continue
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
	toRobotBytes, err := monitor.snmpClient.GetMultiple(monitor.toRobotOids)
	if err != nil {
		return err
	}
	fromRobotBytes, err := monitor.snmpClient.GetMultiple(monitor.fromRobotOids)
	if err != nil {
		return err
	}

	// Calculate the bandwidth usage over time.
	monitor.updateStationBandwidth("R1", 0, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("R2", 1, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("R3", 2, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("B1", 3, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("B2", 4, toRobotBytes, fromRobotBytes)
	monitor.updateStationBandwidth("B3", 5, toRobotBytes, fromRobotBytes)

	monitor.lastToRobotBytes = toRobotBytes
	monitor.lastFromRobotBytes = fromRobotBytes
	monitor.lastBytesTime = time.Now()
	return nil
}

func (monitor *BandwidthMonitor) updateStationBandwidth(station string, oidIndex int, toRobotBytes map[string]interface{},
	fromRobotBytes map[string]interface{}) {
	dsConn := (*monitor.allianceStations)[station].DsConn
	if dsConn == nil {
		// No team assigned; just skip it.
		return
	}
	secondsSinceLast := time.Now().Sub(monitor.lastBytesTime).Seconds()

	toOid := monitor.toRobotOids[oidIndex].String()
	if _, ok := toRobotBytes[toOid]; !ok {
		log.Printf("Error: OID %s not present in new to-robot stats %v.", toOid, toRobotBytes)
		return
	}
	toRobotBytesForPort := uint32(toRobotBytes[toOid].(wapsnmp.Counter))
	if _, ok := monitor.lastToRobotBytes[toOid]; !ok {
		// This may be the first time reading.
		return
	}
	lastToRobotBytesForPort := uint32(monitor.lastToRobotBytes[toOid].(wapsnmp.Counter))
	dsConn.MBpsToRobot = float64(toRobotBytesForPort-lastToRobotBytesForPort) / 1024 / 128 / secondsSinceLast

	fromOid := monitor.fromRobotOids[oidIndex].String()
	if _, ok := fromRobotBytes[fromOid]; !ok {
		log.Printf("Error: OID %s not present in new from-robot stats %v.", fromOid, fromRobotBytes)
		return
	}
	fromRobotBytesForPort := uint32(fromRobotBytes[fromOid].(wapsnmp.Counter))
	if _, ok := monitor.lastFromRobotBytes[fromOid]; !ok {
		// This may be the first time reading.
		return
	}
	lastFromRobotBytesForPort := uint32(monitor.lastFromRobotBytes[fromOid].(wapsnmp.Counter))
	dsConn.MBpsFromRobot = float64(fromRobotBytesForPort-lastFromRobotBytesForPort) / 1024 / 128 / secondsSinceLast
}
