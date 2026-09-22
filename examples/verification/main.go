// Command verification checks a draft answer against its source, a guardrail
// pattern that catches unsupported or fabricated claims.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/shanehull/go-jev"
)

func main() {
	client, err := jev.New()
	if err != nil {
		log.Fatal(err)
	}

	source := "The notice reports a unit trip in South Australia at 16:42. " +
		"No load shedding was required."
	draft := "A unit tripped in South Australia at 16:42, causing load shedding."

	state := jev.Value(map[string]any{"source": source, "draft": draft})

	res, err := client.SystemOne(context.Background(), state,
		jev.Questions{
			"supported":  jev.Noul("Every claim in the draft is supported by the source"),
			"fabricated": jev.Noul("The draft introduces information not present in the source"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	supported, _ := res.Noul("supported")
	fabricated, _ := res.Noul("fabricated")
	fmt.Printf("supported:  %.2f\n", supported.Noul)
	fmt.Printf("fabricated: %.2f\n", fabricated.Noul)
	if fabricated.Noul > 0.5 {
		fmt.Println("verdict: review, the draft may contain unsupported claims")
	}
}
