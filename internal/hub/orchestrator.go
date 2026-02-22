package hub

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"

	"github.com/yklcs/agent-hub-mcp/internal/config"
	"github.com/yklcs/agent-hub-mcp/internal/db"
)

// Config holds the orchestrator configuration.
type Config struct {
	PollInterval     time.Duration // How often to check for new messages
	SummaryThreshold int           // Number of messages before triggering a summary
	InactivityTimeout time.Duration // Time of no activity before nudging
	// LLM Configuration
	Model         string // Gemini model to use for summarization
	APIKey        string // API key for Gemini (overrides env vars)
}

// DefaultConfig returns the default orchestrator configuration.
func DefaultConfig() *Config {
	return &Config{
		PollInterval:     5 * time.Second,
		SummaryThreshold: 5,
		InactivityTimeout: 5 * time.Minute,
		Model:            "gemini-2.0-flash-lite",
	}
}

// getAPIKey returns the API key based on priority:
// 1. Config File: ~/.config/agent-hub-mcp/config.json (Field: api_key)
// 2. Config.APIKey (explicitly set)
// 3. HUB_MASTER_API_KEY (tool-specific)
// 4. GEMINI_API_KEY (generic)
// Also returns the source name for logging.
func (c *Config) getAPIKey() (key string, source string) {
	// 1. Try config file
	configPath := config.DefaultConfigPath()
	if data, err := os.ReadFile(configPath); err == nil {
		var config struct {
			APIKey string `json:"api_key"`
		}
		if json.Unmarshal(data, &config) == nil && config.APIKey != "" {
			return config.APIKey, "Config File (~/.config/agent-hub-mcp/config.json)"
		}
	}

	// 2. Try explicit config
	if c.APIKey != "" {
		return c.APIKey, "Config.APIKey"
	}

	// 3. Try HUB_MASTER_API_KEY
	if key := os.Getenv("HUB_MASTER_API_KEY"); key != "" {
		return key, "HUB_MASTER_API_KEY"
	}

	// 4. Try GEMINI_API_KEY
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key, "GEMINI_API_KEY"
	}

	return "", "none"
}

// Orchestrator monitors the BBS and provides autonomous services.
type Orchestrator struct {
	db     *db.DB
	config *Config
	client *genai.Client

	// Track state per topic
	mu            sync.Mutex
	lastSeenMsgID map[int64]int64     // topicID -> messageID
	topicMsgCount map[int64]int        // topicID -> message count since last summary
	lastActivity  map[int64]time.Time  // topicID -> last message time
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

// Initialize initializes the Orchestrator, setting up the LLM client.
func (o *Orchestrator) Initialize(ctx context.Context) error {
	apiKey, source := o.config.getAPIKey()
	if apiKey == "" {
		log.Println("Warning: No API key found (HUB_MASTER_API_KEY or GEMINI_API_KEY), using mock summarizer")
		return nil
	}

	log.Printf("Using API key from: %s", source)

	// Create Gemini client with API key
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}

	o.client = client
	log.Printf("Gemini client initialized (model: %s)", o.config.Model)
	return nil
}

// Close closes the Orchestrator's resources.
func (o *Orchestrator) Close() error {
	// genai.Client does not require explicit closing
	return nil
}

// Start begins the orchestrator monitoring loop.
func (o *Orchestrator) Start(ctx context.Context) error {
	log.Println("Orchestrator started")

	// Initialize LLM client
	if err := o.Initialize(ctx); err != nil {
		return err
	}
	defer o.Close()

	// Initialize topic tracking
	if err := o.initializeTopics(); err != nil {
		return fmt.Errorf("failed to initialize topics: %w", err)
	}

	// Start polling loop
	ticker := time.NewTicker(o.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Orchestrator stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := o.pollOnce(ctx); err != nil {
				log.Printf("Poll error: %v", err)
			}
		}
	}
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
func (o *Orchestrator) pollOnce(ctx context.Context) error {
	topics, err := o.db.ListTopics()
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if err := o.checkTopic(ctx, int64(topic.ID)); err != nil {
			log.Printf("Error checking topic %d: %v", topic.ID, err)
		}
	}

	return nil
}

// checkTopic examines a single topic for new activity.
func (o *Orchestrator) checkTopic(ctx context.Context, topicID int64) error {
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
			// Parse timestamp
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
			if err := o.generateSummary(ctx, topicID); err != nil {
				log.Printf("Failed to generate summary for topic %d: %v", topicID, err)
			}
		}
	}

	return nil
}

