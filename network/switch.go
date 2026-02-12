// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for Fortinet Switch Support

package network

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Team254/cheesy-arena/model"
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
		username:              "admin",
		password:              password,
		configBackoffDuration: switchConfigBackoffDurationSec * time.Second,
		configPauseDuration:   switchConfigPauseDurationSec * time.Second,
		Status:                "UNKNOWN",
	}
}

func (sw *Switch) ConfigureTeamEthernet(teams [6]*model.Team) error {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.Status = "CONFIGURING"

	// 1. 移除舊的 DHCP Server 設定
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

		addTeamVlansCommand += fmt.Sprintf(
			"config system interface\n"+
				"edit \"vlan%d\"\n"+
				"set vlanid %d\n"+
				"set interface \"internal\"\n"+
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
			vlan, vlan,
			teamPartialIp, switchTeamGatewayAddress,
			vlan, vlan, teamPartialIp, switchTeamGatewayAddress,
			teamPartialIp, teamPartialIp,
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
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", sw.address, sw.port), 5*time.Second)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// 處理 Telnet 協商的 Helper (解決封包 463 的問題)
	// 當收到 IAC (255) 時，根據協議簡單回覆，讓 Switch 願意說話
	handleNegotiation := func(c net.Conn) {
		buf := make([]byte, 3)
		for {
			c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, err := c.Read(buf)
			if err != nil || n < 3 || buf[0] != 255 {
				return
			}
			// 簡單邏輯：收到 DO (253) 回 WON'T (252)；收到 WILL (251) 回 DON'T (254)
			if buf[1] == 253 {
				c.Write([]byte{255, 252, buf[2]})
			} else if buf[1] == 251 {
				c.Write([]byte{255, 254, buf[2]})
			}
		}
	}

	// 1. 先處理握手
	handleNegotiation(conn)
	time.Sleep(500 * time.Millisecond)

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	send := func(s string, delay time.Duration) {
		writer.WriteString(s + "\r")
		writer.Flush()
		time.Sleep(delay)
	}

	// 2. 登入 (即使沒看到藍字也強行送出，但增加間隔)
	send(sw.username, 500*time.Millisecond)
	send(sw.password, 1*time.Second)

	// 3. 環境初始化
	send("config system console", 200*time.Millisecond)
	send("set output standard", 200*time.Millisecond)
	send("end", 500*time.Millisecond)

	// 4. 執行配置
	for _, line := range strings.Split(command, "\n") {
		if clean := strings.TrimSpace(line); clean != "" {
			send(clean, 150*time.Millisecond)
		}
	}
	send("exit", 200*time.Millisecond)

	var result bytes.Buffer
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	result.ReadFrom(reader)
	return result.String(), nil
}
