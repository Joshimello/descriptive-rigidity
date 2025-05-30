#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    export $(cat .env | grep -v '^#' | xargs)
fi

echo "Starting descriptive-rigidity server..."
echo "OPENAI_API_KEY is set: $([ -n "$OPENAI_API_KEY" ] && echo "Yes" || echo "No")"
echo ""

go run main.go