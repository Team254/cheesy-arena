-- +goose Up
ALTER TABLE event_settings ADD COLUMN fmsapidownloadenabled BOOLEAN;
ALTER TABLE event_settings ADD COLUMN fmsapiusername VARCHAR(255);
ALTER TABLE event_settings ADD COLUMN fmsapiauthkey VARCHAR(255);
