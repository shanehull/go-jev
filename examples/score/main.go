// Command score rates an incident note on an ordered severity rubric.
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

	note := "Checkout is returning 500s for every request. No orders have processed in twenty minutes."

	res, err := client.SystemOne(context.Background(), jev.State(note),
		jev.Questions{
			"severity": jev.Score("How severe is the incident", []string{
				"Cosmetic or minor",
				"Degraded but usable",
				"Major function unavailable",
				"Total outage",
			}),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	severity, _ := res.Score("severity")
	fmt.Printf("score: %.0f (%.2f confidence)\n", severity.Score, severity.Confidence)
	fmt.Println("label:", severity.Legend[fmt.Sprintf("%.0f", severity.Score)])
}
