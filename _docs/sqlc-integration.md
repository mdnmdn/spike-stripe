# Implementations: Adding tables and developing with SQLc

This guide explains how to extend the database schema (add tables) and how to build Go data-access code using SQLc in this project. It assumes you skimmed `_docs/sqlc.md` for an overview.

## TL;DR
- Add a new migration file under `db/migrations/` (e.g., `0002_orders.sql`). Use idempotent SQL (CREATE TABLE IF NOT EXISTS, etc.) because the built-in runner re-applies all files on every run.
- Add one or more query files under `db/queries/` (e.g., `orders.sql`) with SQLc annotations (`-- name: ... :one|:many|:exec|:execrows|:batchexec`).
- Run `sqlc generate` to produce/update Go code in `internal/db`.
- Use `internal/db.NewConnection()` to obtain a `*sql.DB`, then `db.New(db)` to get a generated `Queries` instance (or use the generated `Querier` interface).
- Apply migrations either:
    - via `go run cmd/migrate/main.go`, or
    - by starting the server with `RUN_MIGRATION=true`.

## Project wiring
- Config: `sqlc.yaml` generates Go into `internal/db` using engine `sqlite`, schema from `db/migrations`, and queries from `db/queries`.
- Connections: `internal/db/connection.go` chooses between:
    - Turso/LibSQL when `TURSO_DATABASE_URL` is set (and `TURSO_AUTH_TOKEN` optionally), or
    - local SQLite file (path from `DB_PATH` or default `_data/db-spike-strip.sqlite3`).
- Migrations: `internal/db/migrate.go` applies all `*.sql` files from `db/migrations` in lexical order. The runner does not track state; make your migrations idempotent.
- Server: `RUN_MIGRATION=true` will run migrations on startup (after dotenv is loaded).

## Add a new table (example: orders)
1) Create a migration file in `db/migrations/` with a higher prefix than existing files, e.g. `0002_orders.sql`:

```sql
-- 0002_orders.sql
-- Orders table (id UUID string for simplicity, created_at as TEXT ISO8601)
CREATE TABLE IF NOT EXISTS orders (
  id         TEXT PRIMARY KEY,
  amount     INTEGER NOT NULL,
  currency   TEXT    NOT NULL,
  status     TEXT    NOT NULL DEFAULT 'pending',
  created_at TEXT    NOT NULL
);

-- Optional indexes
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
```

Notes:
- Use `IF NOT EXISTS` (and similar idempotent patterns) so rerunning migrations doesn’t error.
- Name files with zero-padded numeric prefixes so lexical order matches intended application order.

2) Add queries in `db/queries/orders.sql` using SQLc annotations and SQLite `?` parameters:

```sql
-- name: CreateOrder :exec
INSERT INTO orders (id, amount, currency, status, created_at)
VALUES (?, ?, ?, ?, ?);

-- name: GetOrder :one
SELECT id, amount, currency, status, created_at
FROM orders
WHERE id = ?
LIMIT 1;

-- name: ListOrders :many
SELECT id, amount, currency, status, created_at
FROM orders
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateOrderStatus :exec
UPDATE orders SET status = ? WHERE id = ?;

-- name: DeleteOrder :exec
DELETE FROM orders WHERE id = ?;
```

3) Generate code:

```bash
sqlc generate
```

The following will be created/updated under `internal/db/`:
- `models.go` (new `Order` struct)
- `orders.sql.go` (methods for the queries)
- `querier.go` (interface includes your new methods)

## Using the generated code

