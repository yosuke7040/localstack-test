# localstack-test

## Requirements

- golang 1.22
- Terraform 1.6以上
- Task(<https://taskfile.dev/>)
- AWS CLI(awslocal)
- LocalStack CLI

## Resources Deploy

```bash
cd terraform
terarform init
terraform apply
```

## Test

```bash
awslocal s3 cp data/sample.csv s3://inventory-updates-bucket
```
