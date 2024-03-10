package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jsmithdenverdev/pager/services/alert"
)

func main() {
	handler := alert.HandleAlert()
	lambda.Start(handler)
}
