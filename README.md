# Event Booking REST API

[![Go Version](https://img.shields.io/github/go-mod/go-version/alejandrolopezj/events-booking-rest-api)](https://go.dev/)
[![License](https://img.shields.io/github/license/alejandrolopezj/events-booking-rest-api)](./LICENSE)
![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0.3-85EA2D?logo=swagger)
[![Test Coverage](https://codecov.io/gh/alejandrolopezj/events-booking-rest-api/branch/main/graph/badge.svg)](https://codecov.io/gh/alejandrolopezj/events-booking-rest-api)

A REST API for user registration, event management, and event registrations.
It is built with Go, Gin, and SQLite.

## Purpose and scope

This repository is an educational demonstration of Go and common API
development practices. It is designed for learning and experimentation, not
for production deployment. A production service would require additional
security hardening, configuration management, operational safeguards, and
review.

## Requirements

- Go 1.27.2, as specified in `go.mod`.

## Installation and execution

Clone the repository and run the application from its root directory:

```bash
git clone https://github.com/alejandrolopezj/events-booking-rest-api.git
cd events-booking-rest-api
go mod download
go run .
```

The server listens at `http://localhost:8080`. On startup, it creates `api.db`
and applies the SQL schemas embedded from `db/schema/`.

## Configuration

The `LEVEL_LOG` environment variable sets the minimum logging level. Supported
values are `DEBUG`, `INFO`, `WARN`, and `ERROR`. If unset or invalid, the
default level is `DEBUG`.

For example, on macOS or Linux:

```bash
LEVEL_LOG=INFO go run .
```

The JWT signing key is currently defined in the source code. Replace it with a
secure external configuration before using the API in production.

## Endpoints

| Method | Route | Authentication | Description |
| --- | --- | --- | --- |
| `POST` | `/signup` | No | Creates a user account |
| `POST` | `/login` | No | Validates credentials and returns a JWT |
| `GET` | `/events` | No | Lists all events |
| `GET` | `/events/:id` | No | Retrieves an event by ID |
| `POST` | `/events` | Yes | Creates an event |
| `PUT` | `/events/:id` | Yes | Updates an event owned by the user |
| `DELETE` | `/events/:id` | Yes | Deletes an event owned by the user |
| `POST` | `/events/:id/registration` | Yes | Registers the user for an event |
| `DELETE` | `/events/:id/registration` | Yes | Cancels the user's event registration |

Protected routes require the JWT in the `Authorization` header:

```text
Authorization: <token>
```

The middleware expects the token directly in the header, without the `Bearer`
prefix.

### Sign up

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secure-password"}'
```

### Log in

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secure-password"}'
```

The response includes a JWT that is valid for two hours. Send it in the
`Authorization` header when calling protected routes.

### Create an event

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -H "Authorization: <token>" \
  -d '{
    "title":"Go Meetup",
    "description":"A Go community meetup",
    "location":"Online",
    "date_time":"2026-10-14T18:00:00Z"
  }'
```

## Tests

Run all package tests from the repository root:

```bash
go test -v -cover ./...
```

GitHub Actions runs the test suite and uploads its coverage report to Codecov.

## API specification

The OpenAPI 3 specification is available in [swagger.yaml](./swagger.yaml).

## Project structure

- `db/`: SQLite initialization and SQL schemas embedded in the binary.
- `middlewares/`: JWT authentication middleware.
- `models/`: data models and persistence operations.
- `routes/`: HTTP route registration and handlers.
- `swagger.yaml`: OpenAPI 3 API specification.
- `utils/`: password, JWT, logging, and error-handling utilities.
- `test-api/`: sample HTTP requests for use with a compatible API client.

## License

This project is distributed under the MIT License. See [LICENSE](./LICENSE).

## Acknowledgements

This project is built upon the core concepts and final project taught in the course [Go - The Complete Guide](https://www.udemy.com/course/go-the-complete-guide/) by [Maximilian Schwarzmüller](https://academind.com/courses).

Building on that foundation, I extended the project by incorporating additional features and production-ready practices, including:
- **Unit & Integration Testing:** Added automated test coverage to ensure API reliability.
- **Structured Logging:** Implemented standardized, structured logging for better observability and debugging.

GitHub Copilot assisted with exploring design options, implementing features,
writing tests, and documenting the project. I reviewed the changes and remain
responsible for the final code.
