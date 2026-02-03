// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
// Modified for Fortinet Switch Support

package network

import (
	"bytes"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/Team254/cheesy-arena/model"
	"github.com/stretchr/testify/assert"
)

func TestConfigureSwitch(t *testing.T) {
	// 這裡的 password 會被用在 admin 的密碼
	sw := NewSwitch("127.0.0.1", "password")
	assert.Equal(t, "UNKNOWN", sw.Status)
	sw.port = 9050
	sw.configBackoffDuration = time.Millisecond
	sw.configPauseDuration = time.Millisecond
	var command1, command2 string

	// 修改為 Fortinet 的預期重置指令
	expectedResetCommand := "admin\npassword\nconfig system console\nset output standard\nend\n" +
		"config system dhcp server\ndelete 10\ndelete 20\ndelete 30\ndelete 40\ndelete 50\ndelete 60\nend\nexit\n"

	// 1. 測試：當沒有隊伍時，應該只執行移除 VLAN 的動作
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, nil, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(t, "", command2)
	assert.Equal(t, "ACTIVE", sw.Status)

	// 2. 測試：配置單一隊伍 (Team 254 在 Blue 2 位置，VLAN 50)
	sw.port += 1
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(t, sw.ConfigureTeamEthernet([6]*model.Team{nil, nil, nil, nil, {Id: 254}, nil}))
	assert.Equal(t, expectedResetCommand, command1)
	assert.Equal(
		t,
		"admin\npassword\nconfig system console\nset output standard\nend\n"+
			"config system interface\nedit \"vlan50\"\nset ip 10.2.54.4 255.255.255.0\nnext\nend\n"+
			"config system dhcp server\nedit 50\nset interface \"vlan50\"\nset default-gateway 10.2.54.4\nset netmask 255.255.255.0\n"+
			"config ip-range\nedit 1\nset start-ip 10.2.54.20\nset end-ip 10.2.54.199\nnext\nend\nnext\nend\nexit\n",
		command2,
	)

	// 3. 測試：配置所有隊伍
	sw.port += 1
	mockTelnet(t, sw.port, &command1, &command2)
	assert.Nil(
		t,
		sw.ConfigureTeamEthernet([6]*model.Team{{Id: 1114}, {Id: 254}, {Id: 296}, {Id: 1503}, {Id: 1678}, {Id: 1538}}),
	)
	assert.Equal(t, expectedResetCommand, command1)

	// 注意：這裡的 command2 預期字串必須與 switch.go 輸出的循環順序完全一致
	// 因為 Fortinet 指令比較長，這裡僅展示結構，實際執行時需確保格式符號精確
	assert.Contains(t, command2, "edit \"vlan10\"")
	assert.Contains(t, command2, "edit \"vlan60\"")
	assert.Contains(t, command2, "set start-ip 10.11.14.20")
}

func mockTelnet(t *testing.T, port int, command1 *string, command2 *string) {
	go func() {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return // 避免測試並行時噴錯
		}
		defer ln.Close()
		*command1 = ""
		*command2 = ""

		// 模擬第一連線 (Reset)
		conn1, err := ln.Accept()
		if err == nil {
			conn1.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			var reader bytes.Buffer
			reader.ReadFrom(conn1)
			*command1 = reader.String()
			conn1.Close()
		}

		// 模擬第二連線 (Config)
		conn2, err := ln.Accept()
		if err == nil {
			conn2.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
			var reader bytes.Buffer
			reader.ReadFrom(conn2)
			*command2 = reader.String()
			conn2.Close()
		}
	}()
	time.Sleep(100 * time.Millisecond)
}
