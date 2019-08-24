-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channel_members` ADD `real_name` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `channel_members` ADD `channel_name` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` DROP `real_name`;
ALTER TABLE `channel_members` DROP `channel_name`;