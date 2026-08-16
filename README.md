# Task Manager API

A RESTful task manager backend built with Go and PostgreSQL. It demonstrates a clean layered architecture, JWT authentication, and a three-level testing strategy (unit, repository integration, and handler/API integration).

## Features

- User registration with bcrypt password hashing
- Login that returns a signed JWT (HS256)
- Task CRUD: list, get-by-id, create, update, delete
- JWT-protected task endpoints with per-user ownership enforcement
- Duplicate email detection (`409 Conflict`)
- Input validation (`400 Bad Request`)
- Docker and Docker Compose support

## Tech Stack

- Go 1.26.5
- Gin (HTTP framework)
- PostgreSQL 16
- `database/sql` with the pgx driver
- golang-jwt/jwt v5
- bcrypt (`golang.org/x/crypto`)
- Docker / Docker Compose

## Architecture

The application follows a layered architecture inside a single binary:

```
HTTP Handler -> Service -> Repository -> PostgreSQL
```

- **Handler** — parses and validates HTTP requests, maps errors to status codes.
- **Service** — holds business rules and depends on repository interfaces.
- **Repository** — executes SQL against PostgreSQL.
- **main package** — application bootstrap, routing, middleware, JWT helpers, configuration, and DB connection.

## API Endpoints

| Method | Path          | Auth | Description                              |
|--------|---------------|------|------------------------------------------|
| POST   | `/register`   | No   | Register a new user                      |
| POST   | `/login`      | No   | Login and receive a JWT                  |
| GET    | `/tasks`      | Yes  | List tasks belonging to the user         |
| GET    | `/tasks/:id`  | Yes  | Get one task (owner only)                |
| POST   | `/tasks`      | Yes  | Create a task                            |
| PUT    | `/tasks/:id`  | Yes  | Update a task (owner only)               |
| DELETE | `/tasks/:id`  | Yes  | Delete a task (owner only)               |

Authenticated endpoints require an `Authorization: Bearer <token>` header.

## Database

The schema is defined in [`migrations/001_create_users_and_tasks.sql`](migrations/001_create_users_and_tasks.sql):

- `users`: `id` (identity primary key), `email` (unique), `password`, `role` (defaults to `'user'`)
- `tasks`: `id` (identity primary key), `title`, `user_id` (foreign key to `users.id`)

The schema is defined in a SQL migration file, but migration execution is currently manual — there is no migration tool in use.

## Environment Variables

Configuration is read from environment variables. For local development, copy `.env.example` to `.env` and fill in the values:

| Variable      | Example       | Description          |
|---------------|---------------|----------------------|
| `JWT_SECRET`  | `your-secret` | Secret used to sign JWTs |
| `PORT`        | `8080`        | API listen port      |
| `DB_HOST`     | `localhost`   | PostgreSQL host      |
| `DB_PORT`     | `5432`        | PostgreSQL port      |
| `DB_USER`     | `postgres`    | PostgreSQL user      |
| `DB_PASSWORD` | `***`         | PostgreSQL password  |
| `DB_NAME`     | `taskmanager` | PostgreSQL database  |

The application loads `.env` if present (for local development) but does not require it — in Docker, configuration is supplied through container environment variables.

## Testing

Three testing levels, all using the real repository and a real PostgreSQL database where integration is involved:

1. **Service unit tests** — validate business rules using fake repositories (`service/`).
2. **Repository integration tests** — verify real SQL against a dedicated `taskmanager_test` database (`repository/`).
3. **Handler/API integration tests** — full HTTP requests through the real router, middleware, and JWT against `taskmanager_test` (root package).

```sh
go test ./... -count=1                          # unit tests; integration skipped
RUN_INTEGRATION=1 go test ./... -count=1        # all tests including integration
```

Integration tests never touch the development database.

## Docker

```sh
docker compose up --build
```

This starts two services:

- **postgres** — PostgreSQL 16 with a named volume for persistent data, a healthcheck, and the `taskmanager` database.
- **api** — the Go API built from the multi-stage `Dockerfile`, published on host port `8080`.

Compose reads configuration values from `.env` (e.g. `DB_USER`, `PORT`) and passes them to the containers as environment variables. Secrets such as `DB_PASSWORD` and `JWT_SECRET` are also passed this way — nothing is hardcoded in `compose.yaml`. The API container connects to PostgreSQL via the Compose service name `postgres` (`DB_HOST=postgres`), not `localhost`.

### Applying the migration in Docker

The schema is not auto-applied. On a fresh database, run:

```sh
docker compose exec -T postgres sh -c 'psql -U "$POSTGRES_USER" -d taskmanager' \
  < migrations/001_create_users_and_tasks.sql
```

`$POSTGRES_USER` is resolved inside the container from the user configured in `compose.yaml` (`POSTGRES_USER: ${DB_USER}`), so the command works without hardcoding the database user.

## Local Development

Requirements: Go 1.26+, a running PostgreSQL instance, and a `.env` file.

```sh
cp .env.example .env   # then edit with real values
go run .               # starts the API on PORT (default 8080)
```

## Notes

- `.env` is git-ignored and must never be committed — secrets stay local.
- The API requires the `users` and `tasks` tables to exist before use.