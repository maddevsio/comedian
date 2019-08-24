-- +goose Up
-- SQL in this section is executed when the migration is applied.

DROP TABLE `standup_time`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

CREATE TABLE `standup_time` (
`id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
`created` DATETIME NOT NULL,
`channel` VARCHAR(255) NOT NULL,
`channel_id` VARCHAR(255) NOT NULL,
`standuptime` BIGINT NOT NULL,
KEY (`standuptime`)
);