-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `users` ADD `status` VARCHAR(255) NOT NULL COLLATE utf8mb4_unicode_ci;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `users` DROP `status`;
