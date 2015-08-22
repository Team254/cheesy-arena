-- +goose Up
ALTER TABLE event_settings ADD COLUMN tbadownloadenabled BOOLEAN;
ALTER TABLE event_settings ADD COLUMN tbaawardsdownloadenabled BOOLEAN;