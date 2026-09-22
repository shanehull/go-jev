// Command batch evaluates many states concurrently with Map.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shanehull/go-jev"
)

func main() {
	client, err := jev.New()
	if err != nil {
		log.Fatal(err)
	}

	states := []jev.Input{
		jev.State("The deployment succeeded and all checks passed."),
		jev.State("The deployment failed and rolled back."),
		jev.State("The deployment is still running."),
	}

	results := client.Map(context.Background(), states,
		jev.Questions{"failure": jev.Noul("The text describes a failure")},
		jev.WithConcurrency(3),
		jev.WithMaxRetries(2),
		jev.WithMinInterval(50*time.Millisecond),
	)

	for i, result := range results {
		if result.Err != nil {
			fmt.Printf("%d: error: %v\n", i, result.Err)
			continue
		}
		failure, _ := result.Response.Noul("failure")
		fmt.Printf("%d: failure=%.2f\n", i, failure.Noul)
	}
}
