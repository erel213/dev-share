-- SQLite does not support DROP COLUMN before 3.35.0; recreate table without the columns.
CREATE TABLE environments_backup AS SELECT id, name, description, created_by, workspace_id, template_id, created_at, updated_at, ttl_seconds FROM environments;
DROP TABLE environments;
ALTER TABLE environments_backup RENAME TO environments;

CREATE INDEX IF NOT EXISTS idx_environments_workspace_id ON environments(workspace_id);
CREATE INDEX IF NOT EXISTS idx_environments_created_by ON environments(created_by);
CREATE INDEX IF NOT EXISTS idx_environments_template_id ON environments(template_id);
