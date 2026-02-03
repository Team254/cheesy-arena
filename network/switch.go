// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for Fortinet Switch Support

package network

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Team254/cheesy-arena/model"
	"net"
	"sync"
	"time"
)

const (
	switchConfigBackoffDurationSec = 5
	switchConfigPauseDurationSec   = 2
	switchTeamGatewayAddress       = 4
	switchTelnetPort               = 23
)

const (
	red1Vlan  = 10
	red2Vlan  = 20
	red3Vlan  = 30
	blue1Vlan = 40
	blue2Vlan = 50
	blue3Vlan = 60
)

type Switch struct {
	address               string
	port                  int
	username              string // Fortinet 通常需要帳號
	password              string
	mutex                 sync.Mutex
	configBackoffDuration time.Duration
	configPauseDuration   time.Duration
	Status                string
}

var ServerIpAddress = "10.0.100.5"

func NewSwitch(address, password string) *Switch {
	return &Switch{
		address:               address,
		port:                  switchTelnetPort,
		username:              "admin", // 預設為 admin
		password:              password,
		configBackoffDuration: switchConfigBackoffDurationSec * time.Second,
		configPauseDuration:   switchConfigPauseDurationSec * time.Second,
		Status:                "UNKNOWN",
	}
}

// ConfigureTeamEthernet 針對 Fortinet 語法進行了重構
func (sw *Switch) ConfigureTeamEthernet(teams [6]*model.Team) error {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.Status = "CONFIGURING"

	// 1. 移除舊的 DHCP Server 設定 (Fortinet 刪除邏輯)
	removeTeamVlansCommand := "config system dhcp server\n"
	for vlan := 10; vlan <= 60; vlan += 10 {
		removeTeamVlansCommand += fmt.Sprintf("delete %d\n", vlan)
	}
	removeTeamVlansCommand += "end\n"

	_, err := sw.runCommand(removeTeamVlansCommand)
	if err != nil {
		sw.Status = "ERROR"
		return err
	}
	time.Sleep(sw.configPauseDuration)

	// 2. 建立新設定
	addTeamVlansCommand := ""
	addTeamVlan := func(team *model.Team, vlan int) {
		if team == nil {
			return
		}
		teamPartialIp := fmt.Sprintf("%d.%d", team.Id/100, team.Id%100)
		
		// FortiOS 指令：先設 Interface IP，再設 DHCP Server
		addTeamVlansCommand += fmt.Sprintf(
			"config system interface\n"+
				"edit \"vlan%d\"\n"+
				"set ip 10.%s.%d 255.255.255.0\n"+
				"next\n"+
			"end\n"+
			"config system dhcp server\n"+
				"edit %d\n"+
				"set interface \"vlan%d\"\n"+
				"set default-gateway 10.%s.%d\n"+
				"set netmask 255.255.255.0\n"+
				"config ip-range\n"+
					"edit 1\n"+
					"set start-ip 10.%s.20\n"+
					"set end-ip 10.%s.199\n"+
					"next\n"+
				"end\n"+
				"next\n"+
			"end\n",
			vlan, teamPartialIp, switchTeamGatewayAddress, // Interface
			vlan, vlan, teamPartialIp, switchTeamGatewayAddress, // DHCP
			teamPartialIp, teamPartialIp, // IP Range
		)
	}

	addTeamVlan(teams[0], red1Vlan)
	addTeamVlan(teams[1], red2Vlan)
	addTeamVlan(teams[2], red3Vlan)
	addTeamVlan(teams[3], blue1Vlan)
	addTeamVlan(teams[4], blue2Vlan)
	addTeamVlan(teams[5], blue3Vlan)

	if len(addTeamVlansCommand) > 0 {
		_, err = sw.runCommand(addTeamVlansCommand)
		if err != nil {
			sw.Status = "ERROR"
			return err
		}
	}

	time.Sleep(sw.configBackoffDuration)
	sw.Status = "ACTIVE"
	return nil
}

// runCommand 處理 Fortinet 的登入與分頁關閉
func (sw *Switch) runCommand(command string) (string, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", sw.address, sw.port))
	if err != nil {
		return "", err
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	
	// Fortinet 登入流程：帳號 -> 密碼 -> 關閉分頁 -> 執行指令
	loginPayload := fmt.Sprintf("%s\n%s\n", sw.username, sw.password)
	disablePaging := "config system console\nset output standard\nend\n"
	
	_, err = writer.WriteString(loginPayload + disablePaging + command + "exit\n")
	if err != nil {
		return "", err
	}
	err = writer.Flush()
	if err != nil {
		return "", err
	}

	var reader bytes.Buffer
	_, err = reader.ReadFrom(conn)
	if err != nil {
		return "", err
	}
	return reader.String(), nil
}

// runConfigCommand 在 Fortinet 中與 runCommand 共用邏輯
func (sw *Switch) runConfigCommand(command string) (string, error) {
	return sw.runCommand(command)
}
