#!/usr/bin/env bash
# smoke-test.sh — Post-deploy smoke tests for doit MCP server.
# Usage: bash scripts/smoke-test.sh [BASE_URL] [API_KEY]
#
# Sends real MCP tool calls via the /mcp endpoint and verifies responses.
# Exits 0 if all checks pass, 1 on first failure.

set -euo pipefail

BASE_URL="${1:-https://din.aoendpoint.com}"
API_KEY="${2:-${DOIT_API_KEY:-}}"

if [ -z "$API_KEY" ]; then
  echo "ERROR: API key required. Pass as arg 2 or set DOIT_API_KEY env var."
  exit 1
fi

PASS=0
FAIL=0

check() {
  local name="$1"
  local tool="$2"
  local args="$3"
  local expect_substring="$4"

  local payload
  payload=$(cat <<EOF
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "$tool",
    "arguments": $args
  }
}
EOF
)

  local response
  response=$(curl -s -X POST "$BASE_URL/mcp" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "$payload" 2>&1) || true

  if echo "$response" | grep -q "$expect_substring"; then
    echo "  PASS: $name"
    PASS=$((PASS + 1))
  else
    echo "  FAIL: $name"
    echo "    Expected substring: $expect_substring"
    echo "    Got: $(echo "$response" | head -c 200)"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Doit Smoke Tests ==="
echo "Endpoint: $BASE_URL"
echo ""

# 1. Health check
echo "--- Health ---"
HEALTH=$(curl -s "$BASE_URL/health" 2>&1) || true
if echo "$HEALTH" | grep -q "ok"; then
  echo "  PASS: /health returns ok"
  PASS=$((PASS + 1))
else
  echo "  FAIL: /health"
  echo "    Got: $HEALTH"
  FAIL=$((FAIL + 1))
fi

# 2. List projects (basic connectivity + auth)
echo "--- Projects ---"
check "doit_list_projects" "doit_list_projects" '{}' '"slug"'

# 3. List issues with null optional fields (the bug that prompted this)
echo "--- Null Handling ---"
check "doit_list_issues with null project" "doit_list_issues" \
  '{"status":"open","issue_type":"","priority":null,"assignee":"","project":null,"limit":1,"sort_by":"priority"}' \
  '"id"'

check "doit_list_lessons with null filters" "doit_list_lessons" \
  '{"project":null,"status":null,"expert":null,"component":null,"severity":null,"limit":5}' \
  '['

check "doit_list_flags with null filters" "doit_list_flags" \
  '{"project":null,"status":null,"severity":null,"issue_id":null,"limit":5}' \
  '['

# 4. Ready endpoint
echo "--- Ready ---"
check "doit_ready with null project" "doit_ready" \
  '{"limit":1,"project":null}' \
  '['

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
