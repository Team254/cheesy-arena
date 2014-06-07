-- +goose Up
CREATE TABLE event_settings (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255),
  code VARCHAR(16),
  displaybackgroundcolor VARCHAR(16),
  numelimalliances int,
  selectionround1order VARCHAR(1),
  selectionround2order VARCHAR(1),
  selectionround3order VARCHAR(1)
);

-- +goose Down
DROP TABLE event_settings;
