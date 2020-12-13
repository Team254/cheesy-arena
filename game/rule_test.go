// Copyright 2020 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package game

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetRuleById(t *testing.T) {
	assert.Nil(t, GetRuleById(0))
	assert.Equal(t, rules[0], GetRuleById(1))
	assert.Equal(t, rules[20], GetRuleById(21))
	assert.Nil(t, GetRuleById(1000))
}

func TestGetAllRules(t *testing.T) {
	allRules := GetAllRules()
	assert.Equal(t, len(rules), len(allRules))
	for _, rule := range rules {
		assert.Equal(t, rule, allRules[rule.Id])
	}
}
