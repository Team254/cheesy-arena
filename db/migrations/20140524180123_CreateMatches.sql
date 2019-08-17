-- +goose Up
CREATE TABLE matches (
  id INTEGER PRIMARY KEY,
  type VARCHAR(16),
  displayname VARCHAR(16),
  time DATETIME,
  elimround int,
  elimgroup int,
  eliminstance int,
  elimredalliance int,
  elimbluealliance int,
  red1 int,
  red1issurrogate bool,
  red2 int,
  red2issurrogate bool,
  red3 int,
  red3issurrogate bool,
  blue1 int,
  blue1issurrogate bool,
  blue2 int,
  blue2issurrogate bool,
  blue3 int,
  blue3issurrogate bool,
  status VARCHAR(16),
  startedat DATETIME,
  scorecommittedat DATETIME,
  winner VARCHAR(16)
);
CREATE UNIQUE INDEX type_displayname ON matches(type, displayname);

-- +goose Down
DROP TABLE matches;
