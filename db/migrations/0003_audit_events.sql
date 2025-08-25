-- Create audit events table for logging generic system events
CREATE TABLE IF NOT EXISTS audit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL DEFAULT (datetime('now')),
    subsystem TEXT NOT NULL,          -- e.g., 'stripe', 'payment', 'user'
    event_type TEXT NOT NULL,         -- e.g., 'webhook.received', 'transaction.created'
    user_id TEXT,                     -- user identifier (nullable)
    information TEXT,                 -- human-readable description
    payload TEXT,                     -- JSON data (nullable)
    ref_id TEXT,                      -- primary reference ID (e.g., payment_intent_id)
    ref_id2 TEXT                      -- secondary reference ID (e.g., session_id)
);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_audit_events_timestamp ON audit_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_audit_events_subsystem ON audit_events(subsystem);
CREATE INDEX IF NOT EXISTS idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX IF NOT EXISTS idx_audit_events_user_id ON audit_events(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_ref_id ON audit_events(ref_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_ref_id2 ON audit_events(ref_id2);
CREATE INDEX IF NOT EXISTS idx_audit_events_subsystem_event_type ON audit_events(subsystem, event_type);