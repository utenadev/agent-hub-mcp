# Spec 004: TUI Dashboard v2 (Summaries & Layout)

## 1. Goal
Upgrade the TUI Dashboard to support the new `topic_summaries` table and provide a more informative 3-pane layout.

## 2. Functional Requirements

### 2.1 3-Pane Layout
- **Left Pane**: Topics List (current functionality).
- **Middle Pane**: Message History (current functionality).
- **Right Pane**: Latest Summary View.
    - Displays the most recent Gemini/Mock summary for the selected topic.
    - If no summary exists, show a "No summary available" message.

### 2.2 Summary Navigation
- Provide a way to toggle focus between panes (e.g., Tab key).
- In the Summary Pane, allow the user to view previous summaries (e.g., using `[` and `]` keys).

### 2.3 Visual Improvements
- Highlight real LLM summaries differently from Mock summaries (e.g., different colors or labels).
- Use `lipgloss` to define clear borders between the three panes.

## 3. Implementation Steps

1.  **Update internal/ui/model.go**:
    - Add fields to store the list of summaries for the selected topic.
    - Add logic to fetch summaries when a topic is selected.
    - Handle pane focus state.
2.  **Update internal/ui/view.go**:
    - Implement the 3-pane layout using `lipgloss.JoinHorizontal`.
    - Create a dedicated renderer for the Summary Pane.
3.  **Keyboard Bindings**:
    - `Tab`: Cycle focus between Topics, Messages, and Summaries.
    - `[` / `]`: Navigate summary history.

## 4. Acceptance Criteria
- [ ] Dashboard displays 3 panes when a topic is selected.
- [ ] Right pane shows the latest summary from `topic_summaries` table.
- [ ] User can scroll/navigate through previous summaries in the right pane.
- [ ] Clear visual distinction between Mock and Gemini summaries.
