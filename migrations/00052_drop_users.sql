-- +goose Up
-- SQL in this section is executed when the migration is applied.
DROP TABLE `users`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
CREATE TABLE `users` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_name` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) NOT NULL,
    `real_name` VARCHAR (255) NOT NULL,
    `team_id` VARCHAR (255) NOT NULL,
    `tz` VARCHAR (255) NOT NULL,
    `tz_offset` INTEGER NOT NULL,
    `status` VARCHAR (255) NOT NULL,
);