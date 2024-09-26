package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jsmithdenverdev/pager/services/agency/internal/handlers"
)

var (
	Version string
)

func main() {
	if err := run(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(stdout io.Writer) error {
	fmt.Fprintf(stdout, "Version %s", Version)

	logger := slog.New(slog.NewJSONHandler(stdout, nil))

	lambda.StartWithOptions(handlers.CreateAgency(logger))
	return nil
}
