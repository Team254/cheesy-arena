-- +goose Up
CREATE TABLE matches (
  id INTEGER PRIMARY KEY,
  type VARCHAR(16),
  displayname VARCHAR(16),
  time DATETIME,
  elimround int,
  elimgroup int,
  eliminstance int,
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
  winner VARCHAR(16),
  reddefense1 VARCHAR(3),
  reddefense2 VARCHAR(3),
  reddefense3 VARCHAR(3),
  reddefense4 VARCHAR(3),
  reddefense5 VARCHAR(3),
  bluedefense1 VARCHAR(3),
  bluedefense2 VARCHAR(3),
  bluedefense3 VARCHAR(3),
  bluedefense4 VARCHAR(3),
  bluedefense5 VARCHAR(3)
);
CREATE UNIQUE INDEX type_displayname ON matches(type, displayname);

-- +goose Down
DROP TABLE matches;
