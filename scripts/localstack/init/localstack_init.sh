#! /usr/bin/env bash
create_bucket() {
    BUCKET=$1
    # Create Private Bucket
    awslocal s3api create-bucket \
        --acl private \
        --region eu-west-1 \
        --create-bucket-configuration LocationConstraint=eu-west-1 \
        --bucket "$BUCKET"

    # Add Public Access Block
    awslocal s3api put-public-access-block \
        --public-access-block-configuration "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true" \
        --bucket "$BUCKET"

    # Add Default Encryption
    awslocal s3api put-bucket-encryption \
        --bucket "$BUCKET" \
        --server-side-encryption-configuration '{ "Rules": [ { "ApplyServerSideEncryptionByDefault": { "SSEAlgorithm": "AES256" } } ] }'

    # Add Encryption Policy
    awslocal s3api put-bucket-policy \
        --policy '{ "Statement": [ { "Sid": "DenyUnEncryptedObjectUploads", "Effect": "Deny", "Principal": { "AWS": "*" }, "Action": "s3:PutObject", "Resource": "arn:aws:s3:::'${BUCKET}'/*", "Condition":  { "StringNotEquals": { "s3:x-amz-server-side-encryption": "AES256" } } }, { "Sid": "DenyUnEncryptedObjectUploads", "Effect": "Deny", "Principal": { "AWS": "*" }, "Action": "s3:PutObject", "Resource": "arn:aws:s3:::'${BUCKET}'/*", "Condition":  { "Bool": { "aws:SecureTransport": false } } } ] }' \
        --bucket "$BUCKET"
}


# S3
create_bucket "opg-backoffice-async-uploads-local"
awslocal s3 cp /scripts/files/feemoto_01042025normal.csv s3://opg-backoffice-async-uploads-local/

create_bucket "opg-backoffice-reports-local"
awslocal s3api put-bucket-versioning --bucket "opg-backoffice-reports-local" --versioning-configuration Status=Enabled

