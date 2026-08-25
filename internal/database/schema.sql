CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,

    username TEXT NOT NULL
                  COLLATE NOCASE
                  UNIQUE,

    password_hash TEXT NOT NULL,

    created_at TEXT NOT NULL
                    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY,

    user_id INTEGER NOT NULL,

    token_hash TEXT NOT NULL
                    UNIQUE,

    expires_at TEXT NOT NULL,

    created_at TEXT NOT NULL
                    DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash
ON sessions(token_hash);