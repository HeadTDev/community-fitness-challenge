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

# Service configuration from environment
API_URL="http://${API_HOST:-api}:${API_PORT:-8080}"
LS_URL="http://${LOCALSTACK_HOST:-localstack}:4566"
DB_CONN="postgresql://${DB_USER:-fc_user}:${DB_PASSWORD:-fc_password}@${DB_HOST:-postgres}/${DB_NAME:-fitchallenge}"

# --- Helper Functions ---
print_header() {
    echo -e "${CYAN}${BOLD}============================================================${NC}"
    echo -e "${CYAN}${BOLD}🚀 COMMUNITY FITNESS CHALLENGE - FULL SYSTEM VERIFICATION${NC}"
    echo -e "${CYAN}${BOLD}📅 Coverage: Day 1 to Day 16 (Challenge Service + S3 Logic)${NC}"
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

# --- Main Script ---
print_header

# --- Phase 1: Infrastructure ---
print_section "Phase 1: Infrastructure Connectivity (Day 1-3)"

# Redis PING
REDIS_PING=$(redis-cli -h ${REDIS_HOST:-redis} ping | tr -d '\r')
if [ "$REDIS_PING" == "PONG" ]; then
    report_status "Redis Connectivity (PING/PONG)" "PASS"
else
    report_status "Redis Connectivity (PING/PONG)" "FAIL"
fi

# --- Phase 2: API Backend ---
print_section "Phase 2: Core Backend Skeleton (Day 4)"

if curl -s "$API_URL/healthz" | grep -q "\"success\":true"; then
    report_status "API Readiness (/healthz)" "PASS"
else
    report_status "API Readiness (/healthz)" "FAIL"
fi

# --- Phase 3: Database & Migrations ---
print_section "Phase 3: Persistence & Domain (Day 5-6)"

MIG_TABLE=$(psql "$DB_CONN" -t -c "SELECT to_regclass('public.schema_migrations');" | xargs || true)
if [[ -n "$MIG_TABLE" ]]; then
    report_status "Database Migrations (schema_migrations)" "PASS"
else
    report_status "Database Migrations (schema_migrations)" "FAIL"
fi

USER_TABLE=$(psql "$DB_CONN" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='users' AND column_name='role';" | xargs || true)
if [[ -n "$USER_TABLE" ]]; then
    report_status "User Schema Integrity (Role column)" "PASS"
else
    report_status "User Schema Integrity (Role column)" "FAIL"
fi

# --- Phase 4: AWS Integration ---
print_section "Phase 4: Cloud Simulation (Day 7-8)"

# S3 Check via LocalStack health API
if curl -s "$LS_URL/_localstack/health" | jq '.services.s3' | grep -qE "available|running"; then
    report_status "S3 Storage Provisioning (Assets)" "PASS"
else
    report_status "S3 Storage Provisioning (Assets)" "FAIL"
fi

# SQS Check
if curl -s "$LS_URL/_localstack/health" | jq '.services.sqs' | grep -qE "available|running"; then
    report_status "SQS Message Queue (Background Jobs)" "PASS"
else
    report_status "SQS Message Queue (Background Jobs)" "FAIL"
fi

# --- Phase 5: Security & Auth ---
print_section "Phase 5: Security & Authentication (Day 9-11)"

AUTH_RESPONSE=$(curl -s -X POST "$API_URL/auth/register-dev")
TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.data.access_token' || true)

if [[ "$TOKEN" != "null" ]] && [[ -n "$TOKEN" ]]; then
    report_status "Developer Auth Flow (Token Generation)" "PASS"
else
    report_status "Developer Auth Flow (Token Generation)" "FAIL"
fi

ME_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/v1/users/me")
if [[ $ME_RESPONSE == *"\"success\":true"* ]]; then
    report_status "Auth Middleware: JWT Verification" "PASS"
else
    report_status "Auth Middleware: JWT Verification" "FAIL"
fi

# --- Phase 6: User Experience ---
print_section "Phase 6: User Experience (Day 12)"

PROFILE_UPDATE=$(curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"display_name":"Senior Dev", "bio":"Go Expert"}' \
    "$API_URL/v1/users/profile")
if [[ $PROFILE_UPDATE == *"\"display_name\":\"Senior Dev\""* ]]; then
    report_status "Profile CRUD: Data Persistence" "PASS"
else
    report_status "Profile CRUD: Data Persistence" "FAIL"
fi

# Avatar upload
echo "fake-image" > /tmp/avatar.jpg
AVATAR_UPLOAD=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
    -F "avatar=@/tmp/avatar.jpg" \
    "$API_URL/v1/users/profile/avatar")
if [[ $AVATAR_UPLOAD == *"avatar_url"* ]]; then
    report_status "Avatar Upload: S3 + DB Integration" "PASS"
else
    report_status "Avatar Upload: S3 + DB Integration" "FAIL"
fi

# --- Phase 7: Hardening ---
print_section "Phase 7: Hardening & Reliability (Day 13)"

VALID_FAIL=$(curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"display_name":""}' \
    "$API_URL/v1/users/profile")
if [[ $VALID_FAIL == *"INVALID_INPUT"* ]]; then
    report_status "Input Validation: Empty Fields Reject" "PASS"
else
    report_status "Input Validation: Empty Fields Reject" "FAIL"
fi

# Rate Limit Test (65 requests to trigger 429)
RL_TRIGGERED=0
for i in $(seq 1 65); do
    RESP=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/healthz")
    if [ "$RESP" == "429" ]; then
        RL_TRIGGERED=1
        break
    fi
done

if [ $RL_TRIGGERED -eq 1 ]; then
    report_status "Rate Limiting: Sliding Window (429)" "PASS"
else
    report_status "Rate Limiting: Sliding Window (429)" "FAIL"
fi

# --- Phase 8: Challenge Schema ---
print_section "Phase 8: Challenge & Competition Schema (Day 14)"

CHALLENGE_TABLE=$(psql "$DB_CONN" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='challenges' AND column_name='title';" | xargs || true)
if [[ -n "$CHALLENGE_TABLE" ]]; then
    report_status "Challenge Table Schema (Title column)" "PASS"
else
    report_status "Challenge Table Schema (Title column)" "FAIL"
fi

# --- Phase 9: Challenge Repository CRUD ---
print_section "Phase 9: Challenge Repository CRUD (Day 15)"

# 1. Create (Insert)
CHALLENGE_ID=$(psql "$DB_CONN" -t -A -q -c "INSERT INTO challenges (id, title, description, type, goal, start_date, end_date) \
    VALUES (gen_random_uuid(), 'Test Challenge', 'Description', 'steps', 10000, now(), now() + interval '7 days') \
    RETURNING id;" | awk '{print $1}' || true)

if [[ -n "$CHALLENGE_ID" && "$CHALLENGE_ID" != "null" ]]; then
    report_status "Repository: Create Challenge" "PASS" "ID: $CHALLENGE_ID"
    
    # 2. Read (Select)
    READ_TITLE=$(psql "$DB_CONN" -t -c "SELECT title FROM challenges WHERE id='$CHALLENGE_ID';" | xargs || true)
    if [[ "$READ_TITLE" == "Test Challenge" ]]; then
        report_status "Repository: Read Challenge" "PASS"
    else
        report_status "Repository: Read Challenge" "FAIL"
    fi

    # 3. Update
    psql "$DB_CONN" -c "UPDATE challenges SET title='Updated Title' WHERE id='$CHALLENGE_ID';" > /dev/null
    UPDATE_TITLE=$(psql "$DB_CONN" -t -c "SELECT title FROM challenges WHERE id='$CHALLENGE_ID';" | xargs || true)
    if [[ "$UPDATE_TITLE" == "Updated Title" ]]; then
        report_status "Repository: Update Challenge" "PASS"
    else
        report_status "Repository: Update Challenge" "FAIL"
    fi

    # 4. Delete
    psql "$DB_CONN" -c "DELETE FROM challenges WHERE id='$CHALLENGE_ID';" > /dev/null
    EXISTS=$(psql "$DB_CONN" -t -c "SELECT count(*) FROM challenges WHERE id='$CHALLENGE_ID';" | xargs || true)
    if [[ "$EXISTS" == "0" ]]; then
        report_status "Repository: Delete Challenge" "PASS"
    else
        report_status "Repository: Delete Challenge" "FAIL"
    fi
else
    report_status "Repository: Create Challenge" "FAIL" "Insert failed"
fi

# --- Phase 10: Challenge Service & Ownership ---
print_section "Phase 10: Challenge Service & Ownership (Day 16)"

CREATOR_COL=$(psql "$DB_CONN" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='challenges' AND column_name='creator_id';" | xargs || true)
if [[ -n "$CREATOR_COL" ]]; then
    report_status "Challenge Ownership (CreatorID column)" "PASS"
else
    report_status "Challenge Ownership (CreatorID column)" "FAIL"
fi

STATUS_DEFAULT=$(psql "$DB_CONN" -t -c "SELECT column_default FROM information_schema.columns WHERE table_name='challenges' AND column_name='status';" | xargs || true)
if [[ "$STATUS_DEFAULT" == "'draft'::character varying" ]] || [[ "$STATUS_DEFAULT" == "'draft'" ]] || [[ "$STATUS_DEFAULT" == "draft::character varying" ]]; then
    report_status "Challenge Lifecycle (Draft status default)" "PASS"
else
    report_status "Challenge Lifecycle (Draft status default)" "FAIL" "Default: $STATUS_DEFAULT"
fi

# --- Summary ---
echo -e "\n${CYAN}${BOLD}============================================================${NC}"
if [ $FAILED_COUNT -eq 0 ]; then
    echo -e "${GREEN}${BOLD}✅ ALL SYSTEMS GO! FULL VERIFICATION SUCCESSFUL${NC}"
    echo -e "${CYAN}Total Checks: $TOTAL_COUNT | Failures: 0${NC}"
    echo -e "${CYAN}${BOLD}============================================================${NC}"
    exit 0
else
    echo -e "${RED}${BOLD}❌ VERIFICATION FAILED${NC}"
    echo -e "${RED}Total Checks: $TOTAL_COUNT | Failures: $FAILED_COUNT${NC}"
    echo -e "${CYAN}${BOLD}============================================================${NC}"
    exit 1
fi
