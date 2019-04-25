-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channels` MODIFY `channel_standup_time` VARCHAR(255) NOT NULL;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channels` MODIFY `channel_standup_time` BIGINT NOT NULL;