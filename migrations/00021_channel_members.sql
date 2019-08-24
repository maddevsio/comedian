-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `channel_members` CHANGE `slack_user_id` `user_id` VARCHAR(255);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` CHANGE `user_id` `slack_user_id` VARCHAR(255);