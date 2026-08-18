# notification-service

Consumes domain events from Kafka and delivers notifications to recipients over
email and in-app channels. Part of the Appraisal CRM (Database-per-Service).

## Responsibilities

- **Consume** domain events from Kafka:
  - `request.events` (`request.created`, `request.status_changed`)
  - `inspect.events` (`inspect.completed`)
  - `review.events` (`report.ready`)
- **Dedup** incoming messages via Redis (`SET NX EX` with TTL) for at-least-once idempotency.
- **Render** branded HTML & Plaintext emails with bilingual support (RU & EN).
- **Deliver** email notifications via SMTP (Mailpit locally, production SMTP in deploy).
- **Serve** in-app notifications API with Keycloak JWT authentication.

## Project docs

Project-wide docs live in the reference repo [`request-service`](https://github.com/appraisal-crm/request-service):

- [Architecture (C4 / Structurizr)](https://github.com/appraisal-crm/request-service/tree/main/docs/architecture)
- [ADRs](https://github.com/appraisal-crm/request-service/tree/main/docs/adr)
- [Business requirements (BRD)](https://github.com/appraisal-crm/request-service/tree/main/docs/brd)

## Layout

```
cmd/server/main.go     # entry point: wires DB, Redis, JWKS, Kafka consumers, HTTP server
config/                # ENV configuration
internal/
  dedup/               # Redis deduplication by event_id
  domain/              # Notification entity, event envelopes, errors
  handler/             # HTTP handlers & Chi router (CORS, JWT auth)
  httputil/            # JSON response helpers
  kafka/               # Kafka consumer with at-least-once commit
  middleware/          # Keycloak JWT validation middleware
  repository/          # PostgreSQL repository for notifications
  sender/              # Sender interface, SMTPSender, MockSender
  service/             # NotificationService dispatcher and business logic
  templates/           # Bilingual (RU/EN) HTML email templates renderer
migrations/            # golang-migrate SQL (up/down)
```

## HTTP API

| Method | Path | Roles | Description |
|---|---|---|---|
| `GET` | `/health` | public | Health check (probes PostgreSQL) |
| `GET` | `/notifications` | authenticated | List notifications for authenticated user |
| `GET` | `/notifications/{id}` | authenticated | Get single notification |
| `PATCH` | `/notifications/{id}/read` | authenticated | Mark notification as read |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | — **(required)** | PostgreSQL connection string |
| `SERVER_PORT` | `8083` | HTTP server port |
| `ALLOWED_ORIGINS` | `*` | Allowed CORS origins |
| `REDIS_ADDR` | `localhost:6381` | Redis host:port for dedup |
| `REDIS_PASSWORD` | `appraisal` | Redis password |
| `DEDUP_TTL` | `24h` | Deduplication key TTL in Redis |
| `KAFKA_BROKERS` | `localhost:9094` | Kafka bootstrap brokers |
| `KAFKA_CONSUMER_GROUP` | `notification-service` | Consumer group name |
| `KAFKA_REQUEST_TOPIC` | `request.events` | Request events topic |
| `KAFKA_INSPECT_TOPIC` | `inspect.events` | Inspect events topic |
| `KAFKA_REVIEW_TOPIC` | `review.events` | Review events topic |
| `SMTP_HOST` | `localhost` | SMTP server host (Mailpit: `localhost` / `mailpit`) |
| `SMTP_PORT` | `1025` | SMTP port |
| `SMTP_FROM` | `noreply@appraisal-crm.ru` | Sender email address |
| `SMTP_ENABLED` | `true` | Enable/disable actual SMTP delivery |
| `JWKS_URL` | `http://localhost:8180/.../certs` | Keycloak JWKS certs endpoint |
| `CLIENT_PORTAL_URL` | `http://localhost:5173` | Client SPA base URL |
| `OFFICE_PORTAL_URL` | `http://localhost:5174` | Office SPA base URL |

## Local development

```bash
# 1. Shared infra (Kafka, Keycloak, Mailpit) from repo root:
docker compose -f ../infra/docker-compose.yml up -d

# 2. Notification DB & Redis:
docker compose up -d

# 3. Apply migrations & run service:
make migrate-up
make run
```

Run unit tests:
```bash
make test
```
