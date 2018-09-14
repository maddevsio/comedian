-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channel_members` DROP `created`;
ALTER TABLE `channel_members` DROP `modified`;
ALTER TABLE `channel_members` DROP `username`;
ALTER TABLE `channel_members` DROP `channel`;
ALTER TABLE `channel_members` DROP `role`;
ALTER TABLE `channel_members` ADD (`standup_time` BIGINT NOT NULL);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` DROP `standup_time`;
ALTER TABLE `channel_members` ADD (`created` DATETIME);
ALTER TABLE `channel_members` ADD (`modified` DATETIME);
ALTER TABLE `channel_members` ADD (`username` VARCHAR (255) NOT NULL);
ALTER TABLE `channel_members` ADD (`channel` VARCHAR (255) NOT NULL);
ALTER TABLE `channel_members` ADD (`role` VARCHAR (255) NOT NULL);

