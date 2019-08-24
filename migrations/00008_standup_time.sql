-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `standup_time` MODIFY `standuptime` BIGINT NOT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `standup_time` MODIFY `standuptime` INTEGER NOT NULL;