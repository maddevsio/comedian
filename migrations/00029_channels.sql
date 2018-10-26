-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channels` MODIFY `channel_name` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channels` MODIFY `channel_name` VARCHAR(255) NOT NULL;