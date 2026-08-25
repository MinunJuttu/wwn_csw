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


CREATE TABLE IF NOT EXISTS characters (
    id INTEGER PRIMARY KEY,

    user_id INTEGER NOT NULL,

    name TEXT NOT NULL,

    level INTEGER NOT NULL
                  DEFAULT 1,

    class TEXT NOT NULL
                 DEFAULT '',

    sheet_version INTEGER NOT NULL
                          DEFAULT 1,

    data TEXT NOT NULL
              DEFAULT '{}',

    created_at TEXT NOT NULL
                    DEFAULT CURRENT_TIMESTAMP,

    updated_at TEXT NOT NULL
                    DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_characters_user_id
ON characters(user_id);