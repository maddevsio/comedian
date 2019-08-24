-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE `controll_pannel` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `notifier_interval` INTEGER NOT NULL,
    `manager_slack_user_id` VARCHAR(255) NOT NULL,
    `reporting_channel` VARCHAR(255) NOT NULL,
    `report_time` VARCHAR(255) NOT NULL,
    `language` VARCHAR(255) NOT NULL,
    `reminder_repeats_max` INTEGER NOT NULL,
    `reminder_time` INTEGER NOT NULL,
    `collector_enabled` BOOLEAN NOT NULL
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE `controll_pannel`;
