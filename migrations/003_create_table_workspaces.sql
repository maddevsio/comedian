-- +goose Up
-- +goose StatementBegin
CREATE TABLE `workspaces` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created_at` INTEGER NOT NULL,
    `notifier_interval` INTEGER NOT NULL,
    `max_reminders` INTEGER NOT NULL,
    `reminder_offset` INTEGER NOT NULL,
    `workspace_id` VARCHAR(255) NOT NULL,
    `workspace_name` VARCHAR(255) NOT NULL,
    `bot_access_token` VARCHAR(255) NOT NULL,
    `bot_user_id` VARCHAR(255) NOT NULL,
    `projects_reports_enabled` TINYINT NOT NULL,
    `reporting_channel` VARCHAR(255) NOT NULL,
    `reporting_time` VARCHAR(255) NOT NULL,
    `language` VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `workspaces`;
-- +goose StatementEnd

