-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channel_members` DROP `standup_time`;
ALTER TABLE `channel_members` ADD `submitted_standup_today` BOOLEAN;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` ADD `standup_time` BIGINT NOT NULL;
ALTER TABLE `channel_members` DROP `submitted_standup_today`;