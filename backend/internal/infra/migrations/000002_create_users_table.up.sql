CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(255),
    password VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    workspace_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT unique_oauth_user UNIQUE (oauth_provider, oauth_id),
    CONSTRAINT check_auth_method CHECK (
        (oauth_provider IS NOT NULL AND oauth_id IS NOT NULL AND password IS NULL) OR
        (oauth_provider IS NULL AND oauth_id IS NULL AND password IS NOT NULL)
    )
);

CREATE INDEX idx_users_workspace_id ON users(workspace_id);
CREATE INDEX idx_users_email ON users(email);
