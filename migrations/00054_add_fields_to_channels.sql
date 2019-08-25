-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channels` ADD `tz` VARCHAR(255) NOT NULL;
ALTER TABLE `channels` ADD `onbording_message` VARCHAR(255) NOT NULL;
ALTER TABLE `channels` ADD `submission_days` VARCHAR(255) NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channels` DROP `tz`;
ALTER TABLE `channels` DROP `onbording_message`;
ALTER TABLE `channels` DROP `submission_days`;