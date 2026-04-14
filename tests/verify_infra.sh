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
    echo -e "${CYAN}${BOLD}📅 Coverage: Day 1 to Day 21 (Refactoring & Unit Tests)${NC}"
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

# --- Phase 7: Data Validation ---
print_section "Phase 7: Data Integrity (Day 13-14)"

VALID_FAIL=$(curl -s -X PUT -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"display_name":""}' \
    "$API_URL/v1/users/profile")
if [[ $VALID_FAIL == *"INVALID_INPUT"* ]]; then
    report_status "Input Validation: Empty Fields Reject" "PASS"
else
    report_status "Input Validation: Empty Fields Reject" "FAIL"
fi

CHALLENGE_TABLE=$(psql "$DB_CONN" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name='challenges' AND column_name='title';" | xargs || true)
if [[ -n "$CHALLENGE_TABLE" ]]; then
    report_status "Challenge Table Schema (Title column)" "PASS"
else
    report_status "Challenge Table Schema (Title column)" "FAIL"
fi

# --- Phase 8: Challenge Repository CRUD ---
print_section "Phase 8: Challenge Repository CRUD (Day 15)"

# Get a valid user ID for creator_id (required by foreign key constraint)
VALID_USER_ID=$(psql "$DB_CONN" -t -c "SELECT id FROM users LIMIT 1;" | xargs || true)

# 1. Create (Insert)
CHALLENGE_ID=$(psql "$DB_CONN" -t -A -q -c "INSERT INTO challenges (id, creator_id, title, description, type, goal, start_date, end_date) \
    VALUES (gen_random_uuid(), '$VALID_USER_ID', 'Test Challenge', 'Description', 'steps', 10000, now(), now() + interval '7 days') \
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

# --- Phase 9: Challenge Service & Ownership ---
print_section "Phase 9: Challenge Service & Ownership (Day 16)"

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

# --- Phase 10: Challenge API Handlers ---
print_section "Phase 10: Challenge API Handlers (Day 17)"

# Test List Challenges (should be accessible by any authenticated user)
LIST_CHALLENGES=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/v1/challenges")
if [[ $LIST_CHALLENGES == *"\"success\":true"* ]]; then
    report_status "API: List Challenges (v1/challenges)" "PASS"
else
    report_status "API: List Challenges (v1/challenges)" "FAIL" "Response: $LIST_CHALLENGES"
fi

# --- Phase 11: Challenge Join-Leave ---
print_section "Phase 11: Challenge Join-Leave & Redis Counter (Day 18)"

    # 1. Create a challenge and publish it with max_participants=1
    START_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    # Alpine/BusyBox compatible date math
    END_DATE=$(date -u -d "@$(($(date +%s) + 604800))" +"%Y-%m-%dT%H:%M:%SZ")
    
    CHALLENGE_JSON=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "title": "Join Test Challenge",
            "description": "Test description",
            "start_date": "'$START_DATE'",
            "end_date": "'$END_DATE'",
            "type": "steps",
            "goal": 50000,
            "max_participants": 1
        }' "$API_URL/v1/challenges")

CHALLENGE_ID=$(echo "$CHALLENGE_JSON" | jq -r '.data.id')

if [[ "$CHALLENGE_ID" != "null" ]] && [[ -n "$CHALLENGE_ID" ]]; then
    # Publish it
    curl -s -X POST -H "Authorization: Bearer $TOKEN" "$API_URL/v1/challenges/$CHALLENGE_ID/publish" > /dev/null

    # 2. Join (First User)
    JOIN_RESP=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$API_URL/v1/challenges/$CHALLENGE_ID/join")
    if [[ $JOIN_RESP == *"Successfully joined challenge"* ]]; then
        report_status "API: Join Challenge" "PASS"
    else
        report_status "API: Join Challenge" "FAIL" "Response: $JOIN_RESP"
    fi

    # 3. Check Redis Counter
    REDIS_COUNT=$(redis-cli -h ${REDIS_HOST:-redis} GET "challenge_count:$CHALLENGE_ID" | tr -d '\r')
    if [ "$REDIS_COUNT" == "1" ]; then
        report_status "Redis: Atomic Participant Counter (INCR)" "PASS"
    else
        report_status "Redis: Atomic Participant Counter (INCR)" "FAIL" "Count: $REDIS_COUNT"
    fi

    # 4. Join (Second User) - Should be FULL (410)
    # Register a second dev user
    TOKEN2=$(curl -s -X POST "$API_URL/auth/register-dev" | jq -r '.data.access_token')
    JOIN_FULL_RESP=$(curl -s -X POST -H "Authorization: Bearer $TOKEN2" "$API_URL/v1/challenges/$CHALLENGE_ID/join")
    if [[ $JOIN_FULL_RESP == *"FULL"* ]]; then
        report_status "API: Join Full Challenge (410)" "PASS"
    else
        report_status "API: Join Full Challenge (410)" "FAIL" "Response: $JOIN_FULL_RESP"
    fi

    # 5. Leave
    LEAVE_RESP=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$API_URL/v1/challenges/$CHALLENGE_ID/leave")
    if [[ $LEAVE_RESP == *"Successfully left challenge"* ]]; then
        report_status "API: Leave Challenge" "PASS"
    else
        report_status "API: Leave Challenge" "FAIL" "Response: $LEAVE_RESP"
    fi

    # 6. Check Redis Counter again
    REDIS_COUNT=$(redis-cli -h ${REDIS_HOST:-redis} GET "challenge_count:$CHALLENGE_ID" | tr -d '\r')
    if [ "$REDIS_COUNT" == "0" ] || [ -z "$REDIS_COUNT" ] || [ "$REDIS_COUNT" == "(nil)" ]; then
        report_status "Redis: Atomic Participant Counter (DECR)" "PASS"
    else
        report_status "Redis: Atomic Participant Counter (DECR)" "FAIL" "Count: $REDIS_COUNT"
    fi
else
    report_status "API: Challenge Join Setup" "FAIL" "Could not create challenge. Resp: $CHALLENGE_JSON"
fi

# --- Phase 12: Prize Management ---
print_section "Phase 12: Prize Management (Day 19)"

# Use current timestamp for unique titles to avoid potential conflicts
TS=$(date +%s)
START_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
END_DATE=$(date -u -d "@$(($(date +%s) + 604800))" +"%Y-%m-%dT%H:%M:%SZ")

DRAFT_CHALLENGE=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Prize Test Challenge '$TS'",
        "description": "Prize test description",
        "start_date": "'$START_DATE'",
        "end_date": "'$END_DATE'",
        "type": "calories",
        "goal": 10000
    }' "$API_URL/v1/challenges")

