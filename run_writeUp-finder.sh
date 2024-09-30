#!/bin/bash

# Proxy address
proxy_url="http://0.0.0.0:8086"

# Run curl with a timeout of 5 seconds
timeout 5 curl --silent --head --fail "$proxy_url" &> /dev/null

# Check the exit status of the curl command
if [ $? -eq 124 ]; then
  # Exit status 124 means the command timed out
  echo "$(date) - Proxy is running (command did not finish within 5 seconds), starting the script..."
  # Run the Go script with proxy
  /usr/local/go/bin/go run main.go -d -t --proxy="$proxy_url"
else
  echo "$(date) - Proxy is not available or finished quickly, skipping script execution."
fi

