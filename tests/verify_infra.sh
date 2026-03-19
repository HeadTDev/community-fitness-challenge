#!/bin/bash
set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

FAILED=0

echo "🔍 Starting Full Project Verification (Day 1-7)..."
echo "------------------------------------------------------------"

# 1. Check Docker Containers Status
echo -n "🐳 Checking Docker containers status: "
CONTAINERS=$(docker compose ps --format json)
if [[ $CONTAINERS == *"healthy"* ]] && [[ $CONTAINERS == *"fc-api"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Some containers are not healthy or missing)${NC}"
    docker compose ps
    FAILED=$((FAILED + 1))
fi

# 2. Check Database Migrations (Day 5)
echo -n "🐘 Checking Database Migrations (schema_migrations): "
MIG_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\dt" | grep schema_migrations || true)
if [[ $MIG_TABLE == *"schema_migrations"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Table 'schema_migrations' not found - Migrations not run?)${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "🐘 Checking PostgreSQL (users table with Day 5 fields): "
# Check if users table has the 'role' column which was added in Day 5 migration
USER_ROLE_CHECK=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\d users" | grep role || true)
if [[ $USER_ROLE_CHECK == *"role"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Table 'users' missing 'role' column from migration)${NC}"
    FAILED=$((FAILED + 1))
fi

# 3. Check User Repository (Day 6)
echo -n "👤 Checking User Repository Integration Test: "
REPO_TEST=$(docker compose exec -T api go test ./internal/adapter/postgres/ -v || true)
if [[ $REPO_TEST == *"PASS"* ]] && [[ $REPO_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (User Repository tests failed)${NC}"
    echo "$REPO_TEST"
    FAILED=$((FAILED + 1))
fi

# 4. Check AWS S3 Client (Day 7)
echo -n "☁️  Checking AWS S3 Client Integration Test: "
S3_TEST=$(docker compose exec -T api go test ./internal/aws/ -v || true)
if [[ $S3_TEST == *"PASS"* ]] && [[ $S3_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (AWS S3 Client tests failed)${NC}"
    echo "$S3_TEST"
    FAILED=$((FAILED + 1))
fi

# 5. Check Redis
echo -n "🔴 Checking Redis (PING): "
REDIS_CHECK=$(docker compose exec -T redis redis-cli ping | tr -d '\r')
if [[ $REDIS_CHECK == "PONG" ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Redis did not respond with PONG, got: '$REDIS_CHECK')${NC}"
    FAILED=$((FAILED + 1))
fi

# 4. Check LocalStack Resources (S3, SQS, Secrets)
echo -n "☁️  Checking LocalStack S3 (fitchallenge-assets): "
S3_CHECK=$(docker compose exec -T localstack awslocal s3 ls | grep fitchallenge-assets || true)
if [[ $S3_CHECK == *"fitchallenge-assets"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (S3 Bucket not found)${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "📨 Checking LocalStack SQS (fitchallenge-jobs): "
SQS_CHECK=$(docker compose exec -T localstack awslocal sqs list-queues | grep fitchallenge-jobs || true)
if [[ $SQS_CHECK == *"fitchallenge-jobs"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (SQS Queue not found)${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "🔐 Checking LocalStack Secrets (jwt-secret): "
SECRET_CHECK=$(docker compose exec -T localstack awslocal secretsmanager list-secrets | grep fitchallenge/jwt-secret || true)
if [[ $SECRET_CHECK == *"fitchallenge/jwt-secret"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Secret 'fitchallenge/jwt-secret' not found)${NC}"
    FAILED=$((FAILED + 1))
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
    FAILED=$((FAILED + 1))
fi

echo -n "🚀 Checking Go Backend (/readyz with DB status): "
API_READY=$(docker compose exec -T localstack curl -s http://api:8080/readyz || true)
if [[ $API_READY == *"\"success\":true"* ]] && [[ $API_READY == *"\"status\":\"ready\""* ]] && [[ $API_READY == *"\"db\":\"ok\""* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (API /readyz not responding correctly or DB is not ok)${NC}"
    echo "Response: $API_READY"
    FAILED=$((FAILED + 1))
fi

# 6. Full Integration Test (Day 7 Optimization)
echo -n "🔄 Checking Full Integration Test (DB + S3): "
INT_TEST=$(docker compose exec -T api go test ./internal/integration/ -v || true)
if [[ $INT_TEST == *"PASS"* ]] && [[ $INT_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL (Full Integration tests failed)${NC}"
    echo "$INT_TEST"
    FAILED=$((FAILED + 1))
fi

echo "------------------------------------------------------------"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed. Day 7 Optimization is solid.${NC}"
    exit 0
else
    echo -e "${RED}❌ $FAILED test(s) failed. Please check the logs above.${NC}"
    exit 1
fi

