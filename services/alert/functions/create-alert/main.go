package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jsmithdenverdev/pager/services/alert"
)

func main() {
	handler := alert.HandleCreate()
	lambda.Start(handler)
}
