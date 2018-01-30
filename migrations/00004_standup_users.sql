-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_users` CHANGE `slack_name` `username` VARCHAR(255) NOT NULL;
ALTER TABLE `standup_users` DROP KEY `full_name`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup_users` CHANGE `username` `slack_name` VARCHAR(255) NOT NULL;
ALTER TABLE `standup_users` ADD KEY (`full_name`, `slack_name`);