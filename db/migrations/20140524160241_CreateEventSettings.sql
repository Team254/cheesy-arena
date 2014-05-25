-- +goose Up
CREATE TABLE event_settings (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255),
  code VARCHAR(16)
);

-- +goose Down
DROP TABLE event_settings;
