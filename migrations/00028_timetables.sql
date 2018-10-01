-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `timetables` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `channel_member_id` INTEGER NOT NULL DEFAULT 0, 
    `created` DATETIME NOT NULL,
    `modified` DATETIME NOT NULL,
    `monday` INTEGER NOT NULL DEFAULT 0,
    `tuesday` INTEGER NOT NULL DEFAULT 0,
    `wednesday` INTEGER NOT NULL DEFAULT 0,
    `thursday` INTEGER NOT NULL DEFAULT 0,
    `friday` INTEGER NOT NULL DEFAULT 0,
    `saturday` INTEGER NOT NULL DEFAULT 0,
    `sunday` INTEGER NOT NULL DEFAULT 0
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `timetables`; 