// Command noul returns the probability that a statement about a text is true.
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

	text := "The report is late again. This is the third delay this quarter."

	res, err := client.SystemOne(context.Background(), jev.State(text),
		jev.Questions{
			"frustrated": jev.Noul("The author is frustrated"),
			"action":     jev.Noul("The text asks for something to be done"),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	frustrated, _ := res.Noul("frustrated")
	action, _ := res.Noul("action")
	fmt.Printf("frustrated: %.2f\n", frustrated.Noul)
	fmt.Printf("action:     %.2f\n", action.Noul)
}
