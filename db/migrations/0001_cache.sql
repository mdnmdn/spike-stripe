-- 0001_cache.sql
-- Create a simple key/value cache table
CREATE TABLE IF NOT EXISTS cache (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
