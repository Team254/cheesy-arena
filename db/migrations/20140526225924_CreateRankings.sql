-- +goose Up
CREATE TABLE rankings (
  teamid INTEGER PRIMARY KEY,
  rank int,
  rankingfieldsjson text
);

-- +goose Down
DROP TABLE rankings;
