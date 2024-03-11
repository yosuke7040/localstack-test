terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.6"
    }
  }

  backend "local" {}
}

provider "aws" {
  region = "us-east-1"
  # profile    = "localstack"
  access_key = "test"
  secret_key = "test"

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true

  endpoints {
    s3       = "http://localhost:4566"
    dynamodb = "http://localhost:4566"
    lambda   = "http://localhost:4566"
    sns      = "http://localhost:4566"
    sqs      = "http://localhost:4566"
    iam      = "http://localhost:4566"
  }
}
