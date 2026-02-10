#!/bin/bash

# ============================================
# Outils API Test Script
# ============================================

API_URL="http://165.227.35.35:8080"
TOKEN="eyJhbGciOiJSUzI1NiIsImtpZCI6IjRiMTFjYjdhYjVmY2JlNDFlOTQ4MDk0ZTlkZjRjNWI1ZWNhMDAwOWUiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vb3V0aWxzLWFwaSIsImF1ZCI6Im91dGlscy1hcGkiLCJhdXRoX3RpbWUiOjE3NzA3NDIzNzAsInVzZXJfaWQiOiJscmdpcjdOOGJPZ1lpcDlHNkJadmpNRWdWenMyIiwic3ViIjoibHJnaXI3TjhiT2dZaXA5RzZCWnZqTUVnVnpzMiIsImlhdCI6MTc3MDc0MjM3MCwiZXhwIjoxNzcwNzQ1OTcwLCJlbWFpbCI6ImVzLnJhamF0eWFkYXZAZ21haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJmaXJlYmFzZSI6eyJpZGVudGl0aWVzIjp7ImVtYWlsIjpbImVzLnJhamF0eWFkYXZAZ21haWwuY29tIl19LCJzaWduX2luX3Byb3ZpZGVyIjoicGFzc3dvcmQifX0.pMx4CI7YChPMukt7Fqd1uTni4G2McrI9YZY5gjxX4Spus-_WNJyUjZ_zCDCDYrQSoVvJyvhs5Q5f7aj_ajxWEum-tOWUJzbwY6UGXhvYI80furoo7Zf_1EyTD2t-_9f4gF3YEsKYVoYC4EGa0m5GpQu8KzBbnTSEKEl5uE_-F46I0DUAYz0MriaLG3stBSGAW56-k2m_IZ5aZi1rvjoi-sQFUGnAnfzPcTcoj2QYJ3ambAMn4GN-3567uvXL_GF5QEgkykKStqQ9R22-MgX3qZoL55SmYwPm-ZIUztjtkMTMDys5KWymbNjKYq3ymjg6MO1ndv0w7ziodOpV7XeaIg"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function
log() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

success() {
    echo -e "${GREEN}[✓ PASS]${NC} $1"
}

error() {
    echo -e "${RED}[✗ FAIL]${NC} $1"
}

# ============================================
# SET YOUR FIREBASE TOKEN HERE
# ============================================
echo "============================================"
echo "Outils API Test Script"
echo "============================================"
echo ""
echo "Get your Firebase token from the test HTML page"
echo "and paste it below:"
echo ""
read -p "Token: " TOKEN

if [ -z "$TOKEN" ]; then
    error "No token provided. Exiting."
    exit 1
fi

echo ""
echo "============================================"
echo "Starting Tests..."
echo "============================================"
echo ""

# ============================================
# Test 1: Health Check
# ============================================
log "Testing health endpoint..."
HEALTH=$(curl -s "$API_URL/health")
echo "$HEALTH" | grep -q "healthy" && success "Health check" || error "Health check"
echo "Response: $HEALTH"
echo ""

# ============================================
# Test 2: Ready Check
# ============================================
log "Testing ready endpoint..."
READY=$(curl -s "$API_URL/ready")
echo "$READY" | grep -q "ready" && success "Ready check" || error "Ready check"
echo "Response: $READY"
echo ""

