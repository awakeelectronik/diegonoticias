package ratelimit

import (
	"sync"
	"time"
)

type DailyLimiter struct {
	mu    sync.Mutex
	day   string
	count int
	max   int
}

func NewDailyLimiter(max int) *DailyLimiter {
	if max <= 0 {
		max = 100
	}
	return &DailyLimiter{max: max}
}

func (l *DailyLimiter) Allow(now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	day := now.Format("2006-01-02")
	if l.day != day {
		l.day = day
		l.count = 0
	}
	if l.count >= l.max {
		return false
	}
	l.count++
	return true
}

