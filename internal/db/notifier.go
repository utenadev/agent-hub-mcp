package db

import (
	"sync"
	"time"
)

// Notification represents a notification event.
type Notification struct {
	AgentID   string
	TopicID   int64
	Message   string
	Timestamp time.Time
}

// Notifier manages notification channels for agents waiting on new messages.
type Notifier struct {
	mu       sync.RWMutex
	channels map[string]chan Notification
}

// NewNotifier creates a new Notifier instance.
func NewNotifier() *Notifier {
	return &Notifier{
		channels: make(map[string]chan Notification),
	}
}

// Register creates a notification channel for an agent.
// Returns a channel that the agent can use to receive notifications.
func (n *Notifier) Register(agentID string) chan Notification {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Close existing channel if any
	if ch, exists := n.channels[agentID]; exists {
		close(ch)
	}

	ch := make(chan Notification, 1)
	n.channels[agentID] = ch
	return ch
}

// Unregister removes the notification channel for an agent.
func (n *Notifier) Unregister(agentID string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if ch, exists := n.channels[agentID]; exists {
		close(ch)
		delete(n.channels, agentID)
	}
}

// Notify sends a notification to a specific agent.
func (n *Notifier) Notify(agentID string, notification Notification) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if ch, exists := n.channels[agentID]; exists {
		select {
		case ch <- notification:
		default:
			// Channel is full or not ready, skip notification
		}
	}
}

// NotifyAll sends a notification to all registered agents.
func (n *Notifier) NotifyAll(notification Notification) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, ch := range n.channels {
		select {
		case ch <- notification:
		default:
			// Channel is full or not ready, skip notification
		}
	}
}

// Wait waits for a notification or times out.
// Returns true if a notification was received, false if timed out.
func (n *Notifier) Wait(agentID string, timeoutSec int) bool {
	n.mu.RLock()
	ch, exists := n.channels[agentID]
	n.mu.RUnlock()

	if !exists {
		return false
	}

	if timeoutSec <= 0 {
		timeoutSec = 180 // Default timeout
	}

	select {
	case <-ch:
		return true
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		return false
	}
}

// Count returns the number of registered agents.
func (n *Notifier) Count() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return len(n.channels)
}
