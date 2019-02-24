-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `controll_pannel` ADD `password` VARCHAR(255) NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `controll_pannel` DROP `password`;
