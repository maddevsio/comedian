-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `controll_pannel` ADD `individual_reporting_status` BOOLEAN NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `controll_pannel` DROP `individual_reporting_status`;
