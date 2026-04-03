CREATE TABLE IF NOT EXISTS teardown_queue (
    environment_id TEXT PRIMARY KEY,
    teardown_at TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%S', 'now')),
    CONSTRAINT fk_teardown_environment
        FOREIGN KEY (environment_id) REFERENCES environments(id) ON DELETE CASCADE
);

CREATE INDEX idx_teardown_queue_due ON teardown_queue(status, teardown_at);
