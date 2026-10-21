package history

import (
	"sync"
	"time"

	"github.com/biryanim/family-expenses-bot/internal/agent"
)

const (
	maxMessages = 6
	ttl         = 15 * time.Minute
)

type entry struct {
	message agent.Message
	at      time.Time
}

type History struct {
	mu      sync.Mutex
	dialogs map[int64][]entry
	sweptAt time.Time
}

func New() *History {
	return &History{dialogs: make(map[int64][]entry)}
}
