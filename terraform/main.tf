################################
############# SQS ##############
################################
resource "aws_sqs_queue" "inventory_updates_queue" {
  name                       = "InventoryUpdatesQueue"
  visibility_timeout_seconds = 300
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.inventory_updates_dlq.arn
    maxReceiveCount     = 5
  })
}

resource "aws_sqs_queue" "inventory_updates_dlq" {
  name                       = "InventoryUpdatesDlq"
  visibility_timeout_seconds = 300
}

################################
########### Dynamo #############
################################
resource "aws_dynamodb_table" "inventory_updates" {
  name         = "InventoryUpdates"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

################################
############# IAM ##############
################################
resource "aws_iam_role" "lambda_execution_role" {
  name = "lambda_execution_role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_policy" "lambda_policy" {
  name = "lambda_policy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "s3:GetObject",
        ]
        Resource = "*"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy_attachment" {
  role       = aws_iam_role.lambda_execution_role.name
  policy_arn = aws_iam_policy.lambda_policy.arn
}

################################
############ Lambda ############
################################
resource "aws_lambda_function" "csv_processing_to_sqs" {
  function_name = "CSVProcessingToSQS"
  handler       = "main"
  runtime       = "provided.al2"
  role          = aws_iam_role.lambda_execution_role.arn
  s3_bucket     = aws_s3_bucket.lambda_code.bucket
  s3_key        = "lambda_code.zip"

  environment {
    variables = {
      SQS_QUEUE_URL = aws_sqs_queue.inventory_updates_queue.url
    }
  }
}

resource "aws_lambda_function" "sqs_to_dynamodb" {
  function_name = "SQSToDynamoDB"
  handler       = "main"
  runtime       = "provided.al2"
  role          = aws_iam_role.lambda_execution_role.arn
  s3_bucket     = aws_s3_bucket.lambda_code.bucket
  s3_key        = "lambda_code.zip"

  environment {
    variables = {
      DYNAMODB_TABLE_NAME = aws_dynamodb_table.inventory_updates.name
    }
  }
}

resource "aws_lambda_event_source_mapping" "sqs_to_lambda" {
  event_source_arn = aws_sqs_queue.inventory_updates_queue.arn
  function_name    = aws_lambda_function.sqs_to_dynamodb.arn
  batch_size       = 10
}

################################
############# S3 ###############
################################
resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.inventory_updates_bucket.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.csv_processing_to_sqs.arn
    events              = ["s3:ObjectCreated:*"]
  }
}

resource "aws_s3_bucket" "inventory_updates_bucket" {
  bucket = "inventory-updates-bucket"
}

resource "aws_s3_bucket" "lambda_code" {
  bucket = "lambda-code-bucket"
}

resource "aws_s3_object" "lambda_code_object" {
  bucket = aws_s3_bucket.lambda_code.bucket
  key    = "lambda_code.zip"
  source = "./dummy.zip"
  # source = "${path.module}/../lambda/lambda_code.zip"
}
