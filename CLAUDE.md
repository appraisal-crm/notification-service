# Appraisal CRM — notification-service

Commercial project built for a real client. Code goes to production — treat it accordingly.

**This repo is `notification-service`.** `request-service` is the reference
implementation — follow its layout and conventions. This file carries the
project-wide charter (each service is a separate repo) plus this service's specifics.

## What it does

CRM for a property appraisal company (apartments, houses, land, vehicles, commercial real estate).
Digitizes the full cycle: client submits request → inspector visits the property → appraiser evaluates → client receives report.

`notification-service` is the delivery edge: it consumes domain events from Kafka
and notifies recipients (clients and staff) over email, SMS, in-app and push.

## Roles

| Role          | What they do in the system                                                         |
|---------------|------------------------------------------------------------------------------------|
| Client        | Submits a request, tracks status, downloads the final report                       |
| Appraiser     | Accepts requests, assigns an inspector, conducts appraisal, sends the report       |
| Inspector     | Receives field assignments, uploads photos and property data                       |
| Administrator | Manages users, monitors the system                                                 |

## Request lifecycle (strictly linear — no going back)

```
New → In Progress → Inspection Scheduled → Inspection Completed → Appraisal → Report Sent → Closed
```

Transitions live in request-service; notification-service only reacts to the resulting events.

## Stack

| Layer              | Technology                               |
|--------------------|------------------------------------------|
| Backend services   | Go (chi, pgx, golang-migrate)            |
| Database per svc   | PostgreSQL (Database-per-Service pattern)|
| Auth               | Keycloak 26 (OAuth2/OIDC)               |
| Events             | Apache Kafka                             |
| Cache / Dedup      | Redis                                    |
| Object storage     | S3 Yandex Cloud                          |
| Frontend (4 SPAs)  | React + TypeScript                       |
| Architecture docs  | Structurizr DSL (C4)                     |

## Notification Service (this repo)

- **Pure consumer** — it produces no events, so there is **no `outbox`** here.
- **Consumes:** `request.created`, `request.status_changed` (topic `request.events`),
  `inspect.completed` (`inspect.events`), `report.ready` (`review.events`).
- **Channels:** `email`, `sms`, `in_app`, `push` (lookup table `notification_channels`).
- **Idempotency:** dedup by `event_id` in **Redis** (`SET NX EX`) — no inbox table in the DB.
- **DB:** `notification_db`. Tables: `notifications` (aggregate) + `notification_channels`
  / `notification_statuses` (lookup, FK from `notifications`).
- **Default port:** `8083`.

### Current state (this repo)

**Done:**
- Database layer — schema (migration `000001`), `internal/domain` (Notification,
  channels, statuses, errors), `internal/repository` (interface + pgx implementation),
  `config`, minimal `cmd/server/main.go` (DB connect + `/health` + graceful shutdown).

**Not yet implemented:**
- Kafka consumer (`internal/kafka`) + Redis dedup
- Channel senders (email / SMS / in-app / push)
- HTTP API (list in-app notifications, mark read) + JWT auth
- Unit tests

## Go module path

Not a monorepo. Every service is a separate repository under the `appraisal-crm`
GitHub organization:
```
github.com/appraisal-crm/notification-service
github.com/appraisal-crm/<name>-service   # pattern for services
```

## Go service structure (request-service is the reference layout)

```
notification-service/
  cmd/server/          # entry point, wire DI
  internal/
    domain/            # entities, domain errors (errors.go)
    repository/        # interface + PostgreSQL implementation
    service/           # business logic (to add)
    kafka/             # consumer group (to add)
    handler/           # HTTP (chi router), DTOs (to add)
    middleware/        # JWT auth, role-based access (to add)
    httputil/          # shared response helpers (to add)
  config/              # ENV config (os.Getenv only)
  migrations/          # SQL files (golang-migrate up/down)
```

## Code rules

- No magic frameworks — chi, pgx, playground/validator only
- Config via `os.Getenv` only — no viper, no cobra
- Migration files: `000001_<description>.up.sql` / `.down.sql` — sequential, always both up and down
- JWT validation via Keycloak JWKS: `MicahParks/keyfunc` + `JWKS_URL` env var
- Domain errors in `domain/errors.go`
- HTTP status codes mapped in the handler layer only
- Kafka consumers MUST be idempotent — dedup by `event_id` (Redis `SET NX EX`) before processing
- Never do a synchronous cross-service call — react to events only
- Unit tests for business logic in `service/`

## Kafka events

**One topic per producing service** — events are told apart by `event_type`, NOT by topic.

| event_type               | Topic            | Producer        | This service |
|--------------------------|------------------|-----------------|--------------|
| `request.created`        | `request.events` | request-service | consumes     |
| `request.status_changed` | `request.events` | request-service | consumes     |
| `inspect.completed`      | `inspect.events` | inspect-service | consumes     |
| `report.ready`           | `review.events`  | review-service  | consumes     |

**Event conventions:**
- **Message key** = aggregate id (e.g. `request_id`) → per-aggregate ordering within a partition
- **Format** = JSON envelope: `event_id`, `event_type`, `version`, `occurred_at`, `request_id`, `data{}`
- **Delivery** = at-least-once → this consumer MUST be idempotent (dedup by `event_id`)
- **Schema evolution** = additive only; bump `version` for breaking changes
- **Broker** = KRaft mode (no Zookeeper)

## Commands

```bash
# Start infrastructure (compose lives in the request-service repo root)
docker compose up -d

# Run the service (DATABASE_URL required — see .env.example)
make run          # or: go run cmd/server/main.go

# Tests
make test         # go test ./...

# Migrations (DB_URL defaults to local notification_db)
make migrate-up
make migrate-down
```

## Hard rules — do not break

- No synchronous cross-service calls between business services — events via Kafka only
- No cross-database JOINs between services
- Do not switch config from `os.Getenv` to viper without explicit agreement
- Never modify already-applied migrations — new additive migrations only
- This service never publishes events — do not add an outbox/producer here

## Workflow

- Tasks tracked in Jira, project **ACRM** (mdrslv.atlassian.net)
- Branch from `dev`: `feature/<scope>` / `fix/<scope>`; PR into `dev`
- `main` is the release branch — updated by merging `dev` → `main`; never PR features directly into `main`
- Conventional commits with the Jira key: `feat(notifications): ... (ACRM-XX)`