```go
package yourpkg

import (
  "context"
  "time"
  "github.com/google/uuid"
  mydb "stripe-go-spike/internal/db"
)

func SaveAndFetchExample() error {
  // Connect (Turso if TURSO_DATABASE_URL is set; otherwise local SQLite file)
  d, err := mydb.NewConnection()
  if err != nil { return err }
  defer d.Close()

  q := mydb.New(d)
  ctx := context.Background()

  id := uuid.NewString()
  if err := q.CreateOrder(ctx, mydb.CreateOrderParams{
    ID:        id,
    Amount:    1000,
    Currency:  "usd",
    Status:    "pending",
    CreatedAt: time.Now().UTC().Format(time.RFC3339),
  }); err != nil {
    return err
  }

  o, err := q.GetOrder(ctx, id)
  if err != nil { return err }
  _ = o
  return nil
}
```

Tip: If you enabled `emit_prepared_queries` (this project does), the generated `Queries` type prepares statements lazily and reuses them, improving performance.

## Transactions
- The generated package provides `(*Queries).WithTx(tx *sql.Tx) *Queries`.
- Typical pattern:

```go
ctx := context.Background()
d, _ := mydb.NewConnection()
defer d.Close()

err := func() error {
  tx, err := d.BeginTx(ctx, nil)
  if err != nil { return err }
  qtx := mydb.New(d).WithTx(tx)

  // use qtx for all operations inside the transaction
  if err := qtx.UpdateOrderStatus(ctx, mydb.UpdateOrderStatusParams{Status: "paid", ID: orderID}); err != nil {
    _ = tx.Rollback()
    return err
  }
  return tx.Commit()
}()
```

## Nulls and types
- For nullable columns, use `NULL` in schema and handle with `sql.NullString`, `sql.NullInt64`, etc. SQLc will infer these types.
- For timestamps, you can store ISO8601 text as shown, or use INTEGER epoch seconds. Choose what suits your needs and keep it consistent.

## Parameter style and engine
- Engine is `sqlite`; use positional `?` placeholders in queries (not `$1`).
- If you later add a different engine, you’ll need a separate `sqlc.yaml` entry or a separate project.

## Migrations: running and strategy
- One-off: `go run cmd/migrate/main.go`
- On server start: set `RUN_MIGRATION=true` (dotenv supported). Example:

```bash
APP_ENV=development RUN_MIGRATION=true ./stripe-go-spike
```

Migration runner behavior in this repo:
- Applies all `*.sql` files in lexical order every time it runs.
- Does not keep a migrations table; therefore use idempotent SQL (`IF NOT EXISTS`, `DROP TABLE IF EXISTS`, `CREATE INDEX IF NOT EXISTS`, etc.)
- If you need non-idempotent, stateful migrations (e.g., data backfills you don’t want to re-run), add your own guard logic within the SQL (e.g., `INSERT ... SELECT ... WHERE NOT EXISTS(...)`), or adopt a full-featured tool (e.g., goose, atlas, flyway) and wire it in.

## Testing tips
- Use a temp SQLite database file per test, or in-memory:

```go
// In-memory with modernc.org/sqlite
dsn := "file:memdb1?mode=memory&cache=shared"
d, err := sql.Open("sqlite", dsn)
```

- Apply migrations within tests using `db.RunMigrations(d, os.DirFS("db/migrations"), "")`.
- Then instantiate `q := db.New(d)` and run assertions on CRUD operations.

## Conventions and style
- Migrations are numbered `0001_*.sql`, `0002_*.sql`, ... to preserve order.
- Keep migrations small and focused; avoid altering the same table in multiple ways in a single file if not necessary.
- Co-locate queries by feature (e.g., `orders.sql`, `payments.sql`) under `db/queries/`.
- Prefer `:exec` for inserts/updates/deletes, `:one` for single-row selects, `:many` for lists.
- Use deterministic sorting and pagination in list queries.

## Troubleshooting
- "table already exists": ensure `IF NOT EXISTS` is used.
- "duplicate index": use `CREATE INDEX IF NOT EXISTS`.
- "no such table/column": verify migration order and that you ran migrations before queries.
- Verify environment: `TURSO_DATABASE_URL` set? `DB_PATH` correct? You can enable server startup migrations with `RUN_MIGRATION=true` during development.
