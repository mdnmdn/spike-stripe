-- name: GetCacheValue :one
SELECT value FROM cache WHERE key = ? LIMIT 1;

-- name: SetCacheValue :exec
INSERT INTO cache (key, value)
VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value=excluded.value;

-- name: DeleteCacheKey :exec
DELETE FROM cache WHERE key = ?;

-- name: ListCache :many
SELECT key, value FROM cache ORDER BY key ASC;
