#!/bin/bash
set -e

# --- Configuration & Colors ---
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

FAILED_COUNT=0
TOTAL_COUNT=0

# --- Helper Functions ---
print_header() {
    echo -e "${CYAN}${BOLD}============================================================${NC}"
    echo -e "${CYAN}${BOLD}🚀 COMMUNITY FITNESS CHALLENGE - FULL PROJECT VERIFICATION${NC}"
    echo -e "${CYAN}${BOLD}📅 Progress: Day 1 to Day 14 (Challenge, Prize, Participation Migrations)${NC}"
    echo -e "${CYAN}${BOLD}============================================================${NC}"
}

print_section() {
    echo -e "\n${YELLOW}${BOLD}--- $1 ---${NC}"
}

report_status() {
    TOTAL_COUNT=$((TOTAL_COUNT + 1))
    local label=$1
    local status=$2
    local extra=$3

    printf "  %-50s " "$label"
    if [ "$status" == "PASS" ]; then
        echo -e "[ ${GREEN}${BOLD}PASS${NC} ] $extra"
    else
        echo -e "[ ${RED}${BOLD}FAIL${NC} ] $extra"
        FAILED_COUNT=$((FAILED_COUNT + 1))
    fi
}

exec_api_test() {
    local pkg=$1
    local res=$(docker compose exec -T api go test "$pkg" -v 2>&1 || true)
    if [[ "$res" == *"PASS"* ]] && [[ "$res" != *"FAIL"* ]]; then
        echo "PASS"
    else
        echo "FAIL"
    fi
}

# --- Main Script ---
print_header

# --- Infrastructure ---
print_section "Phase 1: Infrastructure & Environment (Day 1-3)"

CONTAINERS=$(docker compose ps --format json)
if [[ $CONTAINERS == *"healthy"* ]] && [[ $CONTAINERS == *"fc-api"* ]]; then
    report_status "Docker Containers Health" "PASS"
else
    report_status "Docker Containers Health" "FAIL" "One or more containers are unhealthy"
fi

REDIS_PING=$(docker compose exec -T redis redis-cli ping | tr -d '\r')
if [ "$REDIS_PING" == "PONG" ]; then
    report_status "Redis Connectivity (PING/PONG)" "PASS"
else
    report_status "Redis Connectivity (PING/PONG)" "FAIL" "Redis is unreachable"
fi

# --- API Backend ---
print_section "Phase 2: Core Backend Skeleton (Day 4)"

