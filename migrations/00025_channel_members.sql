-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channel_members` ADD (`created` DATETIME);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` DROP `created`;