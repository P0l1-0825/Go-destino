#!/bin/bash
# tenant-guard.sh — Warns when editing repository files without tenant_id in SQL queries
# GoDestino specific: all SQL queries MUST include tenant_id

INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')
FILE_PATH=""

case "$TOOL" in
  Edit|Write)
    FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
    ;;
  *)
    exit 0
    ;;
esac

[ -z "$FILE_PATH" ] && exit 0

# Only check repository files (where SQL queries live)
case "$FILE_PATH" in
  *repository*|*repo*) ;;
  *) exit 0 ;;
esac

# Only check Go files
case "$FILE_PATH" in
  *.go) ;;
  *) exit 0 ;;
esac

# Skip test files
case "$FILE_PATH" in
  *_test.go) exit 0 ;;
esac

# Check if the new content contains SQL without tenant_id
NEW_STRING=$(echo "$INPUT" | jq -r '.tool_input.new_string // .tool_input.content // empty')

if [ -n "$NEW_STRING" ]; then
  # Check for SQL operations without tenant_id
  if echo "$NEW_STRING" | grep -qiE "(SELECT|INSERT|UPDATE|DELETE)" 2>/dev/null; then
    if ! echo "$NEW_STRING" | grep -qi "tenant_id" 2>/dev/null; then
      echo "⚠️  TENANT GUARD: SQL query detected without 'tenant_id' in repository file." >&2
      echo "   File: $FILE_PATH" >&2
      echo "   GoDestino requires ALL SQL queries to filter by tenant_id." >&2
      echo "   Please verify this is intentional (e.g., tenants table itself)." >&2
      echo "" >&2
      # Warning only, don't block
      exit 0
    fi
  fi
fi

exit 0
