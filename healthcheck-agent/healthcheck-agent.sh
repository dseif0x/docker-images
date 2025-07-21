#!/bin/sh
set -e

BASE_URL="${BASE_URL:-https://healthchecks.io/api/v3}"
echo "Using Healthchecks.io base URL: $BASE_URL"
HOSTNAME=$(hostname)
API_KEY=$(cat /etc/healthchecks/api_token)
PROJECT_SLUG="${PROJECT_SLUG:-k8s-homelab}"

# Get all checks in the project
CHECKS_JSON=$(curl -s -H "X-Api-Key: $API_KEY" "$BASE_URL/checks/")

# Check if a check with this name already exists
PING_URL=$(echo "$CHECKS_JSON" | jq -r --arg NAME "$HOSTNAME" '.checks[] | select(.name == $NAME) | .ping_url')
if [ -n "$PING_URL" ]; then
  echo "Check already exists for $HOSTNAME (URL: $PING_URL), continuing..."
else
  echo "Creating new healthcheck for $HOSTNAME"
  PING_URL=$(curl -s -X POST "$BASE_URL/checks/" \
    -H "X-Api-Key: $API_KEY" \
    -H "Content-Type: application/json" \
    -d @- <<EOF | jq -r .ping_url
{
  "name": "$HOSTNAME",
  "tags": "k8s-node",
  "timeout": 3600,
  "grace": 600,
  "project": "$PROJECT_SLUG"
}
EOF
  )
fi

# Start pinging loop
while true; do
  echo "[$(date)] Pinging $PING_URL"
  curl -fsS --retry 3 "$PING_URL"
  sleep 1800
done
