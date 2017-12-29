-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `standup` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created` DATETIME NOT NULL,
    `modified` DATETIME NOT NULL,
    `channel` VARCHAR(255) NOT NULL,
    `channel_id` VARCHAR(255) NOT NULL,
    `username_id` VARCHAR(255) NOT NULL,
    `username` VARCHAR(255) NOT NULL,
    `full_name` VARCHAR(255) NOT NULL,
    `comment` VARCHAR(255) NOT NULL,
    `message_ts` VARCHAR(255) NOT NULL,
    KEY (`created`, `username`)
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `standup`;