-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_users` ADD (
`channel` VARCHAR (255) NOT NULL,
`channel_id` VARCHAR (255) NOT NULL
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup_users` DROP `channel`, DROP `channel_id`;
