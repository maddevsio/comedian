-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channel_members` ADD `role_in_channel` VARCHAR(255) DEFAULT '';
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` DROP `role_in_channel`;