// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigureCatalyst(t *testing.T) {
	catalystTelnetPort = 9050
	eventSettings = &EventSettings{SwitchAddress: "127.0.0.1", SwitchPassword: "password"}
	var command string

	// Should do nothing if current configuration is blank.
	mockTelnet(t, catalystTelnetPort, "", &command)
	assert.Nil(t, ConfigureTeamEthernet(nil, nil, nil, nil, nil, nil))
	assert.Equal(t, "", command)

	// Should remove any existing teams but not other SSIDs.
	catalystTelnetPort += 1
	mockTelnet(t, catalystTelnetPort,
		"interface Vlan2\nip address 10.0.100.2\ninterface Vlan15\nip address 10.2.54.61\n", &command)
	assert.Nil(t, ConfigureTeamEthernet(nil, nil, nil, nil, nil, nil))
	assert.Equal(t, "password\nenable\npassword\nterminal length 0\nconfig terminal\ninterface Vlan15\nno ip"+
		" address\nno access-list 115\nend\ncopy running-config startup-config\n\nexit\n", command)

	// Should configure new teams and leave existing ones alone if still needed.
	catalystTelnetPort += 1
	mockTelnet(t, catalystTelnetPort, "interface Vlan15\nip address 10.2.54.61\n", &command)
	assert.Nil(t, ConfigureTeamEthernet(nil, &Team{Id: 1114}, nil, nil, &Team{Id: 254}, nil))
	assert.Equal(t, "password\nenable\npassword\nterminal length 0\nconfig terminal\nno access-list 112\n"+
		"access-list 112 permit ip 10.11.14.0 0.0.0.255 host 10.0.100.5\ninterface Vlan12\nip address "+
		"10.11.14.61 255.255.255.0\nend\ncopy running-config startup-config\n\nexit\n", command)
}
