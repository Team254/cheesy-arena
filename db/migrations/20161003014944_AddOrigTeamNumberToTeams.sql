-- +goose Up
ALTER TABLE teams ADD COLUMN origteamnumber INTEGER DEFAULT 0;

-- +goose Down
ALTER TABLE teams DROP COLUMN origteamnumber;
