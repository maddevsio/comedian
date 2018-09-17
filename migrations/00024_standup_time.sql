-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channels` ADD (`channel_standup_time` BIGINT NOT NULL);


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channels` DROP `channel_standup_time`;