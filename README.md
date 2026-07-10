# notification-service

Consumes domain events from Kafka and delivers notifications to recipients over
email, SMS, in-app and push channels. Part of the Appraisal CRM (Database-per-Service).

Status: **database layer only** — schema, domain model and repository are in place;
the Kafka consumer, channel senders and HTTP API are added on top of this in the next steps.

## Layout

```
cmd/server/main.go     # entry point: connects DB, health server, graceful shutdown
config/                # ENV config (os.Getenv only)
internal/
  domain/              # Notification entity, channels, statuses, domain errors
  repository/          # NotificationRepository interface + PostgreSQL implementation
migrations/            # golang-migrate SQL (up/down)
```

## Schema

- `notifications` — one queued/dispatched message per recipient per channel
- `notification_channels` / `notification_statuses` — lookup tables (FK from `notifications`)

Consumer idempotency (dedup by `event_id`) is handled in Redis (`SET NX EX`), not in the DB.

## ENV

| Var            | Required | Default | Description                          |
|----------------|----------|---------|--------------------------------------|
| `DATABASE_URL` | yes      | —       | PostgreSQL DSN (`notification_db`)   |
| `SERVER_PORT`  | no       | `8083`  | HTTP port (currently `/health` only) |

## Run

```bash
# infra (Postgres) runs from the request-service compose at the repo root
createdb notification_db     # once, if not provisioned by infra/postgres/init
make migrate-up              # apply migrations
make run                     # DATABASE_URL must be set (see .env.example)
```

## Migrations

```bash
make migrate-up      # apply all
make migrate-down    # roll back one
```
