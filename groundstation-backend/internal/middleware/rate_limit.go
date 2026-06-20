package middleware

import (
	"net/http"
	"sync"
	"time"
	"groundstation-backend/pkg/utils"
	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*requestInfo
	limit    int
	window   time.Duration
}

type requestInfo struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*requestInfo),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.requests {
			if now.After(info.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[key]

	if !exists || now.After(info.resetTime) {
		rl.requests[key] = &requestInfo{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if info.count >= rl.limit {
		return false
	}

	info.count++
	return true
}

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, 429001, "请求过于频繁，请稍后再试", nil)
			c.Abort()
			return
		}
		c.Next()
	}
}
