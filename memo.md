# MEMO

## Command memo

```bash
curl -s localhost:4566/_localstack/init | jq .

awslocal s3 mb s3://localstack-bucket
awslocal s3 ls
awslocal s3 ls s3://lambda-code-bucket
awslocal lambda list-functions
awslocal logs describe-log-groups
```

## Tips

ログの確認は一度Lambdaを起動しないと出来ない。起動後にlog groupが作られる

## TODO

goのdynamo部分の書き方の修正
<https://docs.aws.amazon.com/ja_jp/code-library/latest/ug/go_2_dynamodb_code_examples.html>
