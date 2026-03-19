#!/bin/bash
set -e

GREEN='\033[0-9;32m'
RED='\033[0-9;31m'
NC='\033[0m' # No Color

echo "🔍 Starting Infrastructure Verification..."
echo "------------------------------------------------------------"

# 1. Check Docker Containers Status
echo -n "🐳 Checking Docker containers status: "
CONTAINERS=$(docker compose ps --format json)
if [[ $CONTAINERS == *"healthy"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Some containers are not healthy)${NC}"
    docker compose ps
fi

# 2. Check PostgreSQL
echo -n "🐘 Checking PostgreSQL (users table): "
PG_CHECK=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\dt" | grep users || true)
if [[ $PG_CHECK == *"users"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Table 'users' not found)${NC}"
fi

# 3. Check Redis
echo -n "🔴 Checking Redis (PING): "
REDIS_CHECK=$(docker compose exec -T redis redis-cli ping | tr -d '\r')
if [[ $REDIS_CHECK == "PONG" ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Redis did not respond with PONG, got: '$REDIS_CHECK')${NC}"
fi

# 4. Check LocalStack S3
echo -n "📦 Checking LocalStack S3 (fitchallenge-assets): "
S3_CHECK=$(docker compose exec -T localstack awslocal s3 ls | grep fitchallenge-assets || true)
if [[ $S3_CHECK == *"fitchallenge-assets"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (S3 Bucket not found)${NC}"
fi

# 5. Check LocalStack SQS
echo -n "📨 Checking LocalStack SQS (fitchallenge-jobs): "
SQS_CHECK=$(docker compose exec -T localstack awslocal sqs list-queues | grep fitchallenge-jobs || true)
if [[ $SQS_CHECK == *"fitchallenge-jobs"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (SQS Queue not found)${NC}"
fi

# 6. Check LocalStack Secrets Manager
echo -n "🔐 Checking LocalStack Secrets (jwt-secret): "
SECRET_CHECK=$(docker compose exec -T localstack awslocal secretsmanager list-secrets | grep fitchallenge/jwt-secret || true)
if [[ $SECRET_CHECK == *"fitchallenge/jwt-secret"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Secret 'fitchallenge/jwt-secret' not found)${NC}"
fi

echo "------------------------------------------------------------"
echo "✅ Verification complete."
