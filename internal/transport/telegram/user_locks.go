package telegram

import "sync"

type userLock struct {
	slot chan struct{}
	refs int
}

type UserLocks struct {
	mu    sync.Mutex
	locks map[int64]*userLock
}

func NewUserLocks() *UserLocks {
	return &UserLocks{locks: make(map[int64]*userLock)}
}
