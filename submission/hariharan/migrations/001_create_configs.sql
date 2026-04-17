CREATE TABLE configs (
    id TEXT PRIMARY KEY,
    host TEXT NOT NULL,
    port INT NOT NULL,
    app_name TEXT NOT NULL,
    log_level TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
