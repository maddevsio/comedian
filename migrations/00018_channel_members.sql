-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_users` RENAME TO `channel_members`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `channel_members` RENAME TO `standup_users`;