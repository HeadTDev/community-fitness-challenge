---
name: fc-aws-localstack
description: AWS Cloud simulation with LocalStack Pro for the Community Fitness Challenge. Use for AWS SDK Go v2 integration (S3, SQS, SES, Secrets Manager), worker polling logic, and local infrastructure setup.
---

# AWS Simulation (LocalStack Pro) - Community Fitness Challenge

This skill ensures professional-grade integration with AWS services via LocalStack Pro, facilitating a seamless transition between local development and production.

## 🛠️ AWS SDK Go v2 Standards

- **Central Configuration (`internal/aws/client.go`)**: Use `NewAWSConfig`.
- **LocalStack Endpoint**: All AWS calls must target `http://localstack:4566` when `AWS_ENDPOINT_URL` is set.
- **S3 Specifics**: Mandatory `UsePathStyle = true` in `s3.Options` for LocalStack compatibility.

## ☁️ Resource Orchestration

- **S3 (Assets)**: Used for user avatars and challenge cover images.
  - Prefix pattern: `avatars/<user_id>/` and `challenges/<challenge_id>/`.
- **SQS (Workers)**: Central message queue for background jobs.
  - **Polling Logic**: 10s Long-polling (`WaitTimeSeconds: 10`), process max 10 messages per batch.
  - **Triggers**: `log_submitted` (Anti-cheat), `send_email` (SES), `gdpr_delete`.
- **SES (Email)**: Simulated via `noreply@fitchallenge.local`. Used for notifications and transactional emails.
- **Secrets Manager**: JWT signing keys and DB credentials must be retrieved during application startup.

## ⚙️ Infrastructure & Tooling

- **LocalStack Pro**: Uses `localstack/localstack-pro:latest` image with `LOCALSTACK_AUTH_TOKEN`.
- **Initialization**: `infra/localstack/init-aws.sh` runs automatically to provision buckets, queues, and secrets.
- **CLI Wrapper**: Use `awslocal` instead of `aws` within the LocalStack container for automatic endpoint targeting.

## 🚀 Transition to Production

- The Go code must be environment-agnostic.
- **Removal of `AWS_ENDPOINT_URL`**: This single change must switch the app from LocalStack to real AWS infrastructure.
