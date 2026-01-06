-- +migrate Up
CREATE TABLE clicks (
                        id VARCHAR(36) PRIMARY KEY,
                        url_code VARCHAR(20) NOT NULL,
                        timestamp TIMESTAMP NOT NULL,
                        ip VARCHAR(45),
                        user_agent TEXT,
                        referer TEXT,
                        created_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- Index: URL code ile sorgulama için
CREATE INDEX idx_clicks_url_code ON clicks(url_code);

-- Index: Zaman bazlı sorgulama için
CREATE INDEX idx_clicks_timestamp ON clicks(timestamp);

-- +migrate Down
-- DROP TABLE IF EXISTS clicks;