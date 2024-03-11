import json
import os
import uuid

import boto3


def handler(event, context):
    messages = event["Records"]
    print(messages)

    dynamodb_client = boto3.client("dynamodb")
    table_name = os.environ["DYNAMODB_TABLE_NAME"]

    for message in messages:
        message_body = json.loads(message["body"])

        record_id = str(uuid.uuid4())

        item = {
            "id": {"S": record_id},
            "product_id": {"S": message_body["product_id"]},
            "location": {"S": message_body["location"]},
            "quantity": {"N": str(message_body["quantity"])},
            "update_date": {"S": message_body["update_date"]},
        }

        dynamodb_client.put_item(TableName=table_name, Item=item)
