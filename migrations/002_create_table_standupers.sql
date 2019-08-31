-- +goose Up
-- +goose StatementBegin
CREATE TABLE `standupers` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `created_at` INTEGER NOT NULL,
    `workspace_id` VARCHAR(255) NOT NULL,
    `channel_id` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) NOT NULL,
    `role` VARCHAR(255) NOT NULL,
    `real_name` VARCHAR(255) NOT NULL,
    `channel_name` VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE `standupers`;
-- +goose StatementEnd
