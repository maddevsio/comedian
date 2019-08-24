-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_edit_time` RENAME `standup_edit_history`;
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup_edit_history` RENAME `standup_edit_time`;