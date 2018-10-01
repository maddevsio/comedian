-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `timetables` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `channel_member_id` INTEGER NOT NULL, 
    `created` DATETIME NOT NULL,
    `modified` DATETIME NOT NULL,
    `monday` INTEGER NOT NULL,
    `tuesday` INTEGER NOT NULL,
    `wednesday` INTEGER NOT NULL,
    `thursday` INTEGER NOT NULL,
    `friday` INTEGER NOT NULL,
    `saturday` INTEGER NOT NULL,
    `sunday` INTEGER NOT NULL,
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `timetables`; 