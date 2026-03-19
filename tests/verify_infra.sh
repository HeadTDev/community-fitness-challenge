#!/bin/bash
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🔍 Starting Full Project Verification (Day 1-4)..."
echo "------------------------------------------------------------"

# 1. Check Docker Containers Status
echo -n "🐳 Checking Docker containers status: "
CONTAINERS=$(docker compose ps --format json)
if [[ $CONTAINERS == *"healthy"* ]] && [[ $CONTAINERS == *"fc-api"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Some containers are not healthy or missing)${NC}"
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

# 4. Check LocalStack Resources (S3, SQS, Secrets)
echo -n "☁️  Checking LocalStack S3 (fitchallenge-assets): "
S3_CHECK=$(docker compose exec -T localstack awslocal s3 ls | grep fitchallenge-assets || true)
if [[ $S3_CHECK == *"fitchallenge-assets"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (S3 Bucket not found)${NC}"
fi

echo -n "📨 Checking LocalStack SQS (fitchallenge-jobs): "
SQS_CHECK=$(docker compose exec -T localstack awslocal sqs list-queues | grep fitchallenge-jobs || true)
if [[ $SQS_CHECK == *"fitchallenge-jobs"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (SQS Queue not found)${NC}"
fi

echo -n "🔐 Checking LocalStack Secrets (jwt-secret): "
SECRET_CHECK=$(docker compose exec -T localstack awslocal secretsmanager list-secrets | grep fitchallenge/jwt-secret || true)
if [[ $SECRET_CHECK == *"fitchallenge/jwt-secret"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Secret 'fitchallenge/jwt-secret' not found)${NC}"
fi

# 5. Check Go Backend API (Day 4)
echo -n "🚀 Checking Go Backend (/healthz): "
# We use localstack container to curl the api container to test internal network
API_HEALTH=$(docker compose exec -T localstack curl -s http://api:8080/healthz || true)
if [[ $API_HEALTH == *"\"success\":true"* ]] && [[ $API_HEALTH == *"\"status\":\"ok\""* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (API /healthz not responding correctly)${NC}"
    echo "Response: $API_HEALTH"
fi

echo -n "🚀 Checking Go Backend (/readyz): "
API_READY=$(docker compose exec -T localstack curl -s http://api:8080/readyz || true)
if [[ $API_READY == *"\"success\":true"* ]] && [[ $API_READY == *"\"status\":\"ready\""* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (API /readyz not responding correctly)${NC}"
    echo "Response: $API_READY"
fi

echo "------------------------------------------------------------"
echo "✅ All tests passed. Day 4 is solid."
