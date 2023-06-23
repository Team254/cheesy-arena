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
	Path                string
	bolt                *bbolt.DB
	allianceTable       *table[Alliance]
	awardTable          *table[Award]
	eventSettingsTable  *table[EventSettings]
	lowerThirdTable     *table[LowerThird]
	matchTable          *table[Match]
	matchResultTable    *table[MatchResult]
	rankingTable        *table[game.Ranking]
	scheduleBlockTable  *table[ScheduleBlock]
	scheduledBreakTable *table[ScheduledBreak]
	sponsorSlideTable   *table[SponsorSlide]
	teamTable           *table[Team]
	userSessionTable    *table[UserSession]
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
	if database.allianceTable, err = newTable[Alliance](&database); err != nil {
		return nil, err
	}
	if database.awardTable, err = newTable[Award](&database); err != nil {
		return nil, err
	}
	if database.eventSettingsTable, err = newTable[EventSettings](&database); err != nil {
		return nil, err
	}
	if database.lowerThirdTable, err = newTable[LowerThird](&database); err != nil {
		return nil, err
	}
	if database.matchTable, err = newTable[Match](&database); err != nil {
		return nil, err
	}
	if database.matchResultTable, err = newTable[MatchResult](&database); err != nil {
		return nil, err
	}
	if database.rankingTable, err = newTable[game.Ranking](&database); err != nil {
		return nil, err
	}
	if database.scheduleBlockTable, err = newTable[ScheduleBlock](&database); err != nil {
		return nil, err
	}
	if database.scheduledBreakTable, err = newTable[ScheduledBreak](&database); err != nil {
		return nil, err
	}
	if database.sponsorSlideTable, err = newTable[SponsorSlide](&database); err != nil {
		return nil, err
	}
	if database.teamTable, err = newTable[Team](&database); err != nil {
		return nil, err
	}
	if database.userSessionTable, err = newTable[UserSession](&database); err != nil {
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
