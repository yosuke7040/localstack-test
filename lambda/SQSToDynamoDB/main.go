// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log/slog"
// 	"os"
// 	"strconv"

// 	"github.com/aws/aws-lambda-go/events"
// 	"github.com/aws/aws-lambda-go/lambda"
// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/credentials"
// 	"github.com/aws/aws-sdk-go/aws/endpoints"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/dynamodb"
// 	"github.com/google/uuid"

// 	"github.com/guregu/dynamo"
// )

// type messageBody struct {
// 	ProductID  string `json:"product_id"`
// 	Location   string `json:"location"`
// 	Quantity   string `json:"quantity"`
// 	UpdateDate string `json:"update_date"`
// }

// var tableName string
// var dynamoDbClient *dynamodb.DynamoDB

// func init() {
// 	tableName = os.Getenv("DYNAMODB_TABLE_NAME")
// 	if tableName == "" {
// 		slog.Error("missing environment variable DYNAMODB_TABLE_NAME\n")
// 	}
// 	// sess := session.Must(session.NewSession())
// 	sess := session.Must(session.NewSession(&aws.Config{
// 		Credentials:      credentials.NewStaticCredentials("test", "test", ""),
// 		S3ForcePathStyle: aws.Bool(true),
// 		Region:           aws.String(endpoints.UsEast1RegionID),
// 		Endpoint:         aws.String("http://localhost:4566"),
// 	}))
// 	dynamoDbClient = dynamodb.New(sess)
// }

// func handleRequest(ctx context.Context, sqsEvent events.SQSEvent) {
// 	for _, message := range sqsEvent.Records {
// 		var messageBody messageBody
// 		err := json.Unmarshal([]byte(message.Body), &messageBody)
// 		if err != nil {
// 			fmt.Printf("error unmarshalling message body: %s\n", err)
// 			continue
// 		}

// 		// 文字列から整数への変換を試みる
// 		quantity, err := strconv.Atoi(messageBody.Quantity)
// 		if err != nil {
// 			fmt.Printf("error converting quantity to int: %s\n", err)
// 			continue
// 		}

// 		recordID := uuid.New().String()

// 		item := map[string]*dynamodb.AttributeValue{
// 			"id":         {S: aws.String(recordID)},
// 			"product_id": {S: aws.String(messageBody.ProductID)},
// 			"location":   {S: aws.String(messageBody.Location)},
// 			// "quantity":    {N: aws.String(messageBody.Quantity)},
// 			"quantity":    {N: aws.String(fmt.Sprintf("%d", quantity))},
// 			"update_date": {S: aws.String(messageBody.UpdateDate)},
// 		}

// 		_, err = dynamoDbClient.PutItem(&dynamodb.PutItemInput{
// 			TableName: &tableName,
// 			Item:      item,
// 		})
// 		if err != nil {
// 			fmt.Printf("failed to put item in dynamoDB: %s\n", err)
// 			continue
// 		}
// 	}
// }

//	func main() {
//		slog.Info("starting lambda function")
//		lambda.Start(handleRequest)
//	}
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
