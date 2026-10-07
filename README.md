# event-proxy

Demonstration and practice for creating a proxy server in Go.

## Internal structure

Endpoint and event features are grouped by domain:

- `internal/endpoints` contains endpoint ports, request handlers, use cases,
  errors, and SQL persistence.
- `internal/events` contains event ports, request handlers, delivery use cases,
  errors, and SQL persistence.
- `internal/database` opens the database and initializes the shared schema.
- `internal/utility` contains shared HTTP helpers.
- `cmd/main.go` wires repositories, services, and handlers together.

Each request flows from its HTTP handler to a domain service. The service
validates input and coordinates the repository plus outbound HTTP behavior.
Repositories own database queries; handlers own HTTP decoding and responses.

## Go documentation and naming conventions

- Start each exported type, function, method, and package-level variable
  comment with the identifier it documents. Write comments as complete,
  concise sentences that explain purpose or behavior.
- Use `UpperCamelCase` for exported identifiers and `lowerCamelCase` for
  package-private identifiers. Keep initialisms uppercase (`HTTP`, `URL`,
  `SQL`).
- Give types singular, role-focused names (`Handler`, `Repository`, `Request`).
  Use descriptive names for parameters and locals; avoid opaque abbreviations.
  Conventional short names such as `ctx`, `err`, and `db` are acceptable.
- Keep receiver names short, consistent, and derived from the receiver type
  (`app` for `Application`, `handler` for `Handler`).
- Put `context.Context` first in function and method parameter lists.
- In tests, name cases by behavior and use table-driven tests when cases share
  setup and assertions.

Send an event with `POST /event/` using a JSON body like
`{"method":"POST","url":"https://example.com/events","data":{"id":123}}`.
The URL must already be registered through the endpoints API. The proxy forwards
the `data` value as JSON, retries failed deliveries up to three times, and
returns the upstream response status, headers, and body.
