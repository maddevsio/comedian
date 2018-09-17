-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standups` CHANGE `username_id` `user_id` VARCHAR(255) NOT NULL;
ALTER TABLE `standups` DROP `channel`;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standups` CHANGE `user_id` `username_id` VARCHAR(255) NOT NULL;
ALTER TABLE `standups` ADD (`channel` VARCHAR (255) NOT NULL);