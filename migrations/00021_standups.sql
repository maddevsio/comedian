-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup` RENAME TO `standups`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standups` RENAME TO `standup`;