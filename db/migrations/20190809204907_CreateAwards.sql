-- +goose Up
CREATE TABLE awards (
  id INTEGER PRIMARY KEY,
  type int,
  awardname VARCHAR(255),
  teamid int,
  personname VARCHAR(255)
);

-- +goose Down
DROP TABLE awards;
