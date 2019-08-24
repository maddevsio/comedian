-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `bot_settings` ADD `individual_reports_on` BOOLEAN NOT NULL;
ALTER TABLE `bot_settings` ADD `reporting_channel` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `bot_settings` ADD `reporting_time` VARCHAR(255) NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `bot_settings` DROP `individual_reports_on`;
ALTER TABLE `bot_settings` DROP `reporting_channel`;
ALTER TABLE `bot_settings` DROP `reporting_time`;