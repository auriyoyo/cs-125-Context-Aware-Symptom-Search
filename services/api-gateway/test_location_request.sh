#!/usr/bin/env bash

# This script sends a test location update event to the local API gateway
# Make sure the API gateway is running on localhost:8080 before executing

API_URL="http://localhost:8080/events/location"

# Hardcoded test payload
JSON_BODY=$(cat <<EOF
{
  "userId": "test_user_789",
  "country": "US",
  "state": "NY",
  "city": "New York",
  "zipCode": "10001"
}
EOF
)

echo "Sending location update to $API_URL..."
echo "Payload: $JSON_BODY"
echo ""

# Send the POST request
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d "$JSON_BODY" \
  -v

echo -e "\n\nDone."
