// Command systemone classifies a support ticket across the three primitives.
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

	res, err := client.SystemOne(context.Background(),
		jev.State("Hi, I've been trying to connect my Stripe account for 3 days and the integration keeps failing. I'm losing sales. Please help ASAP."),
		jev.Questions{
			"department": jev.Choice("Which team should handle this", map[string]string{
				"billing":   "Payment or subscription issues",
				"technical": "Bugs or integration problems",
				"sales":     "Pricing or account questions",
			}),
			"frustration": jev.Score("How frustrated the customer appears", []string{
				"Calm, just stating facts",
				"Frustrated but civil",
				"Very angry, strong language",
			}),
			"is_urgent": jev.Noul("The message conveys urgency or time-sensitivity"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	department, _ := res.Choice("department")
	frustration, _ := res.Score("frustration")
	urgent, _ := res.Noul("is_urgent")
	fmt.Printf("department:  %s (%.2f)\n", department.Choice, department.Confidence)
	fmt.Printf("frustration: %.0f (%.2f)\n", frustration.Score, frustration.Confidence)
	fmt.Printf("urgent:      %.2f\n", urgent.Noul)
}
