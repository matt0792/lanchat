package app

import (
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type RateLimiter struct {
	mu       sync.Mutex
	messages map[peer.ID][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		messages: make(map[peer.ID][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(peerID peer.ID) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	messages := []time.Time{}
	for _, t := range rl.messages[peerID] {
		if t.After(cutoff) {
			messages = append(messages, t)
		}
	}

	if len(messages) >= rl.limit {
		return false
	}

	messages = append(messages, now)
	rl.messages[peerID] = messages
	return true
}
