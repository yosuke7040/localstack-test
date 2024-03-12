package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/guregu/dynamo"
)

type messageBody struct {
	ProductID  string `json:"product_id"`
	Location   string `json:"location"`
	Quantity   string `json:"quantity"`
	UpdateDate string `json:"update_date"`
}

type InventoryItem struct {
	ID         string `dynamo:"id"`
	ProductID  string `dynamo:"product_id"`
	Location   string `dynamo:"location"`
	Quantity   int    `dynamo:"quantity"`
	UpdateDate string `dynamo:"update_date"`
}

var (
	tableName string
	db        *dynamo.DB
)

func init() {
	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		log.Fatal("Missing environment variable: DYNAMODB_TABLE_NAME")
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		// Profile:           "localstack",
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Endpoint:   aws.String("http://localstack:4566"),
			DisableSSL: aws.Bool(true),
		},
	}))
	db = dynamo.New(sess)
}

func handleRequest(ctx context.Context, sqsEvent events.SQSEvent) {
	table := db.Table(tableName)

	for _, message := range sqsEvent.Records {
		var mb messageBody
		err := json.Unmarshal([]byte(message.Body), &mb)
		if err != nil {
			log.Printf("error unmarshalling message body: %s\n", err)
			continue
		}

		quantity, err := strconv.Atoi(mb.Quantity)
		if err != nil {
			log.Printf("error converting quantity to int: %s\n", err)
			continue
		}

		item := InventoryItem{
			ID:         uuid.New().String(),
			ProductID:  mb.ProductID,
			Location:   mb.Location,
			Quantity:   quantity,
			UpdateDate: mb.UpdateDate,
		}

		err = table.Put(item).Run()
		if err != nil {
			log.Printf("failed to put item in DynamoDB: %s\n", err)
			continue
		}
	}
}

func main() {
	log.Println("Starting lambda function")
	lambda.Start(handleRequest)
}
