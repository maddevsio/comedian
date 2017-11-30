-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `standup` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created` DATETIME NOT NULL,
    `modified` DATETIME NOT NULL,
    `username` VARCHAR(255) NOT NULL,
    `comment` VARCHAR(255) NOT NULL,
    KEY (`created`, `username`)
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `standup`;