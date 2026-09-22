# go-jev

<p align="center">
  <img src="https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_LightBlue.png" height="80" alt="Go">
  <br>
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/typesafe-dark.png">
    <img src="assets/typesafe-light.png" height="40" alt="TypeSafe AI">
  </picture>
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/shanehull/go-jev.svg)](https://pkg.go.dev/github.com/shanehull/go-jev)
[![Go Report Card](https://goreportcard.com/badge/github.com/shanehull/go-jev)](https://goreportcard.com/report/github.com/shanehull/go-jev)
[![Go](https://img.shields.io/github/go-mod/go-version/shanehull/go-jev)](https://go.dev/dl/)
[![CI](https://github.com/shanehull/go-jev/actions/workflows/test.yaml/badge.svg)](https://github.com/shanehull/go-jev/actions/workflows/test.yaml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Go client for the [TypeSafe System One API](https://docs.typesafe.ai). Standard library only, zero dependencies.

It sends a state and a set of typed questions, and returns calibrated answers: a choice from a list, a score on an ordered rubric, or the probability that a statement is true.

## Install

```sh
go get github.com/shanehull/go-jev
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/shanehull/go-jev"
)

func main() {
	client, err := jev.New() // uses $TYPESAFE_API_KEY
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.SystemOne(context.Background(),
		jev.State("I was charged twice, please fix this ASAP."),
		jev.Questions{
			"category": jev.Choice("What is this about", map[string]string{
				"billing":   "Payment or subscription issues",
				"technical": "Bugs or integration problems",
				"other":     "Anything else",
			}),
			"urgent": jev.Noul("The message conveys urgency"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	category, _ := res.Choice("category")
	urgent, _ := res.Noul("urgent")
	fmt.Println(category.Choice, urgent.Noul)
}
```

## API Key

Get a key from the [TypeSafe dashboard](https://console.typesafe.ai/keys).

```go
// Environment variable (recommended)
client, _ := jev.New()

// Explicit option (overrides env)
client, _ := jev.New(jev.WithAPIKey("your-key"))
```

## Usage

### Primitives

A question is one of three primitives.

```go
// Choice: the model picks one option and returns a probability for each.
jev.Choice("Which team should handle this", map[string]string{
    "billing":   "Payment or subscription issues",
    "technical": "Bugs or integration problems",
})

// Score: the model places the state on an ordered rubric.
jev.Score("How frustrated the customer appears", []string{
    "Calm, just stating facts",
    "Frustrated but civil",
    "Very angry, strong language",
})

// Noul: the model returns the probability that the statement is true.
jev.Noul("The message conveys urgency")
```

### Answers

`SystemOne` returns a `Response`. Answers are typed. Use the accessors, or a type switch over the `Answer` interface.

```go
res, err := client.SystemOne(ctx, jev.State(text), questions)

category, _ := res.Choice("category")       // ChoiceAnswer
frustration, _ := res.Score("frustration")  // ScoreAnswer
urgent, _ := res.Noul("urgent")             // NoulAnswer

category.Choice                 // "billing"
category.Confidence             // 0.78
category.Probabilities["billing"] // 0.85

frustration.Score               // 1.0
frustration.Legend["1"]         // "Frustrated but civil"

urgent.Noul                     // 1.0
```

### States

`State` sends text. `Value` sends any JSON-marshalable value when the state is structured.

```go
client.SystemOne(ctx, jev.State("plain text"), questions)
client.SystemOne(ctx, jev.Value(map[string]any{"text": "...", "region": "NSW1"}), questions)
```

### Fan-out and batching

`SystemOne` answers every question for one state in a single request. `Map` evaluates the same questions against many states with bounded concurrency, retries, and optional rate limiting.

```go
results := client.Map(ctx, states, questions,
    jev.WithConcurrency(8),
    jev.WithMaxRetries(2),
    jev.WithMinInterval(50*time.Millisecond),
)
for _, result := range results {
    if result.Err != nil {
        log.Print(result.Err)
        continue
    }
    // result.Response
}
```

### Models

The default model is `jev-latest`. Override it per client or per call.

```go
client, _ := jev.New(jev.WithModel("jev-latest"))
client.SystemOne(ctx, state, questions, jev.WithRequestModel("jev-latest"))
```

### Low-level requests

`SystemOne` covers the common case. `Evaluate` sends a full `Request` when you need to build the state, model, and questions ahead of time.

```go
req := jev.Request{
    State:     jev.Value(map[string]any{"text": text, "region": "NSW1"}),
    Model:     "jev-latest",
    Questions: questions,
}
res, err := client.Evaluate(ctx, req)
```

### Retries

Transient failures are retried with exponential backoff. `SystemOne` and `Evaluate` retry HTTP 429, 529, and 5xx up to `WithMaxAttempts` times (default 3). Other 4xx responses are returned immediately.

```go
client, _ := jev.New(jev.WithMaxAttempts(5))
```

`APIError.Retryable` reports whether a status code is worth retrying.

### Error handling

```go
_, err := client.SystemOne(ctx, state, questions)
var apiErr *jev.APIError
if errors.As(err, &apiErr) {
    fmt.Printf("TypeSafe error %d: %s\n", apiErr.StatusCode, apiErr.Message)
    if apiErr.Retryable() {
        // back off and try again
    }
}
```

## API Coverage

| Area       | Functions                                                                                         |
| ---------- | ------------------------------------------------------------------------------------------------- |
| Client     | `New`, `WithAPIKey`, `WithBaseURL`, `WithHTTPClient`, `WithModel`, `WithMaxAttempts`              |
| Requests   | `SystemOne`, `Evaluate`, `Request`, `WithRequestModel`                                            |
| Batching   | `Map`, `WithConcurrency`, `WithMaxRetries`, `WithMinInterval`, `WithMapModel`                     |
| Primitives | `Choice`, `Score`, `Noul`, `Questions`                                                            |
| States     | `State`, `Value`                                                                                  |
| Answers    | `Response.Choice`, `Response.Score`, `Response.Noul`, `ChoiceAnswer`, `ScoreAnswer`, `NoulAnswer` |

## Examples

See [examples/](examples/) for runnable programs:

- [systemone](examples/systemone/main.go) — classify a support ticket across all three primitives
- [choice](examples/choice/main.go) — route a message to a team
- [score](examples/score/main.go) — rate severity on a rubric
- [noul](examples/noul/main.go) — detect properties in text
- [batch](examples/batch/main.go) — evaluate many states concurrently with `Map`
- [structured](examples/structured/main.go) — send a JSON state with `Value`
- [verification](examples/verification/main.go) — check a draft answer against its source

The package also carries godoc examples in [example_test.go](example_test.go) for every primitive and accessor. Run an example with:

```sh
TYPESAFE_API_KEY=... go run ./examples/systemone
```

## Testing

Unit tests run offline with a mock server. Live API tests are opt-in.

```sh
go test ./...
TYPESAFE_API_KEY=... go test ./...
```

## License

MIT. See [LICENSE](LICENSE).

The TypeSafe AI logo is a trademark of TypeSafe AI, used only to identify the service this library calls.
