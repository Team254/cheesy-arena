-- +goose Up
CREATE TABLE teams (
  id INTEGER PRIMARY KEY,
  name VARCHAR(1000),
  nickname VARCHAR(255),
  city VARCHAR(255),
  stateprov VARCHAR(255),
  country VARCHAR(255),
  rookieyear int,
  robotname VARCHAR(255),
  accomplishments VARCHAR(1000),
  wpakey VARCHAR(16),
  yellowcard bool,
  hasconnected bool
);

-- +goose Down
DROP TABLE teams;
