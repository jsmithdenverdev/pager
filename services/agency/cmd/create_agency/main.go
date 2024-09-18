package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jsmithdenverdev/pager/services/agency/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
		// force build
	}
}

func run() error {
	lambda.StartWithOptions(handlers.CreateAgency())
	// test
	return nil
}
