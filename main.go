package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, event json.RawMessage) error {
	log.Printf("Successfully processed. Event: %s", string(event))
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
