-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `bot_settings` DROP `password`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `bot_settings` ADD `password` VARCHAR(255) NOT NULL;
