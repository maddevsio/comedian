-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `controll_pannel` RENAME `bot_settings`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `bot_settings` RENAME `controll_pannel`;