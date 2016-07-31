-- +goose Up
CREATE TABLE match_results (
  id INTEGER PRIMARY KEY,
  matchid int,
  playnumber int,
  matchtype VARCHAR(16),
  redscorejson text,
  bluescorejson text,
  redcardsjson text,
  bluecardsjson text
);
CREATE UNIQUE INDEX matchid_playnumber ON match_results(matchid, playnumber);

-- +goose Down
DROP TABLE match_results;
