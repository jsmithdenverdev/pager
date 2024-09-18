package main

import (
	"fmt"
	"os"
)

var (
	Version string
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	fmt.Fprintf(os.Stdout, "Version %s", Version)
	// lambda.StartWithOptions(handlers.CreateAgency())
	return nil
}
