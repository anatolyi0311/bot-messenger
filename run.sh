#!/bin/bash

echo "Running the application with the following configuration:"
# echo "Environment: $ENV"
# echo "Database URL: $DATABASE_URL"
# echo "API Key: $API_KEY"
# echo "Log Level: $LOG_LEVEL"
# echo "Debug Mode: $DEBUG_MODE"
# echo "---------------------------------------------"
# echo "Setting up the environment variables..."
# export ENV='production'
# export DATABASE_URL='postgres://user:password@localhost:5432/mydb'
# export API_KEY='your-api-key'
# export LOG_LEVEL='info'
# export DEBUG_MODE='false'
# echo "Environment variables set successfully."
export CONFIG='data/config.json'
echo "Configuration file set to: $CONFIG"
echo "Configuration file: $CONFIG"
echo "Checking if the configuration file exists..."
if [ -f "$CONFIG" ]; then
    echo "Configuration file found: $CONFIG"
    echo "Starting the application..."
    go run ./cmd/main.go
else
    echo "Error: Configuration file not found at $CONFIG"
    exit 1
fi
