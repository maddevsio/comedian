-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup` ADD `channel` VARCHAR (255) NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup` DROP `channel`;