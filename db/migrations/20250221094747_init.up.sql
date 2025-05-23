CREATE TABLE IF NOT EXISTS users (
    id TEXT NOT NULL PRIMARY KEY,
    created_at INTEGER NOT NULL,
    user_name TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT,
    FOREIGN KEY (role) REFERENCES roles (name) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS items (
    id TEXT NOT NULL PRIMARY KEY,
    created_at INTEGER NOT NULL,
    content TEXT NOT NULL,
    user_id TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS roles (name TEXT NOT NULL PRIMARY KEY);

INSERT
OR IGNORE INTO roles (name)
VALUES
    ("admin");

INSERT
OR IGNORE INTO roles (name)
VALUES
    ("user");

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT NOT NULL PRIMARY KEY,
    expires_at INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
