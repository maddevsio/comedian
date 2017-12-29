-- +goose Up
-- SQL in this section is executed when the migration is applied.
DROP TABLE IF EXISTS `posts`;
CREATE TABLE `posts` (
  `id` INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `post` TEXT NOT NULL,
  `slack_user_id` VARCHAR(100) NOT NULL,
  `slack_channel_id` VARCHAR(200) NOT NULL,
  `date` DATE NOT NULL
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE posts;