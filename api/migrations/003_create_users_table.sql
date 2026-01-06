-- +migrate Up
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    tier VARCHAR(20) DEFAULT 'free' NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP
);

-- Index: Email ile hızlı arama
CREATE INDEX idx_users_email ON users(email);

-- +migrate Down
-- DROP TABLE IF EXISTS users;