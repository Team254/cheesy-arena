-- +goose Up
CREATE TABLE rankings (
  teamid INTEGER PRIMARY KEY,
  qualificationscore int,
  assistpoints int,
  autopoints int,
  trusscatchpoints int,
  goalfoulpoints int,
  random REAL,
  wins int,
  losses int,
  ties int,
  disqualifications int,
  played int
);

-- +goose Down
DROP TABLE rankings;
