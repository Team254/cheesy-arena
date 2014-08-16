-- +goose Up
CREATE TABLE lower_thirds (
  id INTEGER PRIMARY KEY,
  toptext VARCHAR(255),
  bottomtext VARCHAR(255)
);

-- +goose Down
DROP TABLE lower_thirds;
