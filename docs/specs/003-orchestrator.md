# Spec 003: BBS Orchestrator (Autonomous Monitor)

## 1. Goal
Implement an autonomous process ("Orchestrator") that monitors the BBS activity, provides summaries, and ensures the multi-agent collaboration stays on track.
This corresponds to **Phase 3** of the Roadmap.

## 2. Technical Requirements

### 2.1 Architecture: Shared DB Model
- The Orchestrator runs as a separate long-lived process or a subcommand.
- Command: `bbs orchestrator`
- It directly accesses the shared `agent-hub.db` using SQLite WAL mode.
- It acts as an "Agent" by posting messages back to the BBS.

### 2.2 Core Functions

#### A. Topic Monitoring (Polling)
- Periodic check of the `messages` table (e.g., every 5-10 seconds).
- Tracks the "last seen message ID" to detect new activity.

#### B. Thread Summarization (LLM Integration)
- **Model**: Use `gemini-2.0-flash-lite`.
- **API Key Priority**:
    1. Config File: `~/.config/agent-hub-mcp/config.json` (Field: `api_key`)
    2. Env Var: `HUB_MASTER_API_KEY`
    3. Env Var: `GEMINI_API_KEY`
- **Behavior**:
    - When threshold is met, send recent message history to Gemini.
    - Prompt should focus on: "What was discussed?", "Current status/consensus", and "Next steps".
    - Post the resulting summary back to the BBS.

#### C. Inactivity / Deadlock Detection
- If a topic has no new messages for X minutes, post a reminder or a nudge.
- Check if agents are "stuck" (e.g., waiting for each other without a clear path).

#### D. Status Reporting
- Periodic "Pulse" updates to a meta-topic or dashboard to report overall system health.

## 3. Implementation Steps

1.  **CLI Subcommand**: Add `orchestrator` to `cmd/bbs/main.go`.
2.  **Watcher Loop**: Implement a background loop in `internal/hub/orchestrator.go`.
3.  **LLM Integration**: 
    - For the first iteration, use a simple template or a mock summarizer.
    - Later, integrate a real LLM provider (Gemini API recommended for this project).
4.  **Logging**: Ensure the Orchestrator's own logs are captured for debugging.

## 4. Acceptance Criteria
- [ ] `bbs orchestrator` runs without errors.
- [ ] It detects new messages in a topic and logs them.
- [ ] (Manual/Mock) It posts a "Summary" message to a topic after 5 messages are posted.
- [ ] It nudges the topic if there is no activity for a configured duration.
