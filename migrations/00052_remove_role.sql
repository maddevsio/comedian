-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `users` DROP `role`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `users` ADD `role` VARCHAR (255) NOT NULL;
