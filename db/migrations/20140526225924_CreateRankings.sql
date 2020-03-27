-- +goose Up
CREATE TABLE rankings (
  teamid INTEGER PRIMARY KEY,
  rank int,
  previousrank int,
  rankingfieldsjson text
);

-- +goose Down
DROP TABLE rankings;
