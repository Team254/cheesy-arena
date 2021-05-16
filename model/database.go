// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Functions for manipulating the per-event Bolt datastore.

package model

import (
	"fmt"
	"github.com/Team254/cheesy-arena/game"
	"go.etcd.io/bbolt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const backupsDir = "db/backups"

var BaseDir = "." // Mutable for testing

type Database struct {
	Path               string
	bolt               *bbolt.DB
	allianceTeamTable  *table
	awardTable         *table
	eventSettingsTable *table
	lowerThirdTable    *table
	matchTable         *table
	matchResultTable   *table
	rankingTable       *table
	scheduleBlockTable *table
	sponsorSlideTable  *table
	teamTable          *table
	userSessionTable   *table
}

// Opens the Bolt database at the given path, creating it if it doesn't exist.
func OpenDatabase(filename string) (*Database, error) {
	database := Database{Path: filename}
	var err error
	database.bolt, err = bbolt.Open(database.Path, 0644, &bbolt.Options{NoSync: true, Timeout: time.Second})
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
	if database.rankingTable, err = database.newTable(game.Ranking{}); err != nil {
		return nil, err
	}
	if database.scheduleBlockTable, err = database.newTable(ScheduleBlock{}); err != nil {
		return nil, err
	}
	if database.sponsorSlideTable, err = database.newTable(SponsorSlide{}); err != nil {
		return nil, err
	}
	if database.teamTable, err = database.newTable(Team{}); err != nil {
		return nil, err
	}
	if database.userSessionTable, err = database.newTable(UserSession{}); err != nil {
		return nil, err
	}

	return &database, nil
}

func (database *Database) Close() error {
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

	dest, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dest.Close()

	if err = database.WriteBackup(dest); err != nil {
		return err
	}
	return nil
}

// Takes a snapshot of Bolt database and writes it to the given writer.
func (database *Database) WriteBackup(writer io.Writer) error {
	return database.bolt.View(func(tx *bbolt.Tx) error {
		_, err := tx.WriteTo(writer)
		return err
	})
}
