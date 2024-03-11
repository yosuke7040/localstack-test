import csv
import json
import os

import boto3


def handler(event, context):
    print(event)
    print("--------------------------------------")

    record = event["Records"][0]

    source_bucket = record["s3"]["bucket"]["name"]
    key = record["s3"]["object"]["key"]

    s3_client = boto3.client("s3")
    response = s3_client.get_object(Bucket=source_bucket, Key=key)
    csv_content = response["Body"].read().decode("utf-8-sig")

    sqs_client = boto3.client("sqs")
    queue_url = os.environ["SQS_QUEUE_URL"]

    csv_reader = csv.DictReader(csv_content.splitlines())
    message_batch = []
    for row in csv_reader:
        json_message = json.dumps(row)

        message_batch.append(
            {"Id": str(len(message_batch) + 1), "MessageBody": json_message}
        )

        if len(message_batch) == 10:
            sqs_client.send_message_batch(QueueUrl=queue_url, Entries=message_batch)
            message_batch = []
            print("Sent messages in batch")

    if message_batch:
        sqs_client.send_message_batch(QueueUrl=queue_url, Entries=message_batch)
