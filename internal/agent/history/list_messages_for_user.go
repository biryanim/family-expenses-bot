package history

import (
	"time"

	"github.com/biryanim/family-expenses-bot/internal/agent"
)

func (h *History) ListMessagesForUser(userID int64, now time.Time) []agent.Message {
	h.mu.Lock()
	defer h.mu.Unlock()

	var messages []agent.Message
	for _, e := range h.dialogs[userID] {
		if now.Sub(e.at) > ttl {
			continue
		}
		messages = append(messages, e.message)
	}
	return messages
}
