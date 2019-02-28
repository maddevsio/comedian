-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channels` ADD `team_id` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `controll_pannel` ADD `team_id` VARCHAR(255) NOT NULL;
ALTER TABLE `standups` ADD `team_id` VARCHAR(255) NOT NULL;
ALTER TABLE `users` ADD `team_id` VARCHAR(255) NOT NULL;
ALTER TABLE `channel_members` ADD `team_id` VARCHAR(255) NOT NULL;
ALTER TABLE `controll_pannel` ADD `team_name` VARCHAR(255) NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channels` DROP `team_id`;
ALTER TABLE `controll_pannel` DROP `team_id`;
ALTER TABLE `standups` DROP `team_id`;
ALTER TABLE `users` DROP `team_id`; 
ALTER TABLE `channel_members` DROP `team_id`;
ALTER TABLE `controll_pannel` DROP `team_name`;