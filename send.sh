#!/bin/bash
#
# send.sh - Send messages to AI agents in tmux panes
#

show_help() {
    cat >&2 << 'EOF'
Usage: send.sh <target_partial_name> <message>

Send a message to an AI agent running in another tmux pane.
The target is matched against pane titles (case-insensitive).

ARGUMENTS:
  target_partial_name    Partial match of the target pane title
  message                Message to send (can contain spaces)

USE CASES:
  1. Notify implementer about new spec:
     send.sh implementer "Spec ready: docs/specs/auth.md"

  2. Ask reviewer to check implementation:
     send.sh reviewer "Please review the auth module"

  3. Send command to another agent:
     send.sh architect "Check PR #42"

  4. Use with AGENT_NAME environment variable:
     AGENT_NAME=claude send.sh implementer "Hello"

NOTES:
  - Sender name is auto-detected from pane title or AGENT_NAME env var
  - If target is not found, available panes are listed
EOF
}

TARGET=${1:-}
shift 2>/dev/null || true
MESSAGE="$*"

if [ -z "$TARGET" ] || [ -z "$MESSAGE" ]; then
    show_help
    exit 1
fi

# Determine Sender Name
# 1. AGENT_NAME env var
# 2. pane_title (cleaned up)
# 3. Default "unknown"
if [ -n "$AGENT_NAME" ]; then
    SENDER="$AGENT_NAME"
else
    RAW_TITLE=$(tmux display-message -p '#{pane_title}')
    # Cleanup: Remove emojis and special characters, take the first meaningful word
    # e.g., "✳ Claude Code" -> "Claude"
    SENDER=$(echo "$RAW_TITLE" | sed 's/[^a-zA-Z0-9 ]//g' | awk '{print $1}' | tr '[:upper:]' '[:lower:]')
    
    if [ -z "$SENDER" ]; then
        SENDER="unknown"
    fi
fi

# Check if TARGET is a pane ID (starts with % or is a number)
if echo "$TARGET" | grep -qE '^%[0-9]+$'; then
    # Direct pane ID (e.g., %2)
    TARGET_PANE="$TARGET"
    TARGET_INFO=$(tmux list-panes -a -F "#{pane_id}:#{pane_title}" | grep "^${TARGET_PANE}:" | head -n1)
    TARGET_TITLE=$(echo "$TARGET_INFO" | cut -d: -f2-)
elif echo "$TARGET" | grep -qE '^[0-9]+$'; then
    # Pane number (e.g., 2) -> find by pane_index in current window
    TARGET_INFO=$(tmux list-panes -s -F "#{pane_index}:#{pane_id}:#{pane_title}" | grep "^${TARGET}:" | head -n1)
    TARGET_PANE=$(echo "$TARGET_INFO" | cut -d: -f2)
    TARGET_TITLE=$(echo "$TARGET_INFO" | cut -d: -f3-)
else
    # Find target pane ID by partial title match
    TARGET_INFO=$(tmux list-panes -s -F "#{pane_id}:#{pane_title}" | grep -i "$TARGET" | head -n1)
    TARGET_PANE=$(echo "$TARGET_INFO" | cut -d: -f1)
    TARGET_TITLE=$(echo "$TARGET_INFO" | cut -d: -f2-)
fi

if [ -z "$TARGET_PANE" ]; then
    echo "Error: Target pane matching '$TARGET' not found."
    # List available panes
    echo "Available panes:"
    tmux list-panes -s -F " - #{pane_title}"
    exit 1
fi

# Send message
# Format: "sender: message"
FULL_MSG="$SENDER: $MESSAGE"

echo "Sending to '$TARGET_TITLE' ($TARGET_PANE) as '$SENDER': $MESSAGE"
# Split into separate commands to ensure Enter is sent properly
tmux send-keys -t "$TARGET_PANE" -- "$FULL_MSG"
sleep 0.05
tmux send-keys -t "$TARGET_PANE" C-m