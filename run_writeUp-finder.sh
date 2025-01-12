#!/bin/bash

source /home/mohammad/Videos/go/proxy.env

# Define variables
SCRIPT_PATH="$HOME/Videos/go/writeup-finder" # Change this to your actual script directory

# Function to check proxy connection
check_proxy() {
    nc -zv "$PROXY_HOST" "$PROXY_PORT" >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# Function to check if Windscribe VPN is up
check_windscribe() {
    pgrep -l windscribe >/dev/null 2>&1
    if [ $? -eq 0 ]; then
        return 0
    else
        return 1
    fi
}

# Main logic
if check_windscribe; then
    echo "Windscribe VPN is up, running the writeup-finder script without proxy..."
    cd "$SCRIPT_PATH" || { echo "Failed to change directory to $SCRIPT_PATH"; exit 1; }
    $HOME/Videos/go/writeup-finder/writeup-finder --database --telegram

elif check_proxy; then
    echo "Proxy is up, running the writeup-finder script with proxy..."
    cd "$SCRIPT_PATH" || { echo "Failed to change directory to $SCRIPT_PATH"; exit 1; }
    $HOME/Videos/go/writeup-finder/writeup-finder --database --telegram --proxy="$PROXY"

else
    echo "Neither Windscribe VPN nor proxy is available. Skipping this attempt."
fi

