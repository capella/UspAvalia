#!/bin/sh
set -e

# Function to update disciplines
update_disciplines() {
    echo "[$(date)] Starting weekly discipline update..."

    # Run fetch-disciplines command to update database
    echo "[$(date)] Fetching disciplines..."
    /app/uspavalia fetch-disciplines --store

    echo "[$(date)] Weekly discipline update completed"
}

# Start the weekly update job in the background
# Run every Sunday at 3:00 AM
(
    while true; do
        # Calculate seconds until next Sunday 3:00 AM
        current_day=$(date +%u)  # 1-7 (Monday is 1, Sunday is 7)
        current_hour=$(date +%H)
        current_minute=$(date +%M)

        # Days until Sunday (7 = Sunday)
        if [ "$current_day" -eq 7 ]; then
            # It's Sunday
            if [ "$current_hour" -lt 3 ] || { [ "$current_hour" -eq 3 ] && [ "$current_minute" -eq 0 ]; }; then
                # Before or at 3:00 AM today
                days_until=0
            else
                # After 3 AM, wait until next Sunday
                days_until=7
            fi
        else
            # Not Sunday yet
            days_until=$((7 - current_day))
        fi

        # Calculate seconds until target time (3:00 AM)
        # Target: days_until days from now at 03:00:00
        current_seconds=$((current_hour * 3600 + current_minute * 60))
        target_seconds=$((3 * 3600))  # 3:00 AM = 3 * 3600 seconds

        if [ "$days_until" -eq 0 ]; then
            # Today - just calculate time difference
            seconds_until=$((target_seconds - current_seconds))
        else
            # Future day - add remaining time today + full days + target time
            seconds_until=$(( (86400 - current_seconds) + ((days_until - 1) * 86400) + target_seconds ))
        fi

        echo "[$(date)] Next discipline update scheduled in $((seconds_until / 3600)) hours"

        # Sleep until the scheduled time
        sleep "$seconds_until"

        # Run the update
        update_disciplines
    done
) &

# Save the background job PID
echo $! > /tmp/update_cron.pid

echo "[$(date)] Starting USP Avalia server..."
echo "[$(date)] Weekly discipline updates are scheduled for Sundays at 3:00 AM"

# Start the main application
exec "$@"
