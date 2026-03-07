CREATE TABLE IF NOT EXISTS template_variables (
    id TEXT PRIMARY KEY,
    template_id TEXT NOT NULL,
    key TEXT NOT NULL,
    description TEXT,
    var_type TEXT NOT NULL DEFAULT 'string',
    default_value TEXT,
    is_sensitive INTEGER NOT NULL DEFAULT 0,
    is_required INTEGER NOT NULL DEFAULT 1,
    validation_regex TEXT,
    is_auto_parsed INTEGER NOT NULL DEFAULT 1,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    FOREIGN KEY (template_id) REFERENCES templates(id) ON DELETE CASCADE,
    UNIQUE(template_id, key)
);

CREATE INDEX idx_template_variables_template_id ON template_variables(template_id);
