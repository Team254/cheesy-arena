-- +goose Up
CREATE TABLE alliance_teams (
  id INTEGER PRIMARY KEY,
  allianceid int,
  pickposition int,
  teamid int
);
CREATE UNIQUE INDEX alliance_position ON alliance_teams(allianceid, pickposition);
CREATE UNIQUE INDEX team ON alliance_teams(teamid);

-- +goose Down
DROP TABLE alliance_teams;
