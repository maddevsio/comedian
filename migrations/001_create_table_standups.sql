-- +goose Up
-- +goose StatementBegin
CREATE TABLE `standups` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created_at` INTEGER NOT NULL,
    `workspace_id` VARCHAR(255) NOT NULL,
    `channel_id` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) NOT NULL,
    `comment` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    `message_ts` VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `standups`;
-- +goose StatementEnd
