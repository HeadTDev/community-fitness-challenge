# Community Fitness Challenge 🚀

A real-time, AWS-powered fitness challenge platform with iOS and Go backend.

## 🏗️ Architecture
- **Backend:** Go 1.26 (Clean Architecture)
- **iOS:** SwiftUI (iOS 17+)
- **Infra:** LocalStack (S3, SQS, SES, Secrets Manager)
- **Database:** PostgreSQL 18 + Redis 8

## 📜 License
This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🛠️ Local Setup
1. Copy `.env.example` to `.env` and fill in `LOCALSTACK_AUTH_TOKEN`.
2. Run `docker compose up -d`.
3. Follow the instructions in `docs/local-setup.md`.