# ============================================
# Test 3: Register User
# ============================================
log "Registering user..."
REGISTER=$(curl -s -X POST "$API_URL/api/register" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","display_name":"Test User"}')
echo "$REGISTER" | grep -q "firebase_uid" && success "Register user" || error "Register user"
echo "Response: $REGISTER"
USER_ID=$(echo "$REGISTER" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "User ID: $USER_ID"
echo ""

# ============================================
# Test 4: Get Me
# ============================================
log "Getting current user..."
ME=$(curl -s "$API_URL/api/me" \
    -H "Authorization: Bearer $TOKEN")
echo "$ME" | grep -q "firebase_uid" && success "Get me" || error "Get me"
echo "Response: $ME"
echo ""

# ============================================
# Test 5: Create Task List
# ============================================
log "Creating task list..."
CREATE_LIST=$(curl -s -X POST "$API_URL/api/tasklists" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"Shopping List"}')
echo "$CREATE_LIST" | grep -q '"id"' && success "Create task list" || error "Create task list"
echo "Response: $CREATE_LIST"
LIST_ID=$(echo "$CREATE_LIST" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Task List ID: $LIST_ID"
echo ""

# ============================================
# Test 6: Get All Task Lists
# ============================================
log "Getting all task lists..."
GET_LISTS=$(curl -s "$API_URL/api/tasklists" \
    -H "Authorization: Bearer $TOKEN")
echo "$GET_LISTS" | grep -q "Shopping List" && success "Get all task lists" || error "Get all task lists"
echo "Response: $GET_LISTS"
echo ""

# ============================================
# Test 7: Get One Task List
# ============================================
log "Getting one task list..."
GET_LIST=$(curl -s "$API_URL/api/tasklist/$LIST_ID" \
    -H "Authorization: Bearer $TOKEN")
echo "$GET_LIST" | grep -q "Shopping List" && success "Get one task list" || error "Get one task list"
echo "Response: $GET_LIST"
echo ""

# ============================================
# Test 8: Create Task 1
# ============================================
log "Creating task 1..."
CREATE_TASK1=$(curl -s -X POST "$API_URL/api/tasklist/$LIST_ID/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Buy milk"}')
echo "$CREATE_TASK1" | grep -q '"id"' && success "Create task 1" || error "Create task 1"
echo "Response: $CREATE_TASK1"
TASK1_ID=$(echo "$CREATE_TASK1" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Task 1 ID: $TASK1_ID"
echo ""

# ============================================
# Test 9: Create Task 2
# ============================================
log "Creating task 2..."
CREATE_TASK2=$(curl -s -X POST "$API_URL/api/tasklist/$LIST_ID/tasks" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Buy eggs"}')
echo "$CREATE_TASK2" | grep -q '"id"' && success "Create task 2" || error "Create task 2"
echo "Response: $CREATE_TASK2"
TASK2_ID=$(echo "$CREATE_TASK2" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Task 2 ID: $TASK2_ID"
echo ""

# ============================================
# Test 10: Get All Tasks
# ============================================
log "Getting all tasks in list..."
GET_TASKS=$(curl -s "$API_URL/api/tasklist/$LIST_ID/tasks" \
    -H "Authorization: Bearer $TOKEN")
echo "$GET_TASKS" | grep -q "Buy milk" && success "Get all tasks" || error "Get all tasks"
echo "Response: $GET_TASKS"
echo ""

# ============================================
# Test 11: Mark Task Complete
# ============================================
log "Marking task 1 complete..."
COMPLETE_TASK=$(curl -s -X PUT "$API_URL/api/tasks/$TASK1_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"completed":true}')
echo "$COMPLETE_TASK" | grep -q '"completed":true' && success "Mark task complete" || error "Mark task complete"
echo "Response: $COMPLETE_TASK"
echo ""

# ============================================
# Test 12: Update Task Title
# ============================================
log "Updating task 2 title..."
UPDATE_TASK=$(curl -s -X PUT "$API_URL/api/tasks/$TASK2_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"title":"Buy organic eggs"}')
echo "$UPDATE_TASK" | grep -q "organic eggs" && success "Update task title" || error "Update task title"
echo "Response: $UPDATE_TASK"
echo ""

# ============================================
# Test 13: Delete Task
# ============================================
log "Deleting task 2..."
DELETE_TASK=$(curl -s -X DELETE "$API_URL/api/tasks/$TASK2_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -w "%{http_code}")
[ "$DELETE_TASK" = "204" ] && success "Delete task" || error "Delete task (status: $DELETE_TASK)"
echo ""

# ============================================
# Test 14: Verify Task Deleted
# ============================================
log "Verifying task deleted..."
GET_TASKS2=$(curl -s "$API_URL/api/tasklist/$LIST_ID/tasks" \
    -H "Authorization: Bearer $TOKEN")
echo "$GET_TASKS2" | grep -q "organic eggs" && error "Task should be deleted" || success "Task deleted verified"
echo "Response: $GET_TASKS2"
echo ""

# ============================================
# Test 15: Create Second Task List
# ============================================
log "Creating second task list..."
CREATE_LIST2=$(curl -s -X POST "$API_URL/api/tasklists" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"name":"Work Tasks"}')
echo "$CREATE_LIST2" | grep -q '"id"' && success "Create second task list" || error "Create second task list"
LIST2_ID=$(echo "$CREATE_LIST2" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
echo "Task List 2 ID: $LIST2_ID"
echo ""

# ============================================
# Test 16: Delete Task List
# ============================================
log "Deleting second task list..."
DELETE_LIST=$(curl -s -X DELETE "$API_URL/api/tasklist/$LIST2_ID" \
    -H "Authorization: Bearer $TOKEN" \
    -w "%{http_code}")
[ "$DELETE_LIST" = "204" ] && success "Delete task list" || error "Delete task list (status: $DELETE_LIST)"
echo ""

# ============================================
# Test 17: Verify Only One List Remains
# ============================================
log "Verifying task list deleted..."
GET_LISTS2=$(curl -s "$API_URL/api/tasklists" \
    -H "Authorization: Bearer $TOKEN")
echo "$GET_LISTS2" | grep -q "Work Tasks" && error "List should be deleted" || success "Task list deleted verified"
echo "Response: $GET_LISTS2"
echo ""

# ============================================
# Summary
# ============================================
echo ""
echo "============================================"
echo "Test Summary"
echo "============================================"
echo ""
echo "Remaining data:"
echo "- User ID: $USER_ID"
echo "- Task List ID: $LIST_ID (Shopping List)"
echo "- Task ID: $TASK1_ID (Buy milk - completed)"
echo ""
echo "All tests completed!"
