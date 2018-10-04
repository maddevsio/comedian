-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_edit_history` MODIFY `standup_text` TEXT COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `standups` MODIFY `comment` TEXT COLLATE utf8mb4_unicode_ci NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup_edit_history` MODIFY `standup_text` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL;
ALTER TABLE `standups` MODIFY `comment` VARCHAR(3000) COLLATE utf8mb4_unicode_ci NOT NULL;
