-- +goose Up
-- +goose StatementBegin
ALTER TABLE event_settings ADD durationauto int;
ALTER TABLE event_settings ADD durationteleop int;
UPDATE event_settings SET durationauto = 15, durationteleop = 135;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event_settings DROP durationauto;
ALTER TABLE event_settings DROP durationteleop;
-- +goose StatementEnd
