-- +goose Up
CREATE TABLE IF NOT EXISTS users (
   id SERIAL PRIMARY KEY,
   current_minor BIGINT DEFAULT 0,
   withdrawn_minor BIGINT DEFAULT 0,
   username VARCHAR(50) UNIQUE NOT NULL,
   password_hash TEXT NOT NULL,
   created_at TIMESTAMP DEFAULT NOW(),
   updated_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;



