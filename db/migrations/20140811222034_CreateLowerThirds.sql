-- +goose Up
CREATE TABLE lower_thirds (
  id INTEGER PRIMARY KEY,
  toptext VARCHAR(255),
  bottomtext VARCHAR(255),
  displayorder int,
  awardid int
);

-- +goose Down
DROP TABLE lower_thirds;
