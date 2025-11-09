#!/bin/bash

# SSH Monitoring Script
# This script monitors the specific SSH session by TTY
# Usage: ./ssh-mon.bash [TTY]

LOG_FILE="/var/log/ssh-monitor.log"
current_USER=$USER

# Get TTY from parameter or SSH_TTY environment variable
if [ -n "$1" ]; then
    MONITOR_TTY="$1"
elif [ -n "$SSH_TTY" ]; then
    MONITOR_TTY="$SSH_TTY"
else
    TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
    echo "[$TIMESTAMP] ERROR: No TTY specified and SSH_TTY not set" >> "$LOG_FILE"
    exit 1
fi

# Log start of monitoring and record start time
START_TIME=$(date +%s)
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
echo "[$TIMESTAMP] Started monitoring SSH session for user $current_USER on TTY $MONITOR_TTY" >> "$LOG_FILE"

# Get the basename of the TTY (e.g., /dev/pts/5 -> pts/5)
TTY_BASENAME=$(echo "$MONITOR_TTY" | sed 's#^/dev/##')

# Function to calculate duration
calculate_duration() {
    local start=$1
    local end=$2
    local duration=$((end - start))
    
    local hours=$((duration / 3600))
    local minutes=$(((duration % 3600) / 60))
    local seconds=$((duration % 60))
    
    if [ $hours -gt 0 ]; then
        echo "${hours}h ${minutes}m ${seconds}s"
    elif [ $minutes -gt 0 ]; then
        echo "${minutes}m ${seconds}s"
    else
        echo "${seconds}s"
    fi
}

# Monitor the SSH session
while true; do
    # Check if the TTY device still exists and has processes attached
    if [ ! -e "$MONITOR_TTY" ]; then
        END_TIME=$(date +%s)
        DURATION=$(calculate_duration $START_TIME $END_TIME)
        TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
        echo "[$TIMESTAMP] SSH session ended - TTY device $MONITOR_TTY no longer exists (Duration: $DURATION)" >> "$LOG_FILE"
        # Delete this script
        rm -f /tmp/ssh-mon.bash
        exit 0
    fi
    
    # Check if there are any bash/shell processes on this TTY
    SHELL_ON_TTY=$(ps aux | grep "$TTY_BASENAME" | grep -E "(bash|sh|zsh)" | grep -v grep | grep "$current_USER")
    
    if [ -z "$SHELL_ON_TTY" ]; then
        END_TIME=$(date +%s)
        DURATION=$(calculate_duration $START_TIME $END_TIME)
        TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
        echo "[$TIMESTAMP] SSH session ended - No shell process found on TTY $MONITOR_TTY (Duration: $DURATION)" >> "$LOG_FILE"
        # Delete this script
        rm -f /tmp/ssh-mon.bash
        exit 0
    fi
    
    sleep 2
done

# This should never be reached, but just in case
END_TIME=$(date +%s)
DURATION=$(calculate_duration $START_TIME $END_TIME)
TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
echo "[$TIMESTAMP] Monitoring script ended for user $current_USER on TTY $MONITOR_TTY (Duration: $DURATION)" >> "$LOG_FILE"
# Delete this script
rm -f /tmp/ssh-mon.bash
exit 0
