package hub

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// Config holds the orchestrator configuration.
type Config struct {
	PollInterval    time.Duration // How often to check for new messages
	SummaryThreshold int           // Number of messages before triggering a summary
	InactivityTimeout time.Duration // Time of no activity before nudging
}

// DefaultConfig returns the default orchestrator configuration.
func DefaultConfig() *Config {
	return &Config{
		PollInterval:     5 * time.Second,
		SummaryThreshold: 5,
		InactivityTimeout: 5 * time.Minute,
	}
}

// Orchestrator monitors the BBS and provides autonomous services.
type Orchestrator struct {
	db     *db.DB
	config *Config

	// Track state per topic
	mu              sync.Mutex
	lastSeenMsgID   map[int64]int64 // topicID -> messageID
	topicMsgCount   map[int64]int    // topicID -> message count since last summary
	lastActivity    map[int64]time.Time // topicID -> last message time
}

// NewOrchestrator creates a new orchestrator instance.
func NewOrchestrator(database *db.DB, config *Config) *Orchestrator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Orchestrator{
		db:              database,
		config:          config,
		lastSeenMsgID:   make(map[int64]int64),
		topicMsgCount:   make(map[int64]int),
		lastActivity:    make(map[int64]time.Time),
	}
}

// Start begins the orchestrator monitoring loop.
func (o *Orchestrator) Start() error {
	log.Println("Orchestrator started")

	// Initialize topic tracking
	if err := o.initializeTopics(); err != nil {
		return fmt.Errorf("failed to initialize topics: %w", err)
	}

	// Start polling loop
	ticker := time.NewTicker(o.config.PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := o.pollOnce(); err != nil {
			log.Printf("Poll error: %v", err)
		}
	}

	return nil
}

// initializeTopics sets up tracking for all existing topics.
func (o *Orchestrator) initializeTopics() error {
	topics, err := o.db.ListTopics()
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	for _, topic := range topics {
		// Get the latest message ID for this topic
		messages, err := o.db.GetMessages(int64(topic.ID), 1)
		if err != nil {
			log.Printf("Failed to get messages for topic %d: %v", topic.ID, err)
			continue
		}

		if len(messages) > 0 {
			o.lastSeenMsgID[int64(topic.ID)] = int64(messages[0].ID)
		} else {
			o.lastSeenMsgID[int64(topic.ID)] = 0
		}
		o.topicMsgCount[int64(topic.ID)] = 0
	}

	log.Printf("Initialized tracking for %d topics", len(topics))
	return nil
}

// pollOnce performs a single poll cycle.
func (o *Orchestrator) pollOnce() error {
	topics, err := o.db.ListTopics()
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if err := o.checkTopic(int64(topic.ID)); err != nil {
			log.Printf("Error checking topic %d: %v", topic.ID, err)
		}
	}

	return nil
}

// checkTopic examines a single topic for new activity.
func (o *Orchestrator) checkTopic(topicID int64) error {
	o.mu.Lock()
	lastSeen := o.lastSeenMsgID[topicID]
	o.mu.Unlock()

	// Get messages newer than last seen
	messages, err := o.db.GetMessages(topicID, 20)
	if err != nil {
		return err
	}

	// Messages come in DESC order, reverse to process chronologically
	reverse(messages)

	newMessages := 0
	var latestMsgID int64 = lastSeen
	var latestTime time.Time

	for _, msg := range messages {
		if int64(msg.ID) > lastSeen {
			newMessages++
			if int64(msg.ID) > latestMsgID {
				latestMsgID = int64(msg.ID)
			}
			// Parse timestamp (simple parsing for now)
			if latestTime.IsZero() || msg.CreatedAt > latestTime.Format(time.RFC3339) {
				latestTime, _ = time.Parse(time.RFC3339, msg.CreatedAt)
			}
		}
	}

	if newMessages > 0 {
		o.mu.Lock()
		o.lastSeenMsgID[topicID] = latestMsgID
		o.topicMsgCount[topicID] += newMessages
		if !latestTime.IsZero() {
			o.lastActivity[topicID] = latestTime
		}
		count := o.topicMsgCount[topicID]
		o.mu.Unlock()

		log.Printf("[Topic %d] %d new messages (total since last summary: %d)",
			topicID, newMessages, count)

		// Check if we should generate a summary
		if count >= o.config.SummaryThreshold {
			if err := o.generateSummary(topicID); err != nil {
				log.Printf("Failed to generate summary for topic %d: %v", topicID, err)
			}
		}
	}

	return nil
}

// generateSummary creates and posts a summary for the topic.
func (o *Orchestrator) generateSummary(topicID int64) error {
	log.Printf("Generating summary for topic %d...", topicID)

	// Get recent messages for summarization
	messages, err := o.db.GetMessages(topicID, 50)
	if err != nil {
		return err
	}

	// Mock summarizer - create a simple summary
	summary := o.mockSummarizer(messages)

	// Post the summary
	_, err = o.db.PostMessage(topicID, "orchestrator", summary)
	if err != nil {
		return err
	}

	// Reset counter
	o.mu.Lock()
	o.topicMsgCount[topicID] = 0
	o.mu.Unlock()

	log.Printf("Summary posted for topic %d", topicID)
	return nil
}

// mockSummarizer creates a simple summary of messages.
// This is a placeholder for future LLM integration.
func (o *Orchestrator) mockSummarizer(messages []db.Message) string {
	if len(messages) == 0 {
		return "No activity to summarize."
	}

	// Reverse to get chronological order
	reverse(messages)

	// Count messages per sender
	senderCounts := make(map[string]int)
	for _, msg := range messages {
		senderCounts[msg.Sender]++
	}

	// Build summary
	summary := fmt.Sprintf("📊 **Orchestrator Summary**\n\n")
	summary += fmt.Sprintf("Activity: %d messages\n", len(messages))
	summary += "Participants:\n"
	for sender, count := range senderCounts {
		summary += fmt.Sprintf("  - %s: %d messages\n", sender, count)
	}
	summary += "\n[This is a mock summary. LLM integration coming soon.]"

	return summary
}

// reverse reverses a slice of messages in place.
func reverse(messages []db.Message) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}
