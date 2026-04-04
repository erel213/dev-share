CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    workspace_id TEXT NOT NULL,
    access_all_templates INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    CONSTRAINT fk_groups_workspace FOREIGN KEY (workspace_id)
        REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT unique_group_name_workspace UNIQUE (workspace_id, name)
);

CREATE INDEX IF NOT EXISTS idx_groups_workspace_id ON groups(workspace_id);

CREATE TABLE IF NOT EXISTS group_memberships (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    PRIMARY KEY (group_id, user_id),
    CONSTRAINT fk_gm_group FOREIGN KEY (group_id)
        REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_gm_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_memberships_user_id ON group_memberships(user_id);

CREATE TABLE IF NOT EXISTS group_template_access (
    group_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    PRIMARY KEY (group_id, template_id),
    CONSTRAINT fk_gta_group FOREIGN KEY (group_id)
        REFERENCES groups(id) ON DELETE CASCADE,
    CONSTRAINT fk_gta_template FOREIGN KEY (template_id)
        REFERENCES templates(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_template_access_template_id ON group_template_access(template_id);
