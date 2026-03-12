CREATE TABLE IF NOT EXISTS environment_variable_values (
    id UUID PRIMARY KEY,
    environment_id  NOT NULL,
    template_variable_id TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE,
    FOREIGN KEY (template_variable_id) REFERENCES template_variables(id) ON DELETE CASCADE,
    UNIQUE(environment_id, template_variable_id)
);

CREATE INDEX idx_env_var_values_environment_id ON environment_variable_values(environment_id);
CREATE INDEX idx_env_var_values_template_variable_id ON environment_variable_values(template_variable_id);
