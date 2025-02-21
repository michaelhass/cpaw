CREATE TABLE users (
    id TEXT PRIMARY KEY,
    created_at INTEGER NOT NULL,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL
);
