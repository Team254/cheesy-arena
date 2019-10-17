// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a user login session.

package model

import "time"

type UserSession struct {
	Id        int
	Token     string
	Username  string
	CreatedAt time.Time
}

func (database *Database) CreateUserSession(session *UserSession) error {
	return database.userSessionMap.Insert(session)
}

func (database *Database) GetUserSessionByToken(token string) (*UserSession, error) {
	session := new(UserSession)
	err := database.userSessionMap.SelectOne(session, "SELECT * FROM user_sessions WHERE token = ?", token)
	if err != nil && err.Error() == "sql: no rows in result set" {
		session = nil
		err = nil
	}
	return session, err
}

func (database *Database) DeleteUserSession(session *UserSession) error {
	_, err := database.userSessionMap.Delete(session)
	return err
}

func (database *Database) TruncateUserSessions() error {
	return database.userSessionMap.TruncateTables()
}
