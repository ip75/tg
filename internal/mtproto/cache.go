package mtproto

import (
	"context"
	"sync/atomic"

	"github.com/gotd/td/session"
)

type SessionCache struct {
	atomic.Value
}

func (c *SessionCache) LoadSession(ctx context.Context) ([]byte, error) {
	s := c.Value.Load()
	if s == nil {
		return nil, session.ErrNotFound
	}
	if s, ok := s.([]byte); ok {
		return s, nil
	}
	return nil, session.ErrNotFound
}

func (c *SessionCache) StoreSession(ctx context.Context, data []byte) error {
	c.Value.Store(data)
	return nil
}
