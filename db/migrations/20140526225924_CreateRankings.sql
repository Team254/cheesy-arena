-- +goose Up
CREATE TABLE rankings (
  teamid INTEGER PRIMARY KEY,
  rank int,
  rankingpoints int,
  autopoints int,
  scalechallengepoints int,
  goalpoints int,
  defensepoints int,
  random REAL,
  disqualifications int,
  played int
);

-- +goose Down
DROP TABLE rankings;
