CREATE TABLE urls
(
    id             VARCHAR(36) PRIMARY KEY,
    long_url       TEXT                    NOT NULL,
    code           VARCHAR(20) UNIQUE      NOT NULL,
    url_expiration TIMESTAMP,
    click_count    INTEGER   DEFAULT 0,
    user_id        VARCHAR(36),
    created_at     TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at     TIMESTAMP,
    deleted_at     TIMESTAMP
);

CREATE INDEX idx_urls_code ON urls (code);

CREATE SEQUENCE url_counter_seq START WITH 1000000;