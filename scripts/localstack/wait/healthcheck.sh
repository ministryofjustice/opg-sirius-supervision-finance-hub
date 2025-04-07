#!/usr/bin/env bash

# S3
buckets=$(awslocal s3 ls)

echo $buckets | grep "opg-backoffice-async-uploads-local" || exit 1
echo $buckets | grep "opg-backoffice-finance-reports-local" || exit 1
