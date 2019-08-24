-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_users` DROP `full_name`;
ALTER TABLE `standup` DROP `full_name`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup_users` ADD (`full_name` VARCHAR (255) NOT NULL);
ALTER TABLE `standup` ADD (`full_name` VARCHAR (255) NOT NULL);
