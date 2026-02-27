CREATE TABLE IF NOT EXISTS urls (
  id                BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  slug              VARCHAR(100) NOT NULL UNIQUE,
  target_url        TEXT         NOT NULL,
  password_hash     VARCHAR(60)  NULL,
  manage_token_hash CHAR(64)     NOT NULL,
  expires_at        TIMESTAMP    NOT NULL,
  created_at        TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

