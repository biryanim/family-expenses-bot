package telegram

import "context"

func (l *UserLocks) AcquireForUser(ctx context.Context, userID int64) (func(), error) {
	l.mu.Lock()
	lock, ok := l.locks[userID]
	if !ok {
		lock = &userLock{slot: make(chan struct{}, 1)}
		l.locks[userID] = lock
	}
	lock.refs++
	l.mu.Unlock()

	release := func() {
		l.mu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(l.locks, userID)
		}
		l.mu.Unlock()
	}

	select {
	case lock.slot <- struct{}{}:
		return func() {
			<-lock.slot
			release()
		}, nil
	case <-ctx.Done():
		release()
		return nil, ctx.Err()
	}
}
