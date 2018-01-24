-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `comedians` (
`id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
`created` DATETIME NOT NULL,
`modified` DATETIME NOT NULL,
`slack_name` VARCHAR(255) NOT NULL,
`full_name` VARCHAR(255) NOT NULL,
KEY (`full_name`, `slack_name`)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `comedians`;
