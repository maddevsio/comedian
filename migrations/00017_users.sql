-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `users` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `user_name` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) NOT NULL,
    `role` VARCHAR (255) NOT NULL
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `users`;