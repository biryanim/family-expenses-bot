package history

import (
	"time"

	"github.com/biryanim/family-expenses-bot/internal/agent"
)

func (h *History) AppendMessagesForUser(userID int64, now time.Time, messages ...agent.Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if now.Sub(h.sweptAt) > ttl {
		for id, dialog := range h.dialogs {
			if len(dialog) == 0 || now.Sub(dialog[len(dialog)-1].at) > ttl {
				delete(h.dialogs, id)
			}
		}
		h.sweptAt = now
	}

	entries := make([]entry, 0, maxMessages+len(messages))
	for _, e := range h.dialogs[userID] {
		if now.Sub(e.at) > ttl {
			continue
		}
		entries = append(entries, e)
	}
	for _, m := range messages {
		entries = append(entries, entry{message: m, at: now})
	}
	if len(entries) > maxMessages {
		entries = entries[len(entries)-maxMessages:]
	}
	h.dialogs[userID] = entries
}
