#!/bin/bash

# Define variables
SCRIPT_PATH="$HOME/Videos/go/writeup-finder" # Change this to your actual script directory
echo $PROXY
echo $PROXY_HOST
echo $PROXY_PORT

# Function to check proxy connection
check_proxy() {
    nc -zv "$PROXY_HOST" "$PROXY_PORT" >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# Now, check the proxy connection
if check_proxy; then
    echo "Proxy is up, running the writeup-finder script..."

    # Run the barcelona-watch script with the proxy
    cd "$SCRIPT_PATH" || { echo "Failed to change directory to $SCRIPT_PATH"; exit 1; }
    
    go run main.go --d --t --proxy="$PROXY"

else
    echo "Proxy is down, skipping this attempt."
fi