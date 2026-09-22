package jev_test

import (
	"context"
	"os"
	"testing"

	"github.com/shanehull/go-jev"
)

func TestLiveSystemOne(t *testing.T) {
	if os.Getenv("TYPESAFE_API_KEY") == "" {
		t.Skip("set TYPESAFE_API_KEY to run live tests")
	}

	client, err := jev.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := client.SystemOne(context.Background(),
		jev.State("I have been trying to connect my account for three days and it keeps failing. I am losing sales."),
		jev.Questions{
			"urgent": jev.Noul("The message conveys urgency or time sensitivity"),
		},
	)
	if err != nil {
		t.Fatalf("SystemOne: %v", err)
	}
	if _, ok := res.Noul("urgent"); !ok {
		t.Fatalf("missing noul answer: %+v", res.Answers)
	}
}
