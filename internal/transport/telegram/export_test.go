package telegram

func (l *UserLocks) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.locks)
}

func (h *Handler) Locks() *UserLocks {
	return h.locks
}
