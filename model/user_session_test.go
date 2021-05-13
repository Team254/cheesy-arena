// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetNonexistentUserSession(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	session, err := db.GetUserSessionByToken("blorpy")
	assert.Nil(t, err)
	assert.Nil(t, session)
}

func TestUserSessionCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	session := UserSession{0, "token1", "Bertha", time.Now()}
	err := db.CreateUserSession(&session)
	assert.Nil(t, err)
	session2, err := db.GetUserSessionByToken("token1")
	assert.Nil(t, err)
	assert.Equal(t, session.Token, session2.Token)
	assert.Equal(t, session.Username, session2.Username)
	assert.True(t, session.CreatedAt.Equal(session2.CreatedAt))

	db.DeleteUserSession(session.Id)
	session2, err = db.GetUserSessionByToken("token1")
	assert.Nil(t, err)
	assert.Nil(t, session2)
}

func TestTruncateUserSessions(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	session := UserSession{0, "token1", "Bertha", time.Now()}
	db.CreateUserSession(&session)
	db.TruncateUserSessions()
	session2, err := db.GetUserSessionByToken("token1")
	assert.Nil(t, err)
	assert.Nil(t, session2)
}
