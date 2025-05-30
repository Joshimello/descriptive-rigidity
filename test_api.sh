#!/bin/bash

# Test script for the descriptive-rigidity API

# Load environment variables from .env file if it exists
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set the API endpoint
API_URL="http://localhost:8080/generate-deformations"

# Test data for single frame animation
TEST_DATA_SINGLE='{
  "control_points": [
    {"id": 0, "role": "left leg", "position": [1.0, 2.0, 0.0]},
    {"id": 1, "role": "right arm", "position": [-1.0, 2.0, 0.0]},
    {"id": 2, "role": "head", "position": [0.0, 7.0, 0.0]},
    {"id": 3, "role": "left arm", "position": [1.0, 2.0, 0.0]},
    {"id": 4, "role": "right leg", "position": [-1.0, 2.0, 0.0]}
  ],
  "prompt": "make the character wave",
  "length": 1
}'

# Test data for multi-frame animation
TEST_DATA_MULTI='{
  "control_points": [
    {"id": 0, "role": "left leg", "position": [1.0, 2.0, 0.0]},
    {"id": 1, "role": "right arm", "position": [-1.0, 2.0, 0.0]},
    {"id": 2, "role": "head", "position": [0.0, 7.0, 0.0]},
    {"id": 3, "role": "left arm", "position": [1.0, 2.0, 0.0]},
    {"id": 4, "role": "right leg", "position": [-1.0, 2.0, 0.0]}
  ],
  "prompt": "make the character walk naturally forward",
  "length": 6
}'

echo "Testing the descriptive-rigidity API..."
echo "API URL: $API_URL"
echo ""
echo "Note: Make sure your Go server is running with:"
echo "  go run main.go"
echo "And that OPENAI_API_KEY is set in your environment or .env file"
echo ""

# Function to test API with given data
test_api() {
    local test_name="$1"
    local test_data="$2"
    
    echo "========================================="
    echo "Test: $test_name"
    echo "========================================="
    echo "Request payload:"
    echo "$test_data" | jq .
    echo ""
    
    # Make the API request
    echo "Sending request..."
    response=$(curl -s -X POST \
      -H "Content-Type: application/json" \
      -d "$test_data" \
      "$API_URL")
    
    # Check if curl command was successful
    if [ $? -eq 0 ]; then
        echo "Raw response:"
        echo "$response"
        echo ""
        echo "Formatted response (if valid JSON):"
        echo "$response" | jq . 2>/dev/null || echo "Response is not valid JSON"
        
        # Check if response is an array and show frame count
        frame_count=$(echo "$response" | jq '. | length' 2>/dev/null)
        if [ "$frame_count" != "null" ] && [ "$frame_count" != "" ]; then
            echo ""
            echo "Number of animation frames: $frame_count"
        fi
    else
        echo "Error: Failed to connect to the API"
        echo "Make sure the server is running on port 8080"
    fi
    echo ""
}

# Test single frame animation
test_api "Single Frame Animation (Wave)" "$TEST_DATA_SINGLE"

# Test multi-frame animation
test_api "Multi-Frame Animation (Walk)" "$TEST_DATA_MULTI"

echo "All tests completed."
