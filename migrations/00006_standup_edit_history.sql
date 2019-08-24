-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `standup_edit_time` (
`id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
`created` DATETIME NOT NULL,
`standup_id` VARCHAR(255) NOT NULL,
`standup_text` VARCHAR(255) NOT NULL,
KEY (`standup_id`)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `standup_edit_time`;