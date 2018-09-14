-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `channels` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `channel_name` VARCHAR(255) NOT NULL,
    `channel_id` VARCHAR(255) NOT NULL
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `channels`;