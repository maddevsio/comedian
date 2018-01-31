-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `standup_time` (
`id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
`created` DATETIME NOT NULL,
`channel` VARCHAR(255) NOT NULL,
`channel_id` VARCHAR(255) NOT NULL,
`standuptime` INTEGER NOT NULL,
KEY (`standuptime`)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `standup_time`;