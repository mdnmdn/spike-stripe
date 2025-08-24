-- name: CreateAuditEvent :exec
INSERT INTO audit_events (
    subsystem,
    event_type,
    user_id,
    information,
    payload
) VALUES (?, ?, ?, ?, ?);

-- name: GetAuditEventsBySubsystem :many
SELECT * FROM audit_events
WHERE subsystem = ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetAuditEventsByEventType :many
SELECT * FROM audit_events
WHERE event_type = ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetAuditEventsByUser :many
SELECT * FROM audit_events
WHERE user_id = ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetAuditEventsInDateRange :many
SELECT * FROM audit_events
WHERE timestamp >= ? AND timestamp <= ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetAllAuditEvents :many
SELECT * FROM audit_events
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: GetAuditEventsBySubsystemAndType :many
SELECT * FROM audit_events
WHERE subsystem = ? AND event_type = ?
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;