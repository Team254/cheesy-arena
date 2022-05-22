// Copyright 2019 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model and datastore CRUD methods for a user login session.

package model

import "time"

type UserSession struct {
	Id        int `db:"id"`
	Token     string
	Username  string
	CreatedAt time.Time
}

func (database *Database) CreateUserSession(session *UserSession) error {
	return database.userSessionTable.create(session)
}

func (database *Database) GetUserSessionByToken(token string) (*UserSession, error) {
	userSessions, err := database.userSessionTable.getAll()
	if err != nil {
		return nil, err
	}

	for _, userSession := range userSessions {
		if userSession.Token == token {
			return &userSession, nil
		}
	}
	return nil, nil
}

func (database *Database) DeleteUserSession(id int) error {
	return database.userSessionTable.delete(id)
}

func (database *Database) TruncateUserSessions() error {
	return database.userSessionTable.truncate()
}