// generateSummary creates and posts a summary for the topic.
func (o *Orchestrator) generateSummary(ctx context.Context, topicID int64) error {
	log.Printf("Generating summary for topic %d...", topicID)

	// Get latest summary to check if chain is broken
	latestSummary, err := o.db.GetLatestSummary(topicID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to get latest summary: %v", err)
	}

	// Get recent messages for summarization
	messages, err := o.db.GetMessages(topicID, 50)
	if err != nil {
		return err
	}

	var summary string
	isMock := false

	if o.client != nil {
		// Check if previous summary was Mock (chain broken)
		if latestSummary != nil && latestSummary.IsMock {
			log.Printf("Previous summary was Mock, doing full history scan")
			summary, err = o.llmSummarizer(ctx, messages)
		} else {
			// Try incremental summarization
			if latestSummary != nil {
				summary, err = o.llmIncrementalSummarizer(ctx, latestSummary.SummaryText, messages)
			} else {
				summary, err = o.llmSummarizer(ctx, messages)
			}
		}

		if err != nil {
			log.Printf("LLM summarization failed, falling back to mock: %v", err)
			summary = o.mockSummarizer(messages)
			isMock = true
		}
	} else {
		// Use mock summarizer
		summary = o.mockSummarizer(messages)
		isMock = true
	}

	// Save summary to topic_summaries table
	_, err = o.db.SaveSummary(topicID, summary, isMock)
	if err != nil {
		log.Printf("Failed to save summary to topic_summaries: %v", err)
	}

	// Post the summary to messages
	_, err = o.db.PostMessage(topicID, "orchestrator", summary)
	if err != nil {
		return err
	}

	// Reset counter
	o.mu.Lock()
	o.topicMsgCount[topicID] = 0
	o.mu.Unlock()

	log.Printf("Summary posted for topic %d (mock=%v)", topicID, isMock)
	return nil
}

// llmIncrementalSummarizer updates an existing summary with new messages.
func (o *Orchestrator) llmIncrementalSummarizer(ctx context.Context, previousSummary string, messages []db.Message) (string, error) {
	if len(messages) == 0 {
		return previousSummary, nil
	}

	// Reverse to get chronological order
	reverse(messages)

	// Build incremental prompt
	var prompt strings.Builder
	prompt.WriteString("You are maintaining a summary of a BBS conversation.\n\n")
	prompt.WriteString("** Previous Summary **\n")
	prompt.WriteString(previousSummary)
	prompt.WriteString("\n\n** New Messages **\n")
	for i, msg := range messages {
		prompt.WriteString(fmt.Sprintf("[%d] %s: %s\n", i+1, msg.Sender, msg.Content))
	}
	prompt.WriteString("\n** Task **\n")
	prompt.WriteString("Please update the summary to incorporate the new messages. ")
	prompt.WriteString("Maintain the structure and important information from the previous summary, ")
	prompt.WriteString("while adding new insights, decisions, or action items from the new messages.")
	prompt.WriteString("Keep it concise and well-organized.")

	// Generate content using new SDK pattern
	contents := genai.Text(prompt.String())

	resp, err := o.client.Models.GenerateContent(ctx, o.config.Model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate incremental summary: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	// Extract text from response
	content := resp.Candidates[0].Content
	if len(content.Parts) == 0 {
		return "", fmt.Errorf("no parts in response content")
	}
	result := content.Parts[0].Text
	if result == "" {
		return "", fmt.Errorf("empty text in response")
	}

	return fmt.Sprintf("📊 **Orchestrator Summary (Gemini)**\n\n%s", result), nil
}

// llmSummarizer uses Gemini to create a summary of messages.
func (o *Orchestrator) llmSummarizer(ctx context.Context, messages []db.Message) (string, error) {
	if len(messages) == 0 {
		return "No activity to summarize.", nil
	}

	// Reverse to get chronological order
	reverse(messages)

	// Build conversation history
	var conversation strings.Builder
	conversation.WriteString("Here is a recent conversation from a BBS topic:\n\n")
	for i, msg := range messages {
		conversation.WriteString(fmt.Sprintf("[%d] %s: %s\n", i+1, msg.Sender, msg.Content))
	}
	conversation.WriteString("\nPlease provide a concise summary covering:\n")
	conversation.WriteString("1. What was discussed?\n")
	conversation.WriteString("2. Current status or consensus\n")
	conversation.WriteString("3. Next steps or action items\n\n")
	conversation.WriteString("Format the response with clear sections and bullet points.")

	// Generate content using new SDK pattern
	contents := genai.Text(conversation.String())

	resp, err := o.client.Models.GenerateContent(ctx, o.config.Model, contents, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no response from model")
	}

	// Extract text from response
	content := resp.Candidates[0].Content
	if len(content.Parts) == 0 {
		return "", fmt.Errorf("no parts in response content")
	}
	result := content.Parts[0].Text
	if result == "" {
		return "", fmt.Errorf("empty text in response")
	}

	// Add header
	return fmt.Sprintf("📊 **Orchestrator Summary (Gemini)**\n\n%s", result), nil
}

// mockSummarizer creates a simple summary of messages.
// This is a fallback when LLM is not available.
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
	summary := fmt.Sprintf("📊 **Orchestrator Summary (Mock)**\n\n")
	summary += fmt.Sprintf("Activity: %d messages\n", len(messages))
	summary += "Participants:\n"
	for sender, count := range senderCounts {
		summary += fmt.Sprintf("  - %s: %d messages\n", sender, count)
	}
	summary += "\n[LLM integration not configured. Set HUB_MASTER_API_KEY or GEMINI_API_KEY to enable AI summarization.]"

	return summary
}

// reverse reverses a slice of messages in place.
func reverse(messages []db.Message) {
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
}
