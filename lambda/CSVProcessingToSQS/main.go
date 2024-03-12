package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// 参考：https://github.com/aws/aws-lambda-go/blob/main/events/README.md
// https://docs.localstack.cloud/user-guide/integrations/sdks/go/

var (
	s3Client  *s3.Client
	sqsClient *sqs.Client
)

func init() {
	// Lambdaは別コンテナとして起動されるので、localhostではなくlocalstackを指定する
	awsEndpoint := "http://localstack:4566"

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			// カスタムエンドポイント使用
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: "us-east-1",
				// これがないとURLにバケット名を付与してしまい、エラーになる
				HostnameImmutable: true,
			}, nil
		}
		// 通常のエンドポイント使用
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))), config.WithCredentialsProvider(aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))),
	)
	if err != nil {
		log.Fatalf("Cannot load the AWS configs: %s", err)
	}

	s3Client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// o.UsePathStyle = true
	})
	sqsClient = sqs.NewFromConfig(awsCfg)
}

func handleRequest(event events.S3Event) {
	for _, record := range event.Records {
		s3Event := record.S3
		sourceBucket := s3Event.Bucket.Name
		key := s3Event.Object.Key

		result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(sourceBucket),
			Key:    aws.String(key),
		})
		defer result.Body.Close()

		if err != nil {
			slog.Error("Error getting object: ",
				err,
				slog.String("sourceBucket", sourceBucket),
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
		slog.Info("Read CSV records", slog.Any("records", records))

		messageBatch := []types.SendMessageBatchRequestEntry{}
		queueURL := os.Getenv("SQS_QUEUE_URL")

		headers := records[0]

		for i, record := range records[1:] {
			messageBody := make(map[string]string)
			for i, value := range record {
				messageBody[headers[i]] = value
			}

			jsonMessage, _ := json.Marshal(messageBody)
			slog.Info("Sending message", slog.String("jsonMessage", string(jsonMessage)))
			messageBatch = append(messageBatch, types.SendMessageBatchRequestEntry{
				Id:          aws.String(strconv.Itoa(i + 1)),
				MessageBody: aws.String(string(jsonMessage)),
			})

			if len(messageBatch) == 10 {
				_, err := sqsClient.SendMessageBatch(
					context.TODO(),
					&sqs.SendMessageBatchInput{
						Entries:  messageBatch,
						QueueUrl: aws.String(queueURL),
					},
				)

				if err != nil {
					slog.Error("Error sending 10 messages batch",
						err,
						slog.String("queueURL", queueURL),
					)
					return
				}

				messageBatch = []types.SendMessageBatchRequestEntry{}
				slog.Info("Sent messages in batch")
			}
		}

		if len(messageBatch) > 0 {
			_, err := sqsClient.SendMessageBatch(
				context.TODO(),
				&sqs.SendMessageBatchInput{
					Entries:  messageBatch,
					QueueUrl: aws.String(queueURL),
				},
			)

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
