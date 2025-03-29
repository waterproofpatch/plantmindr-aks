#!/bin/bash

# Check if email and password are provided as arguments
if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <email> <password>"
  exit 1
fi

email="$1"
password="$2"

# Step 1: Login and get the auth token
response=$(curl -s 'https://strider.azurewebsites.net/api/login' \
  -H 'Content-Type: application/json' \
  --data-raw "{\"email\":\"$email\",\"password\":\"$password\"}")

auth_token=$(echo "$response" | jq -r '.token')

if [[ -z "$auth_token" || "$auth_token" == "null" ]]; then
  echo "Error: Failed to retrieve auth token."
  echo "Response: $response"
  exit 1
fi

echo "Authorization Token: $auth_token"

# Step 2: Fetch plants using the auth token
plants=$(curl -s 'https://strider.azurewebsites.net/api/plants' \
  -H "Authorization: Bearer $auth_token" \
  -H 'Content-Type: application/json')

if [[ -z "$plants" || "$plants" == "null" ]]; then
  echo "Error: Failed to retrieve plants."
  echo "Response: $plants"
  exit 1
fi

echo "Done."
echo "$plants"

# Step 3: Loop through each plant and print details
echo "Plants:"
echo "$plants" | jq -c '.[]' | while read -r plant; do
  plant_name=$(echo "$plant" | jq -r '.name')

  if [[ -z "$plant_name" || "$plant_name" == "null" ]]; then
    echo "Error: Failed to extract plant name."
    echo "Plant: $plant"
    continue
  fi

  # echo "Name: $plant_name"
  echo "Plant: $plant"
done