PRIZE_CHALLENGE_ID=$(echo "$DRAFT_CHALLENGE" | jq -r '.data.id')

if [[ "$PRIZE_CHALLENGE_ID" != "null" ]] && [[ -n "$PRIZE_CHALLENGE_ID" ]]; then
    # 2. Add a prize to draft
    ADD_PRIZE_RESP=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "title": "Gold Medal",
            "description": "For the first place",
            "rank_required": 1
        }' "$API_URL/v1/challenges/$PRIZE_CHALLENGE_ID/prizes")
    
    if [[ $ADD_PRIZE_RESP == *"Gold Medal"* ]]; then
        report_status "API: Add Prize to Draft" "PASS"
    else
        report_status "API: Add Prize to Draft" "FAIL" "Resp: $ADD_PRIZE_RESP (URL: $API_URL/v1/challenges/$PRIZE_CHALLENGE_ID/prizes)"
    fi

    # 3. List prizes
    LIST_PRIZES=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/v1/challenges/$PRIZE_CHALLENGE_ID/prizes")
    if [[ $LIST_PRIZES == *"Gold Medal"* ]]; then
        report_status "API: List Prizes" "PASS"
    else
        report_status "API: List Prizes" "FAIL" "Resp: $LIST_PRIZES"
    fi

    # 4. Publish challenge
    PUBLISH_RESP=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" "$API_URL/v1/challenges/$PRIZE_CHALLENGE_ID/publish")

    # 5. Try to add prize to published (Should FAIL with 400)
    ADD_LATE_PRIZE=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "title": "Silver Medal",
            "rank_required": 2
        }' "$API_URL/v1/challenges/$PRIZE_CHALLENGE_ID/prizes")
    
    if [[ $ADD_LATE_PRIZE == *"BAD_REQUEST"* ]]; then
        report_status "API: Add Prize to Published Rejected" "PASS"
    else
        report_status "API: Add Prize to Published Rejected" "FAIL" "Resp: $ADD_LATE_PRIZE"
    fi
else
    report_status "API: Prize Management Setup" "FAIL" "Could not create draft challenge. Resp: $DRAFT_CHALLENGE"
fi

# --- Phase 13: Seed Data Verification ---
print_section "Phase 13: Seed Data & Documentation (Day 20)"

USER_COUNT=$(psql "$DB_CONN" -t -c "SELECT count(*) FROM users WHERE email LIKE '%@fitchallenge.com';" | xargs || true)
if [ "$USER_COUNT" -ge 5 ]; then
    report_status "Seed: User Generation (>= 5)" "PASS" "Count: $USER_COUNT"
else
    report_status "Seed: User Generation (>= 5)" "FAIL" "Count: $USER_COUNT"
fi

CHALLENGE_COUNT=$(psql "$DB_CONN" -t -c "SELECT count(*) FROM challenges WHERE creator_id IN (SELECT id FROM users WHERE role='creator');" | xargs || true)
if [ "$CHALLENGE_COUNT" -ge 3 ]; then
    report_status "Seed: Challenge Generation (>= 3)" "PASS" "Count: $CHALLENGE_COUNT"
else
    report_status "Seed: Challenge Generation (>= 3)" "FAIL" "Count: $CHALLENGE_COUNT"
fi

PRIZE_COUNT=$(psql "$DB_CONN" -t -c "SELECT count(*) FROM prizes;" | xargs || true)
if [ "$PRIZE_COUNT" -ge 6 ]; then
    report_status "Seed: Prize Generation (>= 6)" "PASS" "Count: $PRIZE_COUNT"
else
    report_status "Seed: Prize Generation (>= 6)" "FAIL" "Count: $PRIZE_COUNT"
fi

if [ -f "/app/backend/docs/community-fitness-challenge.postman_collection.json" ]; then
    report_status "Docs: Postman Collection Export" "PASS"
else
    report_status "Docs: Postman Collection Export" "FAIL"
fi

# --- Phase 14: Security Hardening (STRESS TEST) ---
print_section "Phase 14: Security Hardening & Rate Limiting (Day 13)"

# Rate Limit Test (65 requests to trigger 429) - PERFORMED LAST
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

# --- Phase 15: Code Refactoring Verification ---
print_section "Phase 15: Code Refactoring & Unit Tests (Day 21)"

if [ -f "/app/backend/internal/app/app.go" ]; then
    report_status "DI: Application Container (app.go)" "PASS"
else
    report_status "DI: Application Container (app.go)" "FAIL"
fi

# Use grep to check if main.go is using the new app package
if grep -q "internal/app" "/app/backend/cmd/api/main.go"; then
    report_status "DI: Main uses App Container" "PASS"
else
    report_status "DI: Main uses App Container" "FAIL"
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
