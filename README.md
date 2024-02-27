# localstack-test

## Requirements

```bash
brew install localstack/tap/localstack-cli
pip3 install awscli-local
```

## Command memo

```bash
curl -s localhost:4566/_localstack/init | jq .

awslocal s3 mb s3://localstack-bucket
awslocal s3 ls
```
