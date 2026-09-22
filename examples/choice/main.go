// Command choice routes a message to a team with a choice question.
package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/shanehull/go-jev"
)

func main() {
	client, err := jev.New()
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.SystemOne(context.Background(),
		jev.State("The invoice shows the wrong GST amount for last month."),
		jev.Questions{
			"team": jev.Choice("Which team should handle this", map[string]string{
				"billing":   "Payment, invoice, or subscription issues",
				"technical": "Bugs, outages, or integration problems",
				"sales":     "Pricing, plans, or account questions",
			}),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	team, _ := res.Choice("team")
	fmt.Printf("chosen: %s (%.2f confidence)\n", team.Choice, team.Confidence)

	options := make([]string, 0, len(team.Probabilities))
	for option := range team.Probabilities {
		options = append(options, option)
	}
	sort.Strings(options)
	for _, option := range options {
		fmt.Printf("  %-10s %.2f\n", option, team.Probabilities[option])
	}
}
