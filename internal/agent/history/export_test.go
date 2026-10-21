package history

func (h *History) DialogCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.dialogs)
}
