-- +goose Up
CREATE TABLE sponsor_slides (
  id INTEGER PRIMARY KEY,
  subtitle VARCHAR(255),
  line1 VARCHAR(255),
  line2 VARCHAR(255),
  image VARCHAR(255),
  displaytimesec int,
  displayorder int
);

-- +goose Down
DROP TABLE sponsor_slides;
