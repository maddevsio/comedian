-- +goose Up
-- SQL in this section is executed when the migration is applied.
DROP TABLE `standup_edit_history`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
CREATE TABLE `standup_edit_history` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created` DATETIME NOT NULL,
    `standup_id` VARCHAR(255) NOT NULL,
    `standup_text` TEXT NOT NULL,
);