ALTER TABLE environments ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE environments ADD COLUMN last_applied_at TEXT;
ALTER TABLE environments ADD COLUMN last_operation TEXT;
ALTER TABLE environments ADD COLUMN last_error TEXT;
