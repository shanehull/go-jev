# AGENTS.md — go-jev

TypeSafe System One API client for Go. Stdlib only. No dependencies.

## Commands

```bash
go build ./...          # always builds clean (no deps)
go vet ./...            # always clean
go test ./...           # unit tests, offline, mock server
TYPESAFE_API_KEY=... go test ./...   # adds live API tests
go test ./... -run TestSystemOneRequest   # run a single test
```

Offline tests use an in-process mock server. Live tests are gated on `TYPESAFE_API_KEY` and skip cleanly when unset.

## Architecture

```
jev.go         Client, New(), WithAPIKey/WithHTTPClient/WithBaseURL/WithModel
errors.go      type APIError = internal.APIError  (type alias, use errors.As)
state.go       Input, State (text), Value (structured)
primitives.go  Choice, Score, Noul, Questions, Question
systemone.go   SystemOne, Evaluate, Request, request encoding, retries
map.go         Map, MapResult, retries, backoff, rate limiting
options.go     RequestOption and MapOption
types.go       Response, Answer, ChoiceAnswer, ScoreAnswer, NoulAnswer, Usage, answer decoding
internal/http.go  Post — all HTTP I/O, bearer auth, error parsing
```

Every public method takes `ctx context.Context` as first argument.

## Gotchas

### API key is required

`New` returns an error when no key is set. Resolution order is `WithAPIKey`, then `$TYPESAFE_API_KEY`.

### Answers are a discriminated union

The wire format tags each answer with a `type` field (`choice`, `score`, `noul`). `Response.UnmarshalJSON` decodes each into its concrete type and stores it in `Answers`. Use `Response.Choice`, `Response.Score`, and `Response.Noul`, or a type switch over the `Answer` interface. A type switch on a value type, not a pointer.

### Criteria shape depends on the primitive

A choice question takes an object of option to description. A score question takes an array of rubric labels. A noul question takes none. `Question.Criteria` is a `json.RawMessage` because the shape varies; the constructors build the right one.

### Fan-out is inherent to SystemOne

`SystemOne` sends every question for one state in a single request. `Map` is the multi-state fan-out, with bounded concurrency, exponential backoff (200ms doubling), and a simple rate limiter.

### Retries

`Evaluate` and `SystemOne` retry network failures and HTTP 429, 529, and 5xx with exponential backoff, up to the client's max attempts (default 3, set with `WithMaxAttempts`). Other 4xx return immediately. `Map` calls `evaluate` with a single attempt and applies its own outer retry, so a `Map` call never retries twice for one state. `APIError.Retryable` reports whether a status code is retryable.

### Usage is returned, output is free

The response carries `Usage` with input and output token counts. Output tokens are not billed by the API, but the counts are exposed so callers can track cost.

## Conventions

- Tests in the `jev_test` external package. Test functions named `Test<Name>`.
- When adding a method, also add a test, an example if it is user-facing, and a row in the README API coverage table.
- Live tests check `os.Getenv("TYPESAFE_API_KEY")` and `t.Skip` when unset.
- Semantic commits for release-please. `feat:` bumps minor (pre-1.0), `fix:` bumps patch.
