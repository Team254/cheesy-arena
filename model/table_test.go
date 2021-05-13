// Copyright 2021 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)

package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type validRecord struct {
	Id         int `db:"id"`
	IntData    int
	StringData string
}

type manualIdRecord struct {
	Id         int `db:"id,manual""`
	StringData string
}

func TestTableSingleCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	table, err := db.newTable(validRecord{})
	if !assert.Nil(t, err) {
		return
	}

	// Test initial create and then read back.
	record := validRecord{IntData: 254, StringData: "The Cheesy Poofs"}
	if assert.Nil(t, table.create(&record)) {
		assert.Equal(t, 1, record.Id)
	}
	var record2 *validRecord
	if assert.Nil(t, table.getById(record.Id, &record2)) {
		assert.Equal(t, record, *record2)
	}

	// Test update and then read back.
	record.IntData = 252
	record.StringData = "Teh Chezy Pofs"
	assert.Nil(t, table.update(&record))
	if assert.Nil(t, table.getById(record.Id, &record2)) {
		assert.Equal(t, record, *record2)
	}

	// Test delete.
	assert.Nil(t, table.delete(record.Id))
	if assert.Nil(t, table.getById(record.Id, &record2)) {
		assert.Nil(t, record2)
	}
}

func TestTableMultipleCrud(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	table, err := db.newTable(validRecord{})
	if !assert.Nil(t, err) {
		return
	}

	// Insert a few test records.
	record1 := validRecord{IntData: 1, StringData: "One"}
	record2 := validRecord{IntData: 2, StringData: "Two"}
	record3 := validRecord{IntData: 3, StringData: "Three"}
	assert.Nil(t, table.create(&record1))
	assert.Nil(t, table.create(&record2))
	assert.Nil(t, table.create(&record3))

	// Read all records.
	var records []validRecord
	assert.Nil(t, table.getAll(&records))
	if assert.Equal(t, 3, len(records)) {
		assert.Equal(t, record1, records[0])
		assert.Equal(t, record2, records[1])
		assert.Equal(t, record3, records[2])
	}

	// Truncate the table and verify that the records no longer exist.
	assert.Nil(t, table.truncate())
	assert.Nil(t, table.getAll(&records))
	assert.Equal(t, 0, len(records))
	var record4 *validRecord
	if assert.Nil(t, table.getById(record1.Id, &record4)) {
		assert.Nil(t, record4)
	}
}

func TestTableWithManualId(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	table, err := db.newTable(manualIdRecord{})
	if !assert.Nil(t, err) {
		return
	}

	// Test initial create and then read back.
	record := manualIdRecord{Id: 254, StringData: "The Cheesy Poofs"}
	if assert.Nil(t, table.create(&record)) {
		assert.Equal(t, 254, record.Id)
	}
	var record2 *manualIdRecord
	if assert.Nil(t, table.getById(record.Id, &record2)) {
		assert.Equal(t, record, *record2)
	}

	// Test update and then read back.
	record.StringData = "Teh Chezy Pofs"
	assert.Nil(t, table.update(&record))
	if assert.Nil(t, table.getById(record.Id, &record2)) {
		assert.Equal(t, record, *record2)
	}

	// Test delete.
	assert.Nil(t, table.delete(record.Id))
	if assert.Nil(t, table.getById(record.Id, &record2)) {
		assert.Nil(t, record2)
	}

	// Test creating a record with a zero ID.
	record.Id = 0
	err = table.create(&record)
	if assert.NotNil(t, err) {
		assert.Equal(
			t, "can't create manualIdRecord with zero ID since table is configured for manual IDs", err.Error(),
		)
	}
}

func TestNewTableErrors(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	// Pass a non-struct as the record type.
	table, err := db.newTable(123)
	assert.Nil(t, table)
	if assert.NotNil(t, err) {
		assert.Equal(t, "record type must be a struct; got int", err.Error())
	}

	// Pass a struct that doesn't have an ID field.
	type recordWithNoId struct {
		StringData string
	}
	table, err = db.newTable(recordWithNoId{})
	assert.Nil(t, table)
	if assert.NotNil(t, err) {
		assert.Equal(t, "struct recordWithNoId has no field tagged as the id", err.Error())
	}

	// Pass a struct that has a field with the wrong type tagged as the ID.
	type recordWithWrongIdType struct {
		Id bool `db:"id"`
	}
	table, err = db.newTable(recordWithWrongIdType{})
	assert.Nil(t, table)
	if assert.NotNil(t, err) {
		assert.Equal(
			t, "field in struct recordWithWrongIdType tagged with 'id' must be an int; got bool", err.Error(),
		)
	}
}

func TestTableCrudErrors(t *testing.T) {
	db := setupTestDb(t)
	defer db.Close()

	table, err := db.newTable(validRecord{})
	if !assert.Nil(t, err) {
		return
	}
	type differentRecord struct {
		StringData string
	}

	// Pass an object of the wrong type when getting a single record.
	var record validRecord
	err = table.getById(record.Id, record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "input must be a ptr; got a struct", err.Error())
	}
	err = table.getById(record.Id, &record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "input must be a ptr -> ptr; got a ptr -> struct", err.Error())
	}
	var recordTriplePointer ***validRecord
	err = table.getById(record.Id, recordTriplePointer)
	if assert.NotNil(t, err) {
		assert.Equal(t, "input must be a ptr -> ptr -> struct; got a ptr -> ptr -> ptr", err.Error())
	}
	var differentRecordPointer *differentRecord
	err = table.getById(record.Id, &differentRecordPointer)
	if assert.NotNil(t, err) {
		assert.Equal(
			t,
			"given record of type model.differentRecord does not match expected type for table validRecord",
			err.Error(),
		)
	}

	// Pass an object of the wrong type when getting all records.
	var records []validRecord
	err = table.getAll(records)
	if assert.NotNil(t, err) {
		assert.Equal(t, "input must be a ptr; got a slice", err.Error())
	}

	// Pass an object of the wrong type when creating or updating a record.
	err = table.create(record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "input must be a ptr; got a struct", err.Error())
	}
	err = table.update(record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "input must be a ptr; got a struct", err.Error())
	}

	// Create a record with a non-zero ID.
	record.Id = 12345
	err = table.create(&record)
	if assert.NotNil(t, err) {
		assert.Equal(
			t,
			"can't create validRecord with non-zero ID since table is configured for autogenerated IDs: 12345",
			err.Error(),
		)
	}

	// Update a record with an ID of zero.
	record.Id = 0
	err = table.update(&record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "can't update validRecord with zero ID", err.Error())
	}

	// Update a nonexistent record.
	record.Id = 12345
	err = table.update(&record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "can't update non-existent validRecord with ID 12345", err.Error())
	}

	// Delete a nonexistent record.
	err = table.delete(12345)
	if assert.NotNil(t, err) {
		assert.Equal(t, "can't delete non-existent validRecord with ID 12345", err.Error())
	}

	// Update a record with an incorrectly constructed table object.
	table.idFieldIndex = nil
	err = table.update(&record)
	if assert.NotNil(t, err) {
		assert.Equal(t, "struct validRecord has no field tagged as the id", err.Error())
	}
}
