package jev_test

import (
	"context"
	"fmt"

	"github.com/shanehull/go-jev"
)

func ExampleClient_SystemOne() {
	client, _ := jev.New()
	res, _ := client.SystemOne(context.Background(), jev.State("I was charged twice, please fix this ASAP."),
		jev.Questions{
			"category": jev.Choice("What is this about", map[string]string{
				"billing":   "Payment or subscription issues",
				"technical": "Bugs or integration problems",
				"other":     "Anything else",
			}),
			"urgent": jev.Noul("The message conveys urgency"),
		},
	)

	category, _ := res.Choice("category")
	urgent, _ := res.Noul("urgent")
	fmt.Println(category.Choice, urgent.Noul)
}

func ExampleClient_Map() {
	client, _ := jev.New()
	states := []jev.Input{
		jev.State("The system is down"),
		jev.State("Thanks for the quick fix"),
	}
	results := client.Map(context.Background(), states,
		jev.Questions{"negative": jev.Noul("The text expresses a negative experience")},
		jev.WithConcurrency(2),
	)
	for _, result := range results {
		if result.Err == nil {
			noul, _ := result.Response.Noul("negative")
			fmt.Println(noul.Noul)
		}
	}
}

func ExampleChoice() {
	client, _ := jev.New()
	res, _ := client.SystemOne(context.Background(), jev.State("I was charged twice."),
		jev.Questions{
			"category": jev.Choice("What is this about", map[string]string{
				"billing":   "Payment or invoice issues",
				"technical": "Bugs or integration problems",
			}),
		})

	category, _ := res.Choice("category")
	fmt.Println(category.Choice, category.Confidence, category.Probabilities)
}

func ExampleScore() {
	client, _ := jev.New()
	res, _ := client.SystemOne(context.Background(), jev.State("The site is completely down."),
		jev.Questions{
			"severity": jev.Score("How severe is the incident", []string{
				"Minor", "Degraded", "Major", "Outage",
			}),
		})

	severity, _ := res.Score("severity")
	fmt.Println(severity.Score, severity.Legend)
}

func ExampleNoul() {
	client, _ := jev.New()
	res, _ := client.SystemOne(context.Background(), jev.State("Please fix this now, customers are blocked."),
		jev.Questions{
			"urgent": jev.Noul("The message conveys urgency"),
		})

	urgent, _ := res.Noul("urgent")
	fmt.Println(urgent.Noul)
}

func ExampleValue() {
	client, _ := jev.New()
	state := jev.Value(map[string]any{"text": "The build failed.", "env": "production"})
	res, _ := client.SystemOne(context.Background(), state,
		jev.Questions{"failure": jev.Noul("The text describes a failure")},
	)

	failure, _ := res.Noul("failure")
	fmt.Println(failure.Noul)
}
