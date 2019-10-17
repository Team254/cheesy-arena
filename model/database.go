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
	Path             string
	db               *sql.DB
	eventSettingsMap *modl.DbMap
	matchMap         *modl.DbMap
	matchResultMap   *modl.DbMap
	rankingMap       *modl.DbMap
	teamMap          *modl.DbMap
	allianceTeamMap  *modl.DbMap
	lowerThirdMap    *modl.DbMap
	sponsorSlideMap  *modl.DbMap
	scheduleBlockMap *modl.DbMap
	awardMap         *modl.DbMap
	userSessionMap   *modl.DbMap
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

	return &database, nil
}

func (database *Database) Close() {
	database.db.Close()
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

	database.eventSettingsMap = modl.NewDbMap(database.db, dialect)
	database.eventSettingsMap.AddTableWithName(EventSettings{}, "event_settings").SetKeys(false, "Id")

	database.matchMap = modl.NewDbMap(database.db, dialect)
	database.matchMap.AddTableWithName(Match{}, "matches").SetKeys(true, "Id")

	database.matchResultMap = modl.NewDbMap(database.db, dialect)
	database.matchResultMap.AddTableWithName(MatchResultDb{}, "match_results").SetKeys(true, "Id")

	database.rankingMap = modl.NewDbMap(database.db, dialect)
	database.rankingMap.AddTableWithName(RankingDb{}, "rankings").SetKeys(false, "TeamId")

	database.teamMap = modl.NewDbMap(database.db, dialect)
	database.teamMap.AddTableWithName(Team{}, "teams").SetKeys(false, "Id")

	database.allianceTeamMap = modl.NewDbMap(database.db, dialect)
	database.allianceTeamMap.AddTableWithName(AllianceTeam{}, "alliance_teams").SetKeys(true, "Id")

	database.lowerThirdMap = modl.NewDbMap(database.db, dialect)
	database.lowerThirdMap.AddTableWithName(LowerThird{}, "lower_thirds").SetKeys(true, "Id")

	database.sponsorSlideMap = modl.NewDbMap(database.db, dialect)
	database.sponsorSlideMap.AddTableWithName(SponsorSlide{}, "sponsor_slides").SetKeys(true, "Id")

	database.scheduleBlockMap = modl.NewDbMap(database.db, dialect)
	database.scheduleBlockMap.AddTableWithName(ScheduleBlock{}, "schedule_blocks").SetKeys(true, "Id")

	database.awardMap = modl.NewDbMap(database.db, dialect)
	database.awardMap.AddTableWithName(Award{}, "awards").SetKeys(true, "Id")

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
