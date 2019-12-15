-- +goose Up
-- +goose StatementBegin
ALTER TABLE event_settings ADD networkconfigrt bool;
UPDATE event_settings SET networkconfigrt = 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event_settings DROP networkconfigrt;
-- +goose StatementEnd
