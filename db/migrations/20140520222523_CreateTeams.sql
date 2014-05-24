-- +goose Up
CREATE TABLE teams (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255),
  nickname VARCHAR(255),
  city VARCHAR(255),
  stateprov VARCHAR(255),
  country VARCHAR(255),
  rookieyear int,
  robotname VARCHAR(255)
);

-- +goose Down
DROP TABLE teams;
