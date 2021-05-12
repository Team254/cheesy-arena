// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for manipulating the per-event SQLite datastore.

package model

import (
	"bitbucket.org/liamstask/goose/lib/goose"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/modl"
	_ "github.com/mattn/go-sqlite3"
	"go.etcd.io/bbolt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const backupsDir = "db/backups"
const migrationsDir = "db/migrations"

var BaseDir = "." // Mutable for testing

type Database struct {
	Path               string
	db                 *sql.DB
	rankingMap         *modl.DbMap
	teamMap            *modl.DbMap
	sponsorSlideMap    *modl.DbMap
	scheduleBlockMap   *modl.DbMap
	userSessionMap     *modl.DbMap
	bolt               *bbolt.DB
	allianceTeamTable  *table
	awardTable         *table
	eventSettingsTable *table
	lowerThirdTable    *table
	matchTable         *table
	matchResultTable   *table
}

// Opens the SQLite database at the given path, creating it if it doesn't exist, and runs any pending
// migrations.
func OpenDatabase(filename string) (*Database, error) {
	// Find and run the migrations using goose. This also auto-creates the DB.
	database := Database{Path: filename}
	migrationsPath := filepath.Join(BaseDir, migrationsDir)
	dbDriver := goose.DBDriver{"sqlite3", database.Path, "github.com/mattn/go-sqlite3", &goose.Sqlite3Dialect{}}
	dbConf := goose.DBConf{MigrationsDir: migrationsPath, Env: "prod", Driver: dbDriver}
	target, err := goose.GetMostRecentDBVersion(migrationsPath)
	if err != nil {
		return nil, err
	}
	err = goose.RunMigrations(&dbConf, migrationsPath, target)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", database.Path)
	if err != nil {
		return nil, err
	}
	database.db = db
	database.mapTables()

	database.bolt, err = bbolt.Open(database.Path+".bolt", 0644, &bbolt.Options{NoSync: true, Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	// Register tables.
	if database.allianceTeamTable, err = database.newTable(AllianceTeam{}); err != nil {
		return nil, err
	}
	if database.awardTable, err = database.newTable(Award{}); err != nil {
		return nil, err
	}
	if database.eventSettingsTable, err = database.newTable(EventSettings{}); err != nil {
		return nil, err
	}
	if database.lowerThirdTable, err = database.newTable(LowerThird{}); err != nil {
		return nil, err
	}
	if database.matchTable, err = database.newTable(Match{}); err != nil {
		return nil, err
	}
	if database.matchResultTable, err = database.newTable(MatchResult{}); err != nil {
		return nil, err
	}

	return &database, nil
}

func (database *Database) Close() error {
	database.db.Close()
	return database.bolt.Close()
}

// Creates a copy of the current database and saves it to the backups directory.
func (database *Database) Backup(eventName, reason string) error {
	backupsPath := filepath.Join(BaseDir, backupsDir)
	err := os.MkdirAll(backupsPath, 0755)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%s/%s_%s_%s.db", backupsPath, strings.Replace(eventName, " ", "_", -1),
		time.Now().Format("20060102150405"), reason)
	src, err := os.Open(database.Path)
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

	database.rankingMap = modl.NewDbMap(database.db, dialect)
	database.rankingMap.AddTableWithName(RankingDb{}, "rankings").SetKeys(false, "TeamId")

	database.teamMap = modl.NewDbMap(database.db, dialect)
	database.teamMap.AddTableWithName(Team{}, "teams").SetKeys(false, "Id")

	database.sponsorSlideMap = modl.NewDbMap(database.db, dialect)
	database.sponsorSlideMap.AddTableWithName(SponsorSlide{}, "sponsor_slides").SetKeys(true, "Id")

	database.scheduleBlockMap = modl.NewDbMap(database.db, dialect)
	database.scheduleBlockMap.AddTableWithName(ScheduleBlock{}, "schedule_blocks").SetKeys(true, "Id")

	database.userSessionMap = modl.NewDbMap(database.db, dialect)
	database.userSessionMap.AddTableWithName(UserSession{}, "user_sessions").SetKeys(true, "Id")
}

func serializeHelper(target *string, source interface{}) error {
	bytes, err := json.Marshal(source)
	if err != nil {
		return err
	}
	*target = string(bytes)
	return nil
}
