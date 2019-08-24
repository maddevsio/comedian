-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup` MODIFY `comment` VARCHAR(3000) COLLATE utf8mb4_unicode_ci NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup` MODIFY `comment` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;