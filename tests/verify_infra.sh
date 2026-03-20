#!/bin/bash
set -e

GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

FAILED=0

echo -e "${CYAN}============================================================${NC}"
echo -e "${CYAN}🚀 COMMUNITY FITNESS CHALLENGE - FULL PROJECT VERIFICATION${NC}"
echo -e "${CYAN}📅 Progress: Day 1 to Day 11${NC}"
echo -e "${CYAN}============================================================${NC}"

# --- Day 1-2-3: Infrastructure Bases ---
echo -e "\n${YELLOW}[Day 1-3: Infrastructure & Environment]${NC}"

echo -n "  🐳 Docker Containers Status: "
CONTAINERS=$(docker compose ps --format json)
if [[ $CONTAINERS == *"healthy"* ]] && [[ $CONTAINERS == *"fc-api"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  🔴 Redis Connection (PING): "
REDIS_CHECK=$(docker compose exec -T redis redis-cli ping | tr -d '\r')
if [[ $REDIS_CHECK == "PONG" ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Day 4: API Skeleton ---
echo -e "\n${YELLOW}[Day 4: Go Backend Skeleton]${NC}"

echo -n "  🚀 Waiting for API to be ready: "
MAX_RETRIES=15
COUNT=0
until $(docker compose exec -T localstack curl -s http://api:8080/healthz | grep -q "\"success\":true"); do
    echo -n "."
    sleep 2
    COUNT=$((COUNT + 1))
    if [ $COUNT -eq $MAX_RETRIES ]; then
        echo -e "${RED} TIMEOUT${NC}"
        FAILED=$((FAILED + 1))
        break
    fi
done
if [ $COUNT -lt $MAX_RETRIES ]; then
    echo -e "${GREEN} READY${NC}"
fi

echo -n "  🚀 API Health Check (/healthz): "
API_HEALTH=$(docker compose exec -T localstack curl -s http://api:8080/healthz || true)
if [[ $API_HEALTH == *"\"success\":true"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Day 5: Database Migrations ---
echo -e "\n${YELLOW}[Day 5: Database Migration System]${NC}"

echo -n "  🐘 Migration Table Exists: "
MIG_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\dt" | grep schema_migrations || true)
if [[ $MIG_TABLE == *"schema_migrations"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  🐘 Users Table Schema (Role column): "
USER_ROLE_CHECK=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\d users" | grep role || true)
if [[ $USER_ROLE_CHECK == *"role"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Day 6: User Domain ---
echo -e "\n${YELLOW}[Day 6: User Domain & Repository]${NC}"

echo -n "  👤 User Repository Integration Tests: "
REPO_TEST=$(docker compose exec -T api go test ./internal/adapter/postgres/ -v || true)
if [[ $REPO_TEST == *"PASS"* ]] && [[ $REPO_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Day 7-8: AWS Integration ---
echo -e "\n${YELLOW}[Day 7-8: AWS SDK & Clients (S3, SQS, SES, Secrets)]${NC}"

echo -n "  ☁️  AWS Clients Integration Tests: "
AWS_TEST=$(docker compose exec -T api go test ./internal/aws/ -v || true)
if [[ $AWS_TEST == *"PASS"* ]] && [[ $AWS_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  📦 LocalStack S3 Bucket: "
S3_CHECK=$(docker compose exec -T localstack awslocal s3 ls | grep fitchallenge-assets || true)
if [[ $S3_CHECK == *"fitchallenge-assets"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  📨 LocalStack SQS Queue: "
SQS_CHECK=$(docker compose exec -T localstack awslocal sqs list-queues | grep fitchallenge-jobs || true)
if [[ $SQS_CHECK == *"fitchallenge-jobs"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Day 9: JWT System ---
echo -e "\n${YELLOW}[Day 9: JWT Token System]${NC}"

echo -n "  🔑 JWT Manager Unit Tests: "
JWT_TEST=$(docker compose exec -T api go test ./internal/pkg/jwt/ -v || true)
if [[ $JWT_TEST == *"PASS"* ]] && [[ $JWT_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Day 10-11: Auth & Middleware & Standard Response ---
echo -e "\n${YELLOW}[Day 10-11: Auth Handler, Middleware & Standard Response]${NC}"

echo -n "  🔐 Register Dev (Get Token): "
AUTH_RESPONSE=$(docker compose exec -T localstack curl -s -X POST http://api:8080/auth/register-dev || true)
TOKEN=$(echo "$AUTH_RESPONSE" | grep -oP '"access_token":"\K[^"]+' || true)
if [[ $AUTH_RESPONSE == *"\"success\":true"* ]] && [[ -n "$TOKEN" ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  🆔 Request ID Middleware (Header): "
RID_HEADER=$(docker compose exec -T localstack curl -s -i http://api:8080/healthz | tr -d '\r' | grep -i "X-Request-ID" || true)
if [[ -n "$RID_HEADER" ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  🆔 Request ID in JSON Response (Meta): "
RID_JSON=$(docker compose exec -T localstack curl -s http://api:8080/healthz || true)
if [[ $RID_JSON == *"\"request_id\":"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  🔐 Protected Route Access (/v1/users/me): "
ME_RESPONSE=$(docker compose exec -T localstack curl -s -H "Authorization: Bearer $TOKEN" http://api:8080/v1/users/me || true)
if [[ $ME_RESPONSE == *"\"success\":true"* ]] && [[ $ME_RESPONSE == *"\"user_id\":"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  🔐 Protected Route Rejection (No Token): "
NO_AUTH_RESPONSE=$(docker compose exec -T localstack curl -s http://api:8080/v1/users/me || true)
if [[ $NO_AUTH_RESPONSE == *"\"success\":false"* ]] && [[ $NO_AUTH_RESPONSE == *"\"code\":\"AUTH_REQUIRED\""* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -n "  ☁️  AWS Status Endpoint: "
AWS_STATUS=$(docker compose exec -T localstack curl -s -H "Authorization: Bearer $TOKEN" http://api:8080/v1/aws-status || true)
if [[ $AWS_STATUS == *"\"s3\":\"ok\""* ]] && [[ $AWS_STATUS == *"\"sqs\":\"ok\""* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

# --- Final Check: Full Integration ---
echo -e "\n${YELLOW}[System Integrity Check]${NC}"

echo -n "  🔄 Full Integration (DB + AWS): "
INT_TEST=$(docker compose exec -T api go test ./internal/integration/ -v || true)
if [[ $INT_TEST == *"PASS"* ]] && [[ $INT_TEST != *"FAIL"* ]]; then
    echo -e "${GREEN}PASS${NC}"
else
    echo -e "${RED}FAIL${NC}"
    FAILED=$((FAILED + 1))
fi

echo -e "\n${CYAN}============================================================${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ VERIFICATION SUCCESSFUL: Day 11 Middleware & Standard Response are solid.${NC}"
    echo -e "${CYAN}============================================================${NC}"
    exit 0
else
    echo -e "${RED}❌ VERIFICATION FAILED: $FAILED check(s) failed.${NC}"
    echo -e "${CYAN}============================================================${NC}"
    exit 1
fi
