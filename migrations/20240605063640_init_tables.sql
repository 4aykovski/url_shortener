-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users
(
  id SERIAL PRIMARY KEY,
  login TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS refresh_sessions 
(
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token TEXT NOT NULL,
  expires_in TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS urls
(
  id SERIAL PRIMARY KEY,
  alias TEXT NOT NULL UNIQUE,
  url TEXT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_sessions;
DROP TABLE IF EXISTS urls;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
