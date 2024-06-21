#!/bin/bash
num_records=$1

# Check if the correct number of arguments is provided
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <number_of_requests>"
  exit 1
fi

generate_random_string() {
  local length=$1
  tr -dc A-Za-z0-9 </dev/urandom | head -c $length
}

for ((i=1; i<=num_records; i++)); do
    record_val=$(generate_random_string 16) # Generate a random string of 16 characters
    payload="{\"record\":{\"value\":\"$record_val\"}}"

    echo "Sending request $i with payload: $payload"
    curl -X POST localhost:8080 -d "$payload" -H "Content-Type: application/json"
done

# ps. technically we're supposed to send base64 encoded strings here but i'm just sending a random set of characters that is 16 bytes long, 
# so that it satisfies the input constraints
# cmd to curl read: curl -X GET localhost:8080 -d \ '{"offset":0}'
