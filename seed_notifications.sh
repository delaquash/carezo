#!/bin/bash
cd "$(dirname "$0")"

API_BASE="https://carezo.onrender.com"

jq -c '.[]' notifications.json | while read -r notification; do
  title=$(echo "$notification" | jq -r '.title')
  echo "Creating: $title"
  curl -s -X POST "$API_BASE/api/admin/notifications" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$notification" | jq -r '.message // .error'
  echo "---"
done
