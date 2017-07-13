-- +goose Up
CREATE TABLE rankings (
  teamid INTEGER PRIMARY KEY,
  rank int,
  rankingpoints int,
  matchpoints int,
  autopoints int,
  rotorpoints int,
  takeoffpoints int,
  pressurepoints int,
  random REAL,
  wins int,
  losses int,
  ties int,
  disqualifications int,
  played int
);

-- +goose Down
DROP TABLE rankings;
