-- +goose Up
CREATE TABLE user_sessions (
  id INTEGER PRIMARY KEY,
  token VARCHAR(255),
  username VARCHAR(255),
  createdat DATETIME
);
CREATE UNIQUE INDEX token ON user_sessions(token);

-- +goose Down
DROP TABLE user_sessions;
