-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE `controll_pannel` DROP `task_done_status`;
ALTER TABLE `controll_pannel` DROP `sprint_report_status`;
ALTER TABLE `controll_pannel` DROP `sprint_report_time`;
ALTER TABLE `controll_pannel` DROP `sprint_report_channel`;
ALTER TABLE `controll_pannel` DROP `sprint_weekdays`;
ALTER TABLE `controll_pannel` DROP `individual_reporting_status`;
ALTER TABLE `controll_pannel` DROP `reporting_channel`;
ALTER TABLE `controll_pannel` DROP `report_time`; 
ALTER TABLE `controll_pannel` DROP `collector_enabled`;


-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE `controll_pannel` ADD `task_done_status` VARCHAR(255) NOT NULL;
ALTER TABLE `controll_pannel` ADD `sprint_report_status` BOOLEAN NOT NULL;
ALTER TABLE `controll_pannel` ADD `sprint_report_time` VARCHAR(255) NOT NULL;
ALTER TABLE `controll_pannel` ADD `sprint_report_channel` VARCHAR(255) NOT NULL;
ALTER TABLE `controll_pannel` ADD `sprint_weekdays` VARCHAR(255) NOT NULL;
ALTER TABLE `controll_pannel` ADD `individual_reporting_status` BOOLEAN NOT NULL;
ALTER TABLE `controll_pannel` ADD `reporting_channel`VARCHAR(255) NOT NULL;
ALTER TABLE `controll_pannel` ADD `report_time`VARCHAR(255) NOT NULL; 
ALTER TABLE `controll_pannel` ADD `collector_enabled`BOOLEAN NOT NULL;