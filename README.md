# print-quote-service

A small Go service that prices custom sticker orders and lets a customer accept a quote while it is still valid.

It is intentionally small, but it is built the way I build production services: the business rules are isolated and fully tested, every error has a defined HTTP meaning, and the edge cases that usually become production incidents (money rounding, expiry boundaries, two requests racing on the same record) are handled on purpose rather than by accident.

Standard library only. No external dependencies.

## Run it

```bash
go run ./cmd/api                     # listens on :8080 (override with PORT)
# or
docker build -t print-quote-service . && docker run -p 8080:8080 print-quote-service
```

```bash
# Create a quote: 100 circle stickers, 2x2 inches
curl -s -X POST localhost:8080/v1/quotes \
  -d '{"product":"circle","widthInches":2,"heightInches":2,"quantity":100}'
# => 201, {"id":"qt_…","subtotalCents":4000,"discountCents":1000,"totalCents":3000,"totalDisplay":"$30.00","status":"pending",…}

curl -s -X POST localhost:8080/v1/quotes/<id>/accept   # 200, status "accepted"
curl -s -X POST localhost:8080/v1/quotes/<id>/accept   # 409, problem+json
curl -s "localhost:8080/v1/quotes?limit=10"            # newest first
```

The full contract is in [`openapi.yaml`](openapi.yaml).

## Pricing rules

| Product   | Price per square inch |
| --------- | --------------------- |
| die-cut   | 12¢                   |
| circle    | 10¢                   |
| rectangle | 9¢                    |

Volume discount on the subtotal: 50+ units 10%, 100+ 25%, 500+ 40%, 1,000+ 50%. A quote is valid for 7 days.

## Design

```
cmd/api                 composition root: wires adapters to use cases, graceful shutdown
internal/domain         pricing, validation and quote lifecycle; no I/O, no framework
internal/application
  port                  interfaces the application needs (repository, clock, ID generator)
  command               use cases that change state (create, accept)
  query                 read-only use cases returning API-shaped views
internal/adapter
  http                  JSON API, the only place that knows about HTTP
  memory                in-memory repository
  clock                 system clock and random IDs
```

Dependencies point inward: adapters depend on the application, the application depends on the domain, and the domain depends on nothing. Swapping the in-memory repository for Postgres means writing one new adapter; no use case changes.

### Decisions worth calling out

- **Money is integer cents.** `float64` cannot represent 0.1 exactly, and the error compounds when multiplied by quantities. Discounts use basis points with explicit half-up rounding, and the tests pin the rounding boundary (49.5¢ rounds to 50¢, 61.2¢ to 61¢).
- **Expiration is derived, not stored.** A quote's status is computed from `ExpiresAt` and the current time. There is no background job to flip a flag and no window where the flag and reality disagree. The boundary is exclusive: at exactly `ExpiresAt` the quote is expired, and there is a test for the nanosecond before.
- **Time is injected.** Use cases read time through a `Clock` port, so expiry tests move a fake clock forward instead of sleeping.
- **Optimistic concurrency on accept.** Every state change bumps `Version`, and the repository rejects a write whose previous version does not match. A test fires 50 concurrent accepts at one quote and asserts that exactly one succeeds. In Postgres this becomes `UPDATE … WHERE id = $1 AND version = $2`.
- **Domain methods return new values.** `Quote.Accept` does not mutate its receiver, so a failed save can never leave a half-applied change in memory.
- **CQRS without ceremony.** Commands return only an ID; reads go through query handlers that return views. That keeps the read shape free to evolve without touching write logic, with no buses or frameworks involved.
- **One place maps errors to HTTP.** Domain errors are sentinel values checked with `errors.Is`, mapped to 400/404/409/422 in a single function, and returned as RFC 9457 `problem+json`. Unknown errors become a generic 500 and are logged, so internal details never leak to clients.
- **Strict input.** Request bodies are capped at 1 MiB and unknown JSON fields are rejected, so a client typo fails loudly instead of being silently ignored.

## Tests

```bash
go test -race ./...
```

- Table-driven domain tests covering every discount tier, the rounding edges and the expiry boundaries.
- Use case tests run against the real in-memory repository with a fake clock. They use fakes, not mocks, so they assert outcomes rather than call sequences.
- HTTP tests start the real router with `httptest` and exercise the full request path, including the error mapping.

CI runs `gofmt`, `go vet`, the race detector and an 85% coverage gate (currently 91%), then builds the container image.
