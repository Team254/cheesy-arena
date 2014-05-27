-- +goose Up
CREATE TABLE match_results (
  id INTEGER PRIMARY KEY,
  matchid int,
  playnumber int,
  redscorejson text,
  bluescorejson text,
  redfoulsjson text,
  bluefoulsjson text,
  cardsjson text
);

-- +goose Down
DROP TABLE match_results;
