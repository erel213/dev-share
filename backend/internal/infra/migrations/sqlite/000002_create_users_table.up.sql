CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    oauth_provider TEXT,
    oauth_id TEXT,
    password TEXT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    workspace_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT unique_oauth_user UNIQUE (oauth_provider, oauth_id),
    CONSTRAINT check_auth_method CHECK (
        (oauth_provider IS NOT NULL AND oauth_id IS NOT NULL AND password IS NULL) OR
        (oauth_provider IS NULL AND oauth_id IS NULL AND password IS NOT NULL)
    )
);

CREATE INDEX idx_users_workspace_id ON users(workspace_id);
CREATE INDEX idx_users_email ON users(email);
