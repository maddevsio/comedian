-- +goose Up
-- +goose StatementBegin
CREATE TABLE `projects` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created_at` INTEGER NOT NULL,
    `workspace_id` VARCHAR(255) NOT NULL,
    `channel_id` VARCHAR(255) NOT NULL,
    `channel_name` VARCHAR(255) NOT NULL,
    `deadline` VARCHAR(255) NOT NULL,
    `tz` VARCHAR(255) NOT NULL,
    `submission_days` VARCHAR(255) NOT NULL,
    `onbording_message` TEXT COLLATE utf8mb4_unicode_ci NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `projects`;
-- +goose StatementEnd