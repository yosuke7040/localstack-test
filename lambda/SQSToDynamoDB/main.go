package main

import (
	"log/slog"

	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest() {
	slog.Info("Hello from Lambda!")
}

func main() {
	lambda.Start(handleRequest)
}