# Wait for API with improved logic
MAX_RETRIES=15
COUNT=0
API_READY="FAIL"
until $(docker compose exec -T localstack curl -s http://api:8080/healthz | grep -q "\"success\":true"); do
    sleep 1
    COUNT=$((COUNT + 1))
    if [ $COUNT -eq $MAX_RETRIES ]; then break; fi
done
[ $COUNT -lt $MAX_RETRIES ] && API_READY="PASS"
report_status "API Readiness (/healthz)" "$API_READY" "Retries: $COUNT"

# --- Database & Migrations ---
print_section "Phase 3: Persistence & Domain (Day 5-6)"

MIG_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\dt" | grep schema_migrations || true)
if [[ -n "$MIG_TABLE" ]]; then
    report_status "Database Migrations (schema_migrations)" "PASS"
else
    report_status "Database Migrations (schema_migrations)" "FAIL" "Migration table not found"
fi

USER_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\d users" | grep role || true)
if [[ -n "$USER_TABLE" ]]; then
    report_status "User Schema Integrity (Role column)" "PASS"
else
    report_status "User Schema Integrity (Role column)" "FAIL" "Users table or role column missing"
fi

USER_REPO_TEST=$(exec_api_test "./internal/adapter/postgres/")
report_status "User Repository Unit/Int Tests" "$USER_REPO_TEST"

# --- AWS Integration ---
print_section "Phase 4: Cloud Simulation (Day 7-8)"

AWS_CLIENT_TEST=$(exec_api_test "./internal/aws/")
report_status "AWS SDK Client Suite" "$AWS_CLIENT_TEST"

S3_BUCKET=$(docker compose exec -T localstack awslocal s3 ls | grep fitchallenge-assets || true)
if [[ -n "$S3_BUCKET" ]]; then
    report_status "S3 Storage Provisioning (Assets)" "PASS"
else
    report_status "S3 Storage Provisioning (Assets)" "FAIL" "Bucket 'fitchallenge-assets' missing"
fi

SQS_QUEUE=$(docker compose exec -T localstack awslocal sqs list-queues | grep fitchallenge-jobs || true)
if [[ -n "$SQS_QUEUE" ]]; then
    report_status "SQS Message Queue (Background Jobs)" "PASS"
else
    report_status "SQS Message Queue (Background Jobs)" "FAIL" "Queue 'fitchallenge-jobs' missing"
fi

# --- Security & Auth ---
print_section "Phase 5: Security & Authentication (Day 9-11)"

JWT_TEST=$(exec_api_test "./internal/pkg/jwt/")
report_status "JWT Logic & Token Generation" "$JWT_TEST"

AUTH_RESPONSE=$(docker compose exec -T localstack curl -s -X POST http://api:8080/auth/register-dev || true)
TOKEN=$(echo "$AUTH_RESPONSE" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p' || true)
USER_ID=$(echo "$AUTH_RESPONSE" | sed -n 's/.*"user_id":"\([^"]*\)".*/\1/p' || true)

if [[ -n "$TOKEN" ]]; then
    report_status "Developer Auth Flow (Token Generation)" "PASS"
else
    report_status "Developer Auth Flow (Token Generation)" "FAIL" "Could not register dev user"
fi

RID_HEADER=$(docker compose exec -T localstack curl -s -i http://api:8080/healthz | tr -d '\r' | grep -i "X-Request-ID" || true)
if [[ -n "$RID_HEADER" ]]; then
    report_status "Middleware: Request Tracking (X-Request-ID)" "PASS"
else
    report_status "Middleware: Request Tracking (X-Request-ID)" "FAIL" "Header missing in response"
fi

ME_RESPONSE=$(docker compose exec -T localstack curl -s -H "Authorization: Bearer $TOKEN" http://api:8080/v1/users/me || true)
if [[ $ME_RESPONSE == *"\"success\":true"* ]]; then
    report_status "Auth Middleware: JWT Verification" "PASS"
else
    report_status "Auth Middleware: JWT Verification" "FAIL" "Protected route rejected valid token"
fi

# --- User Profile & Avatar ---
print_section "Phase 6: User Experience (Day 12)"

PROFILE_UPDATE=$(docker compose exec -T localstack curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"display_name":"Senior Dev", "bio":"Go Expert"}' \
    http://api:8080/v1/users/profile || true)
if [[ $PROFILE_UPDATE == *"\"display_name\":\"Senior Dev\""* ]]; then
    report_status "Profile CRUD: Data Persistence" "PASS"
else
    report_status "Profile CRUD: Data Persistence" "FAIL" "Update failed or returned incorrect data"
fi

# Dummy avatar upload
docker compose exec -T localstack sh -c "echo 'fake-image' > /tmp/avatar.jpg"
AVATAR_UPLOAD=$(docker compose exec -T localstack curl -s -X POST -H "Authorization: Bearer $TOKEN" \
    -F "avatar=@/tmp/avatar.jpg" \
    http://api:8080/v1/users/profile/avatar || true)
if [[ $AVATAR_UPLOAD == *"avatar_url"* ]]; then
    report_status "Avatar Upload: S3 + DB Integration" "PASS"
else
    report_status "Avatar Upload: S3 + DB Integration" "FAIL" "Upload flow failed"
fi

# --- Validation & Rate Limiting ---
print_section "Phase 7: Hardening & Reliability (Day 13)"

VALID_FAIL=$(docker compose exec -T localstack curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"display_name":""}' \
    http://api:8080/v1/users/profile || true)
if [[ $VALID_FAIL == *"INVALID_INPUT"* ]]; then
    report_status "Input Validation: Empty Fields Reject" "PASS"
else
    report_status "Input Validation: Empty Fields Reject" "FAIL" "Server accepted invalid empty display_name"
fi

# Optimized Rate Limit Test
# Since limit is 60/min, we need to send 61 requests.
# We'll use a loop to send them in parallel if possible, or just fast.
RL_TRIGGERED=0
for i in $(seq 1 65); do
    RESP=$(docker compose exec -T localstack curl -s -o /dev/null -w "%{http_code}" http://api:8080/healthz || true)
    if [ "$RESP" == "429" ]; then
        RL_TRIGGERED=1
        break
    fi
done

if [ $RL_TRIGGERED -eq 1 ]; then
    report_status "Rate Limiting: Sliding Window (429)" "PASS" "429 Triggered successfully"
else
    report_status "Rate Limiting: Sliding Window (429)" "FAIL" "Did not trigger 429 after 65 requests"
fi

# --- Day 14: Challenge, Prize & Participation Migrations ---
print_section "Phase 8: Challenge & Competition Schema (Day 14)"

CHALLENGE_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\d challenges" | grep title || true)
if [[ -n "$CHALLENGE_TABLE" ]]; then
    report_status "Challenge Table Schema (Title column)" "PASS"
else
    report_status "Challenge Table Schema (Title column)" "FAIL" "Challenges table or title column missing"
fi

PRIZE_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\d prizes" | grep challenge_id || true)
if [[ -n "$PRIZE_TABLE" ]]; then
    report_status "Prize Table Schema (Challenge ID Ref)" "PASS"
else
    report_status "Prize Table Schema (Challenge ID Ref)" "FAIL" "Prizes table or challenge_id column missing"
fi

PARTICIPATION_TABLE=$(docker compose exec -T postgres psql -U fc_user -d fitchallenge -c "\d participations" | grep current_score || true)
if [[ -n "$PARTICIPATION_TABLE" ]]; then
    report_status "Participation Table Schema (Score column)" "PASS"
else
    report_status "Participation Table Schema (Score column)" "FAIL" "Participations table or score column missing"
fi

# --- Summary ---
echo -e "\n${CYAN}${BOLD}============================================================${NC}"
if [ $FAILED_COUNT -eq 0 ]; then
    echo -e "${GREEN}${BOLD}✅ ALL SYSTEMS GO! DAY 14 VERIFICATION SUCCESSFUL${NC}"
    echo -e "${CYAN}Total Checks: $TOTAL_COUNT | Failures: 0${NC}"
    echo -e "${CYAN}${BOLD}============================================================${NC}"
    exit 0
else
    echo -e "${RED}${BOLD}❌ VERIFICATION FAILED${NC}"
    echo -e "${RED}Total Checks: $TOTAL_COUNT | Failures: $FAILED_COUNT${NC}"
    echo -e "${CYAN}${BOLD}============================================================${NC}"
    exit 1
fi
