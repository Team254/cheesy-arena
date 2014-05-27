-- +goose Up
CREATE TABLE matches (
  id INTEGER PRIMARY KEY,
  type VARCHAR(16),
  displayname VARCHAR(16),
  time DATETIME,
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
  startedat DATETIME
);

-- +goose Down
DROP TABLE matches;
