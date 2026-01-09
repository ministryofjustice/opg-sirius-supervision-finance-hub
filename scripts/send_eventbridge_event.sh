#!/bin/zsh
# Script to mimic sending AWS Eventbridge events to the local API
# Usage: send_eventbridge_event.sh <source> <detail-type> <detail-json> [override-json] [API_URL]

set -e

SOURCE="$1"
DETAIL_TYPE="$2"
DETAIL="$3"
OVERRIDE="$4"
API_URL="${5:-http://localhost:8181/events}"

if [[ -z "$SOURCE" || -z "$DETAIL_TYPE" || -z "$DETAIL" ]]; then
  echo "Usage: $0 <source> <detail-type> <detail-json> [override-json] [API_URL]"
  exit 1
fi

if [[ -n "$OVERRIDE" ]]; then
  # Merge override into detail
  DETAIL=$(jq -n --argjson body "$DETAIL" --argjson override "$OVERRIDE" '$body + {override: $override}')
fi

# Compose the event JSON
EVENT_JSON=$(jq -n --arg src "$SOURCE" --arg name "$DETAIL_TYPE" --argjson body "$DETAIL" '{source: $src, "detail-type": $name, detail: $body}')

# Send the event
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test" \
  -d "$EVENT_JSON"

echo "\nEvent sent to $API_URL:"
echo "$EVENT_JSON"
