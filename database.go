// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for manipulating the per-event SQLite datastore.

package main

import (
	"bitbucket.org/liamstask/goose/lib/goose"
	"database/sql"
	"fmt"
	"github.com/jmoiron/modl"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"os"
	"strings"
	"time"
)

const backupsDir = "db/backups"
const migrationsDir = "db/migrations"

type Database struct {
	path             string
	db               *sql.DB
	eventSettingsMap *modl.DbMap
	matchMap         *modl.DbMap
	matchResultMap   *modl.DbMap
	rankingMap       *modl.DbMap
	teamMap          *modl.DbMap
	allianceTeamMap  *modl.DbMap
	lowerThirdMap    *modl.DbMap
	sponsorSlideMap  *modl.DbMap
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

// Creates a copy of the current database and saves it to the backups directory.
func (database *Database) Backup(reason string) error {
	err := os.MkdirAll(backupsDir, 0755)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%s/%s_%s_%s.db", backupsDir, strings.Replace(eventSettings.Name, " ", "_", -1),
		time.Now().Format("20060102150405"), reason)
	src, err := os.Open(database.path)
	if err != nil {
		return err
	}
	defer src.Close()
	dest, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, src); err != nil {
		return err
	}
	return nil
}

// Sets up table-object associations.
func (database *Database) mapTables() {
	dialect := new(modl.SqliteDialect)

	database.eventSettingsMap = modl.NewDbMap(database.db, dialect)
	database.eventSettingsMap.AddTableWithName(EventSettings{}, "event_settings").SetKeys(false, "Id")

	database.matchMap = modl.NewDbMap(database.db, dialect)
	database.matchMap.AddTableWithName(Match{}, "matches").SetKeys(true, "Id")

	database.matchResultMap = modl.NewDbMap(database.db, dialect)
	database.matchResultMap.AddTableWithName(MatchResultDb{}, "match_results").SetKeys(true, "Id")

	database.rankingMap = modl.NewDbMap(database.db, dialect)
	database.rankingMap.AddTableWithName(Ranking{}, "rankings").SetKeys(false, "TeamId")

	database.teamMap = modl.NewDbMap(database.db, dialect)
	database.teamMap.AddTableWithName(Team{}, "teams").SetKeys(false, "Id")

	database.allianceTeamMap = modl.NewDbMap(database.db, dialect)
	database.allianceTeamMap.AddTableWithName(AllianceTeam{}, "alliance_teams").SetKeys(true, "Id")

	database.lowerThirdMap = modl.NewDbMap(database.db, dialect)
	database.lowerThirdMap.AddTableWithName(LowerThird{}, "lower_thirds").SetKeys(true, "Id")

	database.sponsorSlideMap = modl.NewDbMap(database.db, dialect)
	database.sponsorSlideMap.AddTableWithName(SponsorSlide{}, "sponsor_slides").SetKeys(true, "Id")
}
