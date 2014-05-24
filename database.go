// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for manipulating the per-event SQLite datastore.

package main

import (
	"bitbucket.org/liamstask/goose/lib/goose"
	"database/sql"
	"github.com/jmoiron/modl"
	_ "github.com/mattn/go-sqlite3"
)

const migrationsDir = "db/migrations"

type Database struct {
	path    string
	db      *sql.DB
	teamMap *modl.DbMap
}

// Opens the SQLite database at the given path, creating it if it doesn't exist, and runs any pending
// migrations.
func OpenDatabase(path string) (*Database, error) {
	// Find and run the migrations using goose. This also auto-creates the DB.
	dbDriver := goose.DBDriver{"sqlite3", path, "github.com/mattn/go-sqlite3", &goose.Sqlite3Dialect{}}
	dbConf := goose.DBConf{migrationsDir, "prod", dbDriver}
	target, err := goose.GetMostRecentDBVersion(migrationsDir)
	if err != nil {
		return nil, err
	}
	err = goose.RunMigrations(&dbConf, migrationsDir, target)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	database := Database{path: path, db: db}
	database.mapTables()

	return &database, nil
}

func (database *Database) Close() {
	database.db.Close()
}

// Sets up table-object associations.
func (database *Database) mapTables() {
	dialect := new(modl.SqliteDialect)

	database.teamMap = modl.NewDbMap(database.db, dialect)
	database.teamMap.AddTableWithName(Team{}, "teams").SetKeys(false, "Id")
}
