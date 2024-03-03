package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"log/slog"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// 参考：https://github.com/aws/aws-lambda-go/blob/main/events/README.md

func handleRequest(event events.S3Event) {
	sess := session.Must(session.NewSession())
	s3Client := s3.New(sess)
	sqsClient := sqs.New(sess)

	for _, record := range event.Records {
		s3Event := record.S3
		sourceBucket := s3Event.Bucket.Name
		key := s3Event.Object.Key

		result, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(sourceBucket),
			Key:    aws.String(key),
		})

		if err != nil {
			slog.Error("Error getting object",
				err,
				slog.String("key", key),
				slog.String("bucket", sourceBucket),
			)
			return
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(result.Body)
		csvContent := buf.String()

		reader := csv.NewReader(bytes.NewBufferString(csvContent))
		records, _ := reader.ReadAll()

		messageBatch := []*sqs.SendMessageBatchRequestEntry{}
		queueURL := os.Getenv("SQS_QUEUE_URL")

		for i, record := range records {
			jsonMessage, _ := json.Marshal(record)
			messageBatch = append(messageBatch, &sqs.SendMessageBatchRequestEntry{
				Id:          aws.String(strconv.Itoa(i + 1)),
				MessageBody: aws.String(string(jsonMessage)),
			})

			if len(messageBatch) == 10 {
				_, err := sqsClient.SendMessageBatch(&sqs.SendMessageBatchInput{
					Entries:  messageBatch,
					QueueUrl: aws.String(queueURL),
				})

				if err != nil {
					slog.Error("Error sending message batch",
						err,
						slog.String("queueURL", queueURL),
					)
					return
				}

				messageBatch = []*sqs.SendMessageBatchRequestEntry{}
				slog.Info("Sent messages in batch")
			}
		}

		if len(messageBatch) > 0 {
			_, err := sqsClient.SendMessageBatch(&sqs.SendMessageBatchInput{
				Entries:  messageBatch,
				QueueUrl: aws.String(queueURL),
			})

			if err != nil {
				slog.Error("Error sending message batch",
					err,
					slog.String("queueURL", queueURL),
				)
				return
			}
		}
	}
}

func main() {
	lambda.Start(handleRequest)
}
