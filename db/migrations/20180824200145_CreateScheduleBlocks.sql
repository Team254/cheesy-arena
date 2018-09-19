-- +goose Up
CREATE TABLE schedule_blocks (
  id INTEGER PRIMARY KEY,
  matchtype VARCHAR(16),
  starttime DATETIME,
  nummatches int,
  matchspacingsec int
);

-- +goose Down
DROP TABLE schedule_blocks;
