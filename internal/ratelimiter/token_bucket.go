package ratelimiter

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type TokenBucketLimiter struct {
	mu                sync.Mutex
	tokens            uint64
	fillRate          float64
	capacity          uint64
	lastTime          time.Time
}

func GetIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("Error parssing IP :%v", err)
		return ""
	}
	return host
}

func NewTokenBucketLimiter(f float64, b uint64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		tokens:   b,
		fillRate: f,
		capacity: b,
		lastTime: time.Now(),
	}
}

func (t *TokenBucketLimiter) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	timePassed := now.Sub(t.lastTime).Seconds()
	tokensToAdd := timePassed * t.fillRate

	if tokensToAdd > 0 {
		t.tokens = min(t.capacity, t.tokens+uint64(tokensToAdd))
		t.lastTime = now
	}

	if t.tokens > 0 {
		t.tokens--
		return true
	}

	return false
}
