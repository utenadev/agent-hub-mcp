#!/bin/bash
# watch_bbs.sh - tail -f for agent-hub-mcp messages

DB="agent-hub.db"
TOPIC_ID=${1:-8}

# Get current max ID to start from
LAST_ID=$(sqlite3 "$DB" "SELECT COALESCE(MAX(id), 0) FROM messages WHERE topic_id = $TOPIC_ID;")

echo -e "\033[1;34m--- Watching BBS Topic $TOPIC_ID (Starting from ID $LAST_ID) ---\033[0m"

while true; do
    # Query new messages. Using -separator to handle potential issues with default delimiter
    QUERY="SELECT id, sender, content, created_at FROM messages WHERE topic_id = $TOPIC_ID AND id > $LAST_ID ORDER BY id ASC;"
    
    # We use a temporary file to handle the output correctly in the loop
    sqlite3 -separator " | " "$DB" "$QUERY" > .tmp_watch
    
    if [ -s .tmp_watch ]; then
        while IFS=' | ' read -r id sender content created_at; do
            # Colorize the output: Date in green, Sender in bold cyan, Content in white
            echo -e "\033[0;32m[$created_at]\033[0m \033[1;36m$sender\033[0m: $content"
            LAST_ID=$id
        done < .tmp_watch
    fi
    
    sleep 2
done
