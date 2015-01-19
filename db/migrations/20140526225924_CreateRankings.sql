-- +goose Up
CREATE TABLE rankings (
  teamid INTEGER PRIMARY KEY,
  rank int,
  qualificationaverage int,
  coopertitionpoints int,
  autopoints int,
  containerpoints int,
  totepoints int,
  litterpoints int,
  random REAL,
  disqualifications int,
  played int
);

-- +goose Down
DROP TABLE rankings;
