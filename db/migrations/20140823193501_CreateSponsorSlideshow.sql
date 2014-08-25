-- +goose Up
CREATE TABLE sponsor_slideshow (
  id INTEGER PRIMARY KEY,
  subtitle VARCHAR(255),
  line1 VARCHAR(255),
  line2 VARCHAR(255),
  image VARCHAR(255),
  priority VARCHAR(255)
);

-- +goose Down
DROP TABLE sponsor_slideshow;
