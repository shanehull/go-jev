// Command structured sends a JSON state with Value.
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

	state := jev.Value(map[string]any{
		"message": "Unit 4 tripped at 16:42 and reserve is thin.",
		"region":  "NSW1",
		"price":   412.5,
	})

	res, err := client.SystemOne(context.Background(), state,
		jev.Questions{
			"material": jev.Noul("The message describes a material supply event"),
			"region": jev.Choice("Which region is affected", map[string]string{
				"NSW1": "New South Wales",
				"VIC1": "Victoria",
				"QLD1": "Queensland",
				"SA1":  "South Australia",
				"TAS1": "Tasmania",
			}),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	material, _ := res.Noul("material")
	region, _ := res.Choice("region")
	fmt.Printf("material: %.2f\n", material.Noul)
	fmt.Printf("region:   %s\n", region.Choice)
}
