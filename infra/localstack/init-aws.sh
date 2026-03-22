#!/bin/bash
set -e

echo "------------------------------------------------------------"
echo "🚀 Initializing LocalStack AWS Resources..."
echo "------------------------------------------------------------"

# --- S3 Buckets ---
echo "📦 Creating S3 buckets..."
awslocal s3 mb s3://fitchallenge-assets

# CORS Configuration for S3
cat <<EOF > /tmp/cors-config.json
{
  "CORSRules": [
    {
      "AllowedOrigins": ["*"],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE", "HEAD"],
      "AllowedHeaders": ["*"],
      "ExposeHeaders": ["ETag"]
    }
  ]
}
EOF
awslocal s3api put-bucket-cors --bucket fitchallenge-assets --cors-configuration file:///tmp/cors-config.json

# --- SQS Queues ---
echo "📨 Creating SQS queues..."
awslocal sqs create-queue --queue-name fitchallenge-jobs-dlq
awslocal sqs create-queue --queue-name fitchallenge-jobs --attributes '{
  "RedrivePolicy": "{\"deadLetterTargetArn\":\"arn:aws:sqs:us-east-1:000000000000:fitchallenge-jobs-dlq\",\"maxReceiveCount\":\"5\"}",
  "VisibilityTimeout": "30"
}'

# --- SNS Topics ---
echo "🔔 Creating SNS topics..."
awslocal sns create-topic --name fitchallenge-notifications

# --- SES Identities ---
echo "📧 Verifying SES email identities..."
awslocal ses verify-email-identity --email-address noreply@fitchallenge.local

# --- Secrets Manager ---
echo "🔐 Provisioning secrets..."
awslocal secretsmanager create-secret \
    --name fitchallenge/jwt-secret \
    --description "JWT signing key for the API" \
    --secret-string "fitchallenge_super_secret_dev_key_2026"

awslocal secretsmanager create-secret \
    --name fitchallenge/db-credentials \
    --description "Database credentials for the API" \
    --secret-string '{"username":"fc_user","password":"fc_password","dbname":"fitchallenge"}'

echo "------------------------------------------------------------"
echo "✅ AWS Resources Provisioned Successfully!"
echo "------------------------------------------------------------"
awslocal s3 ls
awslocal sqs list-queues
awslocal secretsmanager list-secrets --query 'SecretList[].Name'
