-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `users` ADD `tz` VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT 'Asia/Bishkek';
ALTER TABLE `users` ADD `tz_offset` INTEGER NOT NULL DEFAULT 21600;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `users` DROP `tz`;
ALTER TABLE `users` DROP `tz_offset`;
